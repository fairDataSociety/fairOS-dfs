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

package file

import (
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"io/ioutil"
	"strconv"
	"testing"
)

func TestStat(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(ioutil.Discard, 0)
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

	t.Run("stat-file", func(t *testing.T) {
		fileObject := NewFile("pod1", mockClient, fd, user, logger)

		// upload a file
		_, err = uploadFile(t, fileObject, "/dir1", "file1", "", 100, 10)
		if err != nil { t.Fatal(err) }

		// stat the file
		stats, err := fileObject.GetStats("pod1", "/dir1/file1")
		if err != nil { t.Fatal(err) }

		// validate state
		if stats.PodName != "pod1" {
			t.Fatalf("invalid pod name in stats")
		}
		if stats.FilePath != "/dir1" {
			t.Fatalf("invalid file path in stats")
		}
		if stats.FileName != "file1" {
			t.Fatalf("invalid file name in stats")
		}
		if stats.FileSize != strconv.FormatUint(100, 10) {
			t.Fatalf("invalid file size in stats")
		}
		if stats.BlockSize != strconv.FormatUint(10, 10) {
			t.Fatalf("invalid block size in stats")
		}
	})

}