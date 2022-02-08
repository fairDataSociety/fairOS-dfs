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

// CreateUser is a controller function which calls the create user function from the user object.
func (d *DfsAPI) CreateUser(userName, passPhrase, mnemonic string, response http.ResponseWriter, sessionId string) (string, string, error) {
	reference, rcvdMnemonic, _, err := d.users.CreateNewUser(userName, passPhrase, mnemonic, response, sessionId)
	if err != nil {
		return reference, rcvdMnemonic, err
	}
	return reference, rcvdMnemonic, nil
}

// ImportUserUsingMnemonic is a controller function which calls the create user using the mnemonic passed.
func (d *DfsAPI) ImportUserUsingMnemonic(userName, passPhrase, mnemonic string, response http.ResponseWriter, sessionId string) (string, error) {
	reference, _, err := d.CreateUser(userName, passPhrase, mnemonic, response, sessionId)
	return reference, err
}

// ImportUserUsingAddress is a controller function which calls the create user using the address passed.
func (d *DfsAPI) ImportUserUsingAddress(userName, passPhrase, address string, response http.ResponseWriter, sessionId string) error {
	return d.users.ImportUsingAddress(userName, passPhrase, address, d.dataDir, d.client, response, sessionId)
}

// LoginUser is a controller function which calles the users login function.
func (d *DfsAPI) LoginUser(userName, passPhrase string, response http.ResponseWriter, sessionId string) error {
	return d.users.LoginUser(userName, passPhrase, d.dataDir, d.client, response, sessionId)
}

// LogoutUser is a controller function which gets the logged in user information and logs it out.
func (d *DfsAPI) LogoutUser(sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	return d.users.LogoutUser(ui.GetUserName(), d.dataDir, sessionId)
}

// DeleteUser is a controller function which deletes a logged in user.
func (d *DfsAPI) DeleteUser(passPhrase, sessionId string, response http.ResponseWriter) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	return d.users.DeleteUser(ui.GetUserName(), d.dataDir, passPhrase, sessionId, response, ui)
}

// IsUserNameAvailable checks if a given user name is available in this dfs server.
func (d *DfsAPI) IsUserNameAvailable(userName string) bool {
	return d.users.IsUsernameAvailable(userName, d.dataDir)
}

// IsUserLoggedIn checks if the given user is logged in
func (d *DfsAPI) IsUserLoggedIn(userName string) bool {
	// check if a given user is logged in
	return d.users.IsUserNameLoggedIn(userName)
}

// GetUserStat gets the information related to the user.
func (d *DfsAPI) GetUserStat(sessionId string) (*user.Stat, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	return d.users.GetUserStat(ui)
}

// ExportUser exports the currently logged in user.
func (d *DfsAPI) ExportUser(sessionId string) (string, string, error) {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return "", "", ErrUserNotLoggedIn
	}
	return d.users.ExportUser(ui)
}
