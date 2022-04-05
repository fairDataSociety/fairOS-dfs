// Package fnm stands for FairOS Namespace Manager in the xDai Chain.
// It initialises an eth and only exposes the essential functionalities for fairOS
package fnm

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
)

type FairOSNamespaceManager interface {
	GetOwner(username string) (common.Address, error)
	RegisterSubdomain(username string, owner common.Address) error
	SetResolver(username string, owner common.Address, key *ecdsa.PrivateKey) error
	SetAll(username string, owner common.Address, key *ecdsa.PrivateKey) error
	GetPublicKey(username string) (*ecdsa.PublicKey, error)
	Fund(owner common.Address) error
}
