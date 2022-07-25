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
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
)

// CreateUserV2 is a controller function which calls the create user function from the user object.
func (d *DfsAPI) CreateUserV2(userName, passPhrase, mnemonic string, sessionId string) (string, string, string, string, *user.Info, error) {
	return d.users.CreateNewUserV2(userName, passPhrase, mnemonic, sessionId)
}

// LoginUserV2 is a controller function which calls the users login function.
func (d *DfsAPI) LoginUserV2(userName, passPhrase string, sessionId string) (*user.Info, string, string, error) {
	return d.users.LoginUserV2(userName, passPhrase, d.client, sessionId)
}

// CreateUser is a controller function which calls the create user function from the user object.
// FOR MIGRATION PURPOSE ONLY
func (d *DfsAPI) CreateUser(userName, passPhrase, mnemonic string, sessionId string) (string, string, *user.Info, error) {
	return d.users.CreateNewUser(userName, passPhrase, mnemonic, sessionId)
}

// LoginUser is a controller function which calls the users login function.
// FOR MIGRATION PURPOSE ONLY
func (d *DfsAPI) LoginUser(userName, passPhrase string, sessionId string) (*user.Info, error) {
	return d.users.LoginUser(userName, passPhrase, d.dataDir, d.client, sessionId)
}

// LogoutUser is a controller function which gets the logged in user information and logs it out.
func (d *DfsAPI) LogoutUser(sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	return d.users.LogoutUser(ui.GetUserName(), sessionId)
}

// DeleteUser is a controller function which deletes a logged in user.
func (d *DfsAPI) DeleteUser(passPhrase, sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	return d.users.DeleteUser(ui.GetUserName(), d.dataDir, passPhrase, sessionId, ui)
}

// DeleteUserV2 is a controller function which deletes a logged in user.
func (d *DfsAPI) DeleteUserV2(passPhrase, sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	return d.users.DeleteUserV2(ui.GetUserName(), passPhrase, sessionId, ui)
}

// IsUserNameAvailable checks if a given user name is available in this dfs server.
func (d *DfsAPI) IsUserNameAvailable(userName string) bool {
	return d.users.IsUsernameAvailable(userName, d.dataDir)
}

// IsUserNameAvailableV2 checks if a given user name is available in this dfs server.
func (d *DfsAPI) IsUserNameAvailableV2(userName string) bool {
	return d.users.IsUsernameAvailableV2(userName)
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

// MigrateUser is a controller function which migrates user credentials to swarm from local storage
func (d *DfsAPI) MigrateUser(username, passPhrase, sessionId string) error {
	// get the logged in user information
	ui := d.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}
	return d.users.MigrateUser(ui.GetUserName(), username, d.dataDir, passPhrase, sessionId, d.client, ui)
}
