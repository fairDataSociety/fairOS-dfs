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
	"sort"
	"strings"
	"testing"
	"time"

	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager/rpc/mock"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	"github.com/plexsysio/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

func TestPodDelete(t *testing.T) {
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
	sm := mock2.NewMockSubscriptionManager()

	pod1 := pod.NewPod(mockClient, fd, acc, tm, sm, logger)

	podName1 := "test1"
	podName2 := "test2"

	t.Run("create-one-pod-and-del", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		_, err := pod1.CreatePod(podName1, "", podPassword)
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}

		pods, _, err := pod1.ListPods()
		if err != nil {
			t.Fatalf("error getting pods")
		}

		if strings.Trim(pods[0], "\n") != podName1 {
			t.Fatalf("podName is not %s", podName1)
		}

		err = pod1.DeleteOwnPod(podName1)
		if err != nil {
			t.Fatal(err)
		}

		err = pod1.DeleteOwnPod(podName1)
		if err == nil {
			t.Fatal("pod should have been deleted")
		}

		pods, _, err = pod1.ListPods()
		if err != nil {
			t.Fatalf("error getting pods")
		}

		if len(pods) > 1 {
			t.Fatalf("delete failed")
		}

		infoGot, _, err := pod1.GetPodInfoFromPodMap(podName1)
		if err == nil {
			t.Fatalf("pod not deleted from map")
		}
		if infoGot != nil {
			t.Fatalf("pod not deleted from map")
		}
	})

	t.Run("create-two-pod-and-del", func(t *testing.T) {
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		_, err := pod1.CreatePod(podName1, "", podPassword)
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}
		_, err = pod1.CreatePod(podName2, "", podPassword)
		if err != nil {
			t.Fatalf("error creating pod %s", podName1)
		}

		pods, _, err := pod1.ListPods()
		if err != nil {
			t.Fatalf("error getting pods")
		}

		sort.Strings(pods)
		if strings.Trim(pods[0], "\n") != podName1 {
			t.Fatalf("podName is not %s, got %s", podName1, pods[0])
		}

		if strings.Trim(pods[1], "\n") != podName2 {
			t.Fatalf("podName is not %s, got %s", podName2, pods[1])
		}

		err = pod1.DeleteOwnPod(podName1)
		if err != nil {
			t.Fatal(err)
		}

		pods, _, err = pod1.ListPods()
		if err != nil {
			t.Fatalf("error getting pods")
		}

		if len(pods) > 1 {
			t.Fatalf("delete failed")
		}

		if strings.Trim(pods[0], "\n") != podName2 {
			t.Fatalf("delete pod failed")
		}

		infoGot, _, err := pod1.GetPodInfoFromPodMap(podName1)
		if err == nil {
			t.Fatalf("pod not deleted from map")
		}
		if infoGot != nil {
			t.Fatalf("pod not deleted from map")
		}

		_, _, err = pod1.GetPodInfoFromPodMap(podName2)
		if err != nil {
			t.Fatalf("removed wrong pod")
		}

	})

	t.Run("create-pod-and-del-with-tables", func(t *testing.T) {
		podName := "delPod"
		for i := 0; i < 10; i++ {
			podPassword, _ := utils.GetRandString(pod.PasswordLength)
			pi, err := pod1.CreatePod(podName, "", podPassword)
			if err != nil {
				t.Fatalf("error creating pod %s", podName)
			}
			dbTables, err := pi.GetDocStore().LoadDocumentDBSchemas(podPassword)
			if err != nil {
				t.Fatalf("err doc list %s", podName)
			}
			if len(dbTables) != 0 {
				t.Fatal("doc tables delete failed while pod delete")
			}
			kvTables, err := pi.GetKVStore().LoadKVTables(podPassword)
			if err != nil {
				t.Fatalf("err kv list %s", podName)
			}
			if len(kvTables) != 0 {
				t.Fatal("kv tables delete failed while pod delete")
			}
			si := make(map[string]collection.IndexType)
			si["first_name"] = collection.StringIndex
			si["age"] = collection.NumberIndex
			err = pi.GetDocStore().CreateDocumentDB("dbName", podPassword, si, true)
			if err != nil {
				t.Fatal(err)
			}

			err = pi.GetKVStore().CreateKVTable("kvName", podPassword, collection.StringIndex)
			if err != nil {
				t.Fatal(err)
			}

			dbTables, err = pi.GetDocStore().LoadDocumentDBSchemas(podPassword)
			if err != nil {
				t.Fatalf("err doc list %s", podName)
			}
			if len(dbTables) != 1 {
				t.Fatal("doc tables create failed while pod delete")
			}
			kvTables, err = pi.GetKVStore().LoadKVTables(podPassword)
			if err != nil {
				t.Fatalf("err kv list %s", podName)
			}
			if len(kvTables) != 1 {
				t.Fatal("kv tables create failed while pod delete")
			}

			err = pod1.DeleteOwnPod(podName)
			if err != nil {
				t.Fatal(err)
			}
		}
	})
}
