package pod

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/acl"
	aclController "github.com/fairdatasociety/fairOS-dfs/pkg/acl/acl"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	c "github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/tyler-smith/go-bip39"
)

const (
	GroupFile = "Groups"
)

var (
	// ErrGroupAlreadyExists is returned when a group already exists
	ErrGroupAlreadyExists = fmt.Errorf("group already exists")

	// ErrGroupDoesNotExist is returned when a group does not exist
	ErrGroupDoesNotExist = fmt.Errorf("group does not exist")

	ErrPermissionDenied = fmt.Errorf("permission denied")
)

// Group is the main struct which acts on groups
type Group struct {
	fd        *feed.API
	acc       *account.Account
	client    blockstore.Client
	logger    logging.Logger
	acl       acl.ACL
	groupsMap map[string]*Info //  podName -> dir
	groupMu   *sync.RWMutex
}

// GroupItem defines the structure for a group
type GroupItem struct {
	Name           string `json:"name"`
	OwnerPublicKey []byte `json:"ownerPublicKey"`
	OwnerAddress   string `json:"ownerAddress"`
	Password       string `json:"password"`
	Secret         []byte `json:"secret"`
}

// GroupList lists all the groups
type GroupList struct {
	Groups       []GroupItem `json:"groups"`
	SharedGroups []GroupItem `json:"sharedGroups"`
}

// NewGroup creates the main group object which has all the methods related to the groups.
func NewGroup(client blockstore.Client, feed *feed.API, account *account.Account, acl acl.ACL, logger logging.Logger) *Group {
	return &Group{
		fd:        feed,
		acc:       account,
		client:    client,
		logger:    logger,
		acl:       acl,
		groupsMap: make(map[string]*Info),
		groupMu:   &sync.RWMutex{},
	}
}

func (g *Group) addPodToPodMap(name string, info *Info) {
	g.groupMu.Lock()
	defer g.groupMu.Unlock()
	g.groupsMap[name] = info
}

func (g *Group) removePodFromPodMap(name string) {
	g.groupMu.Lock()
	defer g.groupMu.Unlock()
	delete(g.groupsMap, name)
}

// GetGroupInfoFromMap returns the group info for the given group name.
func (g *Group) GetGroupInfoFromMap(name string) (*Info, string, error) {
	g.groupMu.Lock()
	defer g.groupMu.Unlock()
	if podInfo, ok := g.groupsMap[name]; ok {
		return podInfo, podInfo.podPassword, nil
	}
	return nil, "", fmt.Errorf("could not find pod: %s", name)
}

// CreateGroup creates a new Group
func (g *Group) CreateGroup(name string) (*Info, error) {
	// sanitise: check name for spaces
	name = strings.TrimSpace(name)

	groups, err := g.ListGroup()
	if err != nil && !errors.Is(err, f.ErrFileNotFound) { // skipcq: TCV-001
		return nil, err
	}
	if g.checkIfPodPresent(groups, name) {
		return nil, ErrGroupAlreadyExists
	}

	// generate a new mnemonic: GroupSecret
	entropy, err := bip39.NewEntropy(128)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	seed, err := hdwallet.NewSeedFromMnemonic(mnemonic)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	// Encrypt the GroupSecret with the user's private key
	key, err := utils.EncryptBytes(crypto.FromECDSA(g.acc.GetUserAccountInfo().GetPrivateKey()), seed)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	address := g.acc.GetUserAccountInfo().GetAddress()
	commonAddr := common.HexToAddress(address.Hex())
	addressStr := commonAddr.Hex()
	podPasswordBytes, _ := utils.GetRandBytes(PasswordLength)
	podPassword := hex.EncodeToString(podPasswordBytes)
	group := &GroupItem{
		Name:           name,
		Secret:         key,
		OwnerPublicKey: crypto.FromECDSAPub(g.acc.GetUserAccountInfo().GetPublicKey()),
		OwnerAddress:   commonAddr.Hex(),
		Password:       podPassword,
	}

	// Save in te groups list
	groups.Groups = append(groups.Groups, *group)

	// Save the groups list
	err = g.store(groups)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	err = g.acl.CreateGroup(name, addressStr)
	if err != nil {
		return nil, err
	}
	acc := account.New(g.logger)
	err = acc.LoadUserAccountFromSeed(seed)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	accountInfo := acc.GetUserAccountInfo()
	// load encrypted private key
	fd := feed.New(accountInfo, g.client, -1, 0, g.logger)
	file := f.NewFile(name, g.client, fd, accountInfo.GetAddress(), nil, g.logger)
	dir := d.NewDirectory(name, g.client, fd, accountInfo.GetAddress(), file, nil, g.logger)
	kvStore := c.NewKeyValueStore(name, fd, accountInfo, accountInfo.GetAddress(), g.client, g.logger)
	docStore := c.NewDocumentStore(name, fd, accountInfo, accountInfo.GetAddress(), file, nil, g.client, g.logger)

	podInfo := &Info{
		podName:     name,
		podPassword: group.Password,
		userAddress: accountInfo.GetAddress(),
		accountInfo: accountInfo,
		feed:        fd,
		dir:         dir,
		file:        file,
		kvStore:     kvStore,
		docStore:    docStore,
	}

	g.addPodToPodMap(podInfo.GetPodName(), podInfo)
	return podInfo, podInfo.GetDirectory().MkRootDir(podInfo.GetPodName(), podPassword, podInfo.GetPodAddress(), podInfo.GetFeed())
}

