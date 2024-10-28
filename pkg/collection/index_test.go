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

package collection_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"testing"

	"github.com/ethersphere/bee/v2/pkg/file/redundancy"
	"github.com/ethersphere/bee/v2/pkg/swarm"

	blockstore "github.com/asabya/swarm-blockstore"
	"github.com/asabya/swarm-blockstore/bee"
	mockpost "github.com/ethersphere/bee/v2/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/v2/pkg/storer/mock"
	"github.com/sirupsen/logrus"

	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"

	"github.com/asabya/swarm-blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func TestIndex(t *testing.T) {
	storer := mockstorer.New()
	beeUrl := mock.NewTestBeeServer(t, mock.TestServerOptions{
		Storer:          storer,
		PreventRedirect: true,
		Post:            mockpost.New(mockpost.WithAcceptAll()),
	})

	logger := logging.New(io.Discard, logrus.DebugLevel)
	mockClient := bee.NewBeeClient(beeUrl, bee.WithStamp(mock.BatchOkStr), bee.WithRedundancy(fmt.Sprintf("%d", redundancy.NONE)), bee.WithPinning(true))
	acc := account.New(logger)
	ai := acc.GetUserAccountInfo()
	_, _, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)
	user := acc.GetAddress(account.UserAccountIndex)
	podPassword, _ := utils.GetRandString(pod.PasswordLength)
	t.Run("create_index", func(t *testing.T) {
		//  create an index
		err := collection.CreateIndex("pod1", "testdb_index_0", "key", podPassword, collection.StringIndex, fd, user, mockClient, true)
		if err != nil {
			t.Fatal(err)
		}

		// check if the index is created
		if !isIndexPresent(t, "pod1", "testdb_index_0", "key", podPassword, fd, user, mockClient) {
			t.Fatalf("index not found")
		}
	})

	t.Run("create_and_open_index", func(t *testing.T) {
		//  create an index
		err := collection.CreateIndex("pod1", "testdb_index_1", "key", podPassword, collection.StringIndex, fd, user, mockClient, true)
		if err != nil {
			t.Fatal(err)
		}

		// Open the index
		_, err = collection.OpenIndex("pod1", "testdb_index_1", "key", podPassword, fd, ai, user, mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("close_and_open_index_from_another_machine", func(t *testing.T) {
		// create a DB and open it
		index := createAndOpenIndex(t, "pod1", "testdb_index_2", podPassword, collection.StringIndex, fd, user, mockClient, ai, logger)
		kvMap := addLotOfDocs(t, index, mockClient)

		// open the index again, simulating like another instance
		index1, err := collection.OpenIndex("pod1", "testdb_index_2", "key", podPassword, fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}

		for k, expectedValue := range kvMap {
			gotValue := getDoc(t, k, index1, mockClient)
			if !bytes.Equal(expectedValue, gotValue) {
				t.Fatalf("expected expectedValue %s got expectedValue %s", expectedValue, gotValue)
			}
		}

		// check if an un-inserted value exists
		gotValue := getDoc(t, "p", index1, mockClient)
		if gotValue != nil {
			t.Fatalf("found data for not inserted key")
		}
	})

	t.Run("create_already_present_index", func(t *testing.T) {
		//  create an index
		err := collection.CreateIndex("pod1", "testdb_index_3", "key", podPassword, collection.StringIndex, fd, user, mockClient, true)
		if err != nil {
			t.Fatal(err)
		}

		//  create an index
		err = collection.CreateIndex("pod1", "testdb_index_3", "key", podPassword, collection.StringIndex, fd, user, mockClient, true)
		if !errors.Is(err, collection.ErrIndexAlreadyPresent) {
			t.Fatal(err)
		}
	})

	t.Run("open_index_without_creating_it", func(t *testing.T) {
		// Open the index
		_, err = collection.OpenIndex("pod1", "testdb_index_4", "key", podPassword, fd, ai, user, mockClient, logger)
		if err != collection.ErrIndexNotPresent {
			t.Fatal(err)
		}
	})

	t.Run("create_and_delete_index", func(t *testing.T) {
		//  create an index
		err := collection.CreateIndex("pod1", "testdb_index_5", "key", podPassword, collection.StringIndex, fd, user, mockClient, true)
		if err != nil {
			t.Fatal(err)
		}

		// Open the index
		idx, err := collection.OpenIndex("pod1", "testdb_index_5", "key", podPassword, fd, ai, user, mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}

		// delete Index
		err = idx.DeleteIndex(podPassword)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("delete_index_without_creating_it", func(t *testing.T) {
		// simulate index not present by creating and deleting it
		err := collection.CreateIndex("pod1", "testdb_index_6", "key", podPassword, collection.StringIndex, fd, user, mockClient, true)
		if err != nil {
			t.Fatal(err)
		}
		idx, err := collection.OpenIndex("pod1", "testdb_index_6", "key", podPassword, fd, ai, user, mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}
		err = idx.DeleteIndex(podPassword)
		if err != nil {
			t.Fatal(err)
		}

		// delete Index which is not present
		err = idx.DeleteIndex(podPassword)
		if err != collection.ErrIndexNotPresent {
			t.Fatal(err)
		}
	})

	t.Run("count_docs", func(t *testing.T) {
		// create index and add some docs
		err := collection.CreateIndex("pod1", "testdb_index_7", "key", podPassword, collection.StringIndex, fd, user, mockClient, true)
		if err != nil {
			t.Fatal(err)
		}

		idx, err := collection.OpenIndex("pod1", "testdb_index_7", "key", podPassword, fd, ai, user, mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}

		// add some documents
		actualCount := uint64(100)
		for i := 0; i < int(actualCount); i++ {
			putDocInIndex(t, idx, "key"+strconv.Itoa(i), "value"+strconv.Itoa(i), collection.StringIndex, false)
		}

		// count and check the count
		count, err := idx.CountIndex(podPassword)
		if err != nil {
			t.Fatal(err)
		}

		if count != actualCount {
			t.Fatalf("invalid count in index, expected %d got %d", actualCount, count)
		}

	})

}

func isIndexPresent(t *testing.T, podName, collectionName, indexName, encryptionPassword string, fd *feed.API, user utils.Address, client blockstore.Client) bool {
	actualIndexName := podName + collectionName + indexName
	topic := utils.HashString(actualIndexName)
	_, addr, err := fd.GetFeedData(topic, user, []byte(encryptionPassword), false)
	if err == nil && len(addr) != 0 {
		r, _, err := client.DownloadBlob(swarm.NewAddress(addr))
		if err != nil {
			return false
		}
		defer r.Close()
		data, err := io.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		var manifest collection.Manifest
		err = json.Unmarshal(data, &manifest)
		if err != nil {
			return false
		}
		if manifest.Name != actualIndexName {
			return false
		}
		return true
	}
	return false
}

func putDocInIndex(t *testing.T, index *collection.Index, key, value string, idxTYpe collection.IndexType, apnd bool) {
	err := index.Put(key, []byte(value), idxTYpe, apnd)
	if err != nil {
		t.Fatalf("could not add doc in index: %s:%s, %v", key, value, err)
	}
}
