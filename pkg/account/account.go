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
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec"
	gethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/fairdatasociety/fairOS-dfs-utils/crypto"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

const (
	UserAccountIndex = -1

	addressLength = 64
	paddingMin    = 300
	paddingMax    = 500
)

type Account struct {
	wallet      *Wallet
	userAccount *Info
	podAccounts map[int]*Info
	logger      logging.Logger
}

type Info struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	address    utils.Address
}

// New create a account object through which the entire account management is done.
// it uses a 12 word BIP-0039 wordlist to create a 12 word mnemonic for every user
// and spawns key pais whenever necessary.
func New(logger logging.Logger) *Account {
	wal := NewWallet("")
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
	wal := NewWallet("")
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
		fmt.Print("Enter password to unlock user account: ")
		password = a.getPassword()
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

// LoadUserAccountV2 loads the user account given the encrypted mnemonic and
// password.
func (a *Account) LoadUserAccountV2(privateKey []byte) error {
	p, err := gethCrypto.ToECDSA(privateKey)
	if err != nil {
		return fmt.Errorf("invalid private key")
	}
	a.userAccount.privateKey = p
	publicKey := p.Public()

	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}
	a.userAccount.publicKey = publicKeyECDSA
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
		fmt.Print("Enter user password to delete a pod: ")
		password = a.getPassword()
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

	password := passPhrase
	if password == "" {
		if createPod {
			fmt.Print("Enter user password to create a pod: ")
		} else {
			fmt.Print("Enter user password to open a pod: ")
		}
		password = a.getPassword()
	}

	plainMnemonic, err := a.wallet.decryptMnemonic(password)
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	path := genericPath + strconv.Itoa(accountId)
	acc, err := a.wallet.CreateAccount(path, plainMnemonic)
	if err != nil {
		return nil, err
	}
	hdw, err := hdwallet.NewFromMnemonic(plainMnemonic)
	if err != nil {
		return nil, err
	}
	accountInfo := &Info{}

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

// CreatePodAccountV2 is used to create a new key pair from the master mnemonic. this key pair is
// used as the base key pair for a newly created pod.
func (a *Account) CreatePodAccountV2(accountId int, privateKey []byte) (*Info, error) {
	if acc, ok := a.podAccounts[accountId]; ok {
		return acc, nil
	}
	hdw, err := hdwallet.NewFromSeed(privateKey)
	if err != nil {
		return nil, err
	}
	path := genericPath + strconv.Itoa(accountId)
	acc, err := a.wallet.CreateAccountFromSeed(path, privateKey)
	if err != nil {
		return nil, err
	}
	accountInfo := &Info{}

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

	password := passPhrase
	if password == "" {
		if createCollection {
			fmt.Print("Enter user password to create a collection: ")
		} else {
			fmt.Print("Enter user password to open a collection: ")
		}
		password = a.getPassword()
	}

	plainMnemonic, err := a.wallet.decryptMnemonic(password)
	if err != nil {
		return fmt.Errorf("invalid password")
	}

	path := genericPath + strconv.Itoa(accountId)
	acc, err := a.wallet.CreateAccount(path, plainMnemonic)
	if err != nil {
		return err
	}
	hdw, err := hdwallet.NewFromMnemonic(plainMnemonic)
	if err != nil {
		return err
	}

	accountInfo := &Info{}

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

func (a *Account) GetUserAccountInfo() *Info {
	return a.userAccount
}

func (*Account) GetEmptyAccountInfo() *Info {
	return &Info{}
}

func (a *Account) GetWallet() *Wallet {
	return a.wallet
}

func (a *Info) IsReadOnlyPod() bool {
	return a.privateKey == nil
}

func (ai *Info) GetAddress() utils.Address {
	return ai.address
}

func (ai *Info) SetAddress(addr utils.Address) {
	ai.address = addr
}

func (ai *Info) GetPrivateKey() *ecdsa.PrivateKey {
	return ai.privateKey
}

func (ai *Info) GetPublicKey() *ecdsa.PublicKey {
	return ai.publicKey
}

func (ai *Info) GetEncryptedPublicKey(passPhrase string) (string, error) {
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(passPhrase), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	publicKeyBytes := gethCrypto.FromECDSAPub(ai.GetPublicKey())
	hashedPublicKeyBytes, err := bcrypt.GenerateFromPassword(publicKeyBytes, bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hashedPasswordBytes) + string(hashedPublicKeyBytes), nil
}

func (ai *Info) EncryptPublicKey(passPhrase string, pubKey *ecdsa.PublicKey) (string, error) {
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(passPhrase), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	publicKeyBytes := gethCrypto.FromECDSAPub(pubKey)
	hashedPublicKeyBytes, err := bcrypt.GenerateFromPassword(publicKeyBytes, bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hashedPasswordBytes) + string(hashedPublicKeyBytes), nil
}

const (
	ChunkSize = 4096
)

func (ai *Info) EncryptPrivateKey(passphrase string) ([]byte, error) {
	pvtK := gethCrypto.FromECDSA(ai.GetPrivateKey())
	rand.Seed(time.Now().UnixNano())
	paddingLength := ChunkSize - aes.BlockSize - len(pvtK)
	randomBytes, err := utils.GetRandBytes(paddingLength)
	if err != nil {
		return nil, err
	}
	chunkData := append(pvtK, randomBytes...)
	aesKey := sha256.Sum256([]byte(passphrase))
	encryptedBytes, err := encryptBytes(aesKey[:], chunkData)
	if err != nil {
		return nil, fmt.Errorf("private key encryption failed: %w", err)
	}
	return encryptedBytes, nil
}

func (ai *Info) DecryptPrivateKey(passphrase string, encryptedPrivateKey []byte) ([]byte, error) {
	aesKey := sha256.Sum256([]byte(passphrase))
	decryptedBytes, err := decryptBytes(aesKey[:], encryptedPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("private key decryption failed: %w", err)
	}
	return decryptedBytes[:32], nil
}

func (*Info) EncryptContent(passphrase, data string) (string, error) {
	password := passphrase
	if password == "" {
		return "", fmt.Errorf("passphrase cannot be blank")
	}
	aesKey := sha256.Sum256([]byte(password))
	rand.Seed(time.Now().UnixNano())
	paddingLength := rand.Intn(paddingMax-paddingMin) + paddingMin
	randomStr, err := utils.GetRandString(paddingLength)
	if err != nil {
		return "", err
	}
	encryptedMessage, err := encrypt(aesKey[:], data+randomStr)
	if err != nil {
		return "", fmt.Errorf("create user account: %w", err)
	}
	return encryptedMessage, nil
}

func (ai *Info) DecryptContent(passphrase, encryptedContent string) (string, error) {
	password := passphrase
	if password == "" {
		return "", fmt.Errorf("passphrase cannot be blank")
	}

	if encryptedContent == "" {
		return "", fmt.Errorf("invalid encrypted content")
	}
	aesKey := sha256.Sum256([]byte(password))

	//decrypt the message
	data, err := decrypt(aesKey[:], encryptedContent)
	if err != nil {
		return "", err
	}
	return data[:addressLength], nil
}

func (a *Account) encryptMnemonic(mnemonic, passPhrase string) (string, error) {
	// get the password and hash it to 256 bits
	password := passPhrase
	if password == "" {
		fmt.Print("Enter password to unlock user account: ")
		password = a.getPassword()
		password = strings.Trim(password, "\n")
	}
	aesKey := sha256.Sum256([]byte(password))

	// encrypt the mnemonic
	encryptedMessage, err := encrypt(aesKey[:], mnemonic)
	if err != nil {
		return "", fmt.Errorf("create user account: %w", err)
	}

	return encryptedMessage, nil
}

func (*Account) getPassword() (password string) {
	// read the pass phrase
	bytePassword, err := term.ReadPassword(0)
	if err != nil {
		log.Fatalf("error reading password")
		return
	}
	fmt.Println("")
	passwd := string(bytePassword)
	password = strings.TrimSpace(passwd)
	return password
}
