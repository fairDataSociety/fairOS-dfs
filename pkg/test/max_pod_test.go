package test_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	mockpost "github.com/ethersphere/bee/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"

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
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	sm := mock3.NewMockSubscriptionManager()

	pod1 := pod.NewPod(mockClient, fd, acc, tm, sm, -1, 0, logger)

	t.Run("create-max-pods", func(t *testing.T) {
		// t.SkipNow()

		maxPodId := 30
		for i := 1; i <= maxPodId; i++ {
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
		name, err := utils.GetRandString(utils.MaxPodNameLength)
		if err != nil {
			t.Fatal(err)
		}
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		_, err = pod1.CreatePod(name, "", podPassword)
		if !errors.Is(err, pod.ErrMaximumPodLimit) {
			t.Fatalf("maximum pod limit should have been reached")
		}
	})
}
