package pod

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
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
)

// Group is the main struct which acts on groups
type Group struct {
	fd     *feed.API
	acc    *account.Account
	client blockstore.Client
	logger logging.Logger
	//groupsMap     map[string]*Info //  podName -> dir
	//groupMu       *sync.RWMutex
}

// GroupItem defines the structure for a group
type GroupItem struct {
	Name           string `json:"name"`
	OwnerPublicKey []byte `json:"ownerPublicKey"`
	Password       string `json:"password"`
	Secret         []byte `json:"secret"`
}

// GroupList lists all the groups
type GroupList struct {
	Groups []GroupItem `json:"groups"`
}

// NewGroup creates the main group object which has all the methods related to the groups.
func NewGroup(client blockstore.Client, feed *feed.API, account *account.Account, logger logging.Logger) *Group {
	return &Group{
		fd:     feed,
		acc:    account,
		client: client,
		logger: logger,
	}
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

	podPasswordBytes, _ := utils.GetRandBytes(PasswordLength)
	podPassword := hex.EncodeToString(podPasswordBytes)
	group := &GroupItem{
		Name:           name,
		Secret:         key,
		OwnerPublicKey: crypto.FromECDSAPub(g.acc.GetUserAccountInfo().GetPublicKey()),
		Password:       podPassword,
	}
	// Save in te groups list
	groups.Groups = append(groups.Groups, *group)

	// Save the groups list
	err = g.store(groups)

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

	return podInfo, podInfo.GetDirectory().MkRootDir(podInfo.GetPodName(), podPassword, podInfo.GetPodAddress(), podInfo.GetFeed())
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

	groups, err := g.ListGroup()
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	gr := &GroupItem{}
	for _, group := range groups.Groups {
		if group.Name == name {
			gr = &group
			break
		}
	}
	if gr == nil {
		return nil, ErrGroupDoesNotExist
	}

	seed, err := utils.DecryptBytes(crypto.FromECDSA(g.acc.GetUserAccountInfo().GetPrivateKey()), gr.Secret)
	if err != nil { // skipcq: TCV-001
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
		podPassword: gr.Password,
		userAddress: accountInfo.GetAddress(),
		accountInfo: accountInfo,
		feed:        fd,
		dir:         dir,
		file:        file,
		kvStore:     kvStore,
		docStore:    docStore,
	}

	return podInfo, nil
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
