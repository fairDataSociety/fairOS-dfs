package mock

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	accountingmock "github.com/ethersphere/bee/pkg/accounting/mock"
	"github.com/ethersphere/bee/pkg/api"
	"github.com/ethersphere/bee/pkg/auth"
	mockauth "github.com/ethersphere/bee/pkg/auth/mock"
	"github.com/ethersphere/bee/pkg/crypto"
	"github.com/ethersphere/bee/pkg/feeds"
	"github.com/ethersphere/bee/pkg/log"
	p2pmock "github.com/ethersphere/bee/pkg/p2p/mock"
	"github.com/ethersphere/bee/pkg/pingpong"
	"github.com/ethersphere/bee/pkg/postage"
	mockbatchstore "github.com/ethersphere/bee/pkg/postage/batchstore/mock"
	mockpost "github.com/ethersphere/bee/pkg/postage/mock"
	"github.com/ethersphere/bee/pkg/postage/postagecontract"
	contractMock "github.com/ethersphere/bee/pkg/postage/postagecontract/mock"
	"github.com/ethersphere/bee/pkg/pss"
	"github.com/ethersphere/bee/pkg/resolver"
	resolverMock "github.com/ethersphere/bee/pkg/resolver/mock"
	"github.com/ethersphere/bee/pkg/settlement/pseudosettle"
	chequebookmock "github.com/ethersphere/bee/pkg/settlement/swap/chequebook/mock"
	"github.com/ethersphere/bee/pkg/settlement/swap/erc20"
	erc20mock "github.com/ethersphere/bee/pkg/settlement/swap/erc20/mock"
	swapmock "github.com/ethersphere/bee/pkg/settlement/swap/mock"
	statestore "github.com/ethersphere/bee/pkg/statestore/mock"
	"github.com/ethersphere/bee/pkg/status"
	"github.com/ethersphere/bee/pkg/steward"
	"github.com/ethersphere/bee/pkg/storage"
	"github.com/ethersphere/bee/pkg/storage/inmemstore"
	"github.com/ethersphere/bee/pkg/storageincentives"
	"github.com/ethersphere/bee/pkg/storageincentives/staking"
	mock2 "github.com/ethersphere/bee/pkg/storageincentives/staking/mock"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"
	"github.com/ethersphere/bee/pkg/swarm"
	"github.com/ethersphere/bee/pkg/topology/lightnode"
	topologymock "github.com/ethersphere/bee/pkg/topology/mock"
	"github.com/ethersphere/bee/pkg/tracing"
	"github.com/ethersphere/bee/pkg/transaction"
	"github.com/ethersphere/bee/pkg/transaction/backendmock"
	transactionmock "github.com/ethersphere/bee/pkg/transaction/mock"
	"github.com/ethersphere/bee/pkg/util/testutil"
)

var (
	batchOk    = make([]byte, 32)
	BatchOkStr string
)

// nolint:gochecknoinits
func init() {
	_, _ = rand.Read(batchOk)

	BatchOkStr = hex.EncodeToString(batchOk)
}

type TestServerOptions struct {
	Storer             api.Storer
	StateStorer        storage.StateStorer
	Resolver           resolver.Interface
	Pss                pss.Interface
	WsPath             string
	WsPingPeriod       time.Duration
	Logger             log.Logger
	PreventRedirect    bool
	Feeds              feeds.Factory
	CORSAllowedOrigins []string
	PostageContract    postagecontract.Interface
	StakingContract    staking.Contract
	Post               postage.Service
	Steward            steward.Interface
	WsHeaders          http.Header
	Authenticator      auth.Authenticator
	DebugAPI           bool
	Restricted         bool
	DirectUpload       bool
	Probe              *api.Probe

	Overlay         swarm.Address
	PublicKey       ecdsa.PublicKey
	PSSPublicKey    ecdsa.PublicKey
	EthereumAddress common.Address
	BlockTime       time.Duration
	P2P             *p2pmock.Service
	Pingpong        pingpong.Interface
	TopologyOpts    []topologymock.Option
	AccountingOpts  []accountingmock.Option
	ChequebookOpts  []chequebookmock.Option
	SwapOpts        []swapmock.Option
	TransactionOpts []transactionmock.Option

	BatchStore postage.Storer
	SyncStatus func() (bool, error)

	BackendOpts         []backendmock.Option
	Erc20Opts           []erc20mock.Option
	BeeMode             api.BeeNodeMode
	RedistributionAgent *storageincentives.Agent
	NodeStatus          *status.Service
}

