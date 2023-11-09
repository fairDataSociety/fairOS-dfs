package test_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"

	"github.com/fairdatasociety/fairOS-dfs/pkg/file"

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

func TestMaxFiles(t *testing.T) {
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
	fd := feed.New(pod1AccountInfo, mockClient, 500, 0, logger)
	user := acc.GetAddress(1)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	sm := mock3.NewMockSubscriptionManager()
	_, _, err = acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}

	pod1 := pod.NewPod(mockClient, fd, acc, tm, sm, 500, 0, logger)
	podPassword, _ := utils.GetRandString(pod.PasswordLength)

	_, err = pod1.CreatePod("test", "", podPassword)
	if err != nil {
		t.Fatal(err)
	}
	fileObject := file.NewFile("test", mockClient, fd, user, tm, logger)
	dirObject := dir.NewDirectory("test", mockClient, fd, user, fileObject, tm, logger)
	err = dirObject.MkRootDir("test", podPassword, user, fd)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("create-max-files", func(t *testing.T) {
		maxfiles := 1000000
		filePath := "/"
		for i := 1; i <= maxfiles; i++ {
			fileName, _ := utils.GetRandString(100)
			compression := ""
			fileSize := int64(1000)
			blockSize := file.MinBlockSize
			_, err = uploadFile(t, fileObject, filePath, fileName, compression, podPassword, fileSize, blockSize)
			if err != nil {
				t.Fatal(err)
			}
			err = dirObject.AddEntryToDir("/", podPassword, fileName, true)
			if err != nil {
				t.Fatal(i, err)
			}
		}
	})
}
