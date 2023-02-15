package subscription_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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

func TestSubscription(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	ens := mock2.NewMockNamespaceManager()

	logger := logging.New(os.Stdout, 0)
	acc1 := account.New(logger)
	_, _, err := acc1.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc1.GetUserAccountInfo(), mockClient, logger)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	pod1 := pod.NewPod(mockClient, fd, acc1, tm, logger)
	addr1 := common.HexToAddress(acc1.GetUserAccountInfo().GetAddress().Hex())
	sm := rpcMock.NewMockSubscriptionManager()
	m := subscription.New(pod1, addr1, ens, sm)

	randomLongPodName1, err := utils.GetRandString(64)
	if err != nil {
		t.Fatalf("error creating pod %s", randomLongPodName1)
	}

	podPassword, _ := utils.GetRandString(pod.PasswordLength)
	_, err = pod1.CreatePod(randomLongPodName1, "", podPassword)
	if err != nil {
		t.Fatal(err)
	}

	err = m.ListPod(randomLongPodName1, 1)
	if err != nil {
		t.Fatal(err)
	}

	acc2 := account.New(logger)
	_, _, err = acc2.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	fd2 := feed.New(acc2.GetUserAccountInfo(), mockClient, logger)
	pod2 := pod.NewPod(mockClient, fd2, acc2, tm, logger)
	addr2 := common.HexToAddress(acc2.GetUserAccountInfo().GetAddress().Hex())

	m2 := subscription.New(pod2, addr2, ens, sm)

	err = m2.RequestSubscription(randomLongPodName1, addr1)
	if err != nil {
		t.Fatal(err)
	}

	err = m.ApproveSubscription(randomLongPodName1, addr2)
	if err != nil {
		t.Fatal(err)
	}

	subs, err := m2.GetSubscriptions()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(subs), 1)
	assert.Equal(t, subs[0].Name, randomLongPodName1)

	randomLongPodName2, err := utils.GetRandString(64)
	if err != nil {
		t.Fatalf("error creating pod %s", randomLongPodName2)
	}

	_, err = pod1.CreatePod(randomLongPodName2, "", podPassword)
	if err != nil {
		t.Fatal(err)
	}

	err = m2.RequestSubscription(randomLongPodName2, addr1)
	if err == nil {
		t.Fatal("pod is not listed")
	}

	err = m.ListPod(randomLongPodName2, 1)
	if err != nil {
		t.Fatal(err)
	}

	err = m.DelistPod(randomLongPodName2)
	if err != nil {
		t.Fatal(err)
	}

	err = m2.RequestSubscription(randomLongPodName2, addr1)
	if err == nil {
		t.Fatal("pod is not listed")
	}
}
