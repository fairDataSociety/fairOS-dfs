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

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

func (u *Users) LoginUser(userName, passPhrase, dataDir string, client blockstore.Client, response http.ResponseWriter, sessionId string) error {
	// basic validations
	if u.IsUserLoggedIn(sessionId) {
		return ErrUserAlreadyLoggedIn
	}
	if !u.IsUsernameAvailable(userName, dataDir) {
		return ErrInvalidUserName
	}

	acc := account.New(u.logger)
	accountInfo := acc.GetUserAccountInfo()
	fd := feed.New(accountInfo, client, u.logger)

	// load address from userName
	address, err := u.getAddressFromUserName(userName, dataDir)
	if err != nil {
		return err
	}

	// load encrypted mnemonic from Swarm
	encryptedMnemonic, err := u.getEncryptedMnemonic(userName, address, fd)
	if err != nil {
		return err
	}

	err = acc.LoadUserAccount(passPhrase, encryptedMnemonic)
	if err != nil {
		if err.Error() == "mnemonic is invalid" {
			return ErrInvalidPassword
		}
		return err
	}

	// Instantiate pod, dir & file objects
	file := f.NewFile(userName, client, fd, accountInfo.GetAddress(), u.logger)
	dir := d.NewDirectory(userName, client, fd, accountInfo.GetAddress(), file, u.logger)
	pod := p.NewPod(u.client, fd, acc, u.logger)
	if sessionId == "" {
		sessionId = cookie.GetUniqueSessionId()
	}

	ui := &Info{
		name:      userName,
		sessionId: sessionId,
		feedApi:   fd,
		account:   acc,
		file:      file,
		dir:       dir,
		pod:       pod,
	}

	// set cookie and add user to map
	return u.Login(ui, response)
}

func (u *Users) Login(ui *Info, response http.ResponseWriter) error {
	if response != nil {
		err := cookie.SetSession(ui.GetSessionId(), response, u.cookieDomain)
		if err != nil {
			return err
		}
	}
	u.addUserToMap(ui)
	return nil
}

func (u *Users) Logout(sessionId string, response http.ResponseWriter) error {
	yes := u.isUserPresentInMap(sessionId)
	if !yes {
		return ErrUserNotLoggedIn
	}

	// remove from the user map
	u.removeUserFromMap(sessionId)
	if response != nil {
		cookie.ClearSession(response)
	}
	return nil
}

func (u *Users) IsUserLoggedIn(sessionId string) bool {
	return u.isUserPresentInMap(sessionId)
}

func (u *Users) GetLoggedInUserInfo(sessionId string) *Info {
	return u.getUserFromMap(sessionId)
}

func (u *Users) IsUserNameLoggedIn(userName string) bool {
	return u.isUserNameInMap(userName)
}
