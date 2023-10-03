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
func (a *API) CreateUserV2(userName, passPhrase, mnemonic, sessionId string) (*user.SignupResponse, error) {
	return a.users.CreateNewUserV2(userName, passPhrase, mnemonic, sessionId, a.tm, a.sm)
}

// LoginUserV2 is a controller function which calls the users login function.
func (a *API) LoginUserV2(userName, passPhrase, sessionId string) (*user.LoginResponse, error) {
	return a.users.LoginUserV2(userName, passPhrase, a.client, a.tm, a.sm, sessionId)
}

// LoadLiteUser is a controller function which loads user from mnemonic and doesn't store any user info on chain
func (a *API) LoadLiteUser(userName, passPhrase, mnemonic, sessionId string) (string, string, *user.Info, error) {
	return a.users.LoadLiteUser(userName, passPhrase, mnemonic, sessionId, a.tm, a.sm)
}

// LogoutUser is a controller function which gets the logged-in user information and logs it out.
func (a *API) LogoutUser(sessionId string) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	ui.GetFeed().CommitFeeds()

	return a.users.LogoutUser(ui.GetUserName(), sessionId)
}

// DeleteUserV2 is a controller function which deletes a logged-in user.
func (a *API) DeleteUserV2(passPhrase, sessionId string) error {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	return a.users.DeleteUserV2(ui.GetUserName(), passPhrase, sessionId, ui)
}

// IsUserNameAvailableV2 checks if a given username is available in this dfs server.
func (a *API) IsUserNameAvailableV2(userName string) bool {
	return a.users.IsUsernameAvailableV2(userName)
}

// IsUserLoggedIn checks if the given user is logged-in
func (a *API) IsUserLoggedIn(userName string) bool {
	// check if a given user is logged-in
	return a.users.IsUserNameLoggedIn(userName)
}

// GetUserStat gets the information related to the user.
func (a *API) GetUserStat(sessionId string) (*user.Stat, error) {
	// get the logged-in user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	return a.users.GetUserStat(ui)
}

// ConnectPortableAccountWithWallet is a controller function which calls the users login function.
func (a *API) ConnectPortableAccountWithWallet(userName, passPhrase, addressHex, signature string) error {
	return a.users.ConnectWallet(userName, passPhrase, addressHex, signature, a.client)
}

// LoginWithWallet is a controller function which calls the users login function.
func (a *API) LoginWithWallet(addressHex, signature, sessionId string) (*user.Info, string, error) {
	return a.users.LoginWithWallet(addressHex, signature, a.client, a.tm, a.sm, sessionId)
}

// GetNameHash returns the nameHash of the username
func (a *API) GetNameHash(sessionId, username string) ([32]byte, error) {
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return [32]byte{}, ErrUserNotLoggedIn
	}

	return a.users.GetNameHash(username)
}
