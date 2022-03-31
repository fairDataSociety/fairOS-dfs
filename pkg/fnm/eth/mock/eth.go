package mock

import (
	"crypto/ecdsa"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type MockNamespaceManager struct {
	storer   map[string]string
	storerMu sync.RWMutex
}

func (c *MockNamespaceManager) RegisterSubdomain(username string, owner common.Address) error {
	c.storerMu.Lock()
	defer c.storerMu.Unlock()
	c.storer[username] = owner.Hex()
	return nil
}

func (c *MockNamespaceManager) SetResolver(username string) error {
	// TODO do something
	return nil
}

func (c *MockNamespaceManager) SetAll(username string, owner common.Address, publicKey *ecdsa.PublicKey) error {
	// TODO do something
	return nil
}

func NewMockNamespaceManager() *MockNamespaceManager {
	return &MockNamespaceManager{
		storer:   make(map[string]string),
		storerMu: sync.RWMutex{},
	}
}

func (c *MockNamespaceManager) GetOwner(username string) (common.Address, error) {
	c.storerMu.Lock()
	defer c.storerMu.Unlock()
	addr := c.storer[username]
	if addr == "" {
		return common.Address{}, nil
	}
	return common.HexToAddress(addr), nil
}
