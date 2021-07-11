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
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func TestIndex(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(ioutil.Discard, 0)
	acc := account.New(logger)
	ai := acc.GetUserAccountInfo()
	_, _, err := acc.CreateUserAccount("password", "")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, logger)
	user := acc.GetAddress(account.UserAccountIndex)

	t.Run("create_index", func(t *testing.T) {
		//  create an index
		err := collection.CreateIndex("pod1", "testdb_index_0", "key", collection.StringIndex, fd, user, mockClient, true)
		if err != nil {
			t.Fatal(err)
		}

		// check if the index is created
		if !isIndexPresent(t, "pod1", "testdb_index_0", "key", fd, user, mockClient) {
			t.Fatalf("index not found")
		}
	})

	t.Run("create_and_open_index", func(t *testing.T) {
		//  create an index
		err := collection.CreateIndex("pod1", "testdb_index_1", "key", collection.StringIndex, fd, user, mockClient, true)
		if err != nil {
			t.Fatal(err)
		}

		//Open the index
		_, err = collection.OpenIndex("pod1", "testdb_index_1", "key", fd, ai, user, mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("close_and_open_index_from_another_machine", func(t *testing.T) {
		// create a DB and open it
		index := createAndOpenIndex(t, "pod1", "testdb_index_2", collection.StringIndex, fd, user, mockClient, ai, logger)
		kvMap := addLotOfDocs(t, index, mockClient)

		// open the index again, simulating like another instance
		index1, err := collection.OpenIndex("pod1", "testdb_index_2", "key", fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient, logger)
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
		err := collection.CreateIndex("pod1", "testdb_index_3", "key", collection.StringIndex, fd, user, mockClient, true)
		if err != nil {
			t.Fatal(err)
		}

		//  create an index
		err = collection.CreateIndex("pod1", "testdb_index_3", "key", collection.StringIndex, fd, user, mockClient, true)
		if !errors.Is(err, collection.ErrIndexAlreadyPresent) {
			t.Fatal(err)
		}
	})

	t.Run("open_index_without_creating_it", func(t *testing.T) {
		//Open the index
		_, err = collection.OpenIndex("pod1", "testdb_index_4", "key", fd, ai, user, mockClient, logger)
		if err != collection.ErrIndexNotPresent {
			t.Fatal(err)
		}
	})

	t.Run("create_and_delete_index", func(t *testing.T) {
		//  create an index
		err := collection.CreateIndex("pod1", "testdb_index_5", "key", collection.StringIndex, fd, user, mockClient, true)
		if err != nil {
			t.Fatal(err)
		}

		//Open the index
		idx, err := collection.OpenIndex("pod1", "testdb_index_5", "key", fd, ai, user, mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}

		// delete Index
		err = idx.DeleteIndex()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("delete_index_without_creating_it", func(t *testing.T) {
		// simulate index not present by creating and deleting it
		err := collection.CreateIndex("pod1", "testdb_index_6", "key", collection.StringIndex, fd, user, mockClient, true)
		if err != nil {
			t.Fatal(err)
		}
		idx, err := collection.OpenIndex("pod1", "testdb_index_6", "key", fd, ai, user, mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}
		err = idx.DeleteIndex()
		if err != nil {
			t.Fatal(err)
		}

		// delete Index which is not present
		err = idx.DeleteIndex()
		if err != collection.ErrIndexNotPresent {
			t.Fatal(err)
		}
	})

	t.Run("count_docs", func(t *testing.T) {
		// create index and add some docs
		err := collection.CreateIndex("pod1", "testdb_index_7", "key", collection.StringIndex, fd, user, mockClient, true)
		if err != nil {
			t.Fatal(err)
		}

		idx, err := collection.OpenIndex("pod1", "testdb_index_7", "key", fd, ai, user, mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}

		// add some documents
		actualCount := uint64(100)
		for i := 0; i < int(actualCount); i++ {
			putDocInIndex(t, idx, "key"+strconv.Itoa(i), "value"+strconv.Itoa(i), collection.StringIndex, false)
		}

		// count and check the count
		count, err := idx.CountIndex()
		if err != nil {
			t.Fatal(err)
		}

		if count != actualCount {
			t.Fatalf("invalid count in index, expected %d got %d", actualCount, count)
		}

	})

}

func isIndexPresent(t *testing.T, podName, collectionName, indexName string, fd *feed.API, user utils.Address, client blockstore.Client) bool {
	actualIndexName := podName + collectionName + indexName
	topic := utils.HashString(actualIndexName)
	_, addr, err := fd.GetFeedData(topic, user)
	if err == nil && len(addr) != 0 {
		data, _, err := client.DownloadBlob(addr)
		if err != nil {
			return false
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
