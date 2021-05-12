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
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	bm "github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	fm "github.com/fairdatasociety/fairOS-dfs/pkg/file/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

func TestStat(t *testing.T) {
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

	t.Run("stat-dir", func(t *testing.T) {
		dirObject := dir.NewDirectory("pod1", mockClient, fd, user, mockFile, logger)

		// make root dir so that other directories can be added
		err = dirObject.MkRootDir()
		if err != nil {
			t.Fatal(err)
		}

		// populate the directory with few directory and files
		err := dirObject.MkDir("/", "dirToStat")
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/dirToStat", "subDir1")
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/dirToStat", "subDir2")
		if err != nil {
			t.Fatal(err)
		}
		// just add dummy file enty as file listing is not tested here
		err = dirObject.AddEntryToDir("/dirToStat", "file1", true)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.AddEntryToDir("/dirToStat", "file2", true)
		if err != nil {
			t.Fatal(err)
		}

		// stat the directory
		dirStats, err := dirObject.DirStat("pod1", "/dirToStat")
		if err != nil {
			t.Fatal(err)
		}

		// verify if the stats are okay
		if dirStats == nil {
			t.Fatalf("invalid stats")
		}
		if dirStats.PodName != "pod1" {
			t.Fatalf("invalid pod name")
		}
		if dirStats.DirPath != "/" {
			t.Fatalf("invalid directory path")
		}
		if dirStats.DirName != "dirToStat" {
			t.Fatalf("invalid directory name")
		}
		if dirStats.NoOfDirectories != strconv.FormatUint(2, 10) {
			t.Fatalf("invalid directory count")
		}
		if dirStats.NoOfFiles != strconv.FormatUint(2, 10) {
			t.Fatalf("invalid files count")
		}
	})
}
