package dir_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/file"

	mockpost "github.com/ethersphere/bee/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
	"github.com/sirupsen/logrus"
)

func TestDirRmAllFromMap(t *testing.T) {
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

	t.Run("dir-rm-all-from-map", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		dirObject := dir.NewDirectory("pod1", mockClient, fd, user, mockFile, tm, logger)

		// make root dir so that other directories can be added
		err = dirObject.MkRootDir("pod1", podPassword, user, fd)
		if err != nil {
			t.Fatal(err)
		}

		// create a new dir
		err := dirObject.MkDir("/baseDir", podPassword, 0)
		if err != nil {
			t.Fatal(err)
		}

		// check if dir is present
		present := dirObject.IsDirectoryPresent("/baseDir", podPassword)
		if !present {
			t.Fatalf("directory is not present")
		}

		dirObject.RemoveAllFromDirectoryMap()
		node, _ := dirObject.GetInode(podPassword, "/baseDir")
		if node == nil {
			t.Fatal("node should not be nil, metadata should be available in blockstore")
		}

		err = dirObject.RmDir("/baseDir", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		node, _ = dirObject.GetInode(podPassword, "/baseDir")
		if node != nil {
			t.Fatal("node should be  nil")
		}
	})
}
