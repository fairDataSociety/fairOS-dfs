package test_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	mockpost "github.com/ethersphere/bee/v2/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/v2/pkg/storer/mock"

	mock3 "github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager/rpc/mock"
	"github.com/sirupsen/logrus"

	"github.com/plexsysio/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func TestMaxPods(t *testing.T) {
	storer := mockstorer.New()
	beeUrl := mock.NewTestBeeServer(t, mock.TestServerOptions{
		Storer:          storer,
		PreventRedirect: true,
		Post:            mockpost.New(mockpost.WithAcceptAll()),
	})

	logger := logging.New(io.Discard, logrus.DebugLevel)
	mockClient := bee.NewBeeClient(beeUrl, mock.BatchOkStr, true, logger)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, 500, time.Second*10, logger)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	sm := mock3.NewMockSubscriptionManager()

	pod1 := pod.NewPod(mockClient, fd, acc, tm, sm, -1, 0, logger)

	t.Run("create-max-pods", func(t *testing.T) {
		maxPodId := 100
		for i := 0; i < maxPodId; i++ {
			name, err := utils.GetRandString(utils.MaxPodNameLength)
			if err != nil {
				t.Fatal(err)
			}
			podPassword, _ := utils.GetRandString(pod.PasswordLength)
			_, err = pod1.CreatePod(name, "", podPassword)
			if err != nil {
				t.Fatalf("error creating pod %s with index %d: %s", name, i, err)
			}
		}

		ownPods, _, err := pod1.ListPods()
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, maxPodId, len(ownPods))
	})
}
