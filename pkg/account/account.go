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
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethersphere/bee/v2/pkg/crypto"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

const (
	// UserAccountIndex is user root account
	UserAccountIndex = -1

	// seedSize is used to determine how much padding we need for portable account SOC
	seedSize           = 64
	usernameLengthSize = 1
)

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

// New create an account object through which the entire account management is done.
// it uses a 12 word BIP-0039 wordlist to create a 12 word mnemonic for every user
// and spawns key pais whenever necessary.
func New(logger logging.Logger) *Account {
	wal := newWallet(nil)
	return &Account{
		wallet:      wal,
		userAccount: &Info{},
		podAccounts: make(map[int]*Info),
		logger:      logger,
	}
}

// CreateRandomKeyPair creates an ecdsa key pair by using the given int64 number
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
func (a *Account) CreateUserAccount(mnemonic string) (string, []byte, error) {
	wal := newWallet(nil)
	a.wallet = wal
	acc, mnemonic, err := wal.LoadMnemonicAndCreateRootAccount(mnemonic)
	if err != nil {
		return "", nil, err
	}
	hdw, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil { // skipcq: TCV-001
		return "", nil, err
	}

	// store publicKey, private key and user
	a.userAccount.privateKey, err = hdw.PrivateKey(acc)
	if err != nil { // skipcq: TCV-001
		return "", nil, err
	}
	a.userAccount.publicKey, err = hdw.PublicKey(acc)
	if err != nil { // skipcq: TCV-001
		return "", nil, err
	}
	addrBytes, err := crypto.NewEthereumAddress(a.userAccount.privateKey.PublicKey)
	if err != nil { // skipcq: TCV-001
		return "", nil, err
	}
	a.userAccount.address.SetBytes(addrBytes)

	seed, err := hdwallet.NewSeedFromMnemonic(mnemonic)
	if err != nil { // skipcq: TCV-001
		return "", nil, err
	}

	return mnemonic, seed, nil
}

// GenerateUserAccountFromSignature create a new master account for a user.
func (a *Account) GenerateUserAccountFromSignature(signature, password string) (string, []byte, error) {
	wal := newWallet(nil)
	a.wallet = wal
	acc, mnemonic, err := wal.GenerateWalletFromSignature(signature, password)
	if err != nil {
		return "", nil, err
	}

	hdw, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil { // skipcq: TCV-001
		return "", nil, err
	}

	// store publicKey, private key and user
	a.userAccount.privateKey, err = hdw.PrivateKey(acc)
	if err != nil { // skipcq: TCV-001
		return "", nil, err
	}
	a.userAccount.publicKey, err = hdw.PublicKey(acc)
	if err != nil { // skipcq: TCV-001
		return "", nil, err
	}
	addrBytes, err := crypto.NewEthereumAddress(a.userAccount.privateKey.PublicKey)
	if err != nil { // skipcq: TCV-001
		return "", nil, err
	}
	a.userAccount.address.SetBytes(addrBytes)

	seed, err := hdwallet.NewSeedFromMnemonic(mnemonic)
	if err != nil { // skipcq: TCV-001
		return "", nil, err
	}

	return mnemonic, seed, nil
}

// LoadUserAccountFromSeed loads the user account given the bip39 seed
func (a *Account) LoadUserAccountFromSeed(seed []byte) error {
	acc, err := a.wallet.CreateAccountFromSeed(rootPath, seed)
	if err != nil {
		return err
	}
	hdw, err := hdwallet.NewFromSeed(seed)
	if err != nil { // skipcq: TCV-001
		return err
	}
	a.userAccount.privateKey, err = hdw.PrivateKey(acc)
	if err != nil { // skipcq: TCV-001
		return err
	}
	a.userAccount.publicKey, err = hdw.PublicKey(acc)
	if err != nil { // skipcq: TCV-001
		return err
	}
	addrBytes, err := crypto.NewEthereumAddress(a.userAccount.privateKey.PublicKey)
	if err != nil { // skipcq: TCV-001
		return err
	}
	a.userAccount.address.SetBytes(addrBytes)
	return nil
}

// CreatePodAccount is used to create a new key pair from the master mnemonic. this key pair is
// used as the base key pair for a newly created pod.
func (a *Account) CreatePodAccount(accountId int, createPod bool) (*Info, error) {
	if acc, ok := a.podAccounts[accountId]; ok { // skipcq: TCV-001
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
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		hdw, err = hdwallet.NewFromSeed(a.wallet.seed)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
	}

	accountInfo.privateKey, err = hdw.PrivateKey(acc)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	accountInfo.publicKey, err = hdw.PublicKey(acc)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	addrBytes, err := crypto.NewEthereumAddress(accountInfo.privateKey.PublicKey)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	accountInfo.address.SetBytes(addrBytes)
	a.podAccounts[accountId] = accountInfo
	return accountInfo, nil
}

