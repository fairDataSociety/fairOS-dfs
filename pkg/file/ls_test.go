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
	"context"
	"io"
	"strconv"
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
)

func TestListFiles(t *testing.T) {
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

	t.Run("list-file", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)

		// upload few files
		_, err = uploadFile(t, fileObject, "/dir1", "file1", "", podPassword, 100, 10)
		if err != nil {
			t.Fatal(err)
		}

		_, err = uploadFile(t, fileObject, "/dir1", "file2", "", podPassword, 200, 20)
		if err != nil {
			t.Fatal(err)
		}

		_, err = uploadFile(t, fileObject, "/dir1", "file3", "", podPassword, 300, 30)
		if err != nil {
			t.Fatal(err)
		}

		// list the files
		fileList := []string{"/dir1/file1", "/dir1/file2", "/dir1/file3"}
		entries, err := fileObject.ListFiles(fileList, podPassword)
		if err != nil {
			t.Fatal(err)
		}

		foundIndex1 := -1
		for i, v := range entries {
			if v.Name == "file1" {
				foundIndex1 = i
			}
		}

		if foundIndex1 < 0 {
			t.Fatal("file1 not found")
		}
		foundIndex2 := -1
		for i, v := range entries {
			if v.Name == "file2" {
				foundIndex2 = i
			}
		}
		if foundIndex2 < 0 {
			t.Fatal("file1 not found")
		}
		foundIndex3 := -1
		for i, v := range entries {
			if v.Name == "file3" {
				foundIndex3 = i
			}
		}
		if foundIndex3 < 0 {
			t.Fatal("file1 not found")
		}

		// validate the entries
		entry := entries[foundIndex1]
		if entry.Name != "file1" {
			t.Fatalf("invalid name")
		}
		if entry.Size != strconv.FormatUint(100, 10) {
			t.Fatalf("invalid file size")
		}
		if entry.BlockSize != strconv.FormatUint(10, 10) {
			t.Fatalf("invalid block size")
		}
		entry = entries[foundIndex2]
		if entry.Name != "file2" {
			t.Fatalf("invalid name")
		}
		if entry.Size != strconv.FormatUint(200, 10) {
			t.Fatalf("invalid file size")
		}
		if entry.BlockSize != strconv.FormatUint(20, 10) {
			t.Fatalf("invalid block size")
		}
		entry = entries[foundIndex3]
		if entry.Name != "file3" {
			t.Fatalf("invalid name")
		}
		if entry.Size != strconv.FormatUint(300, 10) {
			t.Fatalf("invalid file size")
		}
		if entry.BlockSize != strconv.FormatUint(30, 10) {
			t.Fatalf("invalid block size")
		}
	})
}
