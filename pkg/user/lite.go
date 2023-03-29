package user

import (
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	p "github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/taskmanager"
)

// LoadLiteUser creates an off chain user, that has no ens or soc in the swarm.
// It only creates the required information to execute user function and stores it in memory.
func (u *Users) LoadLiteUser(userName, _, mnemonic, sessionId string, tm taskmanager.TaskManagerGO, sm subscriptionManager.SubscriptionManager) (string, string, *Info, error) {
	if !isUserNameValid(userName) {
		return "", "", nil, ErrInvalidUserName
	}

	acc := account.New(u.logger)
	accountInfo := acc.GetUserAccountInfo()
	fd := feed.New(accountInfo, u.client, u.logger)
	// create a new base user account with the mnemonic
	mnemonic, _, err := acc.CreateUserAccount(mnemonic)
	if err != nil { // skipcq: TCV-001
		return "", "", nil, err
	}

	// Instantiate pod, dir & file objects
	file := f.NewFile(userName, u.client, fd, accountInfo.GetAddress(), tm, u.logger)
	dir := d.NewDirectory(userName, u.client, fd, accountInfo.GetAddress(), file, tm, u.logger)
	pod := p.NewPod(u.client, fd, acc, tm, sm, u.logger)
	if sessionId == "" {
		sessionId = cookie.GetUniqueSessionId()
	}

	ui := &Info{
		name:       userName,
		sessionId:  sessionId,
		feedApi:    fd,
		account:    acc,
		file:       file,
		dir:        dir,
		pod:        pod,
		openPods:   make(map[string]*p.Info),
		openPodsMu: &sync.RWMutex{},
	}

	// set cookie and add user to map
	err = u.addUserAndSessionToMap(ui)
	if err != nil {
		return "", "", nil, err
	}

	privateKeyBytes := crypto.FromECDSA(accountInfo.GetPrivateKey())
	return mnemonic, hexutil.Encode(privateKeyBytes)[2:], ui, nil
}
