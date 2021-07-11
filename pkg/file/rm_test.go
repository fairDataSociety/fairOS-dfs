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
	"io/ioutil"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

func TestRemoveFile(t *testing.T) {
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

	t.Run("delete-file", func(t *testing.T) {
		fileObject := file.NewFile("pod1", mockClient, fd, user, logger)

		// upload few files
		_, err = uploadFile(t, fileObject, "/dir1", "file1", "", 100, 10)
		if err != nil {
			t.Fatal(err)
		}

		_, err = uploadFile(t, fileObject, "/dir1", "file2", "", 200, 20)
		if err != nil {
			t.Fatal(err)
		}

		// remove file2
		err = fileObject.RmFile("/dir1/file2")
		if err != nil {
			t.Fatal(err)
		}

		// validate file deletion
		meta := fileObject.GetFromFileMap(utils.CombinePathAndFile("pod1", "/dir1", "file2"))
		if meta != nil {
			t.Fatalf("file is not removed")
		}

		// check if other file is present
		meta = fileObject.GetFromFileMap(utils.CombinePathAndFile("pod1", "/dir1", "file1"))
		if meta == nil {
			t.Fatalf("file is not present")
		}
		if meta.Name != "file1" {
			t.Fatalf("retrieved invalid file name")
		}
	})
}
