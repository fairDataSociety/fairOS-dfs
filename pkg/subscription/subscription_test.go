package subscription_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/plexsysio/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/subscription"
	rpcMock "github.com/fairdatasociety/fairOS-dfs/pkg/subscription/rpc/mock"
)

func TestNew(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	ens := mock2.NewMockNamespaceManager()

	logger := logging.New(os.Stdout, 0)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, logger)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	pod1 := pod.NewPod(mockClient, fd, acc, tm, logger)
	addr := common.HexToAddress(acc.GetUserAccountInfo().GetAddress().Hex())
	sm := rpcMock.NewMockSubscriptionManager()
	m := subscription.New(pod1, addr, ens, sm)

	randomLongPodName1, err := utils.GetRandString(64)
	if err != nil {
		t.Fatalf("error creating pod %s", randomLongPodName1)
	}
	err = m.ListPod(randomLongPodName1, 1)
	if err == nil {
		t.Fatal("pod should not be present")
	}

	// check too long pod name

	podPassword, _ := utils.GetRandString(pod.PasswordLength)
	_, err = pod1.CreatePod(randomLongPodName1, "", podPassword)
	if err != nil {
		t.Fatal(err)
	}

	err = m.ListPod(randomLongPodName1, 1)
	if err != nil {
		t.Fatal(err)
	}

	err = m.DelistPod(randomLongPodName1)
	if err != nil {
		t.Fatal(err)
	}
}
