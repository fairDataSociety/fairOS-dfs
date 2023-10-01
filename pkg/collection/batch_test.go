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
	"io"
	"testing"

	mockpost "github.com/ethersphere/bee/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"
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

func TestBatchIndex(t *testing.T) {
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
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, logger)
	user := acc.GetAddress(account.UserAccountIndex)
	podPassword, _ := utils.GetRandString(pod.PasswordLength)
	t.Run("batch-add-docs", func(t *testing.T) {
		// create a DB and open it
		index := createAndOpenIndex(t, "pod1", "testdb_batch_0", podPassword, collection.StringIndex, fd, user, mockClient, ai, logger)
		// batch load and delete

		batch, err := collection.NewBatch(index)
		if err != nil {
			t.Fatal(err)
		}

		batchDocs := addBatchDocs(t, batch, mockClient)
		_, err = batch.Write("")
		if err != nil {
			t.Fatal(err)
		}

		// create the iterator
		itr, err := index.NewStringIterator("", "", 100)
		if err != nil {
			t.Fatal(err)
		}

		// iterate through the keys and check for the values returned
		count := 0
		for itr.Next() {
			value := getValue(t, itr.Value(), mockClient)
			if !bytes.Equal(batchDocs[itr.StringKey()], value) {
				t.Fatalf("expected value %s but got %s for the key %s", string(batchDocs[itr.StringKey()]), string(value), itr.StringKey())
			}
			count++
		}

		if len(batchDocs) != count {
			t.Fatalf("number of elements mismatch in iteration")
		}
	})

	t.Run("batch-add-docs", func(t *testing.T) {
		// create a DB and open it
		index := createAndOpenIndex(t, "pod1", "testdb_batch_1", podPassword, collection.StringIndex, fd, user, mockClient, ai, logger)

		// batch load and delete
		batch, err := collection.NewBatch(index)
		if err != nil {
			t.Fatal(err)
		}

		batchDocs := addBatchDocs(t, batch, mockClient)
		_, err = batch.Write("")
		if err != nil {
			t.Fatal(err)
		}

		// create the iterator
		itr, err := index.NewStringIterator("", "", 100)
		if err != nil {
			t.Fatal(err)
		}

		// iterate through the keys and check for the values returned
		count := 0
		for itr.Next() {
			value := getValue(t, itr.Value(), mockClient)
			if !bytes.Equal(batchDocs[itr.StringKey()], value) {
				t.Fatalf("expected value %s but got %s for the key %s", string(batchDocs[itr.StringKey()]), string(value), itr.StringKey())
			}
			count++
		}

		if len(batchDocs) != count {
			t.Fatalf("number of elements mismatch in iteration")
		}
	})

	t.Run("batch-add-del-docs", func(t *testing.T) {
		// create a DB and open it
		index := createAndOpenIndex(t, "pod1", "testdb_batch_2", podPassword, collection.StringIndex, fd, user, mockClient, ai, logger)

		// batch load and delete
		batch, err := collection.NewBatch(index)
		if err != nil {
			t.Fatal(err)
		}
		_ = addBatchDocs(t, batch, mockClient)
		_, err = batch.Write("")
		if err != nil {
			t.Fatal(err)
		}
		// create the iterator
		itr, err := index.NewStringIterator("", "", 100)
		if err != nil {
			t.Fatal(err)
		}

		// iterate through the keys and check for the values returned
		for itr.Next() {
			_, err = batch.Del(itr.StringKey())
			if err != nil {
				t.Fatal(err)
			}
		}
		_, err = batch.Write("")
		if err != nil {
			t.Fatal(err)
		}
		index2, err := collection.OpenIndex("pod1", "testdb_batch_2", "key", podPassword, fd, ai, user, mockClient, logger)
		if err != nil {
			t.Fatal(err)
		}
		// create the iterator
		itr2, err := index2.NewStringIterator("", "", 100)
		if err != nil {
			t.Fatal(err)
		}
		if itr2.Next() {
			t.Fatal("should be not element")
		}
	})
}
