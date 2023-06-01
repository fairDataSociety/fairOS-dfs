package mock

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	goens "github.com/wealdtech/go-ens/v3"

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

// NamespaceManager is a mock ens provider
type NamespaceManager struct {
	storer         map[string]string
	publicResolver map[string]info
	storerMu       sync.RWMutex
}

func (c *NamespaceManager) GetInfoFromNameHash(node [32]byte) (common.Address, *ecdsa.PublicKey, string, error) {
	c.storerMu.Lock()
	defer c.storerMu.Unlock()

	for username, i := range c.publicResolver {
		nh, err := goens.NameHash(username)
		if err != nil {
			return common.Address{}, nil, "", err
		}
		if nh == node {
			addr := c.storer[username]
			if addr == "" {
				return common.Address{}, nil, "", fmt.Errorf("username not available")
			}
			x := new(big.Int)
			x.SetBytes(i.X[:])

			y := new(big.Int)
			y.SetBytes(i.Y[:])
			pub := new(ecdsa.PublicKey)
			pub.X = x
			pub.Y = y

			pub.Curve = btcec.S256()
			return common.HexToAddress(addr), pub, utils.Encode(nh[:]), nil
		}
	}
	return common.Address{}, nil, "", fmt.Errorf("info not available")
}

func (*NamespaceManager) GetNameHash(username string) ([32]byte, error) {
	return goens.NameHash(username)
}

// GetInfo returns the public key of the user
func (c *NamespaceManager) GetInfo(username string) (*ecdsa.PublicKey, string, error) {
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

// RegisterSubdomain registers the username
func (c *NamespaceManager) RegisterSubdomain(username string, owner common.Address, _ *ecdsa.PrivateKey) error {
	c.storerMu.Lock()
	defer c.storerMu.Unlock()
	c.storer[username] = owner.Hex()
	return nil
}

// SetResolver sets the resolver for the username
func (*NamespaceManager) SetResolver(string, common.Address, *ecdsa.PrivateKey) (string, error) {
	// TODO do something
	return "", nil
}

// SetAll sets all the necessary information of the user
func (c *NamespaceManager) SetAll(username string, owner common.Address, key *ecdsa.PrivateKey) error {
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

// NewMockNamespaceManager returns a new mock ENS manager Client
func NewMockNamespaceManager() *NamespaceManager {
	return &NamespaceManager{
		storer:         make(map[string]string),
		publicResolver: make(map[string]info),
		storerMu:       sync.RWMutex{},
	}
}

// GetOwner returns the owner of the username
func (c *NamespaceManager) GetOwner(username string) (common.Address, error) {
	c.storerMu.Lock()
	defer c.storerMu.Unlock()
	addr := c.storer[username]
	if addr == "" {
		return common.Address{}, nil
	}
	return common.HexToAddress(addr), nil
}