// RemoveGroup removes a group
func (g *Group) RemoveGroup(groupName string) error {
	// check if group exists
	groupName = strings.TrimSpace(groupName)

	groups, err := g.ListGroup()
	if err != nil && !errors.Is(err, f.ErrFileNotFound) { // skipcq: TCV-001
		return err
	}
	found := false
	for index, pod := range groups.Groups {
		if pod.Name == groupName {
			groups.Groups = append(groups.Groups[:index], groups.Groups[index+1:]...)
			found = true
		}
	}
	if !found {
		return ErrGroupDoesNotExist
	}

	gi, err := g.OpenGroup(groupName)
	if err != nil {
		return err
	}
	err = gi.GetDocStore().DeleteAllDocumentDBs(gi.GetPodPassword())
	if err != nil {
		return err
	}

	err = gi.GetKVStore().DeleteAllKVTables(gi.GetPodPassword())
	if err != nil {
		return err
	}

	address := g.acc.GetUserAccountInfo().GetAddress()
	addressStr := address.Hex()
	err = g.acl.RemoveGroup(groupName, addressStr)
	if err != nil {
		return err
	}
	g.removePodFromPodMap(groupName)

	return g.store(groups)
}

// RemoveSharedGroup removes a group from sharedGroup list
func (g *Group) RemoveSharedGroup(groupName string) error {
	// check if group exists
	groupName = strings.TrimSpace(groupName)

	groups, err := g.ListGroup()
	if err != nil && !errors.Is(err, f.ErrFileNotFound) { // skipcq: TCV-001
		return err
	}
	found := false
	for index, group := range groups.SharedGroups {
		if group.Name == groupName {
			groups.SharedGroups = append(groups.SharedGroups[:index], groups.SharedGroups[index+1:]...)
			found = true
		}
	}
	if !found {
		return ErrGroupDoesNotExist
	}

	g.removePodFromPodMap(groupName)

	return g.store(groups)
}

func (*Group) checkIfPodPresent(groups *GroupList, name string) bool {
	if groups == nil || groups.Groups == nil {
		return false
	}
	for _, group := range groups.Groups {
		if group.Name == name {
			return true
		}
	}
	for _, group := range groups.SharedGroups {
		if group.Name == name {
			return true
		}
	}
	return false
}

func (g *Group) ListGroup() (*GroupList, error) {
	// load groups from GroupsFile
	return g.load()
}

