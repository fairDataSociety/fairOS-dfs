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
	"fmt"
	"io/ioutil"
	"math/rand"
	"sort"
	"strconv"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

func TestIndexIterator(t *testing.T) {
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

	t.Run("iterate_all_string_keys", func(t *testing.T) {
		// create a DB and open it
		idx := createAndOpenIndex(t, "testdb0", collection.StringIndex, fd, user, mockClient, ai, logger)

		// add some documents and sort them lexicograpically
		actualCount := uint64(100)
		keys, values := addDocsForStringIteration(t, idx, actualCount)
		sortedKeys, sortedValues := sortLexicographically(t, keys, values)

		// iterate with "no end key" and "no limit"
		itr, err := idx.NewStringIterator("0", "", -1)
		if err != nil {
			t.Fatal(err)
		}

		// check the iteration is in order
		for i := 0; i < int(actualCount); i++ {
			if itr.Next() {
				key := sortedKeys[i]
				value := sortedValues[i]
				if itr.StringKey() != key {
					t.Fatalf("invalid key, expected %s got %s", key, itr.StringKey())
				}
				if string(itr.Value()) != value {
					t.Fatalf("invalid key, expected %s got %s", value, string(itr.Value()))
				}
			}
		}
	})

	t.Run("iterate_all_random_string_keys", func(t *testing.T) {
		// create a DB and open it
		idx := createAndOpenIndex(t, "testdb1", collection.StringIndex, fd, user, mockClient, ai, logger)

		// add some documents and sort them lexicograpically
		actualCount := uint64(100)
		keys, values := addDocsForRandomStringIteration(t, idx, actualCount)
		sortedKeys, sortedValues := sortLexicographically(t, keys, values)

		// iterate with "no end key" and "no limit"
		itr, err := idx.NewStringIterator("0", "", -1)
		if err != nil {
			t.Fatal(err)
		}

		// check the iteration is in order
		for i := 0; i < int(actualCount); i++ {
			if itr.Next() {
				key := sortedKeys[i]
				value := sortedValues[i]
				if itr.StringKey() != key {
					t.Fatalf("invalid key, expected %s got %s", key, itr.StringKey())
				}
				if string(itr.Value()) != value {
					t.Fatalf("invalid key, expected %s got %s", value, string(itr.Value()))
				}
			}
		}
	})

	t.Run("iterate_with_string_end_key", func(t *testing.T) {
		// create a DB and open it
		idx := createAndOpenIndex(t, "testdb2", collection.StringIndex, fd, user, mockClient, ai, logger)

		// add some documents and sort them lexicograpically
		actualCount := uint64(100)
		keys, values := addDocsForStringIteration(t, idx, actualCount)
		sortedKeys, sortedValues := sortLexicographically(t, keys, values)

		// iterate with "no end key" and "no limit"
		itr, err := idx.NewStringIterator("0", "20", -1)
		if err != nil {
			t.Fatal(err)
		}

		// check the iteration is in order until the end key
		for i := 0; i < 14; i++ {
			if itr.Next() {
				key := sortedKeys[i]
				value := sortedValues[i]
				if itr.StringKey() != key {
					t.Fatalf("invalid key, expected %s got %s", key, itr.StringKey())
				}
				if string(itr.Value()) != value {
					t.Fatalf("invalid key, expected %s got %s", value, string(itr.Value()))
				}
			}
		}

		// do a ite.Next() after end key..to see that it should not return anything
		if itr.Next() {
			t.Fatalf("iterating beyond end key")
		}
	})

	t.Run("iterate_with_string_end_key", func(t *testing.T) {
		// create a DB and open it
		idx := createAndOpenIndex(t, "testdb3", collection.StringIndex, fd, user, mockClient, ai, logger)

		// add some documents and sort them lexicograpically
		actualCount := uint64(100)
		keys, values := addDocsForStringIteration(t, idx, actualCount)
		sortedKeys, sortedValues := sortLexicographically(t, keys, values)

		// iterate with "no end key" and "no limit"
		itr, err := idx.NewStringIterator("00", "20", -1)
		if err != nil {
			t.Fatal(err)
		}

		// check the iteration is in order until the end key
		//skip the first key since "0" is lexicographically smaller than "00"
		for i := 1; i < 14; i++ {
			if itr.Next() {
				key := sortedKeys[i]
				value := sortedValues[i]
				if itr.StringKey() != key {
					t.Fatalf("invalid key, expected %s got %s", key, itr.StringKey())
				}
				if string(itr.Value()) != value {
					t.Fatalf("invalid key, expected %s got %s", value, string(itr.Value()))
				}
			}
		}

		// do a ite.Next() after end key..to see that it should not return anything
		if itr.Next() {
			t.Fatalf("iterating beyond end key")
		}
	})

	t.Run("iterate_with_string_keys_with_limit", func(t *testing.T) {
		// create a DB and open it
		idx := createAndOpenIndex(t, "testdb4", collection.StringIndex, fd, user, mockClient, ai, logger)

		// add some documents and sort them lexicograpically
		actualCount := uint64(100)
		keys, values := addDocsForStringIteration(t, idx, actualCount)
		sortedKeys, sortedValues := sortLexicographically(t, keys, values)

		// iterate with "no end key" and "no limit"
		itr, err := idx.NewStringIterator("0", "20", 10)
		if err != nil {
			t.Fatal(err)
		}

		// check the iteration is in order until the end key
		for i := 0; i < 10; i++ {
			if itr.Next() {
				key := sortedKeys[i]
				value := sortedValues[i]
				if itr.StringKey() != key {
					t.Fatalf("invalid key, expected %s got %s", key, itr.StringKey())
				}
				if string(itr.Value()) != value {
					t.Fatalf("invalid key, expected %s got %s", value, string(itr.Value()))
				}
			}
		}

		// do a ite.Next() after end key..to see that it should not return anything
		if itr.Next() {
			t.Fatalf("iterating beyng limit")
		}
	})

	t.Run("iterate_all_number_keys", func(t *testing.T) {
		// create a DB and open it
		idx := createAndOpenIndex(t, "testdb5", collection.NumberIndex, fd, user, mockClient, ai, logger)

		// add some documents and sort them lexicograpically
		actualCount := uint64(100)
		_, _ = addDocsForNumberIteration(t, idx, actualCount)

		// iterate with "no end key" and "no limit"
		itr, err := idx.NewIntIterator(0, -1, -1)
		if err != nil {
			t.Fatal(err)
		}

		// check the iteration is in order
		for i := 0; i < int(actualCount); i++ {
			if itr.Next() {
				key := strconv.Itoa(i)
				value := "value" + strconv.Itoa(i)
				if itr.IntegerKey() != int64(i) {
					t.Fatalf("invalid key, expected %s got %d", key, itr.IntegerKey())
				}
				if string(itr.Value()) != value {
					t.Fatalf("invalid key, expected %s got %s", value, string(itr.Value()))
				}
			}
		}
	})

	t.Run("iterate_all_number_random_keys", func(t *testing.T) {
		// create a DB and open it
		idx := createAndOpenIndex(t, "testdb6", collection.NumberIndex, fd, user, mockClient, ai, logger)

		// add some documents and sort them lexicograpically
		actualCount := uint64(100)
		keys, values := addDocsForRandomNumberIteration(t, idx, actualCount)
		sort.Ints(keys)
		sort.Ints(values)

		// iterate with "no end key" and "no limit"
		itr, err := idx.NewIntIterator(0, -1, -1)
		if err != nil {
			t.Fatal(err)
		}

		// check the iteration is in order
		for i, k := range keys {
			if itr.Next() {
				key := fmt.Sprintf("%d", k)
				value := fmt.Sprintf("%d", values[i])
				if itr.IntegerKey() != int64(k) {
					t.Fatalf("invalid key, expected %s got %s", key, itr.StringKey())
				}
				if string(itr.Value()) != value {
					t.Fatalf("invalid key, expected %s got %s", value, string(itr.Value()))
				}
			}
		}
	})

	t.Run("iterate_with_numbers_end_key", func(t *testing.T) {
		// create a DB and open it
		idx := createAndOpenIndex(t, "testdb7", collection.NumberIndex, fd, user, mockClient, ai, logger)

		// add some documents and sort them lexicograpically
		actualCount := uint64(100)
		_, _ = addDocsForNumberIteration(t, idx, actualCount)

		// iterate with "no end key" and "no limit"
		itr, err := idx.NewIntIterator(0, 22, -1)
		if err != nil {
			t.Fatal(err)
		}

		// check the iteration is in order
		for i := 0; i <= 22; i++ {
			if itr.Next() {
				key := fmt.Sprintf("%d", i)
				value := "value" + strconv.Itoa(i)
				if itr.IntegerKey() != int64(i) {
					t.Fatalf("invalid key, expected %s got %d", key, itr.IntegerKey())
				}
				if string(itr.Value()) != value {
					t.Fatalf("invalid key, expected %s got %s", value, string(itr.Value()))
				}
			}
		}

		// do a ite.Next() after end key..to see that it should not return anything
		if itr.Next() {
			t.Fatalf("iterating beyond end key")
		}
	})

	t.Run("iterate_with_numbers_keys_with_limit", func(t *testing.T) {
		// create a DB and open it
		idx := createAndOpenIndex(t, "testdb8", collection.NumberIndex, fd, user, mockClient, ai, logger)

		// add some documents and sort them lexicograpically
		actualCount := uint64(100)
		_, _ = addDocsForNumberIteration(t, idx, actualCount)

		// iterate with "no end key" and "no limit"
		itr, err := idx.NewIntIterator(0, 22, 10)
		if err != nil {
			t.Fatal(err)
		}

		// check the iteration is in order
		for i := 0; i < 10; i++ {
			if itr.Next() {
				key := strconv.Itoa(i)
				value := "value" + strconv.Itoa(i)
				if itr.IntegerKey() != int64(i) {
					t.Fatalf("invalid key, expected %s got %d", key, itr.IntegerKey())
				}
				if string(itr.Value()) != value {
					t.Fatalf("invalid key, expected %s got %s", value, string(itr.Value()))
				}
			}
		}

		// do a ite.Next() after end key..to see that it should not return anything
		if itr.Next() {
			t.Fatalf("iterating beyond end key")
		}
	})

}

