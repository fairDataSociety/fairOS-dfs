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
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

func TestKeyValueStore(t *testing.T) {
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
	kvStore := collection.NewKeyValueStore(fd, ai, user, mockClient, logger)

	t.Run("create_kv_table_with_string_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_0", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}

		tables, err := kvStore.LoadKVTables()
		if err != nil {
			t.Fatal(err)
		}

		idxType, ok := tables["kv_table_0"]
		if !ok {
			t.Fatalf("table %s not found", "kv_table_0")
		}

		if idxType[0] != collection.StringIndex.String() {
			t.Fatalf("invalid index type")
		}

		// delete so that they dont show up in other testcases
		err = kvStore.DeleteKVTable("kv_table_0")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("create_kv_table_with_number_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_1", collection.NumberIndex)
		if err != nil {
			t.Fatal(err)
		}

		tables, err := kvStore.LoadKVTables()
		if err != nil {
			t.Fatal(err)
		}

		idxType, ok := tables["kv_table_1"]
		if !ok {
			t.Fatalf("table %s not found", "kv_table_1")
		}

		if idxType[0] != collection.NumberIndex.String() {
			t.Fatalf("invalid index type")
		}

		// delete so that they dont show up in other testcases
		err = kvStore.DeleteKVTable("kv_table_1")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("check_delete", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_2", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}

		err = kvStore.DeleteKVTable("kv_table_2")
		if err != nil {
			t.Fatal(err)
		}

		tables, err := kvStore.LoadKVTables()
		if err != nil {
			t.Fatal(err)
		}

		_, ok := tables["kv_table_2"]
		if ok {
			t.Fatalf("table %s  found", "kv_table_2")
		}
	})

	t.Run("create_multiple_kv_tables_and_delete", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_31", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.CreateKVTable("kv_table_32", collection.NumberIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.CreateKVTable("kv_table_33", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}

		tables, err := kvStore.LoadKVTables()
		if err != nil {
			t.Fatal(err)
		}

		// check all 3 tables for existance
		if len(tables) != 3 {
			t.Fatalf("tables length is not proper. expected %d got %d", 3, len(tables))
		}

		idxType, ok := tables["kv_table_31"]
		if !ok {
			t.Fatalf("table %s not found", "kv_table_31")
		}
		if idxType[0] != collection.StringIndex.String() {
			t.Fatalf("invalid index type")
		}

		idxType, ok = tables["kv_table_32"]
		if !ok {
			t.Fatalf("table %s not found", "kv_table_32")
		}
		if idxType[0] != collection.NumberIndex.String() {
			t.Fatalf("invalid index type")
		}

		idxType, ok = tables["kv_table_33"]
		if !ok {
			t.Fatalf("table %s not found", "kv_table_33")
		}
		if idxType[0] != collection.StringIndex.String() {
			t.Fatalf("invalid index type")
		}

		// delete the last table
		err = kvStore.DeleteKVTable("kv_table_33")
		if err != nil {
			t.Fatal(err)
		}

		tables, err = kvStore.LoadKVTables()
		if err != nil {
			t.Fatal(err)
		}

		// check the remaining tables
		if len(tables) != 2 {
			t.Fatalf("tables length is not proper. expected %d got %d", 2, len(tables))
		}
		idxType, ok = tables["kv_table_31"]
		if !ok {
			t.Fatalf("table %s not found", "kv_table_31")
		}
		if idxType[0] != collection.StringIndex.String() {
			t.Fatalf("invalid index type")
		}

		idxType, ok = tables["kv_table_32"]
		if !ok {
			t.Fatalf("table %s not found", "kv_table_32")
		}
		if idxType[0] != collection.NumberIndex.String() {
			t.Fatalf("invalid index type")
		}
	})

	t.Run("create_open_and_delete", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_4", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}

		// open the table
		err = kvStore.OpenKVTable("kv_table_4")
		if err != nil {
			t.Fatal(err)
		}

		// delete the opened table
		err = kvStore.DeleteKVTable("kv_table_4")
		if err != nil {
			t.Fatal(err)
		}

	})

	t.Run("delete_without_create", func(t *testing.T) {
		// delete the last table
		err = kvStore.DeleteKVTable("kv_table_5")
		if !errors.Is(err, collection.ErrKVTableNotPresent) {
			t.Fatal("was able to delete table without creating it")
		}
	})

	t.Run("open_table", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_6", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_6")
		if err != nil {
			t.Fatal(err)
		}

		// do a put to see if the table is opened
		err = kvStore.KVPut("kv_table_6", "key1", []byte("value1"))
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("open_without_create", func(t *testing.T) {
		err = kvStore.OpenKVTable("kv_table_7")
		if !errors.Is(err, collection.ErrKVTableNotPresent) {
			t.Fatal("was able to open table without creating it")
		}
	})

	t.Run("put_string_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_8", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_8")
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.KVPut("kv_table_8", "key1", []byte("value1"))
		if err != nil {
			t.Fatal(err)
		}

		// get the value to see if it is present
		columns, value, err := kvStore.KVGet("kv_table_8", "key1")
		if err != nil {
			t.Fatal(err)
		}
		if columns != nil {
			t.Fatalf("columns present without setting")
		}
		if bytes.Compare(value, []byte("value1")) != 0 {
			t.Fatal(err)
		}
	})

	t.Run("put_chinese_string_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_9", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_9")
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.KVPut("kv_table_9", "立顯榮朝士", []byte("value1"))
		if err != nil {
			t.Fatal(err)
		}

		// get the value to see if it is present
		columns, value, err := kvStore.KVGet("kv_table_9", "立顯榮朝士")
		if err != nil {
			t.Fatal(err)
		}
		if columns != nil {
			t.Fatalf("columns present without setting")
		}
		if bytes.Compare(value, []byte("value1")) != 0 {
			t.Fatal(err)
		}
	})

	t.Run("put_string_in_number_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_10", collection.NumberIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_10")
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.KVPut("kv_table_10", "key1", []byte("value1"))
		if !errors.Is(err, collection.ErrKVKeyNotANumber) {
			t.Fatal("invalid number given as key for a number index")
		}
	})

	t.Run("put_get_del_get_string_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_11", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_11")
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.KVPut("kv_table_11", "key1", []byte("value1"))
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.KVPut("kv_table_11", "key2", []byte("value2"))
		if err != nil {
			t.Fatal(err)
		}

		// get the value to see if it is present
		columns, value, err := kvStore.KVGet("kv_table_11", "key1")
		if err != nil {
			t.Fatal(err)
		}
		if columns != nil {
			t.Fatalf("columns present without setting")
		}
		if bytes.Compare(value, []byte("value1")) != 0 {
			t.Fatal(err)
		}

		// delete the key
		_, err = kvStore.KVDelete("kv_table_11", "key1")
		if err != nil {
			t.Fatal(err)
		}

		// get it again and make sure it is not there
		columns, value, err = kvStore.KVGet("kv_table_11", "key1")
		if !errors.Is(err, collection.ErrEntryNotFound) {
			t.Fatalf("found the deleted entry")
		}

	})

	t.Run("put_without_opening_table", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_12", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.KVPut("kv_table_12", "key1", []byte("value1"))
		if !errors.Is(err, collection.ErrKVTableNotOpened) {
			t.Fatalf("could put without opening the table")
		}
	})

	t.Run("get_non_existant_string_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_13", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_13")
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.KVPut("kv_table_13", "key1", []byte("value1"))
		if err != nil {
			t.Fatal(err)
		}

		// delete a non existant key
		_, err = kvStore.KVDelete("kv_table_13", "key2")
		if !errors.Is(err, collection.ErrEntryNotFound) {
			t.Fatalf("found a non existant entry")
		}
	})

	t.Run("batch_without_open", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_batch_1", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		columns := []string{"c1", "c2", "c3"}
		_, err = kvStore.KVBatch("kv_table_batch_1", columns)
		if !errors.Is(err, collection.ErrKVTableNotOpened) {
			t.Fatalf("found a non existant entry")
		}
	})

	t.Run("batch_columns_and_get_values", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_batch_2", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_batch_2")
		if err != nil {
			t.Fatal(err)
		}

		columns := []string{"c1", "c2", "c3"}
		batch, err := kvStore.KVBatch("kv_table_batch_2", columns)
		if err != nil {
			t.Fatal(err)
		}

		value := []byte("v1,v2,v3")
		err = batch.Put("key1", value)
		if err != nil {
			t.Fatal(err)
		}

		err = batch.Write()
		if err != nil {
			t.Fatal(err)
		}

		gotColumns, gotValue, err := kvStore.KVGet("kv_table_batch_2", "key1")
		if err != nil {
			t.Fatal(err)
		}

		// check the columns returned
		for i, c := range columns {
			if c != gotColumns[i] {
				t.Fatal(err)
			}
		}

		// also check the values returned
		if bytes.Compare(value, gotValue) != 0 {
			t.Fatal(err)
		}
	})

	t.Run("Iterate_string_keys", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_Itr_0", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_Itr_0")
		if err != nil {
			t.Fatal(err)
		}

		keys, values, err := addRandomStrings(t, kvStore, 1000, "kv_table_Itr_0")
		if err != nil {
			t.Fatal(err)
		}
		sortedKeys, sortedValues := sortLexicographically(t, keys, values)

		itr, err := kvStore.KVSeek("kv_table_Itr_0", "", "", -1)
		if err != nil {
			t.Fatal(err)
		}

		// check the order of the keys
		for i := 0; i < 1000; i++ {
			itr.Next()
			if itr.StringKey() != sortedKeys[i] {
				t.Fatal(err)
			}
			if bytes.Compare(itr.Value(), []byte(sortedValues[i])) != 0 {
				t.Fatal(err)
			}
		}
	})

	t.Run("Iterate_seek_limit_string_keys", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_Itr_1", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_Itr_1")
		if err != nil {
			t.Fatal(err)
		}

		keys, values, err := addRandomStrings(t, kvStore, 1000, "kv_table_Itr_1")
		if err != nil {
			t.Fatal(err)
		}
		sortedKeys, sortedValues := sortLexicographically(t, keys, values)

		itr, err := kvStore.KVSeek("kv_table_Itr_1", "B", "", 10)
		if err != nil {
			t.Fatal(err)
		}

		startIndex := 0
		for i := 0; i < 1000; i++ {
			if strings.HasPrefix(keys[i], "B") {
				startIndex = i
				break
			}
		}

		// check the order of the keys
		for i := startIndex; i < startIndex+10; i++ {
			itr.Next()
			if itr.StringKey() != sortedKeys[i] {
				t.Fatal(err)
			}
			if bytes.Compare(itr.Value(), []byte(sortedValues[i])) != 0 {
				t.Fatal(err)
			}
		}

		// do a ite.Next() after limit..to see that it should not return anything
		if itr.Next() {
			t.Fatalf("iterating beyond limit")
		}

	})

	t.Run("Iterate_seek_start_end_string_keys", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_Itr_2", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_Itr_2")
		if err != nil {
			t.Fatal(err)
		}

		keys, values, err := addRandomStrings(t, kvStore, 1000, "kv_table_Itr_2")
		if err != nil {
			t.Fatal(err)
		}
		sortedKeys, sortedValues := sortLexicographically(t, keys, values)

		itr, err := kvStore.KVSeek("kv_table_Itr_2", "B", "C", -1)
		if err != nil {
			t.Fatal(err)
		}

		startIndex := 0
		endIndex := 0
		for i := 0; i < 1000; i++ {
			if startIndex == 0 && strings.HasPrefix(keys[i], "B") {
				startIndex = i
			}
			if strings.HasPrefix(keys[i], "C") {
				endIndex = i
				break
			}
		}

		// check the order of the keys
		for i := startIndex; i < endIndex+1; i++ {
			itr.Next()
			if itr.StringKey() != sortedKeys[i] {
				t.Fatal(err)
			}
			if bytes.Compare(itr.Value(), []byte(sortedValues[i])) != 0 {
				t.Fatal(err)
			}
		}

		// do a ite.Next() after end..to see that it should not return anything
		if itr.Next() {
			t.Fatalf("iterating beyond end")
		}

	})

	t.Run("Iterate_string_of_numbers_keys", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_Itr_3", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_Itr_3")
		if err != nil {
			t.Fatal(err)
		}

		keys, values, err := addRandomNumbersAsString(t, kvStore, 1000, "kv_table_Itr_3")
		if err != nil {
			t.Fatal(err)
		}
		sortedKeys, sortedValues := sortLexicographically(t, keys, values)

		itr, err := kvStore.KVSeek("kv_table_Itr_3", "", "", -1)
		if err != nil {
			t.Fatal(err)
		}

		// check the order of the keys
		for i := 0; i < 1000; i++ {
			itr.Next()
			if itr.StringKey() != sortedKeys[i] {
				t.Fatal(err)
			}
			if bytes.Compare(itr.Value(), []byte(sortedValues[i])) != 0 {
				t.Fatal(err)
			}
		}
	})

	t.Run("Iterate_numbers_keys", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_Itr_4", collection.NumberIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_Itr_4")
		if err != nil {
			t.Fatal(err)
		}

		keys, values, err := addRandomNumbers(t, kvStore, 1000, "kv_table_Itr_4")
		if err != nil {
			t.Fatal(err)
		}
		sort.Ints(keys)
		sort.Ints(values)

		itr, err := kvStore.KVSeek("kv_table_Itr_4", "-1", "-1", -1)
		if err != nil {
			t.Fatal(err)
		}

		// check the order of the keys
		for i := 0; i < 1000; i++ {
			itr.Next()
			if itr.IntegerKey() != int64(keys[i]) {
				t.Fatal(err)
			}
			if bytes.Compare(itr.Value(), []byte(strconv.Itoa(values[i]))) != 0 {
				t.Fatal(err)
			}
		}
	})

	t.Run("Iterate_numbers_start_end_keys", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_Itr_5", collection.NumberIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_Itr_5")
		if err != nil {
			t.Fatal(err)
		}

		keys, values, err := addRandomNumbers(t, kvStore, 1000, "kv_table_Itr_5")
		if err != nil {
			t.Fatal(err)
		}
		sort.Ints(keys)
		sort.Ints(values)

		itr, err := kvStore.KVSeek("kv_table_Itr_5", "100", "200", -1)
		if err != nil {
			t.Fatal(err)
		}

		startIndex := 0
		endIndex := 0
		for i := 0; i < 1000; i++ {
			if startIndex == 0 && keys[i] >= 100 {
				startIndex = i
			}
			if keys[i] > 200 {
				endIndex = i
				break
			}
		}

		// check the order of the keys
		for i := startIndex; i < endIndex; i++ {
			itr.Next()
			if itr.IntegerKey() != int64(keys[i]) {
				t.Fatal(err)
			}
			if bytes.Compare(itr.Value(), []byte(strconv.Itoa(values[i]))) != 0 {
				t.Fatal(err)
			}
		}

		// do a ite.Next() after end..to see that it should not return anything
		if itr.Next() {
			t.Fatalf("iterating beyond end")
		}
	})

	t.Run("Iterate_numbers_start_and_limit_keys", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_Itr_6", collection.NumberIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_Itr_6")
		if err != nil {
			t.Fatal(err)
		}

		keys, values, err := addRandomNumbers(t, kvStore, 1000, "kv_table_Itr_6")
		if err != nil {
			t.Fatal(err)
		}
		sort.Ints(keys)
		sort.Ints(values)

		itr, err := kvStore.KVSeek("kv_table_Itr_6", "100", "-1", 100)
		if err != nil {
			t.Fatal(err)
		}

		startIndex := 0
		for i := 0; i < 1000; i++ {
			if startIndex == 0 && keys[i] >= 100 {
				startIndex = i
				break
			}
		}

		// check the order of the keys
		for i := startIndex; i < startIndex+100; i++ {
			itr.Next()
			if itr.IntegerKey() != int64(keys[i]) {
				t.Fatal(err)
			}
			if bytes.Compare(itr.Value(), []byte(strconv.Itoa(values[i]))) != 0 {
				t.Fatal(err)
			}
		}

		// do a ite.Next() after limit..to see that it should not return anything
		if itr.Next() {
			t.Fatalf("iterating beyond limit")
		}
	})

}

