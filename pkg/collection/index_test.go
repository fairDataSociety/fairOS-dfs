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
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

func TestIndex(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(ioutil.Discard, 0)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("password", "")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, logger)

	t.Run("sync_index", func(t *testing.T) {
		//  create and populate the index
		err := collection.CreateIndex("testdb0", "key", fd, acc.GetAddress(account.UserAccountIndex), mockClient)
		if err != nil {
			t.Fatal(err)
		}
		index, err := collection.OpenIndex("testdb0", "key", fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}
		kvMap := addLotOfDocs(t, index, mockClient)

		// open the index again, simulating like another instance
		index1, err := collection.OpenIndex("testdb0", "key", fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}

		for k, expectedValue := range kvMap {
			gotValue := getDoc(t, k, index1, mockClient)
			if !bytes.Equal(expectedValue, gotValue) {
				t.Fatalf("expected expectedValue %s got expectedValue %s", expectedValue, gotValue)
			}
		}

		// check if anuuninserted value exists
		gotValue := getDoc(t, "p", index1, mockClient)
		if gotValue != nil {
			t.Fatalf("found data for not inserted key")
		}

		// delete the index
		err = index1.DeleteIndex()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("get-doc", func(t *testing.T) {
		//  create and populate the index
		err := collection.CreateIndex("testdb1", "key", fd, acc.GetAddress(account.UserAccountIndex), mockClient)
		if err != nil {
			t.Fatal(err)
		}
		index, err := collection.OpenIndex("testdb1", "key", fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}
		kvMap := addLotOfDocs(t, index, mockClient)

		// get the expectedValue of keys and check against its actual expectedValue
		for k, expectedValue := range kvMap {
			gotValue := getDoc(t, k, index, mockClient)
			if !bytes.Equal(expectedValue, gotValue) {
				t.Fatalf("expected expectedValue %s got expectedValue %s", expectedValue, gotValue)
			}
		}

		gotValue := getDoc(t, "p", index, mockClient)
		if gotValue != nil {
			t.Fatalf("found data for not inserted key")
		}

		// delete the index
		err = index.DeleteIndex()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("get-doc-del-doc-get-doc", func(t *testing.T) {
		//  create and populate the index
		err := collection.CreateIndex("testdb2", "key", fd, acc.GetAddress(account.UserAccountIndex), mockClient)
		if err != nil {
			t.Fatal(err)
		}
		index, err := collection.OpenIndex("testdb2", "key", fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}
		kvMap := addLotOfDocs(t, index, mockClient)

		// get the value of the key just to check
		value := kvMap["aa"]
		gotValue := getDoc(t, "aa", index, mockClient)
		if !bytes.Equal(value, gotValue) {
			t.Fatalf("expected value %s got value %s", value, gotValue)
		}

		// delete the key from the index
		deletedValue := delDoc(t, "aa", index, mockClient)
		if !bytes.Equal(value, deletedValue) {
			t.Fatalf("expected value %s got value %s", value, gotValue)
		}

		// check if the key is not in the index
		afterDeletedValue := getDoc(t, "aa", index, mockClient)
		if afterDeletedValue != nil {
			t.Fatalf("shuld not have got any value")
		}

		// delete the index
		err = index.DeleteIndex()
		if err != nil {
			t.Fatal(err)
		}

	})

	t.Run("add-docs-iterrate", func(t *testing.T) {
		//  create and populate the index
		err := collection.CreateIndex("testdb3", "key", fd, acc.GetAddress(account.UserAccountIndex), mockClient)
		if err != nil {
			t.Fatal(err)
		}
		index, err := collection.OpenIndex("testdb3", "key", fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}
		kvMap := addLotOfDocs(t, index, mockClient)

		// create the iterator
		itr, err := index.NewIterator("", "", 100)
		if err != nil {
			t.Fatal(err)
		}

		// iterate through the keys and check for the values returned
		count := 0
		for itr.Next() {
			value := getDoc(t, itr.Key(), index, mockClient)
			if !bytes.Equal(kvMap[itr.Key()], value) {
				t.Fatalf("expected value %s but got %s for the key %s", string(kvMap[itr.Key()]), string(value), itr.Key())
			}
			count++
		}

		if len(kvMap) != count {
			t.Fatalf("number of elements mismatch in iteration")
		}

		// delete the index
		err = index.DeleteIndex()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add-docs-seek-iterate", func(t *testing.T) {
		//  create and populate the index
		err := collection.CreateIndex("testdb4", "key", fd, acc.GetAddress(account.UserAccountIndex), mockClient)
		if err != nil {
			t.Fatal(err)
		}
		index, err := collection.OpenIndex("testdb4", "key", fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}
		kvMap := addLotOfDocs(t, index, mockClient)

		// create the iterator
		itr, err := index.NewIterator("abc", "bbb", 100)
		if err != nil {
			t.Fatal(err)
		}

		// iterate through the keys and check for the values returned
		count := 0
		for itr.Next() {
			value := getValue(t, itr.Value(), mockClient)
			if !bytes.Equal(kvMap[itr.Key()], value) {
				t.Fatalf("expected value %s but got %s for the key %s", string(kvMap[itr.Key()]), string(value), itr.Key())
			}
			count++
		}

		if count != 7 {
			t.Fatalf("number of elements mismatch in iteration")
		}

		// delete the index
		err = index.DeleteIndex()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add-docs-seek-iterrate-with-limit", func(t *testing.T) {
		//  create and populate the index
		err := collection.CreateIndex("testdb5", "key", fd, acc.GetAddress(account.UserAccountIndex), mockClient)
		if err != nil {
			t.Fatal(err)
		}
		index, err := collection.OpenIndex("testdb5", "key", fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}
		kvMap := addLotOfDocs(t, index, mockClient)

		// create the iterator
		itr, err := index.NewIterator("abc", "bbb", 5)
		if err != nil {
			t.Fatal(err)
		}

		// iterate through the keys and check for the values returned
		count := 0
		for itr.Next() {
			value := getValue(t, itr.Value(), mockClient)
			if !bytes.Equal(kvMap[itr.Key()], value) {
				t.Fatalf("expected value %s but got %s for the key %s", string(kvMap[itr.Key()]), string(value), itr.Key())
			}
			count++
		}

		if count != 5 {
			t.Fatalf("number of elements mismatch in iteration")
		}

		// delete the index
		err = index.DeleteIndex()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("batch-add-docs", func(t *testing.T) {
		//  create and populate the index
		err := collection.CreateIndex("testdb6", "key", fd, acc.GetAddress(account.UserAccountIndex), mockClient)
		if err != nil {
			t.Fatal(err)
		}
		index, err := collection.OpenIndex("testdb6", "key", fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}

		// batch load and delete
		batch, err := index.Batch()
		if err != nil {
			t.Fatal(err)
		}

		batchDocs := addBatchDocs(t, batch, mockClient)
		err = batch.Write()
		if err != nil {
			t.Fatal(err)
		}

		// create the iterator
		itr, err := index.NewIterator("", "", 100)
		if err != nil {
			t.Fatal(err)
		}

		// iterate through the keys and check for the values returned
		count := 0
		for itr.Next() {
			value := getValue(t, itr.Value(), mockClient)
			if !bytes.Equal(batchDocs[itr.Key()], value) {
				t.Fatalf("expected value %s but got %s for the key %s", string(batchDocs[itr.Key()]), string(value), itr.Key())
			}
			count++
		}

		if len(batchDocs) != count {
			t.Fatalf("number of elements mismatch in iteration")
		}

		// delete the index
		err = index.DeleteIndex()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("batch-add-del-docs", func(t *testing.T) {
		//  create and populate the index
		err := collection.CreateIndex("testdb7", "key", fd, acc.GetAddress(account.UserAccountIndex), mockClient)
		if err != nil {
			t.Fatal(err)
		}
		index, err := collection.OpenIndex("testdb7", "key", fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}

		// batch load and delete
		batch, err := index.Batch()
		if err != nil {
			t.Fatal(err)
		}

		batchDocs := addBatchDocs(t, batch, mockClient)
		err = batch.Write()
		if err != nil {
			t.Fatal(err)
		}

		// create the iterator
		itr, err := index.NewIterator("", "", 100)
		if err != nil {
			t.Fatal(err)
		}

		// iterate through the keys and check for the values returned
		count := 0
		for itr.Next() {
			value := getValue(t, itr.Value(), mockClient)
			if !bytes.Equal(batchDocs[itr.Key()], value) {
				t.Fatalf("expected value %s but got %s for the key %s", string(batchDocs[itr.Key()]), string(value), itr.Key())
			}
			count++
		}

		if len(batchDocs) != count {
			t.Fatalf("number of elements mismatch in iteration")
		}

		// delete the index
		err = index.DeleteIndex()
		if err != nil {
			t.Fatal(err)
		}
	})
}

func addDoc(t *testing.T, key string, value []byte, index *collection.Index, client *mock.MockBeeClient) {
	ref, err := client.UploadBlob(value, false, false)
	if err != nil {
		t.Fatalf("could not add doc %s:%s, %v", key, value, err)
	}
	err = index.Put(key, ref)
	if err != nil {
		t.Fatalf("could not add doc in index: %s:%s, %v", key, ref, err)
	}
}

func getDoc(t *testing.T, key string, index *collection.Index, client *mock.MockBeeClient) []byte {
	ref, err := index.Get(key)
	if err != nil {
		if errors.Is(err, collection.ErrEntryNotFound) {
			return nil
		}
		t.Fatal(err)
	}
	data, respCode, err := client.DownloadBlob(ref)
	if err != nil {
		t.Fatal(err)
	}
	if respCode != http.StatusOK {
		t.Fatalf("invalid response code")
	}
	return data
}

func getValue(t *testing.T, ref []byte, client *mock.MockBeeClient) []byte {
	data, respCode, err := client.DownloadBlob(ref)
	if err != nil {
		t.Fatal(err)
	}
	if respCode != http.StatusOK {
		t.Fatalf("invalid response code")
	}
	return data
}

func delDoc(t *testing.T, key string, index *collection.Index, client *mock.MockBeeClient) []byte {
	ref, err := index.Delete(key)
	if err != nil {
		t.Fatal(err)
	}
	data, respCode, err := client.DownloadBlob(ref)
	if err != nil {
		t.Fatal(err)
	}
	if respCode != http.StatusOK {
		t.Fatalf("invalid response code")
	}
	return data

}

func addLotOfDocs(t *testing.T, index *collection.Index, client *mock.MockBeeClient) map[string][]byte {
	// Initialize the values
	kvMap := make(map[string][]byte)
	kvMap["key1"] = []byte("value1")
	kvMap["key11"] = []byte("value11")
	kvMap["aa"] = []byte("doc2")
	kvMap["abc"] = []byte("doc4")
	kvMap["abcd"] = []byte("doc5")
	kvMap["aca"] = []byte("doc6")
	kvMap["ac"] = []byte("doc7")
	kvMap["acb"] = []byte("doc8")
	kvMap["aaa"] = []byte("doc9")
	kvMap["az"] = []byte("doc10")
	kvMap["acdb"] = []byte("doc11")
	kvMap["file1"] = []byte("doc12")
	kvMap["file1.jpg"] = []byte("doc13")

	// add the documents
	for k, v := range kvMap {
		addDoc(t, k, v, index, client)
	}

	kvMap["ab"] = []byte("doc3")
	addDoc(t, "ab", kvMap["ab"], index, client)
	kvMap["a"] = []byte("doc1")
	addDoc(t, "a", kvMap["a"], index, client)
	kvMap["key3"] = []byte("value3")
	addDoc(t, "key3", kvMap["key3"], index, client)
	kvMap["key2"] = []byte("value2")
	addDoc(t, "key2", kvMap["key2"], index, client)

	return kvMap
}

func addBatchDocs(t *testing.T, batch *collection.Batch, client *mock.MockBeeClient) map[string][]byte {
	kvMap := make(map[string][]byte)
	kvMap["key1"] = []byte("value1")
	kvMap["key11"] = []byte("value11")
	kvMap["aa"] = []byte("doc2")
	kvMap["abc"] = []byte("doc4")
	kvMap["abcd"] = []byte("doc5")
	kvMap["aca"] = []byte("doc6")
	kvMap["ac"] = []byte("doc7")
	kvMap["acb"] = []byte("doc8")
	kvMap["aaa"] = []byte("doc9")
	kvMap["az"] = []byte("doc10")
	kvMap["acdb"] = []byte("doc11")
	kvMap["file1"] = []byte("doc12")
	kvMap["file1.jpg"] = []byte("doc13")
	kvMap["ab"] = []byte("doc3")
	kvMap["a"] = []byte("doc1")
	kvMap["key3"] = []byte("value3")
	kvMap["key2"] = []byte("value2")
	kvMap["ke77"] = []byte("batch doc1")
	kvMap["ke79"] = []byte("batch doc2")
	kvMap["a94"] = []byte("batch doc3")

	// add the documents
	for k, v := range kvMap {
		ref, err := client.UploadBlob(v, false, false)
		if err != nil {
			t.Fatalf("could not add doc %s:%s, %v", k, ref, err)
		}
		err = batch.Put(k, ref)
		if err != nil {
			t.Fatal(err)
		}
	}
	return kvMap
}
