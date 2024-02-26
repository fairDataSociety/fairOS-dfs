package acl

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	NoPermission      uint8 = 0
	PermissionRead    uint8 = 1
	PermissionWrite   uint8 = 2
	PermissionExecute uint8 = 4
)

type ACL struct {
	c       blockstore.Client
	logger  logging.Logger
	f       *feed.API
	lock    sync.Mutex
	listMap map[string]map[string]uint8
}

func (a *ACL) GetGroupMembers(groupName, ownerAddress string) (map[string]uint8, error) {
	err := a.loadPermissions(groupName, ownerAddress)
	if err != nil {
		return nil, err
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.listMap[groupName], nil
}

func (a *ACL) UpdatePermission(groupName, ownerAddress, memberAddress string, permission uint8) error {
	a.lock.Lock()
	a.listMap[groupName][memberAddress] = permission
	a.lock.Unlock()

	return a.storePermissions(groupName, ownerAddress)
}

func (a *ACL) GetPermission(groupName, ownerAddress, memberAddress string) (uint8, error) {
	err := a.loadPermissions(groupName, ownerAddress)
	if err != nil {
		return 0, err
	}
	a.lock.Lock()
	defer a.lock.Unlock()

	return a.listMap[groupName][memberAddress], nil
}

func (a *ACL) CreateGroup(groupName, ownerAddress string) error {
	return a.storePermissions(groupName, ownerAddress)
}

func (a *ACL) AddMember(groupName, ownerAddress, memberAddress string, permission uint8) error {
	a.lock.Lock()
	a.listMap[groupName][memberAddress] = permission
	a.lock.Unlock()
	return a.storePermissions(groupName, ownerAddress)
}

func (a *ACL) RemoveMember(groupName, ownerAddress, memberAddress string) error {
	a.lock.Lock()
	delete(a.listMap[groupName], memberAddress)
	a.lock.Unlock()
	return a.storePermissions(groupName, ownerAddress)
}

func (a *ACL) RemoveGroup(groupName, ownerAddress string) error {
	a.lock.Lock()
	delete(a.listMap, groupName)
	a.lock.Unlock()
	return a.storePermissions(groupName, ownerAddress)
}

func NewACL(c blockstore.Client, f *feed.API, logger logging.Logger) *ACL {
	return &ACL{
		c:       c,
		f:       f,
		logger:  logger,
		listMap: map[string]map[string]uint8{},
	}
}

func (a *ACL) loadPermissions(group, ownerAddress string) error {
	f2 := f.NewFile("", a.c, a.f, utils.HexToAddress(ownerAddress), nil, a.logger)
	topicString := utils.CombinePathAndFile(ownerAddress, group)
	r, _, err := f2.Download(topicString, "")
	if err != nil { // skipcq: TCV-001
		return err
	}
	permissions := map[string]uint8{}
	data, err := io.ReadAll(r)
	if err != nil { // skipcq: TCV-001
		return err
	}

	if len(data) == 0 {
		a.listMap[group] = permissions
		return nil
	}

	err = json.Unmarshal(data, &permissions)
	if err != nil { // skipcq: TCV-001
		return err
	}

	a.lock.Lock()
	defer a.lock.Unlock()
	a.listMap[group] = permissions
	return nil
}

func (a *ACL) storePermissions(group, ownerAddress string) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	permissions := map[string]uint8{}
	if _, ok := a.listMap[group]; ok {
		permissions = a.listMap[group]
	} else {
		permissions[ownerAddress] = PermissionWrite
		a.listMap[group] = permissions
	}
	data, err := json.Marshal(&permissions)
	if err != nil {
		return err
	}

	f2 := f.NewFile("", a.c, a.f, utils.HexToAddress(ownerAddress), nil, a.logger)
	topicString := utils.CombinePathAndFile(ownerAddress, group)
	return f2.Upload(bytes.NewReader(data), topicString, int64(len(data)), f.MinBlockSize, 0, "/", "gzip", "")
}
