package dir_test

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	bm "github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	fm "github.com/fairdatasociety/fairOS-dfs/pkg/file/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
)

func TestChmod(t *testing.T) {
	mockClient := bm.NewMockBeeClient()
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
	mockFile := fm.NewMockFile()
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()

	t.Run("chmod-dir", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PodPasswordLength)
		dirObject := dir.NewDirectory("pod1", mockClient, fd, user, mockFile, tm, logger)

		// make root dir so that other directories can be added
		err = dirObject.MkRootDir("pod1", podPassword, user, fd)
		if err != nil {
			t.Fatal(err)
		}

		// populate the directory with few directory and files
		err := dirObject.MkDir("/dirToChmod", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/dirToChmod/subDir1", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/dirToChmod/subDir2", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		// just add dummy file enty as file listing is not tested here
		err = dirObject.AddEntryToDir("/dirToChmod", podPassword, "file1", true)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.AddEntryToDir("/dirToChmod", podPassword, "file2", true)
		if err != nil {
			t.Fatal(err)
		}

		// stat the directory
		dirStats, err := dirObject.DirStat("pod1", podPassword, "/dirToChmod")
		if err != nil {
			t.Fatal(err)
		}

		if fmt.Sprintf("%o", dir.S_IFDIR|0777) != fmt.Sprintf("%o", dirStats.Mode) {
			t.Fatal("default mode mismatch")
		}

		err = dirObject.Chmod("/dirToChmod", podPassword, 0664)
		if err != nil {
			t.Fatal(err)
		}

		dirStats, err = dirObject.DirStat("pod1", podPassword, "/dirToChmod")
		if err != nil {
			t.Fatal(err)
		}

		if fmt.Sprintf("%o", dir.S_IFDIR|0664) != fmt.Sprintf("%o", dirStats.Mode) {
			t.Fatal("updated mode mismatch")
		}
	})
}
