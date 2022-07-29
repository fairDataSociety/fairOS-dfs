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

package account

import (
	"bytes"
	"crypto/aes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/fairdatasociety/fairOS-dfs-utils/crypto"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/tyler-smith/go-bip39"
)

const (
	// UserAccountIndex is user root account
	UserAccountIndex = -1

	// chunkSize is used to set chunk size of the portable account SOC
	chunkSize = 4096

	// seedSize is used to determine how much padding we need for portable account SOC
	seedSize = 64
)

var ErrBlankPassword = errors.New("password cannot be blank")

// Account is used for keeping authenticated logged-in user info in the session
type Account struct {
	wallet      *Wallet
	userAccount *Info
	podAccounts map[int]*Info
	logger      logging.Logger
}

// Info is for keeping account info
type Info struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	address    utils.Address
}

// New create a account object through which the entire account management is done.
// it uses a 12 word BIP-0039 wordlist to create a 12 word mnemonic for every user
// and spawns key pais whenever necessary.
func New(logger logging.Logger) *Account {
	wal := NewWalletFromMnemonic("")
	return &Account{
		wallet:      wal,
		userAccount: &Info{},
		podAccounts: make(map[int]*Info),
		logger:      logger,
	}
}

// CreateRandomKeyPair creates a ecdsa key pair by using the given int64 number
// as the random number.
func CreateRandomKeyPair(now int64) (*ecdsa.PrivateKey, error) {
	randBytes := make([]byte, 40)
	binary.LittleEndian.PutUint64(randBytes, uint64(now))
	randReader := bytes.NewReader(randBytes)
	return ecdsa.GenerateKey(btcec.S256(), randReader)
}

// CreateUserAccount create a new master account for a user. if a valid mnemonic is
// provided it is used, otherwise a new mnemonic is generated. The generated mnemonic is
// AES encrypted using the password provided.
func (a *Account) CreateUserAccount(passPhrase, mnemonic string) (string, string, error) {
	wal := NewWalletFromMnemonic("")
	a.wallet = wal
	acc, mnemonic, err := wal.LoadMnemonicAndCreateRootAccount(mnemonic)
	if err != nil {
		return "", "", err
	}

	hdw, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return "", "", err
	}

	// store publicKey, private key and user
	a.userAccount.privateKey, err = hdw.PrivateKey(acc)
	if err != nil {
		return "", "", err
	}
	a.userAccount.publicKey, err = hdw.PublicKey(acc)
	if err != nil {
		return "", "", err
	}
	addrBytes, err := crypto.NewEthereumAddress(a.userAccount.privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	a.userAccount.address.SetBytes(addrBytes)

	// store the mnemonic
	encryptedMnemonic, err := a.encryptMnemonic(mnemonic, passPhrase)
	if err != nil {
		return "", "", err
	}
	a.wallet.encryptedmnemonic = encryptedMnemonic

	return mnemonic, encryptedMnemonic, nil
}

// LoadUserAccount loads the user account given the encrypted mnemonic and
// password.
func (a *Account) LoadUserAccount(passPhrase, encryptedMnemonic string) error {
	password := passPhrase
	if password == "" {
		return ErrBlankPassword
	}

	a.wallet.encryptedmnemonic = encryptedMnemonic
	plainMnemonic, err := a.wallet.decryptMnemonic(password)
	if err != nil {
		return fmt.Errorf("invalid password")
	}

	acc, err := a.wallet.CreateAccount(rootPath, plainMnemonic)
	if err != nil {
		return err
	}

	hdw, err := hdwallet.NewFromMnemonic(plainMnemonic)
	if err != nil {
		return err
	}
	a.userAccount.privateKey, err = hdw.PrivateKey(acc)
	if err != nil {
		return err
	}
	a.userAccount.publicKey, err = hdw.PublicKey(acc)
	if err != nil {
		return err
	}
	addrBytes, err := crypto.NewEthereumAddress(a.userAccount.privateKey.PublicKey)
	if err != nil {
		return err
	}
	a.userAccount.address.SetBytes(addrBytes)
	return nil
}

