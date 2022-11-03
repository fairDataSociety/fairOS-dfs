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
	"io"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"

	"github.com/plexsysio/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func TestDownload(t *testing.T) {
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
	t.Run("download-small-file", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PodPasswordLength)

		filePath := "/dir1"
		fileName := "file1"
		compression := ""
		fileSize := int64(100)
		blockSize := uint32(10)
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)

		// file existent check
		podFile := utils.CombinePathAndFile(filePath, fileName)
		if fileObject.IsFileAlreadyPresent(podFile) {
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

		// Download the file and read from reader
		reader, rcvdSize, err := fileObject.Download(podFile, podPassword)
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

	t.Run("download-small-file-gzip", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PodPasswordLength)
		filePath := "/dir1"
		fileName := "file1"
		compression := "gzip"
		fileSize := int64(100)
		blockSize := uint32(164000)
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)

		// file existent check
		podFile := utils.CombinePathAndFile(filePath, fileName)
		if fileObject.IsFileAlreadyPresent(podFile) {
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

		// Download the file and read from reader
		reader, rcvdSize, err := fileObject.Download(podFile, podPassword)
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
