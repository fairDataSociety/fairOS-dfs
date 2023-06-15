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
	"github.com/fairdatasociety/fairOS-dfs/pkg/auth"
	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager"
	"github.com/fairdatasociety/fairOS-dfs/pkg/taskmanager"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	minPasswordLength = 12
	zeroAddressHex    = "0x0000000000000000000000000000000000000000"
)

// SignupResponse is the response of a successful signup
type SignupResponse struct {
	Address   string `json:"address"`
	Mnemonic  string `json:"mnemonic"`
	NameHash  string `json:"nameHash"`
	PublicKey string `json:"publicKey"`
	UserInfo  *Info  `json:"userInfo"`
}

// CreateNewUserV2 creates a new user with the given username and password. if a mnemonic is passed
// then it is used instead of creating a new one.
func (u *Users) CreateNewUserV2(userName, passPhrase, mnemonic, sessionId string, tm taskmanager.TaskManagerGO, sm subscriptionManager.SubscriptionManager) (*SignupResponse, error) {
	if userName == "" {
		return nil, ErrBlankUsername
	}
	if passPhrase == "" {
		return nil, ErrBlankPassword
	}

	// check password length
	if len(passPhrase) < minPasswordLength {
		return nil, ErrPasswordTooSmall
	}

	// Check username validity
	if !isUserNameValid(userName) {
		u.logger.Errorf("user: create: user name is not valid")
		return nil, ErrInvalidUserName
	}

	// get the owner of the username
	ownerAddress, err := u.ens.GetOwner(userName)
	if err != nil {
		u.logger.Errorf("user: create: get owner failed for user %s: %v", userName, err)
		return nil, err
	}

	acc := account.New(u.logger)
	accountInfo := acc.GetUserAccountInfo()
	fd := feed.New(accountInfo, u.client, u.logger)

	// create a new base user account with the mnemonic
	mnemonic, seed, err := acc.CreateUserAccount(mnemonic)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	address := accountInfo.GetAddress()
	// username availability
	if ownerAddress.Hex() != zeroAddressHex &&
		ownerAddress.Hex() != address.Hex() {
		return nil, ErrUserAlreadyPresent
	}
	signUp := &SignupResponse{
		Address:  address.Hex(),
		Mnemonic: mnemonic,
	}
	nameHash := ""
	if ownerAddress.Hex() == zeroAddressHex {
		// create ens subdomain and store mnemonic
		nameHash, err = u.createENS(userName, accountInfo)
		if err != nil { // skipcq: TCV-001
			u.logger.Errorf("user: create: create ens failed for user %s: %v", userName, err)
			if err == eth.ErrInsufficientBalance { // skipcq: TCV-001
				return signUp, err
			}
			return nil, err // skipcq: TCV-001
		}
	} else {
		_, nameHash, err = u.ens.GetInfo(userName)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
	}
	signUp.NameHash = nameHash
	key, err := accountInfo.PadSeed(seed, passPhrase)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	if err = u.uploadPortableAccount(accountInfo, userName, passPhrase, key, fd); err != nil { // skipcq: TCV-001
		return nil, err
	}
	// Instantiate pod, dir & file objects
	file := f.NewFile(userName, u.client, fd, accountInfo.GetAddress(), tm, u.logger)
	dir := d.NewDirectory(userName, u.client, fd, accountInfo.GetAddress(), file, tm, u.logger)
	pod := p.NewPod(u.client, fd, acc, tm, sm, u.logger)
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
	signUp.UserInfo = ui
	signUp.PublicKey = utils.Encode(crypto.FromECDSAPub(accountInfo.GetPublicKey()))
	return signUp, nil
}

func isUserNameValid(username string) bool {
	if username == "" {
		return false
	}
	pattern := `^[a-z0-9_-]*$`
	matches, err := regexp.MatchString(pattern, username)
	if err != nil { // skipcq: TCV-001
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
	address := accountInfo.GetAddress()
	err := u.ens.RegisterSubdomain(userName, common.HexToAddress(address.Hex()), accountInfo.GetPrivateKey())
	if err != nil { // skipcq: TCV-001
		return "", err
	}

	nameHash, err := u.ens.SetResolver(userName, common.HexToAddress(address.Hex()), accountInfo.GetPrivateKey())
	if err != nil { // skipcq: TCV-001
		return "", err
	}

	err = u.ens.SetAll(userName, common.HexToAddress(address.Hex()), accountInfo.GetPrivateKey())
	if err != nil { // skipcq: TCV-001
		return "", err
	}
	return nameHash, nil
}
