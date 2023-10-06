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
	"context"
	"errors"
	"io"
	"sort"
	"testing"
	"time"

	mockpost "github.com/ethersphere/bee/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/sirupsen/logrus"

	"github.com/plexsysio/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	fm "github.com/fairdatasociety/fairOS-dfs/pkg/file/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

func TestListDirectory(t *testing.T) {
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
	mockFile := fm.NewMockFile()
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()

	t.Run("list-dir", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		dirObject := dir.NewDirectory("pod1", mockClient, fd, user, mockFile, tm, logger)

		// make root dir so that other directories can be added
		err = dirObject.MkRootDir("pod1", podPassword, user, fd)
		if err != nil {
			t.Fatal(err)
		}
		err := dirObject.MkDir("/", podPassword, 0)
		if !errors.Is(err, dir.ErrInvalidDirectoryName) {
			t.Fatal("invalid dir name", err)
		}
		longDirName, err := utils.GetRandString(101)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/"+longDirName, podPassword, 0)
		if !errors.Is(err, dir.ErrTooLongDirectoryName) {
			t.Fatal("dir name too long")
		}

		// create some dir and files
		err = dirObject.MkDir("/parentDir", podPassword, 0)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/parentDir", podPassword, 0)
		if !errors.Is(err, dir.ErrDirectoryAlreadyPresent) {
			t.Fatal("dir already present")
		}
		// populate the directory with few directory and files
		err = dirObject.MkDir("/parentDir/subDir1", podPassword, 0)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/parentDir/subDir2", podPassword, 0)
		if err != nil {
			t.Fatal(err)
		}

		err = dirObject.AddEntryToDir("", podPassword, "file1", true)
		if !errors.Is(err, dir.ErrInvalidDirectoryName) {
			t.Fatal("invalid dir name")
		}
		err = dirObject.AddEntryToDir("/parentDir", podPassword, "", true)
		if !errors.Is(err, dir.ErrInvalidFileOrDirectoryName) {
			t.Fatal("invalid file or dir name")
		}
		err = dirObject.AddEntryToDir("/parentDir-not-available", podPassword, "file1", true)
		if !errors.Is(err, dir.ErrDirectoryNotPresent) {
			t.Fatal("parent not available")
		}

		// just add dummy file entry as file listing is not tested here
		err = dirObject.AddEntryToDir("/parentDir", podPassword, "file1", true)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.AddEntryToDir("/parentDir", podPassword, "file2", true)
		if err != nil {
			t.Fatal(err)
		}

		// validate dir listing
		dirEntries, files, err := dirObject.ListDir("/parentDir", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		var dirs []string

		for _, v := range dirEntries {
			dirs = append(dirs, v.Name)
		}

		if len(dirs) != 2 {
			t.Fatalf("invalid directory entry count")
		}

		sort.Strings(dirs)
		sort.Strings(files)
		// validate entry names
		if dirs[0] != "subDir1" {
			t.Fatalf("invalid directory name")
		}
		if dirs[1] != "subDir2" {
			t.Fatalf("invalid directory name")
		}
		if files[0] != "/parentDir/file1" {
			t.Fatalf("invalid file name")
		}
		if files[1] != "/parentDir/file2" {
			t.Fatalf("invalid file name")
		}
	})

	t.Run("list-dir-from-different-dir-object", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		dirObject := dir.NewDirectory("pod1", mockClient, fd, user, mockFile, tm, logger)

		// make root dir so that other directories can be added
		err = dirObject.MkRootDir("pod1", podPassword, user, fd)
		if err != nil {
			t.Fatal(err)
		}

		// create dir
		err = dirObject.MkDir("/parentDir", podPassword, 0)
		if err != nil {
			t.Fatal(err)
		}
		// populate the directory with few directory and files
		err = dirObject.MkDir("/parentDir/subDir1", podPassword, 0)
		if err != nil {
			t.Fatal(err)
		}
		dirObject2 := dir.NewDirectory("pod1", mockClient, fd, user, mockFile, tm, logger)
		err = dirObject2.AddRootDir("pod1", podPassword, user, fd)
		if err != nil {
			t.Fatal(err)
		}
		// validate dir listing
		dirs, _, err := dirObject2.ListDir("/parentDir", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		if len(dirs) != 1 {
			t.Fatalf("invalid directory entry count")
		}

		// validate entry names
		if dirs[0].Name != "subDir1" {
			t.Fatalf("invalid directory name")
		}
	})
}
