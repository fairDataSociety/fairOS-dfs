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

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

func TestBatchIndex(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
	acc := account.New(logger)
	ai := acc.GetUserAccountInfo()
	_, _, err := acc.CreateUserAccount("password", "")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, logger)
	user := acc.GetAddress(account.UserAccountIndex)

	t.Run("batch-add-docs", func(t *testing.T) {
		// create a DB and open it
		index := createAndOpenIndex(t, "pod1", "testdb_batch_0", collection.StringIndex, fd, user, mockClient, ai, logger)
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
		index := createAndOpenIndex(t, "pod1", "testdb_batch_1", collection.StringIndex, fd, user, mockClient, ai, logger)

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
}