func NewTestBeeServer(t *testing.T, o TestServerOptions) string {
	t.Helper()
	pk, _ := crypto.GenerateSecp256k1Key()
	signer := crypto.NewDefaultSigner(pk)

	if o.Logger == nil {
		o.Logger = log.Noop
	}
	if o.Resolver == nil {
		o.Resolver = resolverMock.NewResolver()
	}
	if o.WsPingPeriod == 0 {
		o.WsPingPeriod = 60 * time.Second
	}
	if o.Post == nil {
		o.Post = mockpost.New()
	}
	if o.BatchStore == nil {
		o.BatchStore = mockbatchstore.New(mockbatchstore.WithAcceptAllExistsFunc()) // default is with accept-all Exists() func
	}
	if o.SyncStatus == nil {
		o.SyncStatus = func() (bool, error) { return true, nil }
	}
	if o.Authenticator == nil {
		o.Authenticator = &mockauth.Auth{
			EnforceFunc: func(_, _, _ string) (bool, error) {
				return true, nil
			},
		}
	}

	topologyDriver := topologymock.NewTopologyDriver(o.TopologyOpts...)
	acc := accountingmock.NewAccounting(o.AccountingOpts...)
	settlement := swapmock.New(o.SwapOpts...)
	chequebook := chequebookmock.NewChequebook(o.ChequebookOpts...)
	ln := lightnode.NewContainer(o.Overlay)

	transaction := transactionmock.New(o.TransactionOpts...)

	storeRecipient := statestore.NewStateStore()
	recipient := pseudosettle.New(nil, o.Logger, storeRecipient, nil, big.NewInt(10000), big.NewInt(10000), o.P2P)

	if o.StateStorer == nil {
		o.StateStorer = storeRecipient
	}
	erc20 := erc20mock.New(o.Erc20Opts...)
	backend := backendmock.New(o.BackendOpts...)

	var extraOpts = api.ExtraOptions{
		TopologyDriver:  topologyDriver,
		Accounting:      acc,
		Pseudosettle:    recipient,
		LightNodes:      ln,
		Swap:            settlement,
		Chequebook:      chequebook,
		Pingpong:        o.Pingpong,
		BlockTime:       o.BlockTime,
		Storer:          o.Storer,
		Resolver:        o.Resolver,
		Pss:             o.Pss,
		FeedFactory:     o.Feeds,
		Post:            o.Post,
		PostageContract: o.PostageContract,
		Steward:         o.Steward,
		SyncStatus:      o.SyncStatus,
		Staking:         o.StakingContract,
		NodeStatus:      o.NodeStatus,
	}

	// By default bee mode is set to full mode.
	if o.BeeMode == api.UnknownMode {
		o.BeeMode = api.FullMode
	}

	s := api.New(o.PublicKey, o.PSSPublicKey, o.EthereumAddress, o.Logger, transaction, o.BatchStore, o.BeeMode, true, true, backend, o.CORSAllowedOrigins, inmemstore.New())
	testutil.CleanupCloser(t, s)

	s.SetP2P(o.P2P)

	if o.RedistributionAgent == nil {
		o.RedistributionAgent, _ = createRedistributionAgentService(t, o.Overlay, o.StateStorer, erc20, transaction, backend, o.BatchStore)
		s.SetRedistributionAgent(o.RedistributionAgent)
	}
	testutil.CleanupCloser(t, o.RedistributionAgent)

	s.SetSwarmAddress(&o.Overlay)
	s.SetProbe(o.Probe)

	noOpTracer, tracerCloser, _ := tracing.NewTracer(&tracing.Options{
		Enabled: false,
	})
	testutil.CleanupCloser(t, tracerCloser)

	s.Configure(signer, o.Authenticator, noOpTracer, api.Options{
		CORSAllowedOrigins: o.CORSAllowedOrigins,
		WsPingPeriod:       o.WsPingPeriod,
		Restricted:         o.Restricted,
	}, extraOpts, 1, erc20)

	if o.DebugAPI {
		s.MountTechnicalDebug()
		s.MountDebug(false)
	} else {
		s.MountAPI()
	}

	ts := httptest.NewServer(s)
	t.Cleanup(ts.Close)
	return ts.URL
}

func createRedistributionAgentService(
	t *testing.T,
	addr swarm.Address,
	storer storage.StateStorer,
	erc20Service erc20.Service,
	tranService transaction.Service,
	backend storageincentives.ChainBackend,
	chainStateGetter postage.ChainStateGetter,
) (*storageincentives.Agent, error) {
	t.Helper()

	const blocksPerRound uint64 = 12
	const blocksPerPhase uint64 = 4
	postageContract := contractMock.New(contractMock.WithExpiresBatchesFunc(func(context.Context) error {
		return nil
	}),
	)
	stakingContract := mock2.New(mock2.WithIsFrozen(func(context.Context, uint64) (bool, error) {
		return true, nil
	}))
	contract := &mockContract{}

	return storageincentives.New(
		addr,
		common.Address{},
		backend,
		contract,
		postageContract,
		stakingContract,
		mockstorer.NewReserve(),
		func() bool { return true },
		time.Millisecond*10,
		blocksPerRound,
		blocksPerPhase,
		storer,
		chainStateGetter,
		erc20Service,
		tranService,
		&mockHealth{},
		log.Noop,
	)
}

type contractCall int

func (c contractCall) String() string {
	switch c {
	case isWinnerCall:
		return "isWinnerCall"
	case revealCall:
		return "revealCall"
	case commitCall:
		return "commitCall"
	case claimCall:
		return "claimCall"
	}
	return "unknown"
}

const (
	isWinnerCall contractCall = iota
	revealCall
	commitCall
	claimCall
)

type mockContract struct {
	callsList []contractCall
	mtx       sync.Mutex
}

func (m *mockContract) Fee(ctx context.Context, txHash common.Hash) *big.Int {
	return big.NewInt(1000)
}

func (m *mockContract) ReserveSalt(context.Context) ([]byte, error) {
	return nil, nil
}

func (m *mockContract) IsPlaying(context.Context, uint8) (bool, error) {
	return true, nil
}

func (m *mockContract) IsWinner(context.Context) (bool, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.callsList = append(m.callsList, isWinnerCall)
	return false, nil
}

func (m *mockContract) Claim(context.Context) (common.Hash, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.callsList = append(m.callsList, claimCall)
	return common.Hash{}, nil
}

func (m *mockContract) Commit(context.Context, []byte, *big.Int) (common.Hash, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.callsList = append(m.callsList, commitCall)
	return common.Hash{}, nil
}

func (m *mockContract) Reveal(context.Context, uint8, []byte, []byte) (common.Hash, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.callsList = append(m.callsList, revealCall)
	return common.Hash{}, nil
}

type mockHealth struct{}

func (m *mockHealth) IsHealthy() bool { return true }
