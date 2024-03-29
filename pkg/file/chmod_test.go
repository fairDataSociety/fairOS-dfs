package file_test

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	mockpost "github.com/ethersphere/bee/v2/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/v2/pkg/storer/mock"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChmod(t *testing.T) {
	storer := mockstorer.New()
	beeUrl := mock.NewTestBeeServer(t, mock.TestServerOptions{
		Storer:          storer,
		PreventRedirect: true,
		Post:            mockpost.New(mockpost.WithAcceptAll()),
	})

	logger := logging.New(io.Discard, logrus.DebugLevel)
	mockClient := bee.NewBeeClient(beeUrl, mock.BatchOkStr, true, 0, logger)
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

	podPassword, _ := utils.GetRandString(pod.PasswordLength)
	t.Run("chmod-file", func(t *testing.T) {
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)

		// upload a file
		_, err = uploadFile(t, fileObject, "/dir1", "file1", "", podPassword, 100, file.MinBlockSize)
		require.NoError(t, err)

		// stat the file
		stats, err := fileObject.GetStats("pod1", "/dir1/file1", podPassword)
		require.NoError(t, err)

		assert.Equal(t, fmt.Sprintf("%o", file.S_IFREG|0600), fmt.Sprintf("%o", stats.Mode))

		err = fileObject.Chmod("/dir1/file2", podPassword, 0777)
		assert.Equal(t, err, file.ErrFileNotFound)

		err = fileObject.Chmod("/dir1/file1", podPassword, 0777)
		require.NoError(t, err)

		stats, err = fileObject.GetStats("pod1", "/dir1/file1", podPassword)
		require.NoError(t, err)

		assert.Equal(t, fmt.Sprintf("%o", file.S_IFREG|0777), fmt.Sprintf("%o", stats.Mode))
	})
}
