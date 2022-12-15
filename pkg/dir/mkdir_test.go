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
	"io"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	"github.com/plexsysio/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	bm "github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	fm "github.com/fairdatasociety/fairOS-dfs/pkg/file/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

func TestMkdir(t *testing.T) {
	mockClient := bm.NewMockBeeClient()
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
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	fd := feed.New(pod1AccountInfo, mockClient, logger)
	user := acc.GetAddress(1)
	mockFile := fm.NewMockFile()

	t.Run("simple-mkdir", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		dirObject := dir.NewDirectory("pod1", mockClient, fd, user, mockFile, tm, logger)

		// make root dir so that other directories can be added
		err = dirObject.MkRootDir("pod1", podPassword, user, fd)
		if err != nil {
			t.Fatal(err)
		}

		// create a new dir
		err := dirObject.MkDir("/baseDir", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		// validate dir
		dirs, _, err := dirObject.ListDir("/", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		if len(dirs) != 1 {
			t.Fatalf("invalid directory count")
		}
		if dirs[0].Name != "baseDir" {
			t.Fatalf("invalid directory name")
		}
	})

	t.Run("complicated-mkdir", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		dirObject := dir.NewDirectory("pod1", mockClient, fd, user, mockFile, tm, logger)

		// make root dir so that other directories can be added
		err = dirObject.MkRootDir("pod1", podPassword, user, fd)
		if err != nil {
			t.Fatal(err)
		}

		// try to create a new dir without creating root
		err := dirObject.MkDir("/baseDir/baseDir2/baseDir3/baseDir4", podPassword)
		if err == nil || err != dir.ErrDirectoryNotPresent {
			t.Fatal(err)
		}

		err = dirObject.MkDir("/baseDir", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		err = dirObject.MkDir("/baseDir/baseDir2", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		err = dirObject.MkDir("/baseDir/baseDir2/baseDir3", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		// validate dir
		dirs, _, err := dirObject.ListDir("/baseDir", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		if len(dirs) != 1 {
			t.Fatalf("invalid directory count")
		}
		if dirs[0].Name != "baseDir2" {
			t.Fatalf("invalid directory name")
		}

		dirs, _, err = dirObject.ListDir("/baseDir/baseDir2", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		if len(dirs) != 1 {
			t.Fatalf("invalid directory count")
		}
		if dirs[0].Name != "baseDir3" {
			t.Fatalf("invalid directory name")
		}
	})
}
