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
	"github.com/ethersphere/bee/pkg/swarm"
	bmtlegacy "github.com/ethersphere/bmt/legacy"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed/lookup"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"golang.org/x/crypto/sha3"
)

const (
	maxuint64 = ^uint64(0)
	idLength  = 32

	maxUpdateRetry = 3
)

var (
	// ErrInvalidTopicSize is returned when a topic is not equal to TopicLength
	ErrInvalidTopicSize = fmt.Errorf("topic is not equal to %d", TopicLength)

	// ErrInvalidPayloadSize is returned when the payload is greater than the chunk size
	ErrInvalidPayloadSize = fmt.Errorf("payload size is too large. maximum payload size is %d bytes", utils.MaxChunkLength)

	// ErrReadOnlyFeed is returned when a feed is read only for a user
	ErrReadOnlyFeed = fmt.Errorf("read only feed")
)

// API handles feed operations
type API struct {
	handler     *Handler
	accountInfo *account.Info
	logger      logging.Logger
}

// request is a custom type that involves in the fairOS feed creation
type request struct {
	ID
	// User   utils.Address
	idAddr swarm.Address // cached chunk address for the update (not serialized, for internal use)

	data       []byte     // actual data payload
	Signature  *Signature // Signature of the payload
	binaryData []byte     // cached serialized data (does not get serialized again!, for efficiency/internal use)
}

// New create the main feed object which is used to create/update/delete feeds.
func New(accountInfo *account.Info, client blockstore.Client, feedCacheSize int, feedCacheTTL time.Duration, logger logging.Logger) *API {
	bmtPool := bmtlegacy.NewTreePool(hashFunc, swarm.Branches, bmtlegacy.PoolSize)
	return &API{
		handler:     NewHandler(accountInfo, client, bmtPool, feedCacheSize, feedCacheTTL, logger),
		accountInfo: accountInfo,
		logger:      logger,
	}
}

func (a *API) CommitFeeds() {
	a.handler.commit()
}

// CreateFeed creates a feed by constructing a single owner chunk. This chunk
// can only be accessed if the pod address is known. Also, no one else can spoof this
// chunk since this is signed by the pod.
func (a *API) CreateFeed(user utils.Address, topic, data, encryptionPassword []byte) error {

	if a.accountInfo.GetPrivateKey() == nil {
		return ErrReadOnlyFeed
	}

	if len(topic) != TopicLength {
		return ErrInvalidTopicSize
	}

	if len(data) > utils.MaxChunkLength {
		return ErrInvalidPayloadSize
	}

	var err error

	encryptedData := data
	if len(encryptionPassword) != 0 { // skipcq: TCV-001
		encryptedData, err = utils.EncryptBytes(encryptionPassword, data)
		if err != nil { // skipcq: TCV-001
			return err
		}
	}

	if a.handler.pool != nil {
		item := &feedItem{
			User:         user,
			AccountInfo:  a.accountInfo,
			Topic:        topic,
			Data:         encryptedData,
			ShouldCreate: true,
		}
		a.handler.putInPool(topic, item)
		return nil
	}
	_, _, err = a.handler.createSoc(user, a.accountInfo, topic, encryptedData)
	if err != nil {
		a.handler.logger.Errorf("failed to createSoc: %v\n", err)
		return err
	}
	return nil
}

// CreateFeedFromTopic creates a soc with the topic as identifier
func (a *API) CreateFeedFromTopic(topic []byte, user utils.Address, data []byte) ([]byte, error) {
	if a.accountInfo.GetPrivateKey() == nil {
		return nil, ErrReadOnlyFeed
	}

	if len(topic) != TopicLength {
		return nil, ErrInvalidTopicSize
	}

	if len(data) > utils.MaxChunkLength {
		return nil, ErrInvalidPayloadSize
	}

	// create the signer and the content addressed chunk
	signer := crypto.NewDefaultSigner(a.accountInfo.GetPrivateKey())
	ch, err := utils.NewChunkWithSpan(data)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	// generate the data to sign
	toSignBytes, err := toSignDigest(topic, ch.Address().Bytes())
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	// sign the chunk
	signature, err := signer.Sign(toSignBytes)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	// send the updated soc chunk to bee
	address, err := a.handler.update(topic, user.ToBytes(), signature, ch.Data())
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	return address, nil
}

// GetSOCFromAddress will download the soc chunk for the given reference
func (a *API) GetSOCFromAddress(address []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	data, err := a.handler.client.DownloadChunk(ctx, address)
	if err != nil {
		return nil, err
	}
	ch := swarm.NewChunk(swarm.NewAddress(address), data)
	return a.handler.rawSignedChunkData(ch)
}