func addDocsForStringIteration(t *testing.T, idx *collection.Index, actualCount uint64) ([]string, []string) {
	var keys []string
	var values []string
	for i := 0; i < int(actualCount); i++ {
		key := strconv.Itoa(i)
		value := "value" + strconv.Itoa(i)
		putDocInIndex(t, idx, key, value, collection.StringIndex)
		keys = append(keys, key)
		values = append(values, value)
	}
	return keys, values
}

func addDocsForNumberIteration(t *testing.T, idx *collection.Index, actualCount uint64) ([]string, []string) {
	var keys []string
	var values []string
	for i := 0; i < int(actualCount); i++ {
		key := strconv.Itoa(i)
		value := "value" + strconv.Itoa(i)
		putDocInIndex(t, idx, key, value, collection.NumberIndex)
		keys = append(keys, key)
		values = append(values, value)
	}
	return keys, values
}

func addDocsForRandomStringIteration(t *testing.T, idx *collection.Index, actualCount uint64) ([]string, []string) {
	var keys []int
	var values []int
	for i := 0; len(keys) < int(actualCount); i++ {
	DUPLICATE:
		a := rand.Intn(1000)
		for _, k := range keys {
			if k == a {
				goto DUPLICATE
			}
		}
		key := strconv.Itoa(a)
		value := strconv.Itoa(a)
		putDocInIndex(t, idx, key, value, collection.StringIndex)
		keys = append(keys, a)
		values = append(values, a)
	}

	var stringKeys []string
	var stringValues []string
	for _, k := range keys {
		stringKeys = append(stringKeys, strconv.Itoa(k))
	}
	for _, v := range values {
		stringValues = append(stringValues, strconv.Itoa(v))
	}
	return stringKeys, stringValues
}

