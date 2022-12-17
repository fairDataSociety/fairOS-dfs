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
func (a *API) CreateUserV2(userName, passPhrase, mnemonic, sessionId string) (string, string, string, string, *user.Info, error) {
	return a.users.CreateNewUserV2(userName, passPhrase, mnemonic, sessionId, a.tm)
}

// LoginUserV2 is a controller function which calls the users login function.
func (a *API) LoginUserV2(userName, passPhrase, sessionId string) (*user.Info, string, string, error) {
	return a.users.LoginUserV2(userName, passPhrase, a.client, a.tm, sessionId)
}

// LoadLiteUser is a controller function which loads user from mnemonic and doesnot store any user info on chain
func (a *API) LoadLiteUser(userName, passPhrase, mnemonic, sessionId string) (string, *user.Info, error) {
	return a.users.LoadLiteUser(userName, passPhrase, mnemonic, sessionId, a.tm)
}

// LogoutUser is a controller function which gets the logged in user information and logs it out.
func (a *API) LogoutUser(sessionId string) error {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	return a.users.LogoutUser(ui.GetUserName(), sessionId)
}

// DeleteUserV2 is a controller function which deletes a logged in user.
func (a *API) DeleteUserV2(passPhrase, sessionId string) error {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	return a.users.DeleteUserV2(ui.GetUserName(), passPhrase, sessionId, ui)
}

// IsUserNameAvailableV2 checks if a given user name is available in this dfs server.
func (a *API) IsUserNameAvailableV2(userName string) bool {
	return a.users.IsUsernameAvailableV2(userName)
}

// IsUserLoggedIn checks if the given user is logged in
func (a *API) IsUserLoggedIn(userName string) bool {
	// check if a given user is logged in
	return a.users.IsUserNameLoggedIn(userName)
}

// GetUserStat gets the information related to the user.
func (a *API) GetUserStat(sessionId string) (*user.Stat, error) {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	return a.users.GetUserStat(ui)
}

/*
// MigrateUser is a controller function which migrates user credentials to swarm from local storage
func (a *API) MigrateUser(username, passPhrase, sessionId string) error {
	// get the logged in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}
	return a.users.MigrateUser(ui.GetUserName(), username, a.dataDir, passPhrase, sessionId, a.client, ui)
}
*/
