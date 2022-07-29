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
	"io"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func TestShare(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("password", "")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, logger)
	pod1 := pod.NewPod(mockClient, fd, acc, logger)
	podName1 := "test1"

	acc2 := account.New(logger)
	_, _, err = acc2.CreateUserAccount("password2", "")
	if err != nil {
		t.Fatal(err)
	}
	fd2 := feed.New(acc2.GetUserAccountInfo(), mockClient, logger)
	pod2 := pod.NewPod(mockClient, fd2, acc2, logger)
	podName2 := "test2"

	acc3 := account.New(logger)
	_, _, err = acc3.CreateUserAccount("password3", "")
	if err != nil {
		t.Fatal(err)
	}
	fd3 := feed.New(acc3.GetUserAccountInfo(), mockClient, logger)
	pod3 := pod.NewPod(mockClient, fd3, acc3, logger)
	podName3 := "test3"

	acc4 := account.New(logger)
	_, _, err = acc4.CreateUserAccount("password4", "")
	if err != nil {
		t.Fatal(err)
	}
	fd4 := feed.New(acc4.GetUserAccountInfo(), mockClient, logger)
	pod4 := pod.NewPod(mockClient, fd4, acc4, logger)
	podName4 := "test4"

	t.Run("share-pod", func(t *testing.T) {
		// create a pod
		info, err := pod1.CreatePod(podName1, "password", "")
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}

		// make root dir so that other directories can be added
		err = info.GetDirectory().MkRootDir("pod1", info.GetPodAddress(), info.GetFeed())
		if err != nil {
			t.Fatal(err)
		}

		// create some dir and files
		addFilesAndDirectories(t, info, pod1, podName1)

		// share pod
		sharingRef, err := pod1.PodShare(podName1, "password", "Alice")
		if err != nil {
			t.Fatal(err)
		}

		// verify if pod is shared
		if sharingRef == "" {
			t.Fatalf("could not share pod")
		}
	})

	t.Run("receive-pod-info", func(t *testing.T) {
		// create a pod
		info, err := pod2.CreatePod(podName2, "password2", "")
		if err != nil {
			t.Fatalf("error creating pod %s", podName2)
		}

		// make root dir so that other directories can be added
		err = info.GetDirectory().MkRootDir("pod1", info.GetPodAddress(), info.GetFeed())
		if err != nil {
			t.Fatal(err)
		}

		// create some dir and files
		addFilesAndDirectories(t, info, pod2, podName2)

		// share pod
		sharingRef, err := pod2.PodShare(podName2, "password2", "Alice")
		if err != nil {
			t.Fatal(err)
		}

		// receive pod info
		ref, err := utils.ParseHexReference(sharingRef)
		if err != nil {
			t.Fatal(err)
		}
		sharingInfo, err := pod2.ReceivePodInfo(ref)
		if err != nil {
			t.Fatal(err)
		}

		// verify the pod info
		if sharingInfo == nil {
			t.Fatalf("could not receive sharing info")
		}
		if sharingInfo.PodName != podName2 {
			t.Fatalf("invalid pod name received")
		}
		if sharingInfo.UserName != "Alice" {
			t.Fatalf("invalid user address received")
		}
	})

	t.Run("receive-pod", func(t *testing.T) {
		// create sending pod and receiving pod
		info, err := pod3.CreatePod(podName3, "password3", "")
		if err != nil {
			t.Fatalf("error creating pod %s", podName3)
		}
		_, err = pod4.CreatePod(podName4, "password4", "")
		if err != nil {
			t.Fatalf("error creating pod %s", podName4)
		}

		// make root dir so that other directories can be added
		err = info.GetDirectory().MkRootDir("pod1", info.GetPodAddress(), info.GetFeed())
		if err != nil {
			t.Fatal(err)
		}

		// create some dir and files
		addFilesAndDirectories(t, info, pod3, podName3)

		// share pod
		sharingRef, err := pod3.PodShare(podName3, "password3", "Alice")
		if err != nil {
			t.Fatal(err)
		}

		// receive pod info
		ref, err := utils.ParseHexReference(sharingRef)
		if err != nil {
			t.Fatal(err)
		}
		podInfo, err := pod4.ReceivePod(ref)
		if err != nil {
			t.Fatal(err)
		}

		// verify the pod info
		if podInfo == nil {
			t.Fatalf("could not receive sharing info")
		}
		if podInfo.GetPodName() != podName3 {
			t.Fatalf("invalid pod name received")
		}

		pods, sharedPods, err := pod4.ListPods()
		if err != nil {
			t.Fatal(err)
		}
		if pods == nil {
			t.Fatalf("invalid pods")
		}
		if len(pods) != 1 && pods[0] != podName1 {
			t.Fatalf("invalid pod name")
		}
		if sharedPods == nil {
			t.Fatalf("invalid shared pods")
		}
		if len(sharedPods) != 1 && sharedPods[0] != podName4 {
			t.Fatalf("invalid pod name")
		}
	})
}
