package dfs

import (
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

// CreateGroup creates a new group
func (a *API) CreateGroup(sessionId, groupName string) (*pod.Info, error) {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	group := ui.GetGroup()

	// create a new group
	return group.CreateGroup(groupName)
}

// RemoveGroup deletes an existing group
func (a *API) RemoveGroup(sessionId, groupName string) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	group := ui.GetGroup()

	return group.RemoveGroup(groupName)
}

// RemoveSharedGroup deletes an existing group from shared list
func (a *API) RemoveSharedGroup(sessionId, groupName string) error {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	group := ui.GetGroup()

	return group.RemoveSharedGroup(groupName)
}

func (a *API) ListGroups(sessionId string) (*pod.GroupList, error) {
	// get the loggedin user information
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	group := ui.GetGroup()

	return group.ListGroup()
}

func (a *API) OpenGroup(sessionId, groupName string) (*pod.Info, error) {
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	group := ui.GetGroup()

	return group.OpenGroup(groupName)
}

func (a *API) CloseGroup(sessionId, groupName string) error {
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	group := ui.GetGroup()

	return group.CloseGroup(groupName)
}

func (a *API) AddMember(sessionId, groupName, username string, permission uint8) ([]byte, error) {
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	group := ui.GetGroup()

	nh, err := a.users.GetNameHash(username)
	if err != nil {
		return nil, err
	}

	addr, pub, err := a.users.GetUserInfoFromENS(nh)
	if err != nil {
		return nil, err
	}
	return group.AddMember(groupName, addr, pub, permission)
}

func (a *API) AcceptGroupInvite(sessionId string, ref []byte) error {
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	group := ui.GetGroup()

	return group.AcceptGroupInvite(ref)
}

func (a *API) RemoveMember(sessionId, groupName, username string) error {
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	group := ui.GetGroup()

	nh, err := a.users.GetNameHash(username)
	if err != nil {
		return err
	}

	addr, _, err := a.users.GetUserInfoFromENS(nh)
	if err != nil {
		return err
	}
	return group.RemoveMember(groupName, addr)
}

func (a *API) UpdatePermission(sessionId, groupName, username string, permission uint8) error {
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return ErrUserNotLoggedIn
	}

	group := ui.GetGroup()

	nh, err := a.users.GetNameHash(username)
	if err != nil {
		return err
	}

	addr, _, err := a.users.GetUserInfoFromENS(nh)
	if err != nil {
		return err
	}
	return group.UpdatePermission(groupName, addr, permission)
}

func (a *API) GetPermission(sessionId, groupName string) (uint8, error) {
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return 0, ErrUserNotLoggedIn
	}

	group := ui.GetGroup()

	return group.GetPermission(groupName)
}

func (a *API) GetGroupMembers(sessionId, groupName string) (map[string]uint8, error) {
	ui := a.users.GetLoggedInUserInfo(sessionId)
	if ui == nil {
		return nil, ErrUserNotLoggedIn
	}

	group := ui.GetGroup()

	return group.GetGroupMembers(groupName)
}
