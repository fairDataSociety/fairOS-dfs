package dfs

import (
	"crypto/ecdsa"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	"github.com/ethersphere/bee/v2/pkg/swarm"
	"github.com/fairdatasociety/fairOS-dfs/pkg/act"
)

func (a *API) CreateGranteePublicKey(sessionId, actName string, publicKey *ecdsa.PublicKey) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}
	actList := ui.GetACTList()
	_, err := actList.CreateUpdateACT(actName, publicKey, nil)
	return err
}

func (a *API) GrantRevokeGranteePublicKey(sessionId, actName string, publicKeyGrant, publicKeyRevoke *ecdsa.PublicKey) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	actList := ui.GetACTList()
	_, err := actList.CreateUpdateACT(actName, publicKeyGrant, publicKeyRevoke)
	return err
}

func (a *API) ListGrantees(sessionId, actName string) ([]string, error) {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	actList := ui.GetACTList()
	return actList.GetGrantees(actName)
}

func (a *API) ACTPodShare(sessionId, podName, actName string) (*act.Content, error) {
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	actList := ui.GetACTList()
	address, err := ui.GetPod().PodShare(podName, "")
	if err != nil {
		return nil, err
	}

	addr, err := swarm.ParseHexAddress(address)
	if err != nil {
		return nil, err
	}

	return actList.GrantAccess(actName, addr)
}

func (a *API) OpenACTPod(sessionId, actName string) error {
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	actList := ui.GetACTList()
	addr, err := actList.GetPodAccess(actName)
	if err != nil {
		return err
	}
	info, err := ui.GetPod().ReceivePodInfo(utils.NewReference(addr.Bytes()))
	if err != nil {
		return err
	}

	_, err = ui.GetPod().OpenActPod(info, actName)
	if err != nil {
		return err
	}

	return err
}

func (a *API) GetACTs(sessionId string) (act.List, error) {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	actList := ui.GetACTList()
	return actList.GetList()
}
func (a *API) SaveACTPod(sessionId, actName string, c *act.Content) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	actList := ui.GetACTList()
	return actList.SaveGrantedPod(actName, c)
}

func (a *API) GetACTContents(sessionId, actName string) ([]*act.Content, error) {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	actList := ui.GetACTList()

	return actList.GetContentList(actName)
}

//func (a *API) GrantRevokeGrantee(sessionId, grantUser, revokeUser, actName string) error {
//	// get the loggedin user information
//	ui := a.users.GetLoggedInUserInfo(sessionId)
//	if ui == nil {
//		return ErrUserNotLoggedIn
//	}
//	grantList := make([]*ecdsa.PublicKey, 1)
//	revokeList := make([]*ecdsa.PublicKey, 1)
//
//	actList := ui.GetACTList()
//	act, err := actList.GetACT(actName)
//	if err != nil { // skipcq: TCV-001
//		return err
//	}
//	historyRef, granteeListRef := act.HistoryRef, act.GranteesRef
//	if grantUser == "" && revokeUser == "" {
//		return fmt.Errorf("grant or revoke user required")
//	}
//	if grantUser != "" {
//		publicKeyGrant, _, err := a.users.GetUserInfo(grantUser)
//		if err != nil { // skipcq: TCV-001
//			return err
//		}
//		if publicKeyGrant == nil {
//			return fmt.Errorf("public key not found")
//		}
//		grantList[0] = publicKeyGrant
//	}
//	if revokeUser != "" {
//		publicKeyRevoke, _, err := a.users.GetUserInfo(revokeUser)
//		if err != nil { // skipcq: TCV-001
//			return err
//		}
//		if publicKeyRevoke == nil {
//			return fmt.Errorf("public key not found")
//		}
//		revokeList[0] = publicKeyRevoke
//	}
//
//	swarmAct := ui.GetACT()
//	createResp, err := swarmAct.RevokeGrant(a.context, granteeListRef, historyRef, grantList, revokeList)
//	if err != nil { // skipcq: TCV-001
//		return err
//	}
//	_, err = actList.CreateUpdateACT(actName, createResp.HistoryReference, createResp.Reference)
//	return err
//}

//func (a *API) CreateGrantee(sessionId, user, actName string) error {
//	// get the loggedin user information
//	ui := a.users.GetLoggedInUserInfo(sessionId)
//	if ui == nil {
//		return ErrUserNotLoggedIn
//	}
//
//	publicKey, _, err := a.users.GetUserInfo(user)
//	if err != nil { // skipcq: TCV-001
//		return err
//	}
//	if publicKey == nil {
//		return fmt.Errorf("public key not found")
//	}
//
//	act := ui.GetACT()
//	addList := []*ecdsa.PublicKey{publicKey}
//	createResp, err := act.CreateGrantee(a.context, swarm.ZeroAddress, addList)
//
//	// TODO save in ACT LIST
//	_ = createResp
//	return err
//}
