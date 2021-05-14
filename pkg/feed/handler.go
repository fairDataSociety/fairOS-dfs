/*
Copyright © 2020 FairOS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package feed

import (
	"bytes"
	"context"
	"crypto"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"sync"
	"sync/atomic"

	beecrypto "github.com/ethersphere/bee/pkg/crypto"
	"github.com/ethersphere/bee/pkg/swarm"
	bmtlegacy "github.com/ethersphere/bmt/legacy"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed/lookup"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"golang.org/x/crypto/sha3"
)

type Handler struct {
	accountInfo *account.Info
	client      blockstore.Client
	hasherPool  *bmtlegacy.TreePool
	HashSize    int
	cache       map[uint64]*CacheEntry
	cacheLock   sync.RWMutex
}

// hashPool contains a pool of ready hashers
var hashPool sync.Pool

// init initializes the package and hashPool
func init() {
	hashPool = sync.Pool{
		New: func() interface{} {
			return crypto.SHA256.New()
		},
	}
}

// NewHandler the main handler object that handles all the feed related functions.
func NewHandler(accountInfo *account.Info, client blockstore.Client, hasherPool *bmtlegacy.TreePool) *Handler {
	fh := &Handler{
		accountInfo: accountInfo,
		client:      client,
		hasherPool:  hasherPool,
		cache:       make(map[uint64]*CacheEntry),
	}
	for i := 0; i < hasherCount; i++ {
		hashfunc := crypto.SHA256.New()
		if fh.HashSize == 0 {
			fh.HashSize = hashfunc.Size()
		}
		hashPool.Put(hashfunc)
	}
	return fh
}

func (h *Handler) update(id, owner, signature, data []byte) ([]byte, error) {
	// send the SOC chunk
	addr, err := h.client.UploadSOC(utils.Encode(owner), utils.Encode(id), utils.Encode(signature), data)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func (h *Handler) deleteChunk(ref []byte) error {
	return h.client.DeleteChunk(ref)
}

// GetContent retrieves the data payload of the last synced update of the feed
func (h *Handler) GetContent(feed *Feed) (swarm.Address, []byte, error) {
	if feed == nil {
		return swarm.ZeroAddress, nil, NewError(ErrInvalidValue, "feed is nil")
	}
	feedUpdate, err := h.get(feed)
	if err != nil {
		return swarm.ZeroAddress, nil, err
	}
	if feedUpdate == nil {
		return swarm.ZeroAddress, nil, NewError(ErrNotFound, "feed update not cached")
	}
	return swarm.NewAddress(feedUpdate.lastKey), feedUpdate.data, nil
}

// Lookup retrieves a specific or latest feed update
// Lookup works differently depending on the configuration of `query`
// See the `query` documentation and helper functions:
// `NewQueryLatest` and `NewQuery`
func (h *Handler) Lookup(ctx context.Context, query *Query) (*CacheEntry, error) {

	timeLimit := query.TimeLimit
	if timeLimit == 0 { // if time limit is set to zero, the user wants to get the latest update
		timeLimit = TimestampProvider.Now().Time
	}

	if query.Hint == lookup.NoClue { // try to use our cache
		entry, err := h.get(&query.Feed)
		if err != nil {
			return nil, err
		}
		if entry != nil && entry.Epoch.Time <= timeLimit { // avoid bad hints
			query.Hint = entry.Epoch
		}
	}

	// we can't look for anything without a store
	if h.client == nil {
		return nil, NewError(ErrInit, "invalid blockstore")
	}

	var readCount int32

	// Invoke the lookup engine.
	// The callback will be called every time the lookup algorithm needs to guess
	requestPtr, err := lookup.Lookup(ctx, timeLimit, query.Hint, func(ctx context.Context, epoch lookup.Epoch, now uint64) (interface{}, error) {
		atomic.AddInt32(&readCount, 1)
		id := ID{
			Feed:  query.Feed,
			Epoch: epoch,
		}
		ctx, cancel := context.WithTimeout(ctx, defaultRetrieveTimeout)
		defer cancel()

		addr, err := h.getAddress(id.Topic, query.Feed.User, epoch)
		if err != nil {
			return nil, err
		}
		data, err := h.client.DownloadChunk(ctx, addr.Bytes())
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || err.Error() == "error downloading data" { // chunk not found
				return nil, nil
			}
			return nil, err
		}
		ch := swarm.NewChunk(addr, data)
		var request Request
		if err := h.fromChunk(ch, &request, query, &id); err != nil {
			return nil, nil
		}
		if request.Time <= timeLimit {
			return &request, nil
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	request, _ := requestPtr.(*Request)
	if request == nil {
		return nil, NewError(ErrNotFound, "no feed updates found")
	}
	return h.updateCache(request)
}

// fromChunk populates this structure from chunk data. It does not verify the signature is valid.
func (h *Handler) fromChunk(chunk swarm.Chunk, r *Request, q *Query, id *ID) error {
	chunkdata := chunk.Data()

	if len(chunkdata) < idLength+signatureLength+utils.SpanLength {
		return fmt.Errorf("invalid chunk data len")
	}

	r.idAddr = swarm.NewAddress(chunk.Address().Bytes())
	r.binaryData = chunkdata
	cursor := idLength
	r.Signature = &Signature{}
	copy(r.Signature[:], chunkdata[cursor:cursor+signatureLength])
	cursor += signatureLength
	span := binary.LittleEndian.Uint64(chunkdata[cursor : cursor+utils.SpanLength])
	cursor += utils.SpanLength
	r.data = make([]byte, span)
	copy(r.data, chunkdata[cursor:uint64(cursor)+span])

	r.Feed = q.Feed
	r.User = q.User
	r.Epoch = id.Epoch
	return nil
}

// update feed updates cache with specified content
func (h *Handler) updateCache(request *Request) (*CacheEntry, error) {
	updateAddr := request.idAddr.Bytes()
	entry, err := h.get(&request.Feed)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		entry = &CacheEntry{}
		err := h.set(&request.Feed, entry)
		if err != nil {
			return nil, err
		}
	}

	// update our rsrcs entry map
	entry.lastKey = updateAddr
	entry.Update.ID = request.ID
	entry.data = request.data
	entry.Reader = bytes.NewReader(entry.data)
	return entry, nil
}

func hashFunc() hash.Hash {
	return sha3.NewLegacyKeccak256()
}

func (h *Handler) getAddress(topic Topic, user utils.Address, epoch lookup.Epoch) (swarm.Address, error) {
	id, err := h.getId(topic, epoch.Time, epoch.Level)
	if err != nil {
		return swarm.ZeroAddress, err
	}
	addr, err := toSignDigest(id, user[:])
	if err != nil {
		return swarm.ZeroAddress, err
	}
	return swarm.NewAddress(addr), nil
}

func (h *Handler) toChunkContent(req *Request, id, payloadId []byte) ([]byte, error) {
	// get the signature, sign(ID, payloadId)
	signaturebytes, _, err := h.getSignature(id, payloadId)
	if err != nil {
		return nil, err
	}
	req.Signature = &Signature{}
	copy(req.Signature[:], signaturebytes)

	// create the entire soc payload
	buf := make([]byte, idLength+signatureLength+utils.SpanLength+len(req.data))
	var cursor int
	copy(buf[cursor:cursor+idLength], id)
	cursor += idLength
	copy(buf[cursor:cursor+signatureLength], signaturebytes)
	cursor += signatureLength
	binary.LittleEndian.PutUint64(buf[cursor:cursor+utils.SpanLength], uint64(len(req.data)))
	cursor += utils.SpanLength
	copy(buf[cursor:cursor+len(req.data)], req.data)
	req.binaryData = buf

	return buf, nil
}

// NewRequest prepares a Request structure with all the necessary information to
// just add the desired data and sign it.
// The resulting structure can then be signed and passed to Handler.Update to be verified and sent
func (h *Handler) NewRequest(ctx context.Context, feed *Feed) (request *Request, err error) {
	if feed == nil {
		return nil, NewError(ErrInvalidValue, "feed cannot be nil")
	}

	now := TimestampProvider.Now().Time
	request = new(Request)

	query := NewQueryLatest(feed, lookup.NoClue)

	feedUpdate, err := h.Lookup(ctx, query)
	if err != nil {
		if err.(*Error).code != ErrNotFound {
			return nil, err
		}
		// not finding updates means that there is a network error
		// or that the feed really does not have updates
	}

	request.Feed = *feed

	// if we already have an update, then find next epoch
	if feedUpdate != nil {
		request.Epoch = lookup.GetNextEpoch(feedUpdate.Epoch, now)
	} else {
		request.Epoch = lookup.GetFirstEpoch(now)
	}

	return request, nil
}

func (h *Handler) getId(topic Topic, time uint64, level uint8) ([]byte, error) {
	bufId := make([]byte, TopicLength+lookup.EpochLength)
	var cursor int
	copy(bufId[cursor:cursor+TopicLength], topic[:TopicLength])
	cursor += TopicLength
	eid := epocId(time, level)
	copy(bufId[cursor:cursor+lookup.EpochLength], eid[:])
	hasher := bmtlegacy.New(h.hasherPool)
	hasher.Reset()
	_, err := hasher.Write(bufId)
	if err != nil {
		return nil, err
	}
	id := hasher.Sum(nil)
	return id, nil
}

func (h *Handler) getPayloadId(data []byte) ([]byte, error) {
	span := len(data)
	hasher := bmtlegacy.New(h.hasherPool)
	hasher.Reset()
	spanBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(spanBytes, uint64(span))
	err := hasher.SetSpanBytes(spanBytes)
	if err != nil {
		return nil, err
	}
	_, err = hasher.Write(data)
	if err != nil {
		return nil, err
	}
	payloadId := hasher.Sum(nil)
	return payloadId, nil
}

func (h *Handler) getSignature(id, payloadId []byte) ([]byte, []byte, error) {
	toSignBytes, err := toSignDigest(id, payloadId)
	if err != nil {
		return nil, nil, err
	}
	signer := beecrypto.NewDefaultSigner(h.accountInfo.GetPrivateKey())
	signature, err := signer.Sign(toSignBytes)
	if err != nil {
		return nil, nil, err
	}
	return signature, toSignBytes, nil
}

func epocId(time uint64, level uint8) lookup.EpochID {
	base := time & (maxuint64 << level)
	var id lookup.EpochID
	binary.LittleEndian.PutUint64(id[:], base)
	id[7] = level
	return id
}

// Retrieves the feed update cache value for the given nameHash
func (h *Handler) get(feed *Feed) (*CacheEntry, error) {
	mapKey, err := feed.mapKey()
	if err != nil {
		return nil, err
	}
	h.cacheLock.RLock()
	defer h.cacheLock.RUnlock()
	feedUpdate := h.cache[mapKey]
	return feedUpdate, nil
}

// Sets the feed update cache value for the given feed
func (h *Handler) set(feed *Feed, feedUpdate *CacheEntry) error {
	mapKey, err := feed.mapKey()
	if err != nil {
		return err
	}
	h.cacheLock.Lock()
	defer h.cacheLock.Unlock()
	h.cache[mapKey] = feedUpdate
	return nil
}

// toSignDigest creates a digest suitable for signing to represent the soc.
func toSignDigest(id, sum []byte) ([]byte, error) {
	h := swarm.NewHasher()
	_, err := h.Write(id)
	if err != nil {
		return nil, err
	}
	_, err = h.Write(sum)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
