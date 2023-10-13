/*
Copyright © 2020 FairOS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package file_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

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

func TestUpload(t *testing.T) {
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

	t.Run("upload-small-file", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)

		filePath := "/dir1"
		fileName := "file1"
		compression := ""
		fileSize := int64(100)
		blockSize := file.MinBlockSize
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)
		_, err = uploadFile(t, fileObject, filePath, fileName, compression, podPassword, fileSize, blockSize)
		if err != nil {
			t.Fatal(err)
		}

		// check for meta
		meta := fileObject.GetInode(podPassword, utils.CombinePathAndFile(filePath, fileName))
		if meta == nil {
			t.Fatalf("file not added in file map")
		}

		// validate meta items
		if meta.Path != filePath {
			t.Fatalf("invalid path in meta")
		}
		if meta.Name != fileName {
			t.Fatalf("invalid file name in meta")
		}
		if meta.Size != uint64(fileSize) {
			t.Fatalf("invalid file size in meta")
		}
		if meta.BlockSize != blockSize {
			t.Fatalf("invalid block size in meta")
		}

		err := fileObject.LoadFileMeta(filePath+"/"+fileName, podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = fileObject.LoadFileMeta(filePath+"/asd"+fileName, podPassword)
		if err == nil {
			t.Fatal("local file meta should fail")
		}

		meat2, err := fileObject.BackupFromFileName(filePath+"/"+fileName, podPassword)
		if err != nil {
			t.Fatal(err)
		}
		if meta.Name == meat2.Name {
			t.Fatal("name should not be same after backup")
		}
	})

	t.Run("upload-small-file-at-root", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)

		filePath := string(os.PathSeparator)
		fileName := "file1"
		compression := ""
		fileSize := int64(100)
		blockSize := file.MinBlockSize
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)
		_, err = uploadFile(t, fileObject, filePath, fileName, compression, podPassword, fileSize, blockSize)
		if err != nil {
			t.Fatal(err)
		}

		// check for meta
		meta := fileObject.GetInode(podPassword, utils.CombinePathAndFile(filepath.ToSlash(filePath), fileName))
		if meta == nil {
			t.Fatalf("file not added in file map")
		}

		// validate meta items
		if meta.Path != filepath.ToSlash(filePath) {
			t.Fatalf("invalid path in meta")
		}
		if meta.Name != fileName {
			t.Fatalf("invalid file name in meta")
		}
		if meta.Size != uint64(fileSize) {
			t.Fatalf("invalid file size in meta")
		}
		if meta.BlockSize != blockSize {
			t.Fatalf("invalid block size in meta")
		}
	})

	t.Run("upload-small-file-at-root-with-blank-filename", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)

		filePath := string(os.PathSeparator)
		fileName := "file1"
		compression := ""
		fileSize := int64(100)
		blockSize := file.MinBlockSize
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)
		_, err = uploadFile(t, fileObject, filePath, fileName, compression, podPassword, fileSize, blockSize)
		if err != nil {
			t.Fatal(err)
		}

		// check for meta
		meta := fileObject.GetInode(podPassword, filepath.ToSlash(utils.CombinePathAndFile(filePath+fileName, "")))
		if meta == nil {
			t.Fatalf("file not added in file map")
		}

		// validate meta items
		if meta.Path != filepath.ToSlash(filePath) {
			t.Fatalf("invalid path in meta")
		}
		if meta.Name != fileName {
			t.Fatalf("invalid file name in meta")
		}
		if meta.Size != uint64(fileSize) {
			t.Fatalf("invalid file size in meta")
		}
		if meta.BlockSize != blockSize {
			t.Fatalf("invalid block size in meta")
		}
	})

	t.Run("upload-small-file-at-root-with-prefix", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		filePath := string(os.PathSeparator)
		fileName, _ := utils.GetRandString(20)
		compression := ""
		fileSize := int64(100)
		blockSize := file.MinBlockSize
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)
		_, err = uploadFile(t, fileObject, filePath, fileName, compression, podPassword, fileSize, blockSize)
		if err != nil {
			t.Fatal(err)
		}

		// check for meta
		meta := fileObject.GetInode(podPassword, utils.CombinePathAndFile(filepath.ToSlash(filePath), filepath.ToSlash(string(os.PathSeparator)+fileName)))
		if meta == nil {
			t.Fatalf("file not added in file map")
		}

		// validate meta items
		if meta.Path != filepath.ToSlash(filePath) {
			t.Fatalf("invalid path in meta")
		}
		if meta.Name != fileName {
			t.Fatalf("invalid file name in meta")
		}
		if meta.Size != uint64(fileSize) {
			t.Fatalf("invalid file size in meta")
		}
		if meta.BlockSize != blockSize {
			t.Fatalf("invalid block size in meta")
		}

		err = fileObject.RmFile(utils.CombinePathAndFile(filepath.ToSlash(filePath), filepath.ToSlash(string(os.PathSeparator)+fileName)), podPassword)
		if err != nil {
			t.Fatal(err)
		}

		meta2 := fileObject.GetInode(podPassword, utils.CombinePathAndFile(filepath.ToSlash(filePath), filepath.ToSlash(string(os.PathSeparator)+fileName)))
		if meta2 != nil {
			t.Fatal("meta2 should be nil")
		}
	})

	t.Run("upload-small-file-at-root-with-prefix-snappy", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		filePath := string(os.PathSeparator)
		fileName, _ := utils.GetRandString(20)
		compression := "snappy"
		fileSize := int64(100)
		blockSize := file.MinBlockSize
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)
		_, err = uploadFile(t, fileObject, filePath, fileName, compression, podPassword, fileSize, blockSize)
		if err != nil {
			t.Fatal(err)
		}

		// check for meta
		meta := fileObject.GetInode(podPassword, utils.CombinePathAndFile(filepath.ToSlash(filePath), filepath.ToSlash(string(os.PathSeparator)+fileName)))
		if meta == nil {
			t.Fatalf("file not added in file map")
		}

		// validate meta items
		if meta.Path != filepath.ToSlash(filePath) {
			t.Fatalf("invalid path in meta")
		}
		if meta.Name != fileName {
			t.Fatalf("invalid file name in meta")
		}
		if meta.Size != uint64(fileSize) {
			t.Fatalf("invalid file size in meta")
		}
		if meta.BlockSize != blockSize {
			t.Fatalf("invalid block size in meta")
		}

		err = fileObject.RmFile(utils.CombinePathAndFile(filepath.ToSlash(filePath), filepath.ToSlash(string(os.PathSeparator)+fileName)), podPassword)
		if err != nil {
			t.Fatal(err)
		}

		meta2 := fileObject.GetInode(podPassword, utils.CombinePathAndFile(filepath.ToSlash(filePath), filepath.ToSlash(string(os.PathSeparator)+fileName)))
		if meta2 != nil {
			t.Fatal("meta2 should be nil")
		}
	})

	t.Run("upload-small-file-at-root-with-prefix-gzip", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		filePath := string(os.PathSeparator)
		fileName, _ := utils.GetRandString(20)
		compression := "gzip"
		fileSize := int64(100)
		blockSize := file.MinBlockSize
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)

		_, err = uploadFile(t, fileObject, filePath, fileName, compression, podPassword, fileSize, uint32(10))
		if !errors.Is(file.ErrInvalidBlockSize, err) {
			t.Fatal("should provide higher block size")
		}

		_, err = uploadFile(t, fileObject, filePath, fileName, compression, podPassword, fileSize, blockSize)
		if err != nil {
			t.Fatal(err)
		}

		// check for meta
		fp := utils.CombinePathAndFile(filepath.ToSlash(filePath), filepath.ToSlash(string(os.PathSeparator)+fileName))
		meta := fileObject.GetInode(podPassword, fp)
		if meta == nil {
			t.Fatalf("file not added in file map")
		}

		// validate meta items
		if meta.Path != filepath.ToSlash(filePath) {
			t.Fatalf("invalid path in meta")
		}
		if meta.Name != fileName {
			t.Fatalf("invalid file name in meta")
		}
		if meta.Size != uint64(fileSize) {
			t.Fatalf("invalid file size in meta")
		}
		if meta.BlockSize != blockSize {
			t.Fatalf("invalid block size in meta")
		}
		reader, _, err := fileObject.Download(fp, podPassword)
		if err != nil {
			t.Fatal(err)
		}
		rcvdBuffer := new(bytes.Buffer)
		_, err = rcvdBuffer.ReadFrom(reader)
		if err != nil {
			t.Fatal(err)
		}
		err = fileObject.RmFile(utils.CombinePathAndFile(filepath.ToSlash(filePath), filepath.ToSlash(string(os.PathSeparator)+fileName)), podPassword)
		if err != nil {
			t.Fatal(err)
		}
		meta2 := fileObject.GetInode(podPassword, fp)
		if meta2 != nil {
			t.Fatal("meta2 should be nil")
		}
	})
}

func uploadFile(t *testing.T, fileObject *file.File, filePath, fileName, compression, podPassword string, fileSize int64, blockSize uint32) ([]byte, error) {
	// create a temp file
	fd, err := os.CreateTemp("", fileName)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fd.Name())

	// write contents to file
	content := make([]byte, fileSize)
	_, err = rand.Read(content)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = fd.Write(content); err != nil {
		t.Fatal(err)
	}

	// close file
	uploadFileName := fd.Name()
	err = fd.Close()
	if err != nil {
		t.Fatal(err)
	}

	// open file to upload
	f1, err := os.Open(uploadFileName)
	if err != nil {
		t.Fatal(err)
	}

	// upload  the temp file
	return content, fileObject.Upload(f1, fileName, fileSize, blockSize, 0, filePath, compression, podPassword)
}
