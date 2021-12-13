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

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	"github.com/ethersphere/bee/pkg/soc"
	"github.com/ethersphere/bee/pkg/swarm"
)

type MockBeeClient struct {
	storer   map[string][]byte
	storerMu sync.RWMutex
}

func NewMockBeeClient() *MockBeeClient {
	return &MockBeeClient{
		storer:   make(map[string][]byte),
		storerMu: sync.RWMutex{},
	}
}

func (*MockBeeClient) CheckConnection() bool {
	return true
}

func (m *MockBeeClient) UploadSOC(owner, id, signature string, data []byte) (address []byte, err error) {
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
	signedChunk, err := soc.NewSignedChunk(idBytes, ch, ownerBytes, signatureBytes)
	if err != nil {
		return nil, err
	}
	if !soc.Valid(signedChunk) {
		return nil, fmt.Errorf("soc chunk failed in validation")
	}
	m.storer[signedChunk.Address().String()] = signedChunk.Data()
	return signedChunk.Address().Bytes(), nil
}

func (m *MockBeeClient) UploadChunk(ch swarm.Chunk, pin bool) (address []byte, err error) {
	m.storerMu.Lock()
	defer m.storerMu.Unlock()
	m.storer[ch.Address().String()] = ch.Data()
	return ch.Address().Bytes(), nil
}

func (m *MockBeeClient) DownloadChunk(ctx context.Context, address []byte) (data []byte, err error) {
	m.storerMu.Lock()
	defer m.storerMu.Unlock()
	if data, ok := m.storer[swarm.NewAddress(address).String()]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("error downloading data")
}

func (m *MockBeeClient) UploadBlob(data []byte, pin, encrypt bool) (address []byte, err error) {
	m.storerMu.Lock()
	defer m.storerMu.Unlock()
	address = make([]byte, 32)
	_, err = rand.Read(address)
	m.storer[swarm.NewAddress(address).String()] = data
	return address, nil
}

func (m *MockBeeClient) DownloadBlob(address []byte) (data []byte, respCode int, err error) {
	m.storerMu.Lock()
	defer m.storerMu.Unlock()
	if data, ok := m.storer[swarm.NewAddress(address).String()]; ok {
		return data, http.StatusOK, nil
	}
	return nil, http.StatusInternalServerError, fmt.Errorf("error downloading data")
}

func (m *MockBeeClient) DeleteChunk(address []byte) error {
	m.storerMu.Lock()
	defer m.storerMu.Unlock()
	if _, found := m.storer[swarm.NewAddress(address).String()]; found {
		delete(m.storer, swarm.NewAddress(address).String())
		return nil
	}
	return errors.New("chunk not found")
}

func (m *MockBeeClient) DeleteBlob(address []byte) error {
	m.storerMu.Lock()
	defer m.storerMu.Unlock()
	if _, found := m.storer[swarm.NewAddress(address).String()]; found {
		delete(m.storer, swarm.NewAddress(address).String())
		return nil
	}
	return errors.New("blob not found")
}

func (*MockBeeClient) GetNewPostageBatch() error {
	return nil
}
