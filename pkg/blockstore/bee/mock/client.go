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

package mock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	"github.com/ethersphere/bee/pkg/soc"
	"github.com/ethersphere/bee/pkg/swarm"
)

// BeeClient is a mock bee client
type BeeClient struct {
	storer    map[string][]byte
	tagStorer map[uint32]int64
	storerMu  sync.RWMutex
}

// NewMockBeeClient returns a mock bee client
func NewMockBeeClient() *BeeClient {
	return &BeeClient{
		storer:    make(map[string][]byte),
		tagStorer: make(map[uint32]int64),
		storerMu:  sync.RWMutex{},
	}
}

// CheckConnection checks connection
func (*BeeClient) CheckConnection() bool {
	return true
}

// UploadSOC uploads soc into swarm
func (m *BeeClient) UploadSOC(owner, id, signature string, data []byte) (address []byte, err error) {
	m.storerMu.Lock()
	defer m.storerMu.Unlock()
	ch, err := utils.NewChunkWithoutSpan(data)
	if err != nil {
		return nil, err
	}
	idBytes, err := hex.DecodeString(id)
	if err != nil {
		return nil, err
	}
	ownerBytes, err := hex.DecodeString(owner)
	if err != nil {
		return nil, err
	}
	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return nil, err
	}
	signed, err := soc.NewSigned(idBytes, ch, ownerBytes, signatureBytes)
	if err != nil {
		return nil, err
	}
	signedChunk, err := signed.Chunk()
	if err != nil {
		return nil, err
	}
	if !soc.Valid(signedChunk) {
		return nil, fmt.Errorf("soc chunk failed in validation")
	}
	m.storer[signedChunk.Address().String()] = signedChunk.Data()
	return signedChunk.Address().Bytes(), nil
}

// UploadChunk into swarm
func (m *BeeClient) UploadChunk(ch swarm.Chunk) (address []byte, err error) {
	m.storerMu.Lock()
	defer m.storerMu.Unlock()
	m.storer[ch.Address().String()] = ch.Data()
	return ch.Address().Bytes(), nil
}

// DownloadChunk from swarm
func (m *BeeClient) DownloadChunk(_ context.Context, address []byte) (data []byte, err error) {
	m.storerMu.Lock()
	defer m.storerMu.Unlock()
	if data, ok := m.storer[swarm.NewAddress(address).String()]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("error downloading data")
}

// UploadBlob into swarm
func (m *BeeClient) UploadBlob(data []byte, tag uint32, _ bool) (address []byte, err error) {
	m.storerMu.Lock()
	defer m.storerMu.Unlock()
	address = make([]byte, 32)
	_, err = rand.Read(address)
	newChunks := int64(len(data) / 4096000)
	if newChunks == 0 {
		newChunks = 1
	}
	chunks := newChunks + m.tagStorer[tag] + 1
	m.tagStorer[tag] = chunks
	m.storer[swarm.NewAddress(address).String()] = data
	return address, nil
}

// DownloadBlob from swarm
func (m *BeeClient) DownloadBlob(address []byte) ([]byte, int, error) {
	m.storerMu.Lock()
	defer m.storerMu.Unlock()
	if data, ok := m.storer[swarm.NewAddress(address).String()]; ok {
		return data, http.StatusOK, nil
	}
	return nil, http.StatusInternalServerError, fmt.Errorf("error downloading data")
}

// DeleteReference unpins chunk in swarm
func (m *BeeClient) DeleteReference(address []byte) error {
	m.storerMu.Lock()
	defer m.storerMu.Unlock()
	if _, found := m.storer[swarm.NewAddress(address).String()]; found {
		delete(m.storer, swarm.NewAddress(address).String())
		return nil
	}
	return errors.New("chunk not found")
}

// CreateTag
func (m *BeeClient) CreateTag(_ []byte) (uint32, error) {
	tag := time.Now().UnixNano()
	m.storerMu.Lock()
	defer m.storerMu.Unlock()
	m.tagStorer[uint32(tag)] = 0
	return uint32(tag), nil
}

// GetTag
func (m *BeeClient) GetTag(tag uint32) (int64, int64, int64, error) {
	m.storerMu.Lock()
	defer m.storerMu.Unlock()
	return m.tagStorer[tag], m.tagStorer[tag], m.tagStorer[tag], nil
}
