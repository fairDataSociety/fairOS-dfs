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

// DeleteUser deletes a user from the Swarm network. Logs him out if he is logged in and remove from all the
// data structures.
// skipcq: TCV-001
func (u *Users) DeleteUser(userName, dataDir, password, sessionId string, ui *Info) error {
	// check if session id and user address present in map
	if !u.IsUserLoggedIn(sessionId) { // skipcq: TCV-001
		return ErrUserNotLoggedIn
	}

	// username validation
	if !u.IsUsernameAvailable(userName, dataDir) { // skipcq: TCV-001
		return ErrInvalidUserName
	}

	// Logout user
	err := u.Logout(sessionId)
	if err != nil {
		return err
	}

	// skipcq: TCV-001
	// remove the user mnemonic file and the user-address mapping file
	address, err := u.getAddressFromUserName(userName, dataDir)
	if err != nil { // skipcq: TCV-001
		return err
	}
	err = u.deleteMnemonic(userName, address, ui.GetFeed(), u.client) // skipcq: TCV-001
	if err != nil {                                                   // skipcq: TCV-001
		return err
	}
	err = u.deleteUserMapping(userName, dataDir)
	if err != nil { // skipcq: TCV-001
		return err
	}
	return nil // skipcq: TCV-001
}

// DeleteUserV2 deletes a user from the Swarm network. Logs him out if he is logged in and remove from all the
// data structures.
func (u *Users) DeleteUserV2(userName, password, sessionId string, ui *Info) error {
	// check if session id and user address present in map
	if !u.IsUserLoggedIn(sessionId) {
		return ErrUserNotLoggedIn
	}

	// username validation
	if !u.IsUsernameAvailableV2(userName) { // skipcq: TCV-001
		return ErrInvalidUserName
	}

	// check for valid password
	userInfo := u.getUserFromMap(sessionId)
	acc := userInfo.account

	err := u.deletePortableAccount(acc.GetUserAccountInfo().GetAddress(), userName, password, ui.GetFeed())
	if err != nil { // skipcq: TCV-001
		return err
	}

	// Logout user
	return u.Logout(sessionId)
}
