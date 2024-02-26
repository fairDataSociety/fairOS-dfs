package mock

import (
	"sync"
)

type ACL struct {
	lock    sync.Mutex
	listMap map[string]map[string]map[string]uint8
}

func (a *ACL) CreateGroup(groupName, ownerAddress string) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.listMap[ownerAddress] = map[string]map[string]uint8{}
	a.listMap[ownerAddress][groupName] = map[string]uint8{}
	return nil
}

func (a *ACL) AddMember(groupName, ownerAddress, memberAddress string, permission uint8) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.listMap[ownerAddress][groupName][memberAddress] = permission
	return nil
}

func (a *ACL) RemoveMember(groupName, ownerAddress, memberAddress string) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	delete(a.listMap[ownerAddress][groupName], memberAddress)
	return nil
}

func (a *ACL) RemoveGroup(groupName, ownerAddress string) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	delete(a.listMap[ownerAddress], groupName)
	return nil
}

func (a *ACL) GetGroupMembers(groupName, ownerAddress string) (map[string]uint8, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.listMap[ownerAddress][groupName], nil
}

func (a *ACL) GetAllGroups(ownerAddress string) (map[string]map[string]uint8, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	return a.listMap[ownerAddress], nil
}

func (a *ACL) UpdatePermission(groupName, ownerAddress, memberAddress string, permission uint8) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.listMap[ownerAddress][groupName][memberAddress] = permission
	return nil
}

func (a *ACL) GetPermission(groupName, ownerAddress, memberAddress string) (uint8, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.listMap[ownerAddress][groupName][memberAddress], nil
}

func NewMockACL() *ACL {
	return &ACL{
		listMap: make(map[string]map[string]map[string]uint8),
	}
}
