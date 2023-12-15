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
	"io"
	"testing"
	"time"

	mockpost "github.com/ethersphere/bee/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager/rpc/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
	"github.com/sirupsen/logrus"
)

func TestClose(t *testing.T) {
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

	fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()

	sm := mock2.NewMockSubscriptionManager()

	pod1 := pod.NewPod(mockClient, fd, acc, tm, sm, -1, 0, logger)
	podName1 := "test1"

	t.Run("close-pod", func(t *testing.T) {
		// create a pod
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		info, err := pod1.CreatePod(podName1, "", podPassword)
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
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

		gotPodInfo, _, err = pod1.GetPodInfo(podName1)
		if err != nil {
			t.Fatalf("pod should be open %s\n", err)
		}
		if gotPodInfo == nil {
			t.Fatalf("pod should be open")
		}
		dirObject := gotPodInfo.GetDirectory()
		dirInode1, _ := dirObject.GetInode(podPassword, "/parentDir/subDir1")
		if dirInode1 == nil {
			t.Fatalf("dir should nil be nil")
		}
		dirInode2, _ := dirObject.GetInode(podPassword, "/parentDir/subDir2")
		if dirInode2 == nil {
			t.Fatalf("dir should nil be nil")
		}
		fileObject := gotPodInfo.GetFile()
		fileMeta1 := fileObject.GetInode(podPassword, "/parentDir/file1")
		if fileMeta1 == nil {
			t.Fatalf("file should nil be nil")
		}
		fileMeta2 := fileObject.GetInode(podPassword, "/parentDir/file2")
		if fileMeta2 == nil {
			t.Fatalf("file should nil be nil")
		}
	})

}
