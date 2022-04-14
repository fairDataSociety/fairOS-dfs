// Package ensm stands for ENS Manager for fairOS.
// It initialises an eth client and only exposes the essential functionalities for fairOS
package ensm

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
)

type ENSManager interface {
	GetOwner(username string) (common.Address, error)
	RegisterSubdomain(username string, owner common.Address) error
	SetResolver(username string, owner common.Address, key *ecdsa.PrivateKey) (string, error)
	SetAll(username string, owner common.Address, key *ecdsa.PrivateKey) error
	GetInfo(username string) (*ecdsa.PublicKey, string, error)
}
