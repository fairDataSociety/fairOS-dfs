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

package pod_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	"github.com/plexsysio/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

func TestClose(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}

	fd := feed.New(acc.GetUserAccountInfo(), mockClient, logger)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	pod1 := pod.NewPod(mockClient, fd, acc, tm, logger)
	podName1 := "test1"

	t.Run("close-pod", func(t *testing.T) {
		// create a pod
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		info, err := pod1.CreatePod(podName1, "", podPassword)
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}

		// make root dir so that other directories can be added
		err = info.GetDirectory().MkRootDir("pod1", podPassword, info.GetPodAddress(), info.GetFeed())
		if err != nil {
			t.Fatal(err)
		}

		// create some dir and files
		addFilesAndDirectories(t, info, pod1, podName1, podPassword)

		// verify if the pod is closed
		gotPodInfo, _, err := pod1.GetPodInfoFromPodMap(podName1)
		if err == nil {
			t.Fatalf("pod not closed")
		}
		if gotPodInfo != nil {
			t.Fatalf("pod not closed")
		}
		dirObject := info.GetDirectory()
		dirInode1 := dirObject.GetDirFromDirectoryMap("/parentDir/subDir1")
		if dirInode1 != nil {
			t.Fatalf("dir not closed properly")
		}
		dirInode2 := dirObject.GetDirFromDirectoryMap("/parentDir/subDir2")
		if dirInode2 != nil {
			t.Fatalf("dir not closed properly")
		}
		fileObject := info.GetFile()
		fileMeta1 := fileObject.GetFromFileMap("/parentDir/file1")
		if fileMeta1 != nil {
			t.Fatalf("file not closed properly")
		}
		fileMeta2 := fileObject.GetFromFileMap("/parentDir/file2")
		if fileMeta2 != nil {
			t.Fatalf("file not closed properly")
		}
	})

}
