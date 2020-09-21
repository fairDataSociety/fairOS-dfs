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
	"strings"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

func TestPod_New(t *testing.T) {
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
	t.Run("create-first-pod", func(t *testing.T) {
		info, err := pod1.CreatePod(podName1, "password")
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}

		if pod1.fd == nil || pod1.acc == nil {
			t.Fatalf("user not initialized")
		}

		if info.GetCurrentPodNameOnly() != podName1 {
			t.Fatalf("invalid pod name: expected %s got %s", podName1, info.GetCurrentPodNameOnly())
		}

		pods, err := pod1.loadUserPods()
		if err != nil {
			t.Fatalf("error getting pods")
		}

		if len(pods) != 1 {
			t.Fatalf("length of pods is not 1")
		}

		if strings.Trim(pods[0], "\n") != podName1 {
			t.Fatalf("podName is not %s", podName1)
		}

		infoGot, err := pod1.GetPodInfoFromPodMap(podName1)
		if err != nil {
			t.Fatalf("could not get pod from podMap")
		}

		if infoGot.GetCurrentPodNameOnly() != podName1 {
			t.Fatalf("invalid pod name: expected %s got %s", podName1, infoGot.GetCurrentPodNameOnly())
		}

		dirInode := info.dir.GetDirFromDirectoryMap(podName1)
		if dirInode == nil {
			t.Fatalf("pod not added as direcory")
		}

	})

	t.Run("create-second-pod", func(t *testing.T) {
		info, err := pod1.CreatePod(podName2, "password")
		if err != nil {
			t.Fatalf("error creating pod %s", podName2)
		}

		if info.GetCurrentPodNameOnly() != podName2 {
			t.Fatalf("invalid pod name: expected %s got %s", podName2, info.GetCurrentPodNameOnly())
		}

		pods, err := pod1.loadUserPods()
		if err != nil {
			t.Fatalf("error getting pods")
		}

		if len(pods) != 2 {
			t.Fatalf("length of pods is not 2")
		}

		if strings.Trim(pods[1], "\n") != podName2 {
			t.Fatalf("podName is not %s", podName2)
		}

		infoGot, err := pod1.GetPodInfoFromPodMap(podName2)
		if err != nil {
			t.Fatalf("could not get pod from podMap")
		}

		if infoGot.GetCurrentPodNameOnly() != podName2 {
			t.Fatalf("invalid pod name: expected %s got %s", podName2, infoGot.GetCurrentPodNameOnly())
		}

		dirInode := info.dir.GetDirFromDirectoryMap(podName2)
		if dirInode == nil {
			t.Fatalf("pod not added as direcory")
		}
	})
}
