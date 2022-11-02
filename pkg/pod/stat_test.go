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
	"strings"
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

func TestStat(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("password", "")
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

	t.Run("pod-stat", func(t *testing.T) {
		_, err := pod1.PodStat(podName1)
		if err == nil {
			t.Fatal("stat should be nil")
		}
		podPassword, _ := utils.GetRandString(pod.PodPasswordLength)
		info, err := pod1.CreatePod(podName1, "password", "", podPassword)
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}

		// get pod stat
		podStat, err := pod1.PodStat(podName1)
		if err != nil {
			t.Fatal(err)
		}

		// verify if the stat is right
		if podStat == nil {
			t.Fatalf("invalid pod stat")
		}
		if podStat.PodName != podName1 {
			t.Fatalf("invalid pod name: expected %s got %s", podName1, podStat.PodName)
		}
		addr := info.GetAccountInfo().GetAddress().Hex()[2:]
		addr = strings.ToLower(addr)
		if podStat.PodAddress != addr {
			t.Fatalf("invalid pod address")
		}

	})

}
