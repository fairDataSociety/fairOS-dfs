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
	"fmt"
	"time"

	"github.com/ethersphere/bee/pkg/crypto"
	"github.com/ethersphere/bee/pkg/soc"
	"github.com/ethersphere/bee/pkg/swarm"
	bmtlegacy "github.com/ethersphere/bmt/legacy"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed/lookup"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	maxuint64 = ^uint64(0)
	idLength  = 32
)

var (
	// ErrInvalidTopicSize is returned when a topic is not equal to TopicLength
	ErrInvalidTopicSize = fmt.Errorf("Topic is not equal to %d", TopicLength)

	// ErrInvalidPayloadSize is returned when the payload is greater than the chunk size
	ErrInvalidPayloadSize = fmt.Errorf("payload size is too large. maximum payload size is %d bytes", utils.MaxChunkLength)

	ErrReadOnlyFeed = fmt.Errorf("read only feed")
)

type API struct {
	handler     *Handler
	accountInfo *account.Info
	logger      logging.Logger
}

type Request struct {
	ID
	//User   utils.Address
	idAddr swarm.Address // cached chunk address for the update (not serialized, for internal use)

	data       []byte     // actual data payload
	Signature  *Signature // Signature of the payload
	binaryData []byte     // cached serialized data (does not get serialized again!, for efficiency/internal use)
}

// New create the main feed object which is used to create/update/delete feeds.
func New(accountInfo *account.Info, client blockstore.Client, logger logging.Logger) *API {
	bmtPool := bmtlegacy.NewTreePool(hashFunc, swarm.Branches, bmtlegacy.PoolSize)
	return &API{
		handler:     NewHandler(accountInfo, client, bmtPool),
		accountInfo: accountInfo,
		logger:      logger,
	}
}

// CreateFeed creates a feed by constructing a single owner chunk. This chunk
// can only be accessed if the pod address is known. Also no one else can spoof this
// chunk since this is signed by the pod.
func (a *API) CreateFeed(topic []byte, user utils.Address, data []byte) ([]byte, error) {
	var req Request

	if a.accountInfo.GetPrivateKey() == nil {
		return nil, ErrReadOnlyFeed
	}

	if len(topic) != TopicLength {
		return nil, ErrInvalidTopicSize
	}

	if len(data) > utils.MaxChunkLength {
		return nil, ErrInvalidPayloadSize
	}
	// fill Feed and Epoc related details
	copy(req.ID.Topic[:], topic)
	req.ID.User = user
	req.Epoch.Level = 31
	req.Epoch.Time = uint64(time.Now().Unix())

	// Add initial feed data
	req.data = data

	// create the id, hash(topic, epoc)
	id, err := a.handler.getId(req.Topic, req.Time, req.Level)
	if err != nil {
		return nil, err
	}

	// get the payload id BMT(span, payload)
	payloadId, err := a.handler.getPayloadId(data)
	if err != nil {
		return nil, err
	}

	// create the signer and the content addressed chunk
	signer := crypto.NewDefaultSigner(a.accountInfo.GetPrivateKey())
	ch, err := utils.NewChunkWithSpan(data)
	if err != nil {
		return nil, err
	}
	s := soc.New(id, ch)
	sch, err := s.Sign(signer)
	if err != nil {
		return nil, err
	}

	// generate the data to sign
	toSignBytes, err := toSignDigest(id, ch.Address().Bytes())
	if err != nil {
		return nil, err
	}

	// sign the chunk
	signature, err := signer.Sign(toSignBytes)
	if err != nil {
		return nil, err
	}

	// set the address and the data for the soc chunk
	req.idAddr = sch.Address()
	req.binaryData = sch.Data()

	// set signature and binary data fields
	_, err = a.handler.toChunkContent(&req, id, payloadId)
	if err != nil {
		return nil, err
	}

	// send the updated soc chunk to bee
	address, err := a.handler.update(id, user.ToBytes(), signature, ch.Data())
	if err != nil {
		return nil, err
	}

	return address, nil
}