// LoadUserAccountFromSeed loads the user account given the bip39 seed
func (a *Account) LoadUserAccountFromSeed(seed []byte) error {
	acc, err := a.wallet.CreateAccountFromSeed(rootPath, seed)
	if err != nil {
		return err
	}
	hdw, err := hdwallet.NewFromSeed(seed)
	if err != nil {
		return err
	}
	a.userAccount.privateKey, err = hdw.PrivateKey(acc)
	if err != nil {
		return err
	}
	a.userAccount.publicKey, err = hdw.PublicKey(acc)
	if err != nil {
		return err
	}
	addrBytes, err := crypto.NewEthereumAddress(a.userAccount.privateKey.PublicKey)
	if err != nil {
		return err
	}
	a.userAccount.address.SetBytes(addrBytes)
	return nil
}

// Authorise is used to check if the given password is valid for an user account.
// this is done by decrypting the mnemonic using the supplied password and checking
// the validity of the mnemonic to see if it confirms to bip-0039 list of words.
func (a *Account) Authorise(password string) bool {
	if password == "" {
		a.logger.Errorf(ErrBlankPassword.Error())
		return false
	}
	plainMnemonic, err := a.wallet.decryptMnemonic(password)
	if err != nil {
		return false
	}
	// check the validity of the mnemonic
	if plainMnemonic == "" {
		return false
	}
	words := strings.Split(plainMnemonic, " ")
	if len(words) != 12 {
		return false
	}
	if !bip39.IsMnemonicValid(plainMnemonic) {
		return false
	}
	return true
}

