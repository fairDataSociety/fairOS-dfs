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

package test_test

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/ethersphere/bee/v2/pkg/file/redundancy"

	mockpost "github.com/ethersphere/bee/v2/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/v2/pkg/storer/mock"

	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/sirupsen/logrus"

	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager/rpc/mock"

	"github.com/asabya/swarm-blockstore/bee"
	"github.com/asabya/swarm-blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
)

func TestSync(t *testing.T) {
	storer := mockstorer.New()
	beeUrl := mock.NewTestBeeServer(t, mock.TestServerOptions{
		Storer:          storer,
		PreventRedirect: true,
		Post:            mockpost.New(mockpost.WithAcceptAll()),
	})

	logger := logging.New(io.Discard, logrus.DebugLevel)
	mockClient := bee.NewBeeClient(beeUrl, bee.WithStamp(mock.BatchOkStr), bee.WithRedundancy(fmt.Sprintf("%d", redundancy.NONE)), bee.WithPinning(true))

	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	sm := mock2.NewMockSubscriptionManager()

	fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)
	pod1 := pod.NewPod(mockClient, fd, acc, tm, sm, -1, 0, logger)
	podName1 := "test1"

	t.Run("sync-pod", func(t *testing.T) {

		err := pod1.SyncPod(podName1)
		if err == nil {
			t.Fatal("sync should fail, pod not opened")
		}
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

		// open the pod ths triggers sync too
		gotInfo, err := pod1.OpenPod(podName1)
		if err != nil {
			t.Fatal(err)
		}

		// validate if the directory and files are synced
		dirObject := gotInfo.GetDirectory()
		dirInode1, _ := dirObject.GetInode(podPassword, "/parentDir/subDir1")
		if dirInode1 == nil {
			t.Fatalf("invalid dir entry")
		}
		if dirInode1.Meta.Path != "/parentDir" {
			t.Fatalf("invalid path entry")
		}
		if dirInode1.Meta.Name != "subDir1" {
			t.Fatalf("invalid dir entry")
		}
		dirInode2, _ := dirObject.GetInode(podPassword, "/parentDir/subDir2")
		if dirInode2 == nil {
			t.Fatalf("invalid dir entry")
		}
		if dirInode2.Meta.Path != "/parentDir" {
			t.Fatalf("invalid path entry")
		}
		if dirInode2.Meta.Name != "subDir2" {
			t.Fatalf("invalid dir entry")
		}

		fileObject := gotInfo.GetFile()
		fileMeta1 := fileObject.GetInode(podPassword, "/parentDir/file1")
		if fileMeta1 == nil {
			t.Fatalf("invalid file meta")
		}
		if fileMeta1.Path != "/parentDir" {
			t.Fatalf("invalid path entry")
		}
		if fileMeta1.Name != "file1" {
			t.Fatalf("invalid file entry")
		}
		if fileMeta1.Size != uint64(100) {
			t.Fatalf("invalid file size")
		}
		if fileMeta1.BlockSize != file.MinBlockSize {
			t.Fatalf("invalid block size")
		}
		fileMeta2 := fileObject.GetInode(podPassword, "/parentDir/file2")
		if fileMeta2 == nil {
			t.Fatalf("invalid file meta")
		}
		if fileMeta2.Path != "/parentDir" {
			t.Fatalf("invalid path entry")
		}
		if fileMeta2.Name != "file2" {
			t.Fatalf("invalid file entry")
		}
		if fileMeta2.Size != uint64(200) {
			t.Fatalf("invalid file size")
		}
		if fileMeta2.BlockSize != file.MinBlockSize {
			t.Fatalf("invalid block size")
		}
	})
}
