package dir_test

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/ethersphere/bee/v2/pkg/file/redundancy"

	"github.com/fairdatasociety/fairOS-dfs/pkg/file"

	"github.com/asabya/swarm-blockstore/bee/mock"

	"github.com/asabya/swarm-blockstore/bee"
	mockpost "github.com/ethersphere/bee/v2/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/v2/pkg/storer/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
	"github.com/sirupsen/logrus"
)

func TestChmod(t *testing.T) {
	storer := mockstorer.New()
	beeUrl := mock.NewTestBeeServer(t, mock.TestServerOptions{
		Storer:          storer,
		PreventRedirect: true,
		Post:            mockpost.New(mockpost.WithAcceptAll()),
	})

	logger := logging.New(io.Discard, logrus.DebugLevel)
	mockClient := bee.NewBeeClient(beeUrl, bee.WithStamp(mock.BatchOkStr), bee.WithRedundancy(fmt.Sprintf("%d", redundancy.NONE)), bee.WithPinning(true))

	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	pod1AccountInfo, err := acc.CreatePodAccount(1, false)
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(pod1AccountInfo, mockClient, -1, 0, logger)
	user := acc.GetAddress(1)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	mockFile := file.NewFile("pod1", mockClient, fd, user, tm, logger)

	t.Run("chmod-dir", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		dirObject := dir.NewDirectory("pod1", mockClient, fd, user, mockFile, tm, logger)
		// make root dir so that other directories can be added
		err = dirObject.MkRootDir("pod1", podPassword, user, fd)
		if err != nil {
			t.Fatal(err)
		}

		// populate the directory with few directory and files
		err := dirObject.MkDir("/dirToChmod", podPassword, 0)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/dirToChmod/subDir1", podPassword, 0)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/dirToChmod/subDir2", podPassword, 0)
		if err != nil {
			t.Fatal(err)
		}
		// just add dummy file entry as file listing is not tested here
		err = dirObject.AddEntryToDir("/dirToChmod", podPassword, "file1", true)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.AddEntryToDir("/dirToChmod", podPassword, "file2", true)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(3)

		// stat the directory
		dirStats, err := dirObject.DirStat("pod1", podPassword, "/dirToChmod")
		if err != nil {
			t.Fatal(err)
		}

		if fmt.Sprintf("%o", dir.S_IFDIR|0700) != fmt.Sprintf("%o", dirStats.Mode) {
			t.Fatal("default mode mismatch")
		}
		fmt.Println(4)

		err = dirObject.Chmod("/dirToChmod", podPassword, 0664)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(5)

		dirStats, err = dirObject.DirStat("pod1", podPassword, "/dirToChmod")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(6)

		if fmt.Sprintf("%o", dir.S_IFDIR|0664) != fmt.Sprintf("%o", dirStats.Mode) {
			t.Fatal("updated mode mismatch")
		}
	})
}
