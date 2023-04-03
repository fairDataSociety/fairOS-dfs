package file_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
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

	t.Run("upload-rename-same-dir-download-small-file", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		filePath := "/dir1"
		fileName := "file1"
		newFileName := "file_new"
		compression := ""
		fileSize := int64(100)
		blockSize := file.MinBlockSize
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)

		// file existent check
		podFile := utils.CombinePathAndFile(filePath, fileName)
		if fileObject.IsFileAlreadyPresent(podPassword, podFile) {
			t.Fatal("file should not be present")
		}
		_, _, err = fileObject.Download(podFile, podPassword)
		if err == nil {
			t.Fatal("file should not be present for download")
		}
		// upload a file
		content, err := uploadFile(t, fileObject, filePath, fileName, compression, podPassword, fileSize, blockSize)
		if err != nil {
			t.Fatal(err)
		}

		newPodFile := utils.CombinePathAndFile(filePath, newFileName)
		_, err = fileObject.RenameFromFileName(podFile, newPodFile, podPassword)
		if err != nil {
			t.Fatal(err)
		}

		// Download the file and read from reader
		present := fileObject.IsFileAlreadyPresent(podPassword, podFile)
		if present {
			t.Fatal("old name should not be present")
		}

		present = fileObject.IsFileAlreadyPresent(podPassword, newPodFile)
		if !present {
			t.Fatal("new name should be present")
		}

		// Download the file and read from reader
		reader, rcvdSize, err := fileObject.Download(utils.CombinePathAndFile(filePath, newFileName), podPassword)
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

	t.Run("upload-rename-diff-dir-download-small-file", func(t *testing.T) {
		filePath := "/dir1"
		newFilePath := "/dir2"
		fileName := "file1"
		compression := ""
		fileSize := int64(100)
		blockSize := file.MinBlockSize
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)
		dirObject := dir.NewDirectory("pod1", mockClient, fd, user, fileObject, tm, logger)
		podPassword, _ := utils.GetRandString(pod.PasswordLength)

		// make root dir so that other directories can be added
		err = dirObject.MkRootDir("pod1", podPassword, user, fd)
		if err != nil {
			t.Fatal(err)
		}

		// populate the directory with few directory and files
		err = dirObject.MkDir(filePath, podPassword, 0)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir(newFilePath, podPassword, 0)
		if err != nil {
			t.Fatal(err)
		}

		// file existent check
		podFile := utils.CombinePathAndFile(filePath, fileName)
		if fileObject.IsFileAlreadyPresent(podPassword, podFile) {
			t.Fatal("file should not be present")
		}

		// upload a file
		content, err := uploadFile(t, fileObject, filePath, fileName, compression, podPassword, fileSize, blockSize)
		if err != nil {
			t.Fatal(err)
		}
		newPodFile := utils.CombinePathAndFile(newFilePath, fileName)
		if fileObject.IsFileAlreadyPresent(podPassword, newPodFile) {
			t.Fatal("file should not be present")
		}
		_, err = fileObject.RenameFromFileName(podFile, newPodFile, podPassword)
		if err != nil {
			t.Fatal(err)
		}

		// Download the file and read from reader
		present := fileObject.IsFileAlreadyPresent(podPassword, podFile)
		if present {
			t.Fatal("old name should not be present")
		}

		present = fileObject.IsFileAlreadyPresent(podPassword, newPodFile)
		if !present {
			t.Fatal("new name should be present")
		}
		// Download the file and read from reader
		reader, rcvdSize, err := fileObject.Download(newPodFile, podPassword)
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
