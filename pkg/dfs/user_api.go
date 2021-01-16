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

package dfs

import (
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
)

func (d *DfsAPI) CreateUser(userName, passPhrase, mnemonic string, response http.ResponseWriter, sessionId string) (string, string, error) {
	if !d.client.CheckConnection() {
		return "", "", ErrBeeClient
	}

	reference, rcvdMnemonic, userInfo, err := d.users.CreateNewUser(userName, passPhrase, mnemonic, d.dataDir, d.client, response, sessionId)
	if err != nil {
		return reference, rcvdMnemonic, err
	}

	err = d.users.CreateRootFeeds(userInfo)
	if err != nil {
		return reference, rcvdMnemonic, err
	}
	return reference, rcvdMnemonic, nil
}

func (d *DfsAPI) ImportUserUsingMnemonic(userName, passPhrase, mnemonic string, response http.ResponseWriter, sessionId string) (string, error) {
	reference, _, err := d.CreateUser(userName, passPhrase, mnemonic, response, sessionId)
	return reference, err
}

func (d *DfsAPI) ImportUserUsingAddress(userName, passPhrase, address string, response http.ResponseWriter, sessionId string) error {
	return d.users.ImportUsingAddress(userName, passPhrase, address, d.dataDir, d.client, response, sessionId)
}

func (d *DfsAPI) LoginUser(userName, passPhrase string, response http.ResponseWriter, sessionId string) error {
	return d.users.LoginUser(userName, passPhrase, d.dataDir, d.client, response, sessionId)
}

func (d *DfsAPI) LogoutUser(sessionId string, response http.ResponseWriter) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	return d.users.LogoutUser(ui.GetUserName(), d.dataDir, sessionId, response)
}

func (d *DfsAPI) DeleteUser(passPhrase, sessionId string, response http.ResponseWriter) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	return d.users.DeleteUser(ui.GetUserName(), d.dataDir, passPhrase, sessionId, response, ui)
}

func (d *DfsAPI) IsUserNameAvailable(userName string) bool {
	return d.users.IsUsernameAvailable(userName, d.dataDir)
}

func (d *DfsAPI) IsUserLoggedIn(userName string) bool {
	// check if a given user is logged in
	return d.users.IsUserNameLoggedIn(userName)
}

func (d *DfsAPI) ListAllUsers() ([]string, error) {
	return d.users.ListAllUsers(d.dataDir)
}

func (d *DfsAPI) SaveAvatar(sessionId string, data []byte) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	return d.users.SaveAvatar(data, ui)
}

func (d *DfsAPI) GetAvatar(sessionId string) ([]byte, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	return d.users.GetAvatar(ui)
}

func (d *DfsAPI) SaveName(firstName, lastName, middleName, surname, sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}
	return d.users.SaveName(firstName, lastName, middleName, surname, ui)
}

func (d *DfsAPI) GetName(sessionId string) (*user.Name, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}
	return d.users.GetName(ui)
}

func (d *DfsAPI) SaveContact(phone, mobile string, address *user.Address, sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}
	return d.users.SaveContacts(phone, mobile, address, ui)
}

func (d *DfsAPI) GetContact(sessionId string) (*user.Contacts, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}
	return d.users.GetContacts(ui)
}

func (d *DfsAPI) GetUserStat(sessionId string) (*user.Stat, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}
	return d.users.GetUserStat(ui)
}

func (d *DfsAPI) GetUserSharingInbox(sessionId string) (*user.Inbox, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}
	return d.users.GetSharingInbox(ui)
}

func (d *DfsAPI) GetUserSharingOutbox(sessionId string) (*user.Outbox, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}
	return d.users.GetSharingOutbox(ui)
}

func (d *DfsAPI) ExportUser(sessionId string) (string, string, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return "", "", ErrUserNotLoggedIn
	}
	return d.users.ExportUser(ui)
}
