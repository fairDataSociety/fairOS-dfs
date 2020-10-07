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

package pod

import (
	"io/ioutil"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func TestPod_RemoveDir(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(ioutil.Discard, 0)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("password", "")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, logger)
	pod1 := NewPod(mockClient, fd, acc, logger)

	podName1 := "test1"
	podName2 := "test2"
	podName3 := "test3"
	podName4 := "test4"
	podName5 := "test5"
	firstDir := "dir1"
	secondDir := "dir2"
	thirdAndFourthDir := "dir3/dir4"
	fifthDir := "/dir5"
	t.Run("rmdir-on-root-of-pod", func(t *testing.T) {
		info, err := pod1.CreatePod(podName1, "password")
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}

		err = pod1.MakeDir(podName1, firstDir)
		if err != nil {
			t.Fatalf("error creating directory %s", firstDir)
		}
		dirPath := utils.PathSeperator + podName1 + utils.PathSeperator + firstDir
		dirInode := info.GetDirectory().GetDirFromDirectoryMap(dirPath)
		if dirInode == nil {
			t.Fatalf("directory not created")
		}

		err = pod1.RemoveDir(podName1, firstDir)
		if err != nil {
			t.Fatalf("error removing directory")
		}
		dirPath = utils.PathSeperator + podName1 + utils.PathSeperator + firstDir
		dirInode = info.GetDirectory().GetDirFromDirectoryMap(dirPath)
		if dirInode != nil {
			t.Fatalf("directory not removed")
		}

		// cleanup pod
		err = pod1.DeletePod(podName1)
		if err != nil {
			t.Fatalf("could not delete pod")
		}
	})

	t.Run("rmdir-second-dir-from-first-dir", func(t *testing.T) {
		info, err := pod1.CreatePod(podName2, "password")
		if err != nil {
			t.Fatalf("error creating pod %s", podName2)
		}

		err = pod1.MakeDir(podName2, firstDir)
		if err != nil {
			t.Fatalf("error creating directory %s", firstDir)
		}
		_, err = pod1.ChangeDir(podName2, firstDir)
		if err != nil {
			t.Fatalf("error changing directory")
		}
		err = pod1.MakeDir(podName2, secondDir)
		if err != nil {
			t.Fatalf("error creating directory %s", secondDir)
		}
		dirPath := utils.PathSeperator + podName2 + utils.PathSeperator + firstDir + utils.PathSeperator + secondDir
		dirInode := info.GetDirectory().GetDirFromDirectoryMap(dirPath)
		if dirInode == nil {
			t.Fatalf("directory not created")
		}

		err = pod1.RemoveDir(podName2, secondDir)
		if err != nil {
			t.Fatalf("error removing directory")
		}
		dirPath = utils.PathSeperator + podName2 + utils.PathSeperator + firstDir + utils.PathSeperator + secondDir
		dirInode = info.GetDirectory().GetDirFromDirectoryMap(dirPath)
		if dirInode != nil {
			t.Fatalf("directory not removed")
		}

		// cleanup pod
		err = pod1.DeletePod(podName2)
		if err != nil {
			t.Fatalf("could not delete pod")
		}
	})

	t.Run("rmdir-second-dir-from-pod", func(t *testing.T) {
		info, err := pod1.CreatePod(podName3, "password")
		if err != nil {
			t.Fatalf("error creating pod %s", podName3)
		}

		err = pod1.MakeDir(podName3, firstDir)
		if err != nil {
			t.Fatalf("error creating directory %s", err)
		}
		err = pod1.MakeDir(podName3, firstDir+utils.PathSeperator+secondDir)
		if err != nil {
			t.Fatalf("error creating directory %s", err)
		}
		dirPath := utils.PathSeperator + podName3 + utils.PathSeperator + firstDir + utils.PathSeperator + secondDir
		dirInode := info.GetDirectory().GetDirFromDirectoryMap(dirPath)
		if dirInode == nil {
			t.Fatalf("directory not created")
		}

		err = pod1.RemoveDir(podName3, firstDir+utils.PathSeperator+secondDir)
		if err != nil {
			t.Fatalf("error removing directory")
		}
		dirInode = info.GetDirectory().GetDirFromDirectoryMap(dirPath)
		if dirInode != nil {
			t.Fatalf("directory not removed")
		}

		dirPath = utils.PathSeperator + podName3 + utils.PathSeperator + firstDir
		dirInode = info.GetDirectory().GetDirFromDirectoryMap(dirPath)
		if dirInode == nil {
			t.Fatalf("directory deleted")
		}

		// cleanup pod
		err = pod1.DeletePod(podName3)
		if err != nil {
			t.Fatalf("could not delete pod")
		}
	})

	t.Run("rmdir-multiple-dirs-from-pod", func(t *testing.T) {
		info, err := pod1.CreatePod(podName4, "password")
		if err != nil {
			t.Fatalf("error creating pod %s", podName4)
		}

		err = pod1.MakeDir(podName4, thirdAndFourthDir)
		if err != nil {
			t.Fatalf("error creating directory %s", thirdAndFourthDir)
		}

		// check /test/dir3
		dirPath := utils.PathSeperator + podName4 + utils.PathSeperator + "dir3"
		dirInode := info.GetDirectory().GetDirFromDirectoryMap(dirPath)
		if dirInode == nil {
			t.Fatalf("directory not created")
		}
		// check /test/dir3/dir4
		dirPath = utils.PathSeperator + podName4 + utils.PathSeperator + thirdAndFourthDir
		dirInode = info.GetDirectory().GetDirFromDirectoryMap(dirPath)
		if dirInode == nil {
			t.Fatalf("directory not created")
		}

		err = pod1.RemoveDir(podName4, "dir3")
		if err != nil {
			t.Fatalf("error removing directory")
		}
		dirPath = utils.PathSeperator + podName4 + utils.PathSeperator + "dir3"
		dirInode = info.GetDirectory().GetDirFromDirectoryMap(dirPath)
		if dirInode != nil {
			t.Fatalf("directory not removed")
		}
		dirPath = utils.PathSeperator + podName4 + utils.PathSeperator + thirdAndFourthDir
		dirInode = info.GetDirectory().GetDirFromDirectoryMap(dirPath)
		if dirInode != nil {
			t.Fatalf("directory not removed")
		}

		// cleanup pod
		err = pod1.DeletePod(podName4)
		if err != nil {
			t.Fatalf("could not delete pod")
		}
	})

	t.Run("rmdir-with-slash-on-pod", func(t *testing.T) {
		info, err := pod1.CreatePod(podName5, "password")
		if err != nil {
			t.Fatalf("error creating pod %s", podName5)
		}

		err = pod1.MakeDir(podName5, fifthDir)
		if err != nil {
			t.Fatalf("error creating directory %s", fifthDir)
		}
		dirPath := utils.PathSeperator + podName5 + fifthDir
		dirInode := info.GetDirectory().GetDirFromDirectoryMap(dirPath)
		if dirInode == nil {
			t.Fatalf("directory not created")
		}

		err = pod1.RemoveDir(podName5, fifthDir)
		if err != nil {
			t.Fatalf("error removing directory")
		}
		dirPath = utils.PathSeperator + podName5 + fifthDir
		dirInode = info.GetDirectory().GetDirFromDirectoryMap(dirPath)
		if dirInode != nil {
			t.Fatalf("directory not deleted")
		}

		// cleanup pod
		err = pod1.DeletePod(podName5)
		if err != nil {
			t.Fatalf("could not delete pod")
		}
	})
}
