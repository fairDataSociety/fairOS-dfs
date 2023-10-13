package file_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	mockpost "github.com/ethersphere/bee/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
	"github.com/sirupsen/logrus"
)

func TestStatus(t *testing.T) {
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

	podPassword, _ := utils.GetRandString(pod.PasswordLength)
	t.Run("sync-status-file", func(t *testing.T) {
		t.Skip()
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)
		numberOfChunks := int64(10)
		// upload a file
		_, err = uploadFile(t, fileObject, "/dir1", "file1", "", podPassword, 3900*numberOfChunks, file.MinBlockSize)
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
		require.Equal(t, total, numberOfChunks)
	})

}
