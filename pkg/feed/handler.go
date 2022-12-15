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
	"context"
	"crypto"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	beeCrypto "github.com/ethersphere/bee/pkg/crypto"
	"github.com/ethersphere/bee/pkg/feeds"
	"github.com/ethersphere/bee/pkg/feeds/factory"
	"github.com/ethersphere/bee/pkg/soc"
	"github.com/ethersphere/bee/pkg/swarm"
	bmtlegacy "github.com/ethersphere/bmt/legacy"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"golang.org/x/crypto/sha3"
)

const (
	hasherCount = 8
)

// Handler takes care of all the feed operations
type Handler struct {
	accountInfo *account.Info
	client      blockstore.Client
	hasherPool  *bmtlegacy.TreePool
	HashSize    int
}

// hashPool contains a pool of ready hashers
var hashPool sync.Pool

// init initializes the package and hashPool
func init() {
	hashPool = sync.Pool{
		New: func() interface{} { // skipcq: TCV-001
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

func (h *Handler) putterUpdate(id []byte, data []byte, signer beeCrypto.Signer) error {
	ctx := context.TODO()
	version := time.Now().Unix()
	var nxtIndex feeds.Index
	owner, err := signer.EthereumAddress()
	if err != nil {
		return err
	}
	_, currIndex, at, err := h.getUpdate(ctx, id, owner)
	if err == nil {
		nxtIndex = currIndex.Next(at, uint64(version))
	} else {
		nxtIndex = new(index)
	}

	putter, err := feeds.NewPutter(h, signer, id)
	if err != nil {
		return err
	}

	err = putter.Put(ctx, nxtIndex, version, data)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) getUpdate(ctx context.Context, id []byte, owner common.Address) ([]byte, feeds.Index, int64, error) {
	lk, err := factory.New(h).NewLookup(feeds.Sequence, feeds.New(id, owner))
	if err != nil {
		return nil, nil, 0, err
	}

	ch, current, _, err := lk.At(ctx, time.Now().Unix(), 0)
	if err != nil {
		return nil, nil, 0, err
	}
	if ch == nil {
		return nil, nil, 0, errors.New("invalid chunk lookup")
	}

	data, ts, err := parseFeedUpdate(ch)
	if err != nil {
		return nil, nil, 0, err
	}
	return data, current, ts, err
}

func parseFeedUpdate(ch swarm.Chunk) ([]byte, int64, error) {
	s, err := soc.FromChunk(ch)
	if err != nil {
		return nil, 0, fmt.Errorf("soc unmarshal: %w", err)
	}
	update := s.WrappedChunk().Data()
	ts := binary.BigEndian.Uint64(update[8:16])
	return update[16:], int64(ts), nil
}

func (h *Handler) deleteChunk(ref []byte) error {
	return h.client.DeleteReference(ref)
}

func (*Handler) rawSignedChunkData(chunk swarm.Chunk) ([]byte, error) {
	chunkdata := chunk.Data()
	if len(chunkdata) < idLength+signatureLength+utils.SpanLength {
		return nil, ErrInvalidPayloadSize
	}
	cursor := idLength + signatureLength + utils.SpanLength

	return chunkdata[cursor:], nil
}

func hashFunc() hash.Hash {
	return sha3.NewLegacyKeccak256()
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

// index replicates the feeds.sequence.Index. This index creation is not exported from
// the package as a result loader doesn't know how to return the first feeds.Index.
// This index will be used to return which will provide a compatible interface to
// feeds.sequence.Index.
type index struct {
	index uint64
}

func (i *index) String() string {
	return strconv.FormatUint(i.index, 10)
}

// MarshalBinary
func (i *index) MarshalBinary() ([]byte, error) {
	indexBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(indexBytes, i.index)
	return indexBytes, nil
}

// Next
func (i *index) Next(_ int64, _ uint64) feeds.Index {
	return &index{i.index + 1}
}
