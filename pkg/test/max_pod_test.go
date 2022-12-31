package test_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/plexsysio/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func TestMaxPods(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
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

	t.Run("create-max-pods", func(t *testing.T) {
		// t.SkipNow()

		maxPodId := 30
		for i := 1; i <= maxPodId; i++ {
			name, err := utils.GetRandString(utils.MaxPodNameLength)
			if err != nil {
				t.Fatal(err)
			}
			podPassword, _ := utils.GetRandString(pod.PodPasswordLength)
			_, err = pod1.CreatePod(name, "", podPassword)
			if err != nil {
				t.Fatalf("error creating pod %s with index %d: %s", name, i, err)
			}
		}
		name, err := utils.GetRandString(utils.MaxPodNameLength)
		if err != nil {
			t.Fatal(err)
		}
		podPassword, _ := utils.GetRandString(pod.PodPasswordLength)
		_, err = pod1.CreatePod(name, "", podPassword)
		if !errors.Is(err, pod.ErrMaximumPodLimit) {
			t.Fatalf("maximum pod limit should have been reached")
		}
	})
}
