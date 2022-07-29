package mock

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common"
)

type info struct {
	Content   [32]byte
	Multihash []byte
	X         [32]byte
	Y         [32]byte
	Name      string
}

type MockNamespaceManager struct {
	storer         map[string]string
	publicResolver map[string]info
	storerMu       sync.RWMutex
}

func (c *MockNamespaceManager) GetInfo(username string) (*ecdsa.PublicKey, string, error) {
	c.storerMu.Lock()
	defer c.storerMu.Unlock()
	i, ok := c.publicResolver[username]
	if !ok {
		return nil, "", fmt.Errorf("no info available for user")
	}
	x := new(big.Int)
	x.SetBytes(i.X[:])

	y := new(big.Int)
	y.SetBytes(i.Y[:])
	pub := new(ecdsa.PublicKey)
	pub.X = x
	pub.Y = y

	pub.Curve = btcec.S256()
	return pub, "", nil
}

func (c *MockNamespaceManager) RegisterSubdomain(username string, owner common.Address, _ *ecdsa.PrivateKey) error {
	c.storerMu.Lock()
	defer c.storerMu.Unlock()
	c.storer[username] = owner.Hex()
	return nil
}

func (*MockNamespaceManager) SetResolver(string, common.Address, *ecdsa.PrivateKey) (string, error) {
	// TODO do something
	return "", nil
}

func (c *MockNamespaceManager) SetAll(username string, owner common.Address, key *ecdsa.PrivateKey) error {
	name := "subdomain-hidden"
	contentStr := "0x0000000000000000000000000000000000000000000000000000000000000000"
	content := [32]byte{}
	copy(content[:], contentStr)
	publicKey := key.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}
	x := [32]byte{}
	copy(x[:], publicKeyECDSA.X.Bytes())
	y := [32]byte{}
	copy(y[:], publicKeyECDSA.Y.Bytes())
	ret := info{
		Content:   content,
		Multihash: nil,
		X:         x,
		Y:         y,
		Name:      name,
	}
	c.storerMu.Lock()
	defer c.storerMu.Unlock()
	c.publicResolver[username] = ret
	return nil
}

func NewMockNamespaceManager() *MockNamespaceManager {
	return &MockNamespaceManager{
		storer:         make(map[string]string),
		publicResolver: make(map[string]info),
		storerMu:       sync.RWMutex{},
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
