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

package user

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"
	"github.com/fairdatasociety/fairOS-dfs/pkg/auth/jwt"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager"
	"github.com/fairdatasociety/fairOS-dfs/pkg/taskmanager"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// LoginResponse is the response of a successful login
type LoginResponse struct {
	Address     string `json:"address"`
	NameHash    string `json:"nameHash"`
	PublicKey   string `json:"publicKey"`
	UserInfo    *Info  `json:"userInfo"`
	AccessToken string `json:"accessToken"`
}

// LoginUserV2 checks if the user is present and logs in the user. It also creates the required information
// to execute user function and stores it in memory.
func (u *Users) LoginUserV2(userName, passPhrase string, client blockstore.Client, tm taskmanager.TaskManagerGO, sm subscriptionManager.SubscriptionManager, sessionId string) (*LoginResponse, error) {
	// check if sessionId is still active
	if u.IsUserLoggedIn(sessionId) { // skipcq: TCV-001
		return nil, ErrUserAlreadyLoggedIn
	}

	// check if username is available (user created)
	if !u.IsUsernameAvailableV2(userName) {
		return nil, ErrUserNameNotFound
	}

	// get owner address from Subdomain registrar
	address, err := u.ens.GetOwner(userName)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	// create account
	acc := account.New(u.logger)
	accountInfo := acc.GetUserAccountInfo()
	// load encrypted private key
	fd := feed.New(accountInfo, client, u.logger)
	key, err := u.downloadPortableAccount(utils.Address(address), userName, passPhrase, fd)
	if err != nil {
		return nil, ErrInvalidPassword
	}

	// load public key from public resolver
	publicKey, nameHash, err := u.ens.GetInfo(userName)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	if publicKey == nil {
		return nil, fmt.Errorf("public key not found")
	}

	// decrypt and remove pad from private key
	seed, err := accountInfo.RemovePadFromSeed(key, passPhrase)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	// load user account
	if err = acc.LoadUserAccountFromSeed(seed); err != nil { // skipcq: TCV-001
		return nil, err
	}

	// Instantiate pod, dir & file objects
	file := f.NewFile(userName, client, fd, accountInfo.GetAddress(), tm, u.logger)
	pod := p.NewPod(u.client, fd, acc, tm, sm, u.logger)
	dir := d.NewDirectory(userName, client, fd, accountInfo.GetAddress(), file, tm, u.logger)
	if sessionId == "" {
		sessionId = auth.GetUniqueSessionId()
	}

	ui := &Info{
		name:       userName,
		sessionId:  sessionId,
		feedApi:    fd,
		account:    acc,
		file:       file,
		dir:        dir,
		pod:        pod,
		openPods:   make(map[string]*p.Info),
		openPodsMu: &sync.RWMutex{},
	}
	// set cookie and add user to map
	if err = u.addUserAndSessionToMap(ui); err != nil { // skipcq: TCV-001
		return nil, err
	}

	token, err := jwt.GenerateToken(sessionId)
	if err != nil {
		u.logger.Errorf("error generating token: %v\n", err)
	}

	return &LoginResponse{
		Address:     address.Hex(),
		NameHash:    nameHash,
		PublicKey:   utils.Encode(crypto.FromECDSAPub(publicKey)),
		UserInfo:    ui,
		AccessToken: token,
	}, nil
}

func (u *Users) addUserAndSessionToMap(ui *Info) error {
	u.addUserToMap(ui)
	return nil
}

// Logout removes the user information from all the data structures and clears the cookie.
func (u *Users) Logout(sessionId string) error {
	// check if session or user present in map
	if !u.isUserPresentInMap(sessionId) { // skipcq: TCV-001
		return ErrUserNotLoggedIn
	}

	ui := u.getUserFromMap(sessionId)
	u.removeUserFromMap(sessionId)

	return ui.feedApi.Close()
}

// IsUserLoggedIn checks if the user is logged-in from sessionID
func (u *Users) IsUserLoggedIn(sessionId string) bool {
	return u.isUserPresentInMap(sessionId)
}

