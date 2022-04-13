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
func (u *Users) MigrateUser(userName, dataDir, password, sessionId string, ui *Info) error {
	// check if session id and user address present in map
	if !u.IsUserLoggedIn(sessionId) {
		return ErrUserNotLoggedIn
	}

	// username availability
	if !u.IsUsernameAvailable(userName, dataDir) {
		return ErrInvalidUserName
	}

	// check for valid password
	userInfo := u.getUserFromMap(sessionId)
	acc := userInfo.account
	if !acc.Authorise(password) {
		return ErrInvalidPassword
	}
	address, err := u.getAddressFromUserName(userName, dataDir)
	if err != nil {
		return err
	}
	accountInfo := acc.GetUserAccountInfo()
	encryptedMnemonic, err := u.getEncryptedMnemonic(userName, address, userInfo.GetFeed())
	if err != nil {
		return err
	}
	// create ens subdomain and store mnemonic
	err = u.fnm.RegisterSubdomain(userName, common.HexToAddress(accountInfo.GetAddress().Hex()))
	if err != nil {
		return err
	}

	_, err = u.fnm.SetResolver(userName, common.Address(accountInfo.GetAddress()), accountInfo.GetPrivateKey())
	if err != nil {
		return err
	}

	err = u.fnm.SetAll(userName, common.HexToAddress(accountInfo.GetAddress().Hex()), accountInfo.GetPrivateKey())
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
