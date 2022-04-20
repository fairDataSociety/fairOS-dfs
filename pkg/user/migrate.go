package user

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// MigrateUser migrates an user credential from local storage to the Swarm network.
// Deletes local information. It also deletes previous mnemonic and stores it in secondary location
// Logs him out if he is logged in.
func (u *Users) MigrateUser(oldUsername, newUsername, dataDir, password, sessionId string, ui *Info) error {
	// check if session id and user address present in map
	if !u.IsUserLoggedIn(sessionId) {
		return ErrUserNotLoggedIn
	}
	if newUsername == "" {
		newUsername = oldUsername
	}
	// username availability
	if !u.IsUsernameAvailable(oldUsername, dataDir) {
		return ErrInvalidUserName
	}

	// username availability for v2
	if u.IsUsernameAvailableV2(newUsername) {
		return ErrUserAlreadyPresent
	}

	// check for valid password
	userInfo := u.getUserFromMap(sessionId)
	acc := userInfo.account
	if !acc.Authorise(password) {
		return ErrInvalidPassword
	}
	address, err := u.getAddressFromUserName(oldUsername, dataDir)
	if err != nil {
		return err
	}
	accountInfo := acc.GetUserAccountInfo()
	encryptedMnemonic, err := u.getEncryptedMnemonic(oldUsername, address, userInfo.GetFeed())
	if err != nil {
		return err
	}
	// create ens subdomain and store mnemonic
	err = u.ens.RegisterSubdomain(newUsername, common.HexToAddress(accountInfo.GetAddress().Hex()))
	if err != nil {
		return err
	}

	_, err = u.ens.SetResolver(newUsername, common.Address(accountInfo.GetAddress()), accountInfo.GetPrivateKey())
	if err != nil {
		return err
	}

	err = u.ens.SetAll(newUsername, common.HexToAddress(accountInfo.GetAddress().Hex()), accountInfo.GetPrivateKey())
	if err != nil {
		return err
	}

	// store the encrypted mnemonic in Swarm
	addr, err := u.uploadEncryptedMnemonicSOC(accountInfo, encryptedMnemonic, userInfo.GetFeed())
	if err != nil {
		return err
	}

	// encrypt and pad the soc address
	encryptedAddress, err := accountInfo.EncryptContent(password, utils.Encode(addr))
	if err != nil {
		return err
	}

	// store encrypted soc address in secondary location
	pb := crypto.FromECDSAPub(accountInfo.GetPublicKey())
	err = u.uploadSecondaryLocationInformation(accountInfo, encryptedAddress, hex.EncodeToString(pb)+password, userInfo.GetFeed())
	if err != nil {
		return err
	}

	// Logout user
	err = u.Logout(sessionId)
	if err != nil {
		return err
	}

	err = u.deleteMnemonic(oldUsername, address, ui.GetFeed(), u.client)
	if err != nil {
		return err
	}

	return u.deleteUserMapping(oldUsername, dataDir)
}
