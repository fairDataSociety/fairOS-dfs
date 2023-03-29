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
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownload(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("")
	require.NoError(t, err)

	pod1AccountInfo, err := acc.CreatePodAccount(1, false)
	require.NoError(t, err)

	fd := feed.New(pod1AccountInfo, mockClient, logger)
	user := acc.GetAddress(1)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()

	t.Run("download-small-file", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)

		filePath := "/dir1"
		fileName := "file1"
		compression := ""
		fileSize := int64(100)
		blockSize := uint32(10)
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)

		// file existent check
		podFile := utils.CombinePathAndFile(filePath, fileName)
		assert.Equal(t, fileObject.IsFileAlreadyPresent(podPassword, podFile), false)

		_, _, err = fileObject.Download(podFile, podPassword)
		assert.Equal(t, err, file.ErrFileNotFound)

		// upload a file
		content, err := uploadFile(t, fileObject, filePath, fileName, compression, podPassword, fileSize, blockSize)
		require.NoError(t, err)

		// Download the file and read from reader
		reader, _, err := fileObject.Download(podFile, podPassword)
		require.NoError(t, err)

		rcvdBuffer := new(bytes.Buffer)
		_, err = rcvdBuffer.ReadFrom(reader)
		require.NoError(t, err)

		// Download the file and read from reader
		reader2, rcvdSize2, err := fileObject.Download(podFile, podPassword)
		require.NoError(t, err)

		rcvdBuffer2 := new(bytes.Buffer)
		_, err = rcvdBuffer2.ReadFrom(reader2)
		require.NoError(t, err)

		// validate the result
		if len(rcvdBuffer2.Bytes()) != len(content) || int(rcvdSize2) != len(content) {
			t.Fatalf("downloaded content size is invalid")
		}
		assert.Equal(t, content, rcvdBuffer2.Bytes())

	})

	t.Run("download-small-file-gzip", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		filePath := "/dir1"
		fileName := "file1"
		compression := "gzip"
		fileSize := int64(100)
		blockSize := uint32(164000)
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)

		// file existent check
		podFile := utils.CombinePathAndFile(filePath, fileName)
		assert.Equal(t, fileObject.IsFileAlreadyPresent(podPassword, podFile), false)

		_, _, err = fileObject.Download(podFile, podPassword)
		assert.Equal(t, err, file.ErrFileNotFound)

		// upload a file
		content, err := uploadFile(t, fileObject, filePath, fileName, compression, podPassword, fileSize, blockSize)
		require.NoError(t, err)

		// Download the file and read from reader
		reader, rcvdSize, err := fileObject.Download(podFile, podPassword)
		require.NoError(t, err)

		rcvdBuffer := new(bytes.Buffer)
		_, err = rcvdBuffer.ReadFrom(reader)
		require.NoError(t, err)

		// validate the result
		if len(rcvdBuffer.Bytes()) != len(content) || int(rcvdSize) != len(content) {
			t.Fatalf("downloaded content size is invalid")
		}
		assert.Equal(t, content, rcvdBuffer.Bytes())
	})

	t.Run("read-seeker-small", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)

		filePath := "/dir1"
		fileName := "file1"
		compression := ""
		fileSize := int64(100)
		blockSize := uint32(10)
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)

		// file existent check
		podFile := utils.CombinePathAndFile(filePath, fileName)
		assert.Equal(t, fileObject.IsFileAlreadyPresent(podPassword, podFile), false)

		_, _, err = fileObject.Download(podFile, podPassword)
		assert.Equal(t, err, file.ErrFileNotFound)

		// upload a file
		content, err := uploadFile(t, fileObject, filePath, fileName, compression, podPassword, fileSize, blockSize)
		require.NoError(t, err)

		reader, size, err := fileObject.ReadSeeker(podFile, podPassword)
		require.NoError(t, err)

		point := size / 2
		half := content[point:]

		n, err := reader.Seek(int64(point), 0)
		require.NoError(t, err)

		assert.Equal(t, fmt.Sprintf("%d", n), fmt.Sprintf("%d", point))

		rcvdBuffer := new(bytes.Buffer)
		_, err = rcvdBuffer.ReadFrom(reader)
		require.NoError(t, err)

		assert.Equal(t, half, rcvdBuffer.Bytes())
	})
}
