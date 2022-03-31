package mock

import (
	"crypto/ecdsa"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type MockNamespaceManager struct {
	storer   map[string]*common.Address
	storerMu sync.RWMutex
}

func (c *MockNamespaceManager) RegisterSubdomain(username string, owner common.Address) error {
	panic("implement me")
}

func (c *MockNamespaceManager) SetResolver(username string) error {
	panic("implement me")
}

func (c *MockNamespaceManager) SetAll(username string, owner common.Address, publicKey *ecdsa.PublicKey) error {
	panic("implement me")
}

func NewMockNamespaceManager() *MockNamespaceManager {
	return &MockNamespaceManager{
		storer:   make(map[string]*common.Address),
		storerMu: sync.RWMutex{},
	}
}

func (c *MockNamespaceManager) GetOwner(username string) (common.Address, error) {
	c.storerMu.Lock()
	defer c.storerMu.Unlock()
	addr := c.storer[username]
	if addr == nil {
		return common.Address{}, nil
	}
	return *addr, nil
}