func addDocsForRandomNumberIteration(t *testing.T, idx *collection.Index, actualCount uint64) ([]int, []int) {
	var keys []int
	var values []int
	for i := 0; len(keys) < int(actualCount); i++ {
	DUPLICATE:
		a := rand.Intn(1000)
		for _, k := range keys {
			if k == a {
				goto DUPLICATE
			}
		}
		key := strconv.Itoa(a)
		value := strconv.Itoa(a)
		putDocInIndex(t, idx, key, value, collection.NumberIndex)
		keys = append(keys, a)
		values = append(values, a)
	}
	return keys, values
}

func sortLexicographically(t *testing.T, keys []string, values []string) ([]string, []string) {
	sort.Slice(keys, func(i int, j int) bool {
		return keys[i] < keys[j]
	})
	sort.Slice(values, func(i int, j int) bool {
		return values[i] < values[j]
	})
	return keys, values
}

func createAndOpenIndex(t *testing.T, collectionName string, indexType collection.IndexType, fd *feed.API, user utils.Address,
	client blockstore.Client, ai *account.Info, logger logging.Logger) *collection.Index {
	err := collection.CreateIndex(collectionName, "key", indexType, fd, user, client)
	if err != nil {
		t.Fatal(err)
	}
	idx, err := collection.OpenIndex(collectionName, "key", fd, ai, user, client, logger)
	if err != nil {
		t.Fatal(err)
	}
	return idx
}
