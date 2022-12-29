package file_test

import (
	"context"
	"fmt"
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

func TestChmod(t *testing.T) {
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
	t.Run("chmod-file", func(t *testing.T) {
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)

		// upload a file
		_, err = uploadFile(t, fileObject, "/dir1", "file1", "", podPassword, 100, 10)
		if err != nil {
			t.Fatal(err)
		}

		// stat the file
		stats, err := fileObject.GetStats("pod1", "/dir1/file1", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		// check default permission
		if fmt.Sprintf("%o", file.S_IFREG|0666) != fmt.Sprintf("%o", stats.Mode) {
			t.Fatal("default mode mismatch")
		}

		err = fileObject.Chmod("/dir1/file1", podPassword, 0777)
		if err != nil {
			t.Fatal(err)
		}

		stats, err = fileObject.GetStats("pod1", "/dir1/file1", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		if fmt.Sprintf("%o", file.S_IFREG|0777) != fmt.Sprintf("%o", stats.Mode) {
			t.Fatal("mode mismatch after chmod")
		}
	})
}
