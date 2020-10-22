package collection_test

import (
	"bytes"
	"errors"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"io/ioutil"
	"net/http"
	"testing"
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
		err := collection.CreateIndex("testdb0", "key", fd, acc.GetAddress(account.UserAccountIndex))
		if err != nil {
			t.Fatal(err)
		}
		index, err := collection.OpenIndex("testdb0", "key", fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient)
		if err != nil {
			t.Fatal(err)
		}
		kvMap := addLotOfDocs(t, index, mockClient)

		// close the index
		//index = nil

		// open the index again
		index1, err := collection.OpenIndex("testdb0", "key", fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient)
		if err != nil {
			t.Fatal(err)
		}

		for k, expectedValue := range kvMap {
			gotValue := getDoc(t, k, index1, mockClient)
			if !bytes.Equal(expectedValue, gotValue) {
				t.Fatalf("expected expectedValue %s got expectedValue %s", expectedValue, gotValue)
			}
		}

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
		err := collection.CreateIndex("testdb1", "key", fd, acc.GetAddress(account.UserAccountIndex))
		if err != nil {
			t.Fatal(err)
		}
		index, err := collection.OpenIndex("testdb1", "key", fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient)
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
		err := collection.CreateIndex("testdb2", "key", fd, acc.GetAddress(account.UserAccountIndex))
		if err != nil {
			t.Fatal(err)
		}
		index, err := collection.OpenIndex("testdb2", "key", fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient)
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
		err := collection.CreateIndex("testdb3", "key", fd, acc.GetAddress(account.UserAccountIndex))
		if err != nil {
			t.Fatal(err)
		}
		index, err := collection.OpenIndex("testdb3", "key", fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient)
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

	t.Run("add-docs-seek-iterrate", func(t *testing.T) {
		//  create and populate the index
		err := collection.CreateIndex("testdb4", "key", fd, acc.GetAddress(account.UserAccountIndex))
		if err != nil {
			t.Fatal(err)
		}
		index, err := collection.OpenIndex("testdb4", "key", fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient)
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
			value := getDoc(t, itr.Key(), index, mockClient)
			if !bytes.Equal(kvMap[itr.Key()], value) {
				t.Fatalf("expected value %s but got %s for the key %s", string(kvMap[itr.Key()]), string(value), itr.Key())
			}
			count++
		}

		if count != 7 {
			t.Fatalf("number of elements mismatch in iteration")
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
	kvMap["key2"] = []byte("value2")
	kvMap["key3"] = []byte("value3")
	kvMap["key11"] = []byte("value11")
	kvMap["a"] = []byte("doc1")
	kvMap["aa"] = []byte("doc2")
	kvMap["ab"] = []byte("doc3")
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
	return kvMap
}
