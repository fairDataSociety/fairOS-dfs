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

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethersphere/bee/pkg/crypto"
	"github.com/ethersphere/bee/pkg/swarm"
	bmtlegacy "github.com/ethersphere/bmt/legacy"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"golang.org/x/crypto/sha3"
)

const (
	idLength    = 32
	topicLength = 32
)

var (
	// ErrInvalidTopicSize is returned when a topic is not equal to topicLength
	ErrInvalidTopicSize = fmt.Errorf("topic is not equal to %d", topicLength)

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

// New create the main feed object which is used to create/update/delete feeds.
func New(accountInfo *account.Info, client blockstore.Client, logger logging.Logger) *API {
	bmtPool := bmtlegacy.NewTreePool(hashFunc, swarm.Branches, bmtlegacy.PoolSize)
	h := NewHandler(accountInfo, client, bmtPool)
	return &API{
		handler:     h,
		accountInfo: accountInfo,
		logger:      logger,
	}
}

func (a *API) putFeed(topic []byte, data []byte, signer crypto.Signer) (err error) {
	return a.handler.putterUpdate(topic, data, signer)
}

// CreateFeedFromTopic creates a soc with the topic as identifier
func (a *API) CreateFeedFromTopic(topic []byte, user utils.Address, data []byte) error {
	if a.accountInfo.GetPrivateKey() == nil {
		return ErrReadOnlyFeed
	}

	if len(topic) != topicLength {
		return ErrInvalidTopicSize
	}

	if len(data) > utils.MaxChunkLength {
		return ErrInvalidPayloadSize
	}
	// create the signer and the content addressed chunk
	signer := crypto.NewDefaultSigner(a.accountInfo.GetPrivateKey())
	ch, err := utils.NewChunkWithSpan(data)
	if err != nil { // skipcq: TCV-001
		return err
	}

	// generate the data to sign
	toSignBytes, err := toSignDigest(topic, ch.Address().Bytes())
	if err != nil { // skipcq: TCV-001
		return err
	}

	// sign the chunk
	signature, err := signer.Sign(toSignBytes)
	if err != nil { // skipcq: TCV-001
		return err
	}

	// send the updated soc chunk to bee
	_, err = a.handler.update(topic, user.ToBytes(), signature, ch.Data())
	if err != nil { // skipcq: TCV-001
		return err
	}
	return nil
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
func (a *API) GetFeedData(topic []byte, user utils.Address, encryptionPassword []byte) ([]byte, error) {
	if len(topic) != topicLength {
		return nil, ErrInvalidTopicSize
	}

	encryptedData, _, _, err := a.handler.getUpdate(context.TODO(), topic, common.BytesToAddress(a.accountInfo.GetAddress().ToBytes()))
	if err != nil {
		a.logger.Errorf("failed looking up key %s", err.Error())
		return nil, fmt.Errorf("feed does not exist or was not updated yet")
	}

	if encryptionPassword == nil || string(encryptedData) == utils.DeletedFeedMagicWord {
		return encryptedData, nil
	}
	data, err := utils.DecryptBytes(encryptionPassword, encryptedData)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	return data, nil
}

// GetFeedDataFromTopic will generate keccak256 reference of the topic+address and download soc
func (a *API) GetFeedDataFromTopic(topic []byte, user utils.Address) ([]byte, []byte, error) {
	if len(topic) != topicLength {
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
func (a *API) UpdateFeed(topic []byte, user utils.Address, data []byte, encryptionPassword []byte) error {
	if a.accountInfo.GetPrivateKey() == nil {
		return ErrReadOnlyFeed
	}

	if len(topic) != topicLength {
		return ErrInvalidTopicSize
	}

	if len(data) > utils.MaxChunkLength {
		return ErrInvalidPayloadSize
	}

	var err error

	encryptedData := data
	if encryptionPassword != nil && string(data) != utils.DeletedFeedMagicWord {
		encryptedData, err = utils.EncryptBytes(encryptionPassword, data)
		if err != nil { // skipcq: TCV-001
			return err
		}
	}
	return a.putFeed(topic, encryptedData, crypto.NewDefaultSigner(a.accountInfo.GetPrivateKey()))
}

// DeleteFeed deleted the feed by updating with no data inside the SOC chunk.
func (a *API) DeleteFeed(topic []byte, user utils.Address) error {
	if a.accountInfo.GetPrivateKey() == nil {
		return ErrReadOnlyFeed
	}
	//
	//_, err := a.GetFeedData(topic, user, nil)
	//if err != nil && err.Error() != "feed does not exist or was not updated yet" { // skipcq: TCV-001
	//	return err
	//}
	//if delRef != nil {
	//	err = a.handler.deleteChunk(delRef)
	//	if err != nil { // skipcq: TCV-001
	//		return err
	//	}
	//}
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
