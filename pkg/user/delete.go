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
	"encoding/hex"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// DeleteUser deletes a user from the Swarm network. Logs him out if he is logged in and remove from all the
// data structures.
func (u *Users) DeleteUser(userName, password, sessionId string, ui *Info) error {
	// check if session id and user address present in map
	if !u.IsUserLoggedIn(sessionId) {
		return ErrUserNotLoggedIn
	}

	// username validation
	if !u.IsUsernameAvailable(userName) {
		return ErrInvalidUserName
	}

	// check for valid password
	userInfo := u.getUserFromMap(sessionId)
	acc := userInfo.account
	if !acc.Authorise(password) {
		return ErrInvalidPassword
	}

	// Logout user
	err := u.Logout(sessionId)
	if err != nil {
		return err
	}

	// get owner address from Subdomain registrar
	owner, err := u.fnm.GetOwner(userName)
	if err != nil {
		return err
	}
	// load public key from public resolver
	publicKey, err := u.fnm.GetPublicKey(userName)
	if err != nil {
		return err
	}
	pb := crypto.FromECDSAPub(publicKey)
	sliAddr, encryptedAddress, err := u.getSecondaryLocationInformation(utils.Address(owner), hex.EncodeToString(pb)+password, ui.GetFeed())
	if err != nil {
		return err
	}
	// decrypt and remove pad the soc address
	accountInfo := acc.GetUserAccountInfo()
	addrStr, err := accountInfo.DecryptContent(password, encryptedAddress)
	if err != nil {
		return err
	}
	addr, err := hex.DecodeString(addrStr)
	if err != nil {
		return err
	}
	err = u.deleteMnemonic(addr, u.client)
	if err != nil {
		return err
	}
	err = u.client.DeleteReference(sliAddr)
	if err != nil {
		return err
	}
	return nil
}
