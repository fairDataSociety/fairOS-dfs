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
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestPod_RemoveFile(t *testing.T) {
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
	firstDir := "dir1"

	t.Run("remove_file", func(t *testing.T) {
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
		podFile := createRandomFileInPod(t, 540, pod1, podName1, dirPath)

		err = pod1.CopyToLocal(podName1, podFile, os.TempDir())
		if err != nil {
			t.Fatalf("error copying file to local dir %s", err.Error())
		}

		fileInfo, err := os.Stat(os.TempDir() + utils.PathSeperator + filepath.Base(podFile))
		if err != nil {
			t.Fatalf("file not copied to local")
		}

		if fileInfo.Size() != 540 {
			t.Fatalf("invalid file size")
		}

		// Delete all the blocj=ks and fileinode
		err = pod1.RemoveFile(podName1, podFile)
		if err != nil {
			t.Fatal(err)
		}

		os.Remove(fileInfo.Name())
		err = pod1.DeletePod(podName1)
		if err != nil {
			t.Fatalf("could not delete pod")
		}
	})
}