// OpenGroup opens a new Group
func (g *Group) OpenGroup(name string) (*Info, error) {
	// sanitise: check name for spaces
	name = strings.TrimSpace(name)

	pi, _, _ := g.GetGroupInfoFromMap(name)
	if pi != nil {
		return pi, nil
	}

	groups, err := g.ListGroup()
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	var gr *GroupItem
	shared := false
	for _, group := range groups.Groups {
		if group.Name == name {
			gr = &group
			break
		}
	}
	if gr == nil {
		for _, group := range groups.SharedGroups {
			if group.Name == name {
				gr = &group
				shared = true
				break
			}
		}
	}

	if gr == nil {
		return nil, ErrGroupDoesNotExist
	}

	var (
		accountInfo *account.Info
		file        *f.File
		fd          *feed.API
		dir         *d.Directory
	)
	if shared {
		permission, err := g.GetPermission(gr.Name)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		if permission != aclController.PermissionRead && permission != aclController.PermissionWrite {
			_ = g.RemoveSharedGroup(gr.Name)
			return nil, ErrPermissionDenied
		}
		ownerPublicKey, err := crypto.UnmarshalPubkey(gr.OwnerPublicKey)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		a, _ := ownerPublicKey.Curve.ScalarMult(ownerPublicKey.X, ownerPublicKey.Y, g.acc.GetUserAccountInfo().GetPrivateKey().D.Bytes())
		secret := sha256.Sum256(a.Bytes())
		seed, err := utils.DecryptBytes(secret[:], gr.Secret)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		acc := account.New(g.logger)
		err = acc.LoadUserAccountFromSeed(seed)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		accountInfo = acc.GetUserAccountInfo()

		if permission == aclController.PermissionRead {
			readAccount := g.acc.GetEmptyAccountInfo()
			readAccount.SetAddress(accountInfo.GetAddress())

			fd = feed.New(readAccount, g.client, -1, 0, g.logger)
			file = f.NewFile(name, g.client, fd, readAccount.GetAddress(), nil, g.logger)
			dir = d.NewDirectory(name, g.client, fd, readAccount.GetAddress(), file, nil, g.logger)
		} else {
			fd = feed.New(accountInfo, g.client, -1, 0, g.logger)
			file = f.NewFile(name, g.client, fd, accountInfo.GetAddress(), nil, g.logger)
			dir = d.NewDirectory(name, g.client, fd, accountInfo.GetAddress(), file, nil, g.logger)
		}
	} else {
		seed, err := utils.DecryptBytes(crypto.FromECDSA(g.acc.GetUserAccountInfo().GetPrivateKey()), gr.Secret)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}

		acc := account.New(g.logger)
		err = acc.LoadUserAccountFromSeed(seed)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		accountInfo = acc.GetUserAccountInfo()

		// load encrypted private key
		fd = feed.New(accountInfo, g.client, -1, 0, g.logger)
		file = f.NewFile(name, g.client, fd, accountInfo.GetAddress(), nil, g.logger)
		dir = d.NewDirectory(name, g.client, fd, accountInfo.GetAddress(), file, nil, g.logger)
	}
	kvStore := c.NewKeyValueStore(name, fd, accountInfo, accountInfo.GetAddress(), g.client, g.logger)
	docStore := c.NewDocumentStore(name, fd, accountInfo, accountInfo.GetAddress(), file, nil, g.client, g.logger)
	podInfo := &Info{
		podName:     name,
		podPassword: gr.Password,
		userAddress: accountInfo.GetAddress(),
		accountInfo: accountInfo,
		feed:        fd,
		dir:         dir,
		file:        file,
		kvStore:     kvStore,
		docStore:    docStore,
	}
	g.addPodToPodMap(podInfo.GetPodName(), podInfo)
	return podInfo, nil
}

// CloseGroup closed an already opened group and removes its information from directory and file
// data structures.
func (g *Group) CloseGroup(podName string) error {
	podInfo, _, err := g.GetGroupInfoFromMap(podName)
	if err != nil { // skipcq: TCV-001
		return err
	}
	if err := podInfo.feed.Close(); err != nil {
		return err
	}
	if err := g.fd.Close(); err != nil {
		return err
	}

	// remove from all thr maps
	podInfo.dir.RemoveAllFromDirectoryMap()
	podInfo.file.RemoveAllFromFileMap()
	g.removePodFromPodMap(podName)
	return nil
}

func (g *Group) load() (*GroupList, error) {
	list := &GroupList{
		Groups: []GroupItem{},
	}
	f2 := f.NewFile("", g.client, g.fd, g.acc.GetAddress(account.UserAccountIndex), nil, g.logger)
	topicString := utils.CombinePathAndFile("", GroupFile)
	privKeyBytes := crypto.FromECDSA(g.acc.GetUserAccountInfo().GetPrivateKey())
	r, _, err := f2.Download(topicString, hex.EncodeToString(privKeyBytes))
	if err != nil { // skipcq: TCV-001
		return list, err
	}

	data, err := io.ReadAll(r)
	if err != nil { // skipcq: TCV-001
		return list, err
	}

	if len(data) == 0 {
		return list, nil
	}

	err = json.Unmarshal(data, list)
	if err != nil { // skipcq: TCV-001
		return list, err
	}

	return list, nil
}

func (g *Group) store(list *GroupList) error {
	data, err := json.Marshal(list)
	if err != nil {
		return err
	}

	f2 := f.NewFile("", g.client, g.fd, g.acc.GetAddress(account.UserAccountIndex), nil, g.logger)
	privKeyBytes := crypto.FromECDSA(g.acc.GetUserAccountInfo().GetPrivateKey())
	return f2.Upload(bytes.NewReader(data), GroupFile, int64(len(data)), f.MinBlockSize, 0, "/", "gzip", hex.EncodeToString(privKeyBytes))
}