// DeletePodAccount unloads/forgets a particular pods key value pair from the memory.
// skipcq: TCV-001
func (a *Account) DeletePodAccount(accountId int) {
	delete(a.podAccounts, accountId)
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

// GetUserAccountInfo returns the user info
// skipcq: TCV-001
func (a *Account) GetUserAccountInfo() *Info {
	return a.userAccount
}

// GetEmptyAccountInfo returns blank user info
// skipcq: TCV-001
func (*Account) GetEmptyAccountInfo() *Info {
	return &Info{}
}

// GetWallet returns the account.Wallet which contains the encrypted mnemonic or seed
// skipcq: TCV-001
func (a *Account) GetWallet() *Wallet {
	return a.wallet
}

// IsReadOnlyPod checks if a pod account info is read only
// skipcq: TCV-001
func (ai *Info) IsReadOnlyPod() bool {
	return ai.privateKey == nil
}

// GetAddress returns the address of the account info
// skipcq: TCV-001
func (ai *Info) GetAddress() utils.Address {
	return ai.address
}

// SetAddress sets the address of the account info
// skipcq: TCV-001
func (ai *Info) SetAddress(addr utils.Address) {
	ai.address = addr
}

// GetPrivateKey returns the private key from the account info
func (ai *Info) GetPrivateKey() *ecdsa.PrivateKey {
	return ai.privateKey
}

// GetPublicKey returns the public key from the accoutn info
func (ai *Info) GetPublicKey() *ecdsa.PublicKey {
	return ai.publicKey
}

// PadSeed pads the given seed with random elements to be a chunk of chunkSize
func (*Info) PadSeed(seed []byte, passphrase string) ([]byte, error) {
	paddingLength := utils.MaxChunkLength - aes.BlockSize - seedSize
	randomBytes, err := utils.GetRandBytes(paddingLength)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	chunkData := make([]byte, 0, utils.MaxChunkLength)
	chunkData = append(chunkData, seed...)
	chunkData = append(chunkData, randomBytes...)
	encryptedBytes, err := utils.EncryptBytes([]byte(passphrase), chunkData)
	if err != nil { // skipcq: TCV-001
		return nil, fmt.Errorf("mnemonic padding failed: %w", err)
	}
	return encryptedBytes, nil
}

// RemovePadFromSeed removes the padding of random elements from the given data and returns the seed
func (*Info) RemovePadFromSeed(paddedSeed []byte, passphrase string) ([]byte, error) {
	decryptedBytes, err := utils.DecryptBytes([]byte(passphrase), paddedSeed)
	if err != nil { // skipcq: TCV-001
		return nil, fmt.Errorf("seed decryption failed: %w", err)
	}

	return decryptedBytes[:seedSize], nil
}

// PadSeedName pads the given seed and name with random elements to be a chunk of chunkSize
func (*Info) PadSeedName(seed []byte, username, passphrase string) ([]byte, error) {
	usernameLength := len(username)
	if usernameLength > 255 {
		return nil, fmt.Errorf("username length is too long")
	}
	usernameLengthBytes := make([]byte, usernameLengthSize)
	usernameLengthBytes[0] = byte(usernameLength)
	paddingLength := utils.MaxChunkLength - aes.BlockSize - seedSize - usernameLengthSize - usernameLength
	randomBytes, err := utils.GetRandBytes(paddingLength)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	chunkData := make([]byte, 0, utils.MaxChunkLength)
	chunkData = append(chunkData, seed...)
	chunkData = append(chunkData, usernameLengthBytes...)
	chunkData = append(chunkData, []byte(username)...)
	chunkData = append(chunkData, randomBytes...)
	encryptedBytes, err := utils.EncryptBytes([]byte(passphrase), chunkData)
	if err != nil { // skipcq: TCV-001
		return nil, fmt.Errorf("mnemonic padding failed: %w", err)
	}
	return encryptedBytes, nil
}

// RemovePadFromSeedName removes the padding of random elements from the given data and returns the seed and name
func (*Info) RemovePadFromSeedName(paddedSeed []byte, passphrase string) ([]byte, string, error) {
	decryptedBytes, err := utils.DecryptBytes([]byte(passphrase), paddedSeed)
	if err != nil { // skipcq: TCV-001
		return nil, "", fmt.Errorf("seed decryption failed: %w", err)
	}
	usernameLength := int(decryptedBytes[seedSize])
	return decryptedBytes[:seedSize], string(decryptedBytes[seedSize+usernameLengthSize : seedSize+usernameLengthSize+usernameLength]), nil
}
