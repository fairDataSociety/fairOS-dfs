package file_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
)

func TestRename(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("password", "")
	if err != nil {
		t.Fatal(err)
	}
	pod1AccountInfo, err := acc.CreatePodAccount(1, "password", false)
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(pod1AccountInfo, mockClient, logger)
	user := acc.GetAddress(1)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	t.Run("upload-rename-download-small-file", func(t *testing.T) {
		filePath := "/dir1"
		fileName := "file1"
		newFileName := "file_new"
		compression := ""
		fileSize := int64(100)
		blockSize := uint32(10)
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)

		// file existent check
		podFile := utils.CombinePathAndFile(filePath, fileName)
		if fileObject.IsFileAlreadyPresent(podFile) {
			t.Fatal("file should not be present")
		}
		_, _, err = fileObject.Download(podFile)
		if err == nil {
			t.Fatal("file should not be present for download")
		}
		// upload a file
		content, err := uploadFile(t, fileObject, filePath, fileName, compression, fileSize, blockSize)
		if err != nil {
			t.Fatal(err)
		}

		_, err = fileObject.RenameFromFileName(podFile, newFileName)
		if err != nil {
			t.Fatal(err)
		}

		//Download the file and read from reader
		present := fileObject.IsFileAlreadyPresent(podFile)
		if present {
			t.Fatal("old name should not be present")
		}

		// Download the file and read from reader
		reader, rcvdSize, err := fileObject.Download(utils.CombinePathAndFile(filePath, newFileName))
		if err != nil {
			t.Fatal(err)
		}
		rcvdBuffer := new(bytes.Buffer)
		_, err = rcvdBuffer.ReadFrom(reader)
		if err != nil {
			t.Fatal(err)
		}

		// validate the result
		if len(rcvdBuffer.Bytes()) != len(content) || int(rcvdSize) != len(content) {
			t.Fatalf("downloaded content size is invalid")
		}
		if !bytes.Equal(content, rcvdBuffer.Bytes()) {
			t.Fatalf("downloaded content is not equal")
		}

	})
}
