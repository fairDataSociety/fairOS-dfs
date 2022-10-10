package pod_test

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
)

func TestMaxPods(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("password", "")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, logger)
	tm := taskmanager.New(1, 10, time.Second*15, logger)

	pod1 := pod.NewPod(mockClient, fd, acc, tm, logger)

	t.Run("create-max-pods", func(t *testing.T) {
		maxPodId := 140
		for i := 1; i <= maxPodId; i++ {
			name, err := utils.GetRandString(25)
			if err != nil {
				t.Fatal(err)
			}
			_, err = pod1.CreatePod(name, "password", "")
			if err != nil {
				t.Fatalf("error creating pod %s with index %d: %s", name, i, err)
			}
		}
		name, err := utils.GetRandString(25)
		if err != nil {
			t.Fatal(err)
		}
		_, err = pod1.CreatePod(name, "password", "")
		if !errors.Is(err, pod.ErrMaximumPodLimit) {
			t.Fatalf("maximum pod limit should have been reached")
		}
	})
}
