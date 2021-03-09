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
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethersphere/bee/pkg/crypto"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/term"
)

const (
	UserAccountIndex = -1
)

type Account struct {
	wallet      *Wallet
	userAcount  *Info
	podAccounts map[int]*Info
	logger      logging.Logger
}

type Info struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	address    utils.Address
}

func New(logger logging.Logger) *Account {
	wallet := NewWallet("")
	return &Account{
		wallet:      wallet,
		userAcount:  &Info{},
		podAccounts: make(map[int]*Info),
		logger:      logger,
	}
}

func CreateRandomKeyPair(now int64) (*ecdsa.PrivateKey, error) {
	randBytes := make([]byte, 40)
	binary.LittleEndian.PutUint64(randBytes, uint64(now))
	randReader := bytes.NewReader(randBytes)
	return ecdsa.GenerateKey(btcec.S256(), randReader)
}

func (a *Account) CreateUserAccount(passPhrase, mnemonic string) (string, string, error) {
	wallet := NewWallet("")
	a.wallet = wallet
	acc, mnemonic, err := wallet.LoadMnemonicAndCreateRootAccount(mnemonic)
	if err != nil {
		return "", "", err
	}

	hdw, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return "", "", err
	}

	// store publicKey, private key and user
	a.userAcount.privateKey, err = hdw.PrivateKey(acc)
	if err != nil {
		return "", "", err
	}
	a.userAcount.publicKey, err = hdw.PublicKey(acc)
	if err != nil {
		return "", "", err
	}
	addrBytes, err := crypto.NewEthereumAddress(a.userAcount.privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	a.userAcount.address.SetBytes(addrBytes)

	// store the mnemonic
	encryptedMnemonic, err := a.encryptMnemonic(mnemonic, passPhrase)
	if err != nil {
		return "", "", err
	}
	a.wallet.encryptedmnemonic = encryptedMnemonic

	return mnemonic, encryptedMnemonic, nil
}

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
	a.userAcount.privateKey, err = hdw.PrivateKey(acc)
	if err != nil {
		return err
	}
	a.userAcount.publicKey, err = hdw.PublicKey(acc)
	if err != nil {
		return err
	}
	addrBytes, err := crypto.NewEthereumAddress(a.userAcount.privateKey.PublicKey)
	if err != nil {
		return err
	}
	a.userAcount.address.SetBytes(addrBytes)
	return nil
}

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

func (a *Account) DeletePodAccount(accountId int) {
	delete(a.podAccounts, accountId)
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

func (a *Account) GetUserPrivateKey(index int) *ecdsa.PrivateKey {
	if index == UserAccountIndex {
		return a.userAcount.privateKey
	} else {
		return a.podAccounts[index].privateKey
	}
}

func (a *Account) GetAddress(index int) utils.Address {
	if index == UserAccountIndex {
		return a.userAcount.address
	} else {
		return a.podAccounts[index].address
	}
}

func (a *Account) GetUserAccountInfo() *Info {
	return a.userAcount
}

func (a *Account) GetEmptyAccountInfo() *Info {
	return &Info{}
}

func (a *Account) GetPodAccountInfo(index int) (*Info, error) {
	if info, found := a.podAccounts[index]; found {
		return info, nil
	}
	return nil, fmt.Errorf("invalid index : %d", index)
}

func (a *Account) getPassword() (password string) {
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
