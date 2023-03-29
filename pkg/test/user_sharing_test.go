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
	"errors"
	"io"
	"strconv"
	"testing"
	"time"

	mock3 "github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager/rpc/mock"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
)

func TestSharing(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)

	acc1 := account.New(logger)
	_, _, err := acc1.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	_, err = acc1.CreatePodAccount(1, false)
	if err != nil {
		t.Fatal(err)
	}
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	sm := mock3.NewMockSubscriptionManager()

	fd1 := feed.New(acc1.GetUserAccountInfo(), mockClient, logger)
	pod1 := pod.NewPod(mockClient, fd1, acc1, tm, sm, logger)
	podName1 := "test1"

	acc2 := account.New(logger)
	_, _, err = acc2.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	_, err = acc2.CreatePodAccount(1, false)
	if err != nil {
		t.Fatal(err)
	}
	fd2 := feed.New(acc2.GetUserAccountInfo(), mockClient, logger)
	pod2 := pod.NewPod(mockClient, fd2, acc2, tm, sm, logger)
	podName2 := "test2"

	t.Run("sharing-user", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()
		// create source user
		userObject1 := user.NewUsers(mockClient, ens, logger)
		_, _, _, _, ui0, err := userObject1.CreateNewUserV2("user1", "password1twelve", "", "", tm, sm)
		if err != nil {
			t.Fatal(err)
		}
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		// create source pod
		info1, err := pod1.CreatePod(podName1, "", podPassword)
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}
		ui0.AddPodName(podName1, info1)

		// make root dir so that other directories can be added
		err = info1.GetDirectory().MkRootDir("pod1", podPassword, info1.GetPodAddress(), info1.GetFeed())
		if err != nil {
			t.Fatal(err)
		}

		// create dir and file
		dirObject1 := info1.GetDirectory()
		err = dirObject1.MkDir("/parentDir1", podPassword, 0)
		if err != nil {
			t.Fatal(err)
		}
		fileObject1 := info1.GetFile()
		_, err = uploadFile(t, fileObject1, "/parentDir1", "file1", "", podPassword, 100, 10)
		if err != nil {
			t.Fatal(err)
		}
		// share file with another user
		sharingRefString, err := userObject1.ShareFileWithUser("pod1", podPassword, "/parentDir1/file1", "user2", ui0, pod1, info1.GetPodAddress())
		if err != nil {
			t.Fatal(err)
		}

		// create destination user
		userObject2 := user.NewUsers(mockClient, ens, logger)
		_, _, _, _, ui, err := userObject2.CreateNewUserV2("user2", "password1twelve", "", "", tm, sm)
		if err != nil {
			t.Fatal(err)
		}

		// create destination pod
		podPassword, _ = utils.GetRandString(pod.PasswordLength)
		info2, err := pod2.CreatePod(podName2, "", podPassword)
		if err != nil {
			t.Fatalf("error creating pod %s", podName2)
		}

		// make root dir so that other directories can be added
		err = info2.GetDirectory().MkRootDir("pod1", podPassword, info2.GetPodAddress(), info2.GetFeed())
		if err != nil {
			t.Fatal(err)
		}

		// create dir and file
		dirObject2 := info2.GetDirectory()
		err = dirObject2.MkDir("/parentDir2", podPassword, 0)
		if err != nil {
			t.Fatal(err)
		}

		// receive file info
		receiveFileInfo, err := userObject2.ReceiveFileInfo(sharingRefString)
		if err != nil {
			t.Fatal(err)
		}

		// validate receive file info
		if receiveFileInfo == nil {
			t.Fatalf("invalid receive file info")
		}
		if receiveFileInfo.FileName != "file1" {
			t.Fatalf("invalid filename received")
		}
		if receiveFileInfo.Size != strconv.FormatUint(100, 10) {
			t.Fatalf("invalid file size received")
		}
		if receiveFileInfo.BlockSize != strconv.FormatUint(10, 10) {
			t.Fatalf("invalid block size received")
		}

		_, err = userObject2.ReceiveFileFromUser("podName2", sharingRefString, ui, pod2, "/parentDir2")
		if err == nil {
			t.Fatal("pod should not exist")
		}

		// receive file
		destinationFilePath, err := userObject2.ReceiveFileFromUser(podName2, sharingRefString, ui, pod2, "/parentDir2")
		if err != nil {
			t.Fatal(err)
		}

		_, err = userObject2.ReceiveFileFromUser(podName2, sharingRefString, ui, pod2, "/parentDir2")
		if !errors.Is(err, file.ErrFileAlreadyPresent) {
			t.Fatal("pod does not supposed tp be open")
		}

		// verify receive
		if destinationFilePath != "/parentDir2/file1" {
			t.Fatalf("invalid destination file name")
		}
		_, files, err := dirObject2.ListDir("/parentDir2", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		if files == nil {
			t.Fatalf("file not imported")
		}
		if len(files) != 1 {
			t.Fatalf("file not imported")
		}
		if files[0] != "/parentDir2/file1" {
			t.Fatalf("file not imported")
		}
		// delete source pod
		err = pod1.DeleteOwnPod(podName1)
		if err != nil {
			t.Fatalf("error deleting pod %s", podName1)
		}
		ui0.RemovePodName(podName1)
	})
}