// GetLoggedInUserInfo returns the user info of the user
func (u *Users) GetLoggedInUserInfo(sessionId string) *Info {
	return u.getUserFromMap(sessionId)
}

// IsUserNameLoggedIn checks if the user is logged-in from username
func (u *Users) IsUserNameLoggedIn(userName string) bool {
	return u.isUserNameInMap(userName)
}

// ConnectWallet connects user with wallet.
func (u *Users) ConnectWallet(userName, passPhrase, walletAddressHex, signature string, client blockstore.Client) error {
	// check if username is available (user created)
	if !u.IsUsernameAvailableV2(userName) {
		return ErrUserNameNotFound
	}

	// get owner address from Subdomain registrar
	address, err := u.ens.GetOwner(userName)
	if err != nil { // skipcq: TCV-001
		return err
	}

	// check if address matches with wallet address
	if address.Hex() != walletAddressHex {
		return fmt.Errorf("wallet doesnot match portable account address")
	}
	// create account
	acc := account.New(u.logger)
	accountInfo := acc.GetUserAccountInfo()
	// load encrypted private key
	fd := feed.New(accountInfo, client, u.logger)
	key, err := u.downloadPortableAccount(utils.Address(address), userName, passPhrase, fd)
	if err != nil {
		u.logger.Errorf(err.Error())
		return err
	}

	// decrypt and remove pad from private ley
	seed, err := accountInfo.RemovePadFromSeed(key, passPhrase)
	if err != nil { // skipcq: TCV-001
		return err
	}

	if err = acc.LoadUserAccountFromSeed(seed); err != nil {
		return err
	}

	key, err = accountInfo.PadSeedName(seed, userName, signature)
	if err != nil { // skipcq: TCV-001
		return err
	}
	return u.uploadPortableAccount(accountInfo, walletAddressHex, signature, key, fd)
}

// LoginWithWallet logs user in with wallet and signature
func (u *Users) LoginWithWallet(addressHex, signature string, client blockstore.Client, tm taskmanager.TaskManagerGO, sm subscriptionManager.SubscriptionManager, sessionId string) (*Info, string, error) {

	address := common.HexToAddress(addressHex)
	// create account

	acc := account.New(u.logger)
	accountInfo := acc.GetUserAccountInfo()
	// load encrypted private key
	fd := feed.New(accountInfo, client, u.logger)
	key, err := u.downloadPortableAccount(utils.Address(address), addressHex, signature, fd)
	if err != nil {
		u.logger.Errorf(err.Error())
		return nil, "", ErrInvalidPassword
	}

	// decrypt and remove pad from private ley
	seed, username, err := accountInfo.RemovePadFromSeedName(key, signature)
	if err != nil { // skipcq: TCV-001
		return nil, "", err
	}

	nameHash, err := u.GetNameHash(username)
	if err != nil { // skipcq: TCV-001
		return nil, "", err
	}
	// load user account
	err = acc.LoadUserAccountFromSeed(seed)
	if err != nil { // skipcq: TCV-001
		return nil, "", err
	}

	if u.IsUserLoggedIn(sessionId) { // skipcq: TCV-001
		return nil, "", ErrUserAlreadyLoggedIn
	}

	// Instantiate pod, dir & file objects
	file := f.NewFile(addressHex, client, fd, accountInfo.GetAddress(), tm, u.logger)
	pod := p.NewPod(u.client, fd, acc, tm, sm, u.logger)
	dir := d.NewDirectory(addressHex, client, fd, accountInfo.GetAddress(), file, tm, u.logger)
	if sessionId == "" {
		sessionId = auth.GetUniqueSessionId()
	}
	ui := &Info{
		name:       username,
		sessionId:  sessionId,
		feedApi:    fd,
		account:    acc,
		file:       file,
		dir:        dir,
		pod:        pod,
		openPods:   make(map[string]*p.Info),
		openPodsMu: &sync.RWMutex{},
	}
	// set cookie and add user to map
	return ui, utils.Encode(nameHash[:]), u.addUserAndSessionToMap(ui)
}
