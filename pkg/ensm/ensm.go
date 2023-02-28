// Package ensm stands for ENS Manager for fairOS.
// It initialises an eth client and only exposes the essential functionalities for fairOS
package ensm

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
)

// ENSManager interface takes care of the ens based account management
type ENSManager interface {
	GetOwner(username string) (common.Address, error)
	RegisterSubdomain(username string, owner common.Address, key *ecdsa.PrivateKey) error
	SetResolver(username string, owner common.Address, key *ecdsa.PrivateKey) (string, error)
	SetAll(username string, owner common.Address, key *ecdsa.PrivateKey) error
	GetInfo(username string) (*ecdsa.PublicKey, string, error)
	GetInfoFromNameHash(node [32]byte) (common.Address, *ecdsa.PublicKey, string, error)
	GetNameHash(username string) ([32]byte, error)
}