func (a *API) GetFeedDataFromAddress(address []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	data, err := a.handler.client.DownloadChunk(ctx, address)
	if err != nil {
		return nil, err
	}
	ch, err := utils.NewChunkWithoutSpan(data)
	if err != nil {
		return nil, err
	}
	return a.handler.rawSignedChunkData(ch)
}

func (a *API) GetFeedData(topic []byte, user utils.Address) ([]byte, []byte, error) {
	if len(topic) != TopicLength {
		return nil, nil, ErrInvalidTopicSize
	}
	ctx := context.Background()
	f := new(Feed)
	f.User = user
	copy(f.Topic[:], topic)

	// create the query from values
	q := &Query{Feed: *f}
	q.TimeLimit = 0
	q.Hint = lookup.NoClue
	_, err := a.handler.Lookup(ctx, q)
	if err != nil {
		return nil, nil, err
	}
	var data []byte
	addr, data, err := a.handler.GetContent(&q.Feed)
	if err != nil {
		return nil, nil, err
	}
	return addr.Bytes(), data, nil
}

// UpdateFeed updates the contents of an already created feed.
func (a *API) UpdateFeed(topic []byte, user utils.Address, data []byte) ([]byte, error) {
	if a.accountInfo.GetPrivateKey() == nil {
		return nil, ErrReadOnlyFeed
	}

	if len(topic) != TopicLength {
		return nil, ErrInvalidTopicSize
	}

	if len(data) > utils.MaxChunkLength {
		return nil, ErrInvalidPayloadSize
	}

	ctx := context.Background()
	f := new(Feed)
	f.User = user
	copy(f.Topic[:], topic)

	// get the existing request from DB
	req, err := a.handler.NewRequest(ctx, f)
	if err != nil {
		return nil, err
	}
	req.Time = uint64(time.Now().Unix())
	req.data = data

	// create the id, hash(topic, epoc)
	id, err := a.handler.getId(req.Topic, req.Time, req.Level)
	if err != nil {
		return nil, err
	}

	// get the payload id BMT(span, payload)
	payloadId, err := a.handler.getPayloadId(data)
	if err != nil {
		return nil, err
	}

	// create the signer and the content addressed chunk
	signer := crypto.NewDefaultSigner(a.accountInfo.GetPrivateKey())
	ch, err := utils.NewChunkWithSpan(data)
	if err != nil {
		return nil, err
	}
	s := soc.New(id, ch)
	sch, err := s.Sign(signer)
	if err != nil {
		return nil, err
	}

	// generate the data to sign
	toSignBytes, err := toSignDigest(id, ch.Address().Bytes())
	if err != nil {
		return nil, err
	}

	// sign the chunk
	signature, err := signer.Sign(toSignBytes)
	if err != nil {
		return nil, err
	}

	// set the address and the data for the soc chunk
	req.idAddr = sch.Address()
	req.binaryData = sch.Data()

	// set signature and binary data fields
	_, err = a.handler.toChunkContent(req, id, payloadId)
	if err != nil {
		return nil, err
	}

	address, err := a.handler.update(id, user.ToBytes(), signature, ch.Data())
	if err != nil {
		return nil, err
	}
	return address, nil
}

// DeleteFeed deleted the feed by updating with no data inside the SOC chunk.
func (a *API) DeleteFeed(topic []byte, user utils.Address) error {
	if a.accountInfo.GetPrivateKey() == nil {
		return ErrReadOnlyFeed
	}

	delRef, _, err := a.GetFeedData(topic, user)
	if err != nil && err.Error() != "feed does not exist or was not updated yet" {
		return err
	}
	if delRef != nil {
		err = a.handler.deleteChunk(delRef)
		if err != nil {
			return err
		}
	}
	return nil
}

// IsReadOnlyFeed if a public pod is imported, the feed can only be read.
// this function check the feed is read only.
func (a *API) IsReadOnlyFeed() bool {
	return a.accountInfo.GetPrivateKey() == nil
}
