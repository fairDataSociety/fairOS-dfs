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
	"sync"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	bm "github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	fm "github.com/fairdatasociety/fairOS-dfs/pkg/file/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestSync(t *testing.T) {
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
	fd := feed.New(pod1AccountInfo, mockClient, logger)
	user := acc.GetAddress(1)
	mockFile := fm.NewMockFile()
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()

	t.Run("sync-dir", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PodPasswordLength)
		dirObject := dir.NewDirectory("pod1", mockClient, fd, user, mockFile, tm, logger)

		// make root dir so that other directories can be added
		err = dirObject.MkRootDir("pod1", podPassword, user, fd)
		if err != nil {
			t.Fatal(err)
		}

		// populate the directory with few directory and files
		err := dirObject.MkDir("/dirToStat", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/dirToStat/subDir1", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.MkDir("/dirToStat/subDir2", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		// just add dummy file entry as file listing is not tested here
		err = dirObject.AddEntryToDir("/dirToStat", podPassword, "file1", true)
		if err != nil {
			t.Fatal(err)
		}
		err = dirObject.AddEntryToDir("/dirToStat", podPassword, "file2", true)
		if err != nil {
			t.Fatal(err)
		}
		dirObject2 := dir.NewDirectory("pod1", mockClient, fd, user, mockFile, tm, logger)
		if dirObject2.GetDirFromDirectoryMap("/") != nil {
			t.Fatal("it should be nil before sync")
		}
		wg := new(sync.WaitGroup)
		err = dirObject2.SyncDirectoryAsync(context.Background(), "/", podPassword, wg)
		if err != nil {
			t.Fatal(err)
		}
		wg.Wait()
		node := dirObject2.GetDirFromDirectoryMap("/")
		if node.GetDirInodePathAndNameForRoot() != utils.PathSeparator {
			t.Fatal("node is root node")
		}
		if node.GetDirInodePathAndName() != utils.PathSeparator {
			t.Fatal("node is root node")
		}
		if !node.IsDirInodeRoot() {
			t.Fatal("node is root node")
		}

		node2 := dirObject2.GetDirFromDirectoryMap("/dirToStat")
		if node2.GetDirInodePathAndName() != "/dirToStat" {
			t.Fatal("node2 is /dirToStat")
		}
		if node2.GetDirInodePathAndNameForRoot() == utils.PathSeparator {
			t.Fatal("node2 is not root node")
		}
		if node2.IsDirInodeRoot() {
			t.Fatal("node2 is not root node")
		}
		if node2.GetDirInodePathOnly() != "/" {
			t.Fatal("node2 path is not \"/\"")
		}
	})
}
