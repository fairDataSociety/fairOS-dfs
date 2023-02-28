package user

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
)

func (u *Users) GetUserInfoFromENS(nameHash [32]byte) (common.Address, *ecdsa.PublicKey, error) {
	addr, publicKey, _, err := u.ens.GetInfoFromNameHash(nameHash)
	return addr, publicKey, err
}

func (u *Users) GetNameHash(username string) ([32]byte, error) {
	return u.ens.GetNameHash(username)
}
