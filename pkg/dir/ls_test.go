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

package dir_test

import (
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	bm "github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	fm "github.com/fairdatasociety/fairOS-dfs/pkg/file/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"io/ioutil"
	"testing"
)

func TestListDirectory(t *testing.T) {
	mockClient := bm.NewMockBeeClient()
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
	mockFile := fm.NewMockFile()

	t.Run("list-dirr", func(t *testing.T) {
		dirObject := dir.NewDirectory("pod1", mockClient, fd, user, mockFile, logger)

		// create some dir and files
		err := dirObject.MkDir("/", "parentDir")
		if err != nil {
			t.Fatal(err)
		}
		// populate the directory with few directory and files
		err = dirObject.MkDir("/parentDir", "subDir1")
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/parentDir", "subDir2")
		if err != nil {
			t.Fatal(err)
		}
		// just add dummy file enty as file listing is not tested here
		err = dirObject.AddEntryToDir("/parentDir", "file1", true)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.AddEntryToDir("/parentDir", "file2", true)
		if err != nil {
			t.Fatal(err)
		}


		// validate dir listing
		dirs, files, err := dirObject.ListDir("/parentDir")
		if err != nil {
			t.Fatal(err)
		}
		if len(dirs) != 2 {
			t.Fatalf("invalid directory entry count")
		}

		// validate entry names
		if dirs[0].Name != "subDir1" {
			t.Fatalf("invalid directory name")
		}
		if dirs[1].Name != "subDir2" {
			t.Fatalf("invalid directory name")
		}
		if files[0] != "/parentDir/file1" {
			t.Fatalf("invalid file name")
		}
		if files[1] != "/parentDir/file2" {
			t.Fatalf("invalid file name")
		}
	})
}