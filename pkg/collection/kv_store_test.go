package collection_test

import (
	"bytes"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"io/ioutil"
	"testing"
)

type document struct {
	Name    string `json:"id"`
	Country string `json:"country"`
}

func TestStore_AddDocument(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(ioutil.Discard, 0)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("password", "")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, logger)
	db, err := collection.NewDB("testdb", fd, acc.GetUserAccountInfo(), acc.GetAddress(account.UserAccountIndex), mockClient)
	if err != nil {
		t.Fatal(err)
	}

	// Initialize the values
	kvMap := make(map[string][]byte)
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
		err := db.Put(k, v)
		if err != nil {
			t.Fatalf("doc not inserted: %s:%s, %v", k, v, err)
		}
	}

	// create the iterator
	itr, err := db.NewIterator("", "", 100)
	if err != nil {
		t.Fatal(err)
	}

	// iterate through the keys and check for the values returned
	for itr.Next() {
		if !bytes.Equal(kvMap[itr.Key()], itr.Value()) {
			t.Fatalf("expected value %s but got %s for the key %s", string(kvMap[itr.Key()]), string(itr.Value()), itr.Key())
		}
	}
}
