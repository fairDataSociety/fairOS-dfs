package act

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"sync"

	"github.com/btcsuite/btcd/btcec/v2"

	swarm_act "github.com/asabya/swarm-act"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	blockstore "github.com/asabya/swarm-blockstore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/taskmanager"
)

var (
	ErrACTAlreadyExists = errors.New("ACT already exists")
	ErrACTDoesNowExist  = errors.New("ACT does not exist")
)

type ACT struct {
	fd     *feed.API
	acc    *account.Account
	client blockstore.Client
	logger logging.Logger
	tm     taskmanager.TaskManagerGO
	act    *swarm_act.ACT
	mu     *sync.RWMutex
}

func NewACT(client blockstore.Client, feed *feed.API, account *account.Account, m taskmanager.TaskManagerGO, logger logging.Logger) *ACT {
	accountInfo := account.GetUserAccountInfo()
	accPrivKey, _ := btcec.PrivKeyFromBytes(accountInfo.GetPrivateKey().D.Bytes())
	act := swarm_act.New(client, accPrivKey.ToECDSA(), "")
	return &ACT{
		fd:     feed,
		acc:    account,
		client: client,
		logger: logger,
		tm:     m,
		act:    act,
		mu:     &sync.RWMutex{},
	}
}

func (t *ACT) storeUserACTs(actList List) error {
	data, err := json.Marshal(actList)
	if err != nil {
		return err
	}

	// store data as file and get metadata
	// This is a very hacky way to store pod data, but it works for now
	// We create a new file object with the user account address and upload the data
	// We use the user private key to encrypt data.
	f2 := f.NewFile("", t.client, t.fd, t.acc.GetAddress(account.UserAccountIndex), t.tm, t.logger)
	privKeyBytes := crypto.FromECDSA(t.acc.GetUserAccountInfo().GetPrivateKey())
	return f2.Upload(bytes.NewReader(data), actFile, int64(len(data)), f.MinBlockSize, 0, "/", "gzip", hex.EncodeToString(privKeyBytes))
}

func (t *ACT) loadUserACTs() (List, error) {
	actList := List{}
	f2 := f.NewFile("", t.client, t.fd, t.acc.GetAddress(account.UserAccountIndex), t.tm, t.logger)
	topicString := utils.CombinePathAndFile("", actFile)
	privKeyBytes := crypto.FromECDSA(t.acc.GetUserAccountInfo().GetPrivateKey())
	r, _, err := f2.Download(topicString, hex.EncodeToString(privKeyBytes))
	if err != nil { // skipcq: TCV-001
		return actList, nil
	}
	data, err := io.ReadAll(r)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	if len(data) == 0 {
		return actList, nil
	}

	err = json.Unmarshal(data, &actList)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	return actList, nil
}