// CreatePodAccount is used to create a new key pair from the master mnemonic. this key pair is
// used as the base key pair for a newly created pod.
func (a *Account) CreatePodAccount(accountId int, passPhrase string, createPod bool) (*Info, error) {
	if acc, ok := a.podAccounts[accountId]; ok {
		return acc, nil
	}
	var (
		hdw         *hdwallet.Wallet
		err         error
		acc         accounts.Account
		accountInfo = &Info{}
	)
	path := genericPath + strconv.Itoa(accountId)
	if a.wallet.seed != nil {
		acc, err = a.wallet.CreateAccountFromSeed(path, a.wallet.seed)
		if err != nil {
			return nil, err
		}
		hdw, err = hdwallet.NewFromSeed(a.wallet.seed)
		if err != nil {
			return nil, err
		}
	} else {
		password := passPhrase
		if password == "" {
			return nil, ErrBlankPassword
		}

		plainMnemonic, err := a.wallet.decryptMnemonic(password)
		if err != nil {
			return nil, fmt.Errorf("invalid password")
		}

		acc, err = a.wallet.CreateAccount(path, plainMnemonic)
		if err != nil {
			return nil, err
		}
		hdw, err = hdwallet.NewFromMnemonic(plainMnemonic)
		if err != nil {
			return nil, err
		}
	}

	accountInfo.privateKey, err = hdw.PrivateKey(acc)
	if err != nil {
		return nil, err
	}
	accountInfo.publicKey, err = hdw.PublicKey(acc)
	if err != nil {
		return nil, err
	}
	addrBytes, err := crypto.NewEthereumAddress(accountInfo.privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	accountInfo.address.SetBytes(addrBytes)
	a.podAccounts[accountId] = accountInfo
	return accountInfo, nil
}

// CreateCollectionAccount is used to create a new key pair for every collection (KV or Doc) created. This
// key pair is again derived from the same master mnemonic of the user.
func (a *Account) CreateCollectionAccount(accountId int, passPhrase string, createCollection bool) error {
	if _, ok := a.podAccounts[accountId]; ok {
		return nil
	}

	var (
		hdw         *hdwallet.Wallet
		err         error
		acc         accounts.Account
		accountInfo = &Info{}
	)
	path := genericPath + strconv.Itoa(accountId)
	if a.wallet.seed != nil {
		acc, err = a.wallet.CreateAccountFromSeed(path, a.wallet.seed)
		if err != nil {
			return err
		}
		hdw, err = hdwallet.NewFromSeed(a.wallet.seed)
		if err != nil {
			return err
		}
	} else {
		password := passPhrase
		if password == "" {
			return ErrBlankPassword
		}

		plainMnemonic, err := a.wallet.decryptMnemonic(password)
		if err != nil {
			return fmt.Errorf("invalid password")
		}

		acc, err = a.wallet.CreateAccount(path, plainMnemonic)
		if err != nil {
			return err
		}
		hdw, err = hdwallet.NewFromMnemonic(plainMnemonic)
		if err != nil {
			return err
		}
	}

	accountInfo.privateKey, err = hdw.PrivateKey(acc)
	if err != nil {
		return err
	}
	accountInfo.publicKey, err = hdw.PublicKey(acc)
	if err != nil {
		return err
	}
	addrBytes, err := crypto.NewEthereumAddress(accountInfo.privateKey.PublicKey)
	if err != nil {
		return err
	}
	accountInfo.address.SetBytes(addrBytes)
	a.podAccounts[accountId] = accountInfo
	return nil
}

// DeletePodAccount unloads/forgets a particular pods key value pair from the memory.
func (a *Account) DeletePodAccount(accountId int) {
	delete(a.podAccounts, accountId)
}

// GetUserPrivateKey retuens the private key of a given account index.
// the index -1 belongs to user root account and other indexes belong to
// the respective pods.
func (a *Account) GetUserPrivateKey(index int) *ecdsa.PrivateKey {
	if index == UserAccountIndex {
		return a.userAccount.privateKey
	} else {
		return a.podAccounts[index].privateKey
	}
}

// GetAddress returns the address of a given account index.
// the index -1 belongs to user root account and other indexes belong to
// the respective pods.
func (a *Account) GetAddress(index int) utils.Address {
	if index == UserAccountIndex {
		return a.userAccount.address
	} else {
		return a.podAccounts[index].address
	}
}

// GetPodAccountInfo returns the accountInfo for a given pod index.
func (a *Account) GetPodAccountInfo(index int) (*Info, error) {
	if info, found := a.podAccounts[index]; found {
		return info, nil
	}
	return nil, fmt.Errorf("invalid index : %d", index)
}

// GetUserAccountInfo returns the user info
func (a *Account) GetUserAccountInfo() *Info {
	return a.userAccount
}

// GetEmptyAccountInfo returns blank user info
func (*Account) GetEmptyAccountInfo() *Info {
	return &Info{}
}

// GetWallet returns the account.Wallet which contains the encrypted mnemonic or seed
func (a *Account) GetWallet() *Wallet {
	return a.wallet
}

// IsReadOnlyPod checks if a pod account info is read only
func (ai *Info) IsReadOnlyPod() bool {
	return ai.privateKey == nil
}

// GetAddress returns the address of the account info
func (ai *Info) GetAddress() utils.Address {
	return ai.address
}

// SetAddress sets the address of the account info
func (ai *Info) SetAddress(addr utils.Address) {
	ai.address = addr
}

// GetPrivateKey returns the private key from the accoutn info
func (ai *Info) GetPrivateKey() *ecdsa.PrivateKey {
	return ai.privateKey
}

// GetPublicKey returns the public key from the accoutn info
func (ai *Info) GetPublicKey() *ecdsa.PublicKey {
	return ai.publicKey
}

// PadSeed pads the given seed with random elements to be a chunk of chunkSize
func (*Info) PadSeed(seed []byte, passphrase string) ([]byte, error) {
	paddingLength := chunkSize - aes.BlockSize - seedSize
	randomBytes, err := utils.GetRandBytes(paddingLength)
	if err != nil {
		return nil, err
	}
	chunkData := make([]byte, 0, chunkSize)
	chunkData = append(chunkData, seed...)
	chunkData = append(chunkData, randomBytes...)
	aesKey := sha256.Sum256([]byte(passphrase))
	encryptedBytes, err := encryptBytes(aesKey[:], chunkData)
	if err != nil {
		return nil, fmt.Errorf("mnemonic padding failed: %w", err)
	}
	return encryptedBytes, nil
}

// RemovePadFromSeed removes the padding of random elements from the given data and returns the seed
func (*Info) RemovePadFromSeed(paddedSeed []byte, passphrase string) ([]byte, error) {
	aesKey := sha256.Sum256([]byte(passphrase))
	decryptedBytes, err := decryptBytes(aesKey[:], paddedSeed)
	if err != nil {
		return nil, fmt.Errorf("seed decryption failed: %w", err)
	}

	return decryptedBytes[:seedSize], nil
}

func (a *Account) encryptMnemonic(mnemonic, passPhrase string) (string, error) {
	// get the password and hash it to 256 bits
	password := passPhrase
	if password == "" {
		return "", ErrBlankPassword
	}
	aesKey := sha256.Sum256([]byte(password))

	// encrypt the mnemonic
	encryptedMessage, err := encrypt(aesKey[:], mnemonic)
	if err != nil {
		return "", fmt.Errorf("create user account: %w", err)
	}

	return encryptedMessage, nil
}