func addRandomStrings(t *testing.T, kvStore *collection.KeyValue, count int, tableName string) ([]string, []string, error) {
	var keys []string
	var values []string
	for i := 0; i < count; i++ {
	DUPLICATE:
		randStrLen := rand.Intn(15)
		key := getRandString(randStrLen)
		for _, k := range keys {
			if k == key {
				goto DUPLICATE
			}
		}

		err := kvStore.KVPut(tableName, key, []byte(key))
		if err != nil {
			return nil, nil, err
		}
		keys = append(keys, key)
		values = append(values, key)
	}
	return keys, values, nil
}

func addRandomNumbersAsString(t *testing.T, kvStore *collection.KeyValue, count int, tableName string) ([]string, []string, error) {
	var keys []string
	var values []string
	for i := 0; i < count; i++ {
	DUPLICATE:
		key := rand.Intn(10000)
		strKey := strconv.Itoa(key)
		for _, k := range keys {
			if k == strKey {
				goto DUPLICATE
			}
		}

		err := kvStore.KVPut(tableName, strKey, []byte(strKey))
		if err != nil {
			return nil, nil, err
		}
		keys = append(keys, strKey)
		values = append(values, strKey)
	}
	return keys, values, nil
}

func addRandomNumbers(t *testing.T, kvStore *collection.KeyValue, count int, tableName string) ([]int, []int, error) {
	var keys []int
	var values []int
	for i := 0; i < count; i++ {
	DUPLICATE:
		key := rand.Intn(10000)
		strKey := strconv.Itoa(key)
		for _, k := range keys {
			if k == key {
				goto DUPLICATE
			}
		}

		err := kvStore.KVPut(tableName, strKey, []byte(strKey))
		if err != nil {
			return nil, nil, err
		}
		keys = append(keys, key)
		values = append(values, key)
	}
	return keys, values, nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func getRandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
