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
	"net/http"
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

// CreateNewUser creates a new user with the given user name and password. if a mnemonic is passed
// then it is used instead of creating a new one.
func (u *Users) CreateNewUser(userName, passPhrase, mnemonic string, response http.ResponseWriter, sessionId string) (string, string, *Info, error) {
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

	// store the ecnrypted mnemonic in Swarm
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
		openPods:   make(map[string]string),
		openPodsMu: &sync.RWMutex{},
	}

	// set cookie and add user to map
	err = u.addUserAndSessionToMap(ui, response)
	if err != nil {
		return "", "", nil, err
	}

	return userAddressString, mnemonic, ui, nil
}
