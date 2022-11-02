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
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/taskmanager"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// LoginUserV2 checks if the user is present and logs in the user. It also creates the required information
// to execute user function and stores it in memory.
func (u *Users) LoginUserV2(userName, passPhrase string, client blockstore.Client, tm taskmanager.TaskManagerGO, sessionId string) (*Info, string, string, error) {
	// check if username is available (user created)
	if !u.IsUsernameAvailableV2(userName) {
		return nil, "", "", ErrInvalidUserName
	}

	// get owner address from Subdomain registrar
	address, err := u.ens.GetOwner(userName)
	if err != nil { // skipcq: TCV-001
		return nil, "", "", err
	}
	// create account
	acc := account.New(u.logger)
	accountInfo := acc.GetUserAccountInfo()
	// load public key from public resolver
	publicKey, nameHash, err := u.ens.GetInfo(userName)
	if err != nil { // skipcq: TCV-001
		return nil, "", "", err
	}
	pb := crypto.FromECDSAPub(publicKey)

	// load encrypted private key
	fd := feed.New(accountInfo, client, u.logger)
	key, err := u.downloadPortableAccount(utils.Address(address), userName, passPhrase, fd)
	if err != nil {
		u.logger.Errorf(err.Error())
		return nil, "", "", ErrInvalidPassword
	}

	// decrypt and remove pad from private ley
	seed, err := accountInfo.RemovePadFromSeed(key, passPhrase)
	if err != nil { // skipcq: TCV-001
		return nil, "", "", err
	}
	// load user account
	err = acc.LoadUserAccountFromSeed(seed)
	if err != nil { // skipcq: TCV-001
		return nil, "", "", err
	}

	if u.IsUserLoggedIn(sessionId) { // skipcq: TCV-001
		return nil, "", "", ErrUserAlreadyLoggedIn
	}

	// Instantiate pod, dir & file objects
	file := f.NewFile(userName, client, fd, accountInfo.GetAddress(), tm, u.logger)
	pod := p.NewPod(u.client, fd, acc, tm, u.logger)
	dir := d.NewDirectory(userName, client, fd, accountInfo.GetAddress(), file, tm, u.logger)
	if sessionId == "" {
		sessionId = cookie.GetUniqueSessionId()
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
	return ui, nameHash, utils.Encode(pb), u.addUserAndSessionToMap(ui)
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

	// remove from the user map
	u.removeUserFromMap(sessionId)

	return nil
}

// IsUserLoggedIn checks if the user is logged in from sessionID
func (u *Users) IsUserLoggedIn(sessionId string) bool {
	return u.isUserPresentInMap(sessionId)
}

// GetLoggedInUserInfo returns the user info of the user
func (u *Users) GetLoggedInUserInfo(sessionId string) *Info {
	return u.getUserFromMap(sessionId)
}

// IsUserNameLoggedIn checks if the user is logged in from username
func (u *Users) IsUserNameLoggedIn(userName string) bool {
	return u.isUserNameInMap(userName)
}
