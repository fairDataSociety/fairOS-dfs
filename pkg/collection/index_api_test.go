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
	"io"
	"net/http"
	"testing"

	mockpost "github.com/ethersphere/bee/v2/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/v2/pkg/storer/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/sirupsen/logrus"

	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

func TestIndexAPI(t *testing.T) {
	storer := mockstorer.New()
	beeUrl := mock.NewTestBeeServer(t, mock.TestServerOptions{
		Storer:          storer,
		PreventRedirect: true,
		Post:            mockpost.New(mockpost.WithAcceptAll()),
	})

	logger := logging.New(io.Discard, logrus.DebugLevel)
	mockClient := bee.NewBeeClient(beeUrl, mock.BatchOkStr, true, logger)
	acc := account.New(logger)
	ai := acc.GetUserAccountInfo()
	_, _, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)
	user := acc.GetAddress(account.UserAccountIndex)
	podPassword, _ := utils.GetRandString(pod.PasswordLength)
	t.Run("get-doc", func(t *testing.T) {
		// create a DB and open it
		index := createAndOpenIndex(t, "pod1", "testdb_api_0", podPassword, collection.StringIndex, fd, user, mockClient, ai, logger)
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
	})

	t.Run("get-count", func(t *testing.T) {
		// create a DB and open it
		index := createAndOpenIndex(t, "pod1", "testdb_api_1", podPassword, collection.StringIndex, fd, user, mockClient, ai, logger)
		kvMap := addLotOfDocs(t, index, mockClient)

		// find the count
		count, err := index.CountIndex(podPassword)
		if err != nil {
			t.Fatal(err)
		}

		if uint64(len(kvMap)) != count {
			t.Fatal(err)
		}
	})

	t.Run("get-doc-del-doc-get-doc", func(t *testing.T) {
		// create a DB and open it
		index := createAndOpenIndex(t, "pod1", "testdb_api_2", podPassword, collection.StringIndex, fd, user, mockClient, ai, logger)
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
	})

	t.Run("get-multiple_docs", func(t *testing.T) {
		// create a DB and open it
		index := createAndOpenIndex(t, "pod1", "testdb_api_3", podPassword, collection.StringIndex, fd, user, mockClient, ai, logger)

		// add multiple values for the same key
		addDoc(t, "key1", []byte("value1"), index, mockClient, true)
		addDoc(t, "key1", []byte("value2"), index, mockClient, true)
		addDoc(t, "key1", []byte("value3"), index, mockClient, true)
		addDoc(t, "key1", []byte("value4"), index, mockClient, true)

		gotValues := getAllDocs(t, "key1", index, mockClient)
		if gotValues == nil {
			t.Fatalf("could not find any value for key")
		}

		if len(gotValues) != 4 {
			t.Fatalf("invalid number of values for given key")
		}

		if !bytes.Equal(gotValues[0], []byte("value1")) {
			t.Fatalf("invalid value")
		}
		if !bytes.Equal(gotValues[1], []byte("value2")) {
			t.Fatalf("invalid value")
		}
		if !bytes.Equal(gotValues[2], []byte("value3")) {
			t.Fatalf("invalid value")
		}
		if !bytes.Equal(gotValues[3], []byte("value4")) {
			t.Fatalf("invalid value")
		}
	})

}

func addDoc(t *testing.T, key string, value []byte, index *collection.Index, client *bee.Client, apnd bool) {
	ref, err := client.UploadBlob(value, 0, false)
	if err != nil {
		t.Fatalf("could not add doc %s:%s, %v", key, value, err)
	}
	err = index.Put(key, ref, collection.StringIndex, apnd)
	if err != nil {
		t.Fatalf("could not add doc in index: %s:%s, %v", key, ref, err)
	}
}

func getDoc(t *testing.T, key string, index *collection.Index, client *bee.Client) []byte {
	ref, err := index.Get(key)
	if err != nil {
		if errors.Is(err, collection.ErrEntryNotFound) {
			return nil
		}
		t.Fatal(err)
	}
	data, respCode, err := client.DownloadBlob(ref[0])
	if err != nil {
		t.Fatal(err)
	}
	if respCode != http.StatusOK {
		t.Fatalf("invalid response code")
	}
	return data
}
func getAllDocs(t *testing.T, key string, index *collection.Index, client *bee.Client) [][]byte {
	refs, err := index.Get(key)
	if err != nil {
		if errors.Is(err, collection.ErrEntryNotFound) {
			return nil
		}
		t.Fatal(err)
	}

	var data [][]byte
	for _, ref := range refs {
		buf, respCode, err := client.DownloadBlob(ref)
		if err != nil {
			t.Fatal(err)
		}
		if respCode != http.StatusOK {
			t.Fatalf("invalid response code")
		}
		data = append(data, buf)
	}
	return data
}

func getValue(t *testing.T, ref []byte, client *bee.Client) []byte {
	data, respCode, err := client.DownloadBlob(ref)
	if err != nil {
		t.Fatal(err)
	}
	if respCode != http.StatusOK {
		t.Fatalf("invalid response code")
	}
	return data
}

func delDoc(t *testing.T, key string, index *collection.Index, client *bee.Client) []byte {
	ref, err := index.Delete(key)
	if err != nil {
		t.Fatal(err)
	}
	data, respCode, err := client.DownloadBlob(ref[0])
	if err != nil {
		t.Fatal(err)
	}
	if respCode != http.StatusOK {
		t.Fatalf("invalid response code")
	}
	return data

}

func addLotOfDocs(t *testing.T, index *collection.Index, client *bee.Client) map[string][]byte {
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
		addDoc(t, k, v, index, client, false)
	}

	kvMap["ab"] = []byte("doc3")
	addDoc(t, "ab", kvMap["ab"], index, client, false)
	kvMap["a"] = []byte("doc1")
	addDoc(t, "a", kvMap["a"], index, client, false)
	kvMap["key3"] = []byte("value3")
	addDoc(t, "key3", kvMap["key3"], index, client, false)
	kvMap["key2"] = []byte("value2")
	addDoc(t, "key2", kvMap["key2"], index, client, false)

	return kvMap
}

func addBatchDocs(t *testing.T, batch *collection.Batch, client *bee.Client) map[string][]byte {
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
		ref, err := client.UploadBlob(v, 0, false)
		if err != nil {
			t.Fatalf("could not add doc %s:%s, %v", k, ref, err)
		}
		err = batch.Put(k, ref, false, false)
		if err != nil {
			t.Fatal(err)
		}
	}
	return kvMap
}
