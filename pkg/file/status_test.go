package file_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
)

func TestStatus(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	pod1AccountInfo, err := acc.CreatePodAccount(1, false)
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(pod1AccountInfo, mockClient, logger)
	user := acc.GetAddress(1)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	podPassword, _ := utils.GetRandString(pod.PodPasswordLength)
	t.Run("sync-status-file", func(t *testing.T) {
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)
		numberOfChunks := int64(10)
		// upload a file
		_, err = uploadFile(t, fileObject, "/dir1", "file1", "", podPassword, 4095000*numberOfChunks, 40960000)
		if err != nil {
			t.Fatal(err)
		}

		_, _, _, err := fileObject.Status("/dir1/file12", podPassword)
		if err == nil {
			t.Fatal("should be error")
		}

		// status the file
		total, _, _, err := fileObject.Status("/dir1/file1", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		if total != numberOfChunks {
			t.Fatal("chunk count mismatch for status")
		}
	})

}