// GetFeedData looks up feed from swarm
func (a *API) GetFeedData(topic []byte, user utils.Address, encryptionPassword []byte, isFeedUpdater bool) ([]byte, []byte, error) {
	if len(topic) != TopicLength {
		return nil, nil, ErrInvalidTopicSize
	}

	hint := lookup.NoClue

	addr, data, err := a.handler.getSoc(topic, user, hint)
	if err != nil {
		return nil, nil, err
	}
	if len(encryptionPassword) == 0 || string(data) == utils.DeletedFeedMagicWord {
		return addr, data, nil
	}
	decryptedData, err := utils.DecryptBytes(encryptionPassword, data)
	if err != nil { // skipcq: TCV-001
		return nil, nil, err
	}
	return addr, decryptedData, nil
}

// GetFeedDataFromTopic will generate keccak256 reference of the topic+address and download soc
func (a *API) GetFeedDataFromTopic(topic []byte, user utils.Address) ([]byte, []byte, error) {
	if len(topic) != TopicLength {
		return nil, nil, ErrInvalidTopicSize
	}
	// generate reference
	h := sha3.NewLegacyKeccak256()
	_, err := h.Write(topic)
	if err != nil { // skipcq: TCV-001
		return nil, nil, err
	}
	_, err = h.Write(user.ToBytes())
	if err != nil { // skipcq: TCV-001
		return nil, nil, err
	}
	hash := h.Sum(nil)

	// download soc from generated reference
	data, err := a.GetSOCFromAddress(hash)
	if err != nil {
		return nil, nil, err
	}
	return hash, data, nil
}

// UpdateFeed updates the contents of an already created feed.
func (a *API) UpdateFeed(user utils.Address, topic, data, encryptionPassword []byte, isFeedUpdater bool) error {
	if a.accountInfo.GetPrivateKey() == nil {
		return ErrReadOnlyFeed
	}

	if len(topic) != TopicLength {
		return ErrInvalidTopicSize
	}

	if len(data) > utils.MaxChunkLength {
		return ErrInvalidPayloadSize
	}

	var err error

	encryptedData := data
	if len(encryptionPassword) != 0 && string(data) != utils.DeletedFeedMagicWord {
		encryptedData, err = utils.EncryptBytes(encryptionPassword, data)
		if err != nil { // skipcq: TCV-001
			return err
		}
	}
	if a.handler.pool != nil {
		item := &feedItem{
			User:         user,
			AccountInfo:  a.accountInfo,
			Topic:        topic,
			Data:         encryptedData,
			ShouldCreate: false,
		}
		a.handler.putInPool(topic, item)
		return nil
	}
	_, _, err = a.handler.updateSoc(user, a.accountInfo, topic, encryptedData)
	if err != nil {
		a.handler.logger.Errorf("failed to updateSoc: %v\n", err)
		return err
	}
	return nil
}

// DeleteFeed deleted the feed by updating with no data inside the SOC chunk.
func (a *API) DeleteFeed(topic []byte, user utils.Address) error {
	if a.accountInfo.GetPrivateKey() == nil {
		return ErrReadOnlyFeed
	}

	delRef, _, err := a.GetFeedData(topic, user, nil, false)
	if err != nil && err.Error() != "feed does not exist or was not updated yet" { // skipcq: TCV-001
		return err
	}
	if delRef != nil {
		err = a.handler.deleteChunk(delRef)
		if err != nil { // skipcq: TCV-001
			return err
		}
	}
	return nil
}

// DeleteFeedFromTopic deleted the feed by updating with no data inside the SOC chunk.
func (a *API) DeleteFeedFromTopic(topic []byte, user utils.Address) error {
	if a.accountInfo.GetPrivateKey() == nil {
		return ErrReadOnlyFeed
	}

	delRef, _, err := a.GetFeedDataFromTopic(topic, user)
	if err != nil && err.Error() != "feed does not exist or was not updated yet" { // skipcq: TCV-001
		return err
	}
	if delRef != nil {
		err = a.handler.deleteChunk(delRef)
		if err != nil { // skipcq: TCV-001
			return err
		}
	}
	return nil
}

// IsReadOnlyFeed if a public pod is imported, the feed can only be read.
// this function check the feed is read only.
// skipcq: TCV-001
func (a *API) IsReadOnlyFeed() bool {
	return a.accountInfo.GetPrivateKey() == nil
}

func (a *API) Close() error {
	a.CommitFeeds()
	return nil
}
