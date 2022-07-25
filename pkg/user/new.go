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
	"regexp"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

// CreateNewUser creates a new user with the given user name and password. if a mnemonic is passed
// then it is used instead of creating a new one.
func (u *Users) CreateNewUser(userName, passPhrase, mnemonic, sessionId string) (string, string, *Info, error) {
	// username validation
	if u.IsUsernameAvailable(userName, u.dataDir) {
		return "", "", nil, ErrUserAlreadyPresent
	}

	acc := account.New(u.logger)
	accountInfo := acc.GetUserAccountInfo()
	fd := feed.New(accountInfo, u.client, u.logger)

	//create a new base user account with the mnemonic
	mnemonic, encryptedMnemonic, err := acc.CreateUserAccount(passPhrase, mnemonic)
	if err != nil {
		return "", "", nil, err
	}

	// store the encrypted mnemonic in Swarm
	err = u.uploadEncryptedMnemonic(userName, accountInfo.GetAddress(), encryptedMnemonic, fd)
	if err != nil {
		return "", "", nil, err
	}

	// store the username -> address mapping locally
	err = u.storeUserNameToAddressFileMapping(userName, u.dataDir, accountInfo.GetAddress())
	if err != nil {
		return "", "", nil, err
	}

	// Instantiate pod, dir & file objects
	file := f.NewFile(userName, u.client, fd, accountInfo.GetAddress(), u.logger)
	dir := d.NewDirectory(userName, u.client, fd, accountInfo.GetAddress(), file, u.logger)
	pod := p.NewPod(u.client, fd, acc, u.logger)
	if sessionId == "" {
		sessionId = cookie.GetUniqueSessionId()
	}

	userAddressString := accountInfo.GetAddress().Hex()
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
	err = u.addUserAndSessionToMap(ui)
	if err != nil {
		return "", "", nil, err
	}

	return userAddressString, mnemonic, ui, nil
}

// CreateNewUserV2 creates a new user with the given user name and password. if a mnemonic is passed
// then it is used instead of creating a new one.
func (u *Users) CreateNewUserV2(userName, passPhrase, mnemonic, sessionId string) (string, string, string, string, *Info, error) {
	// Check username validity
	if !isUserNameValid(userName) {
		return "", "", "", "", nil, ErrInvalidUserName
	}
	// username availability
	if u.IsUsernameAvailableV2(userName) {
		return "", "", "", "", nil, ErrUserAlreadyPresent
	}

	acc := account.New(u.logger)
	accountInfo := acc.GetUserAccountInfo()
	fd := feed.New(accountInfo, u.client, u.logger)
	//create a new base user account with the mnemonic
	mnemonic, _, err := acc.CreateUserAccount(passPhrase, mnemonic)
	if err != nil {
		return "", "", "", "", nil, err
	}

	// create ens subdomain and store mnemonic
	nameHash, err := u.createENS(userName, accountInfo)
	if err != nil {
		if err == eth.ErrInsufficientBalance {
			return accountInfo.GetAddress().Hex(), mnemonic, "", "", nil, err
		}
		return "", "", "", "", nil, err
	}
	seed, err := hdwallet.NewSeedFromMnemonic(mnemonic)
	if err != nil {
		return "", "", "", "", nil, err
	}
	key, err := accountInfo.PadSeed(seed, passPhrase)
	if err != nil {
		return "", "", "", "", nil, err
	}
	if err := u.uploadPortableAccount(accountInfo, userName, passPhrase, key, fd); err != nil {
		return "", "", "", "", nil, err
	}
	// Instantiate pod, dir & file objects
	file := f.NewFile(userName, u.client, fd, accountInfo.GetAddress(), u.logger)
	dir := d.NewDirectory(userName, u.client, fd, accountInfo.GetAddress(), file, u.logger)
	pod := p.NewPod(u.client, fd, acc, u.logger)
	if sessionId == "" {
		sessionId = cookie.GetUniqueSessionId()
	}

	userAddressString := accountInfo.GetAddress().Hex()
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
	err = u.addUserAndSessionToMap(ui)
	if err != nil {
		return "", "", "", "", nil, err
	}
	// store encrypted soc address in secondary location
	pb := crypto.FromECDSAPub(accountInfo.GetPublicKey())
	return userAddressString, mnemonic, nameHash, utils.Encode(pb), ui, nil
}

func isUserNameValid(username string) bool {
	if username == "" {
		return false
	}
	pattern := `^[a-z0-9_-]*$`
	matches, err := regexp.MatchString(pattern, username)
	if err != nil {
		return false
	}
	pattern2 := `^[A-Z]*$`
	matches2, err := regexp.MatchString(pattern2, username)
	if err != nil {
		return false
	}
	if matches && !matches2 {
		return true
	} else {
		return false
	}
}

func (u *Users) createENS(userName string, accountInfo *account.Info) (string, error) {
	err := u.ens.RegisterSubdomain(userName, common.HexToAddress(accountInfo.GetAddress().Hex()), accountInfo.GetPrivateKey())
	if err != nil {
		return "", err
	}

	nameHash, err := u.ens.SetResolver(userName, common.HexToAddress(accountInfo.GetAddress().Hex()), accountInfo.GetPrivateKey())
	if err != nil {
		return "", err
	}

	err = u.ens.SetAll(userName, common.HexToAddress(accountInfo.GetAddress().Hex()), accountInfo.GetPrivateKey())
	if err != nil {
		return "", err
	}
	return nameHash, nil
}
