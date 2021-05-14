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
)

// DeleteUser deletes a user from the Swarm network. Logs him out if he is logged in and remove from all the
// data structures.
func (u *Users) DeleteUser(userName, dataDir, password, sessionId string, response http.ResponseWriter, ui *Info) error {
	// check if session id and user address present in map
	if !u.IsUserLoggedIn(sessionId) {
		return ErrUserNotLoggedIn
	}

	// username validation
	if !u.IsUsernameAvailable(userName, dataDir) {
		return ErrInvalidUserName
	}

	// check for valid password
	userInfo := u.getUserFromMap(sessionId)
	acc := userInfo.account
	if !acc.Authorise(password) {
		return ErrInvalidPassword
	}

	// Logout user
	err := u.Logout(sessionId, response)
	if err != nil {
		return err
	}

	// remove the user mnemonic file and the user-address mapping file
	address, err := u.getAddressFromUserName(userName, dataDir)
	if err != nil {
		return err
	}
	err = u.deleteMnemonic(userName, address, ui.GetFeed(), u.client)
	if err != nil {
		return err
	}
	err = u.deleteUserMapping(userName, dataDir)
	if err != nil {
		return err
	}
	return nil
}
