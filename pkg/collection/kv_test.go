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
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func TestKeyValueStore(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
	acc := account.New(logger)
	ai := acc.GetUserAccountInfo()
	_, _, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, logger)
	user := acc.GetAddress(account.UserAccountIndex)
	kvStore := collection.NewKeyValueStore("pod1", fd, ai, user, mockClient, logger)
	podPassword, _ := utils.GetRandString(pod.PodPasswordLength)
	t.Run("table_not_opened", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_1314", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}

		err = kvStore.CreateKVTable("kv_table_1314", podPassword, collection.StringIndex)
		if !errors.Is(err, collection.ErrKvTableAlreadyPresent) {
			t.Fatal("table should be already present")
		}

		_, _, _, err = kvStore.KVGetNext("kv_table_1314")
		if !errors.Is(err, collection.ErrKVTableNotOpened) {
			t.Fatal("open table")
		}
		err = kvStore.OpenKVTable("kv_table_1314", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		// delete so that they dont show up in other testcases
		err = kvStore.DeleteKVTable("kv_table_1314", podPassword)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("nil_itr", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_1312", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_1312", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		_, _, _, err = kvStore.KVGetNext("kv_table_1312")
		if !errors.Is(err, collection.ErrKVNilIterator) {
			t.Fatal("found iterator")
		}

		// delete so that they dont show up in other testcases
		err = kvStore.DeleteKVTable("kv_table_1312", podPassword)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("create_kv_table_with_string_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_0", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}

		tables, err := kvStore.LoadKVTables(podPassword)
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
		err = kvStore.DeleteKVTable("kv_table_0", podPassword)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("create_kv_table_with_number_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_1", podPassword, collection.NumberIndex)
		if err != nil {
			t.Fatal(err)
		}

		tables, err := kvStore.LoadKVTables(podPassword)
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
		err = kvStore.DeleteKVTable("kv_table_1", podPassword)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("check_delete", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_2", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}

		err = kvStore.DeleteKVTable("kv_table_2", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		tables, err := kvStore.LoadKVTables(podPassword)
		if err != nil {
			t.Fatal(err)
		}

		_, ok := tables["kv_table_2"]
		if ok {
			t.Fatalf("table %s  found", "kv_table_2")
		}
	})

	t.Run("create_multiple_kv_tables_and_delete", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_31", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.CreateKVTable("kv_table_32", podPassword, collection.NumberIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.CreateKVTable("kv_table_33", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}

		tables, err := kvStore.LoadKVTables(podPassword)
		if err != nil {
			t.Fatal(err)
		}

		// check all 3 tables for existence
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
		err = kvStore.DeleteKVTable("kv_table_33", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		tables, err = kvStore.LoadKVTables(podPassword)
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
		err := kvStore.CreateKVTable("kv_table_4", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}

		// open the table
		err = kvStore.OpenKVTable("kv_table_4", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		// delete the opened table
		err = kvStore.DeleteKVTable("kv_table_4", podPassword)
		if err != nil {
			t.Fatal(err)
		}

	})

	t.Run("delete_without_create", func(t *testing.T) {
		// delete the last table
		err = kvStore.DeleteKVTable("kv_table_5", podPassword)
		if !errors.Is(err, collection.ErrKVTableNotPresent) {
			t.Fatal("was able to delete table without creating it")
		}
	})

	t.Run("open_table", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_6", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_6", podPassword)
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
		err = kvStore.OpenKVTable("kv_table_7", podPassword)
		if !errors.Is(err, collection.ErrKVTableNotPresent) {
			t.Fatal("was able to open table without creating it")
		}
	})

	t.Run("put_string_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_8", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_8", podPassword)
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
		if !bytes.Equal(value, []byte("value1")) {
			t.Fatal(err)
		}

		countObject, err := kvStore.KVCount("kv_table_8", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		if countObject.Count != 1 {
			t.Fatal("kv count value should be one")
		}
	})

	t.Run("put_bytes_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_bytes", podPassword, collection.BytesIndex)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = kvStore.KVGet("kv_table_bytes", "key1")
		if !errors.Is(err, collection.ErrKVTableNotOpened) {
			t.Fatal("kv table open")
		}
		err = kvStore.OpenKVTable("kv_table_bytes", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.KVPut("kv_table_bytes", "key1", []byte("value1"))
		if err != nil {
			t.Fatal(err)
		}

		// get the value to see if it is present
		columns, value, err := kvStore.KVGet("kv_table_bytes", "key1")
		if err != nil {
			t.Fatal(err)
		}
		if columns != nil {
			t.Fatalf("columns present without setting")
		}
		if !bytes.Equal(value, []byte("value1")) {
			t.Fatal(err)
		}

		countObject, err := kvStore.KVCount("kv_table_bytes", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		if countObject.Count != 1 {
			t.Fatal("kv count value should be one")
		}
	})

	t.Run("put_chinese_string_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_9", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_9", podPassword)
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
		if !bytes.Equal(value, []byte("value1")) {
			t.Fatal("values do not match", string(value), "value1")
		}
	})

	t.Run("put_string_in_number_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_10", podPassword, collection.NumberIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_10", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.KVPut("kv_table_10", "key1", []byte("value1"))
		if !errors.Is(err, collection.ErrKVKeyNotANumber) {
			t.Fatal("invalid number given as key for a number index")
		}
	})

	t.Run("put_get_del_get_string_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_11", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_11", podPassword)
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
		if !bytes.Equal(value, []byte("value1")) {
			t.Fatal("values do not match", string(value), "value1")
		}

		// delete the key
		_, err = kvStore.KVDelete("kv_table_11", "key1")
		if err != nil {
			t.Fatal(err)
		}

		// get it again and make sure it is not there
		_, _, err = kvStore.KVGet("kv_table_11", "key1")
		if !errors.Is(err, collection.ErrEntryNotFound) {
			t.Fatalf("found the deleted entry")
		}

	})

	t.Run("put_without_opening_table", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_12", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.KVPut("kv_table_12", "key1", []byte("value1"))
		if !errors.Is(err, collection.ErrKVTableNotOpened) {
			t.Fatalf("could put without opening the table")
		}
	})

	t.Run("delete_non_existent_string_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_13", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_13", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.KVPut("kv_table_13", "key1", []byte("value1"))
		if err != nil {
			t.Fatal(err)
		}

		// delete a non existent key
		_, err = kvStore.KVDelete("kv_table_13", "key2")
		if !errors.Is(err, collection.ErrEntryNotFound) {
			t.Fatalf("found a non existent entry")
		}
	})

	t.Run("batch_without_open", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_batch_1", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		columns := []string{"c1", "c2", "c3"}
		_, err = kvStore.KVBatch("kv_table_batch_1", columns)
		if !errors.Is(err, collection.ErrKVTableNotOpened) {
			t.Fatalf("found a non existent entry")
		}
	})

	t.Run("batch_columns_and_get_values", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_batch_2", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_batch_2", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		columns := []string{"c1", "c2", "c3"}
		batch, err := kvStore.KVBatch("kv_table_batch_2", columns)
		if err != nil {
			t.Fatal(err)
		}

		value := []byte("v1,v2,v3")
		err = batch.Put("key1", value, false, false)
		if err != nil {
			t.Fatal(err)
		}

		_, err = batch.Write("")
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
				t.Fatal("columns do not match", c, gotColumns[i])
			}
		}

		// also check the values returned
		if !bytes.Equal(value, gotValue) {
			t.Fatal("values do not match", string(value), string(gotValue))
		}
	})

	t.Run("batch_put_columns_and_get_values", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_batch_9", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_batch_9", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		columns := []string{"c1", "c2", "c3"}
		batch, err := kvStore.KVBatch("kv_table_batch_9", columns)
		if err != nil {
			t.Fatal(err)
		}

		value := []byte("v1,v2,v3")
		err = kvStore.KVBatchPut(batch, "key1", value)
		if err != nil {
			t.Fatal(err)
		}

		err = batch.Put("key1", value, false, false)
		if err != nil {
			t.Fatal(err)
		}

		err = kvStore.KVBatchWrite(batch)
		if err != nil {
			t.Fatal(err)
		}

		gotColumns, gotValue, err := kvStore.KVGet("kv_table_batch_9", "key1")
		if err != nil {
			t.Fatal(err)
		}

		// check the columns returned
		for i, c := range columns {
			if c != gotColumns[i] {
				t.Fatal("columns do not match", c, gotColumns[i])
			}
		}

		// also check the values returned
		if !bytes.Equal(value, gotValue) {
			t.Fatal("values do not match", string(value), string(gotValue))
		}
	})

	t.Run("count_columns_and_get_values", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_batch_count", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		countObject, err := kvStore.KVCount("kv_table_batch_count", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		if countObject.Count != 0 {
			t.Fatal("count should be zero")
		}
	})

	t.Run("Iterate_string_keys", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_Itr_0", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_Itr_0", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		keys, values, err := addRandomStrings(t, kvStore, 100, "kv_table_Itr_0")
		if err != nil {
			t.Fatal(err)
		}
		sortedKeys, sortedValues := sortLexicographically(t, keys, values)

		itr, err := kvStore.KVSeek("kv_table_Itr_0", "", "", -1)
		if err != nil {
			t.Fatal(err)
		}

		// check the order of the keys
		for i := 0; i < 100; i++ {
			itr.Next()
			if itr.StringKey() != sortedKeys[i] {
				t.Fatal("keys do not match", itr.StringKey(), sortedKeys[i])
			}
			if !bytes.Equal(itr.Value(), []byte(sortedValues[i])) {
				t.Fatal("values do not match", string(itr.Value()), sortedValues[i])
			}
		}
	})

	t.Run("Iterate_seek_limit_string_keys", func(t *testing.T) {
		tableNo := 0
	research:
		tableNo++
		err := kvStore.CreateKVTable(fmt.Sprintf("kv_table_Itr_01%d", tableNo), podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable(fmt.Sprintf("kv_table_Itr_01%d", tableNo), podPassword)
		if err != nil {
			t.Fatal(err)
		}
		keys, values, err := addRandomStrings(t, kvStore, 100, fmt.Sprintf("kv_table_Itr_01%d", tableNo))
		if err != nil {
			t.Fatal(err)
		}
		sortedKeys, sortedValues := sortLexicographically(t, keys, values)

		itr, err := kvStore.KVSeek(fmt.Sprintf("kv_table_Itr_01%d", tableNo), "B", "", 10)
		if err != nil {
			t.Fatal(err)
		}
		matched := false
		startIndex := 0
		for i := 0; i < 100; i++ {
			if strings.HasPrefix(keys[i], "B") {
				matched = true
				startIndex = i
				break
			}
		}
		if !matched {
			goto research
		}

		// check the order of the keys
		for i := startIndex; i < startIndex+10; i++ {
			itr.Next()
			if itr.StringKey() != sortedKeys[i] {
				t.Fatalf("key mismatch: %s : %s\n", itr.StringKey(), sortedKeys[i])
			}
			if !bytes.Equal(itr.Value(), []byte(sortedValues[i])) {
				t.Fatalf("value mismatch: %s : %s\n", itr.StringKey(), sortedKeys[i])
			}
		}

		// do a ite.Next() after limit..to see that it should not return anything
		if itr.Next() {
			t.Fatalf("iterating beyond limit")
		}

	})

	t.Run("Iterate_seek_start_end_string_keys", func(t *testing.T) {
		tableNo := 0
	research:
		tableNo++
		err := kvStore.CreateKVTable(fmt.Sprintf("kv_table_Itr_1%d", tableNo), podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable(fmt.Sprintf("kv_table_Itr_1%d", tableNo), podPassword)
		if err != nil {
			t.Fatal(err)
		}

		keys, values, err := addRandomStrings(t, kvStore, 100, fmt.Sprintf("kv_table_Itr_1%d", tableNo))
		if err != nil {
			t.Fatal(err)
		}
		sortedKeys, sortedValues := sortLexicographically(t, keys, values)

		matched := false
		startIndex := 0
		endIndex := 0

		startPrefix := "B"
		endPrefix := "C"
		for i := 0; i < 100; i++ {
			if startIndex == 0 && strings.HasPrefix(keys[i], startPrefix) {
				matched = true
				startIndex = i
			}
			if strings.HasPrefix(keys[i], endPrefix) {
				matched = true
				if startIndex == 0 {
					startIndex = i
					startPrefix = endPrefix
					endPrefix = "E"
				} else {
					endIndex = i
					break
				}
			}
		}
		if !matched {
			goto research
		}
		itr, err := kvStore.KVSeek(fmt.Sprintf("kv_table_Itr_1%d", tableNo), startPrefix, endPrefix, -1)
		if err != nil {
			t.Fatal(err)
		}
		if startIndex > endIndex {
			goto research
		}
		// check the order of the keys
		for i := startIndex; i < endIndex; i++ {
			itr.Next()
			if itr.StringKey() != sortedKeys[i] {
				t.Fatal("keys do not match", itr.StringKey(), sortedKeys[i])
			}
			if !bytes.Equal(itr.Value(), []byte(sortedValues[i])) {
				t.Fatal("values do not match", string(itr.Value()), sortedValues[i])
			}
		}

		// do a ite.Next() after end..to see that it should not return anything
		if itr.Next() {
			t.Fatalf("iterating beyond end %s %v", itr.StringKey(), string(itr.Value()))
		}

	})

	t.Run("Iterate_seek_start_end_string_keys_over_a_known_failing_keys", func(t *testing.T) {
		tableNo := 486
		err := kvStore.CreateKVTable(fmt.Sprintf("kv_table_Itr_1%d", tableNo), podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable(fmt.Sprintf("kv_table_Itr_1%d", tableNo), podPassword)
		if err != nil {
			t.Fatal(err)
		}
		list := []string{
			"0bL9qTuIq",
			"1KxCHfroi",
			"4",
			"4AwtTa0",
			"4RVqksE",
			"5EQ3A5OEu3Vgn",
			"5U",
			"6zJ",
			"7UKzdnTrve5",
			"7lRJm1js",
			"94ieVmIfkv",
			"97MFodQlrV9p",
			"B9KnfkYw",
			"BmizOfhSl",
			"C",
			"D2IsxTBXGzs5",
			"DQCUdYBL2xDT",
		}
		for _, i := range list {
			err = kvStore.KVPut(fmt.Sprintf("kv_table_Itr_1%d", tableNo), i, []byte(i))
			if err != nil {
				t.Fatal(err)
			}
		}

		startIndex := 0
		endIndex := 0

		startPrefix := "B"
		endPrefix := "C"
		for i := 0; i < 100; i++ {
			if startIndex == 0 && strings.HasPrefix(list[i], startPrefix) {
				startIndex = i
			}
			if strings.HasPrefix(list[i], endPrefix) {
				if startIndex == 0 {
					startIndex = i
					startPrefix = endPrefix
					endPrefix = "E"
				} else {
					endIndex = i
					break
				}
			}
		}
		itr, err := kvStore.KVSeek(fmt.Sprintf("kv_table_Itr_1%d", tableNo), startPrefix, endPrefix, -1)
		if err != nil {
			t.Fatal(err)
		}
		// check the order of the keys
		for i := startIndex; i < endIndex; i++ {
			itr.Next()
			if itr.StringKey() != list[i] {
				t.Fatal("keys do not match", itr.StringKey(), list[i])
			}
			if !bytes.Equal(itr.Value(), []byte(list[i])) {
				t.Fatal("values do not match", string(itr.Value()), list[i])
			}
		}

		// do a ite.Next() after end..to see that it should not return anything
		if itr.Next() {
			t.Fatalf("iterating beyond end %s %v", itr.StringKey(), string(itr.Value()))
		}

	})

	t.Run("Iterate_string_of_numbers_keys", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_Itr_3", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_Itr_3", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		keys, values, err := addRandomNumbersAsString(t, kvStore, 100, "kv_table_Itr_3")
		if err != nil {
			t.Fatal(err)
		}
		sortedKeys, sortedValues := sortLexicographically(t, keys, values)

		itr, err := kvStore.KVSeek("kv_table_Itr_3", "", "", -1)
		if err != nil {
			t.Fatal(err)
		}

		// check the order of the keys
		for i := 0; i < 100; i++ {
			itr.Next()
			if itr.StringKey() != sortedKeys[i] {
				t.Fatal("keys do not match", itr.StringKey(), sortedKeys[i])
			}
			if !bytes.Equal(itr.Value(), []byte(sortedValues[i])) {
				t.Fatal("values do not match", string(itr.Value()), sortedValues[i])
			}
		}
	})

	t.Run("Iterate_numbers_keys", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_Itr_4", podPassword, collection.NumberIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_Itr_4", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		keys, values, err := addRandomNumbers(t, kvStore, 100, "kv_table_Itr_4")
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
		for i := 0; i < 100; i++ {
			itr.Next()
			if itr.IntegerKey() != int64(keys[i]) {
				t.Fatal("keys do not match", itr.StringKey(), keys[i])
			}
			if !bytes.Equal(itr.Value(), []byte(strconv.Itoa(values[i]))) {
				t.Fatal("values do not match", string(itr.Value()), keys[i])
			}
		}
	})

	t.Run("Iterate_numbers_start_end_keys", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_Itr_5", podPassword, collection.NumberIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_Itr_5", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		keys, values, err := addRandomNumbers(t, kvStore, 100, "kv_table_Itr_5")
		if err != nil {
			t.Fatal(err)
		}
		sort.Ints(keys)
		sort.Ints(values)

		itr, err := kvStore.KVSeek("kv_table_Itr_5", "10", "200", -1)
		if err != nil {
			t.Fatal(err)
		}

		startIndex := 0
		endIndex := 0
		startIndexDone := false
		for i := 0; i < 10; i++ {
			if !startIndexDone && keys[i] >= 10 {
				startIndex = i
				startIndexDone = true
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
				t.Fatal("keys do not match", itr.StringKey(), keys[i])
			}
			if !bytes.Equal(itr.Value(), []byte(strconv.Itoa(values[i]))) {
				t.Fatal("values do not match", string(itr.Value()), keys[i])
			}
		}

		// do a ite.Next() after end..to see that it should not return anything
		if itr.Next() {
			t.Fatalf("iterating beyond end")
		}
	})

	t.Run("Iterate_numbers_start_and_limit_keys", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_Itr_6", podPassword, collection.NumberIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_Itr_6", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		keys, values, err := addRandomNumbers(t, kvStore, 100, "kv_table_Itr_6")
		if err != nil {
			t.Fatal(err)
		}
		sort.Ints(keys)
		sort.Ints(values)

		startIndex := 0
		for i := 0; i < 10; i++ {
			if startIndex == 0 && keys[i] >= 50 {
				startIndex = i
				break
			}
		}

		itr, err := kvStore.KVSeek("kv_table_Itr_6", "50", "-1", 10)
		if err != nil && !errors.Is(err, collection.ErrEntryNotFound) {
			t.Fatal(err)
		}

		// check the order of the keys
		for i := startIndex; i < startIndex+10; i++ {
			itr.Next()
			if itr.IntegerKey() != int64(keys[i]) {
				t.Fatal("keys do not match", itr.StringKey(), keys[i])
			}
			if !bytes.Equal(itr.Value(), []byte(strconv.Itoa(values[i]))) {
				t.Fatal("values do not match", string(itr.Value()), keys[i])
			}
		}

		// do a ite.Next() after limit..to see that it should not return anything
		if itr.Next() {
			t.Fatalf("iterating beyond limit")
		}
	})

	t.Run("get_non_existent_string_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_1313", podPassword, collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_1313", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.KVPut("kv_table_1313", "key1", []byte("value1"))
		if err != nil {
			t.Fatal(err)
		}

		_, err = kvStore.KVSeek("kv_table_1313", "key1", "", -1)
		if err != nil {
			t.Fatal(err)
		}
		// this should have value
		_, _, _, err = kvStore.KVGetNext("kv_table_1313")
		if err != nil {
			t.Fatal(err)
		}
		// this should not have value
		_, _, _, err = kvStore.KVGetNext("kv_table_1313")
		if !errors.Is(err, collection.ErrNoNextElement) {
			t.Fatal("found a nonexistent key")
		}
	})

	t.Run("err_byte_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_1316", podPassword, collection.BytesIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_1316", podPassword)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.KVPut("kv_table_1316", "key1", []byte("value1"))
		if err != nil {
			t.Fatal(err)
		}

		_, err = kvStore.KVSeek("kv_table_1316", "key1", "", -1)
		if !errors.Is(err, collection.ErrKVIndexTypeNotSupported) {
			t.Fatal("unsupported index")
		}
	})

	t.Run("err_seek_list_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_1317", podPassword, collection.ListIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_1317", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		_, err = kvStore.KVSeek("kv_table_1317", "key1", "", -1)
		if !errors.Is(err, collection.ErrKVInvalidIndexType) {
			t.Fatal("invalid index")
		}
	})

	t.Run("err_seek_map_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_1318", podPassword, collection.MapIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_1318", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		_, err = kvStore.KVSeek("kv_table_1318", "key1", "", -1)
		if !errors.Is(err, collection.ErrKVInvalidIndexType) {
			t.Fatal("invalid index")
		}
	})

	t.Run("err_seek_invalid_index", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_1319", podPassword, collection.InvalidIndex)
		if err != nil {
			t.Fatal(err)
		}
		err = kvStore.OpenKVTable("kv_table_1319", podPassword)
		if err != nil {
			t.Fatal(err)
		}

		_, err = kvStore.KVSeek("kv_table_1319", "key1", "", -1)
		if !errors.Is(err, collection.ErrKVInvalidIndexType) {
			t.Fatal("invalid index")
		}
	})

	t.Run("seek_unopened_table", func(t *testing.T) {
		err := kvStore.CreateKVTable("kv_table_1320", podPassword, collection.ListIndex)
		if err != nil {
			t.Fatal(err)
		}

		_, err = kvStore.KVSeek("kv_table_1320", "key1", "", -1)
		if !errors.Is(err, collection.ErrKVTableNotOpened) {
			t.Fatal("table open")
		}
	})
}

func addRandomStrings(t *testing.T, kvStore *collection.KeyValue, count int, tableName string) ([]string, []string, error) {
	var keys []string
	var values []string
	for i := 0; i < count; i++ {
	DUPLICATE:
		bi, err := rand.Int(rand.Reader, big.NewInt(15))
		if err != nil {
			return nil, nil, err
		}
		randStrLen := int(bi.Int64())
		key, err := utils.GetRandString(randStrLen)
		if err != nil {
			return nil, nil, err
		}
		for _, k := range keys {
			if k == key {
				goto DUPLICATE
			}
		}

		err = kvStore.KVPut(tableName, key, []byte(key))
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
		bi, err := rand.Int(rand.Reader, big.NewInt(10000))
		if err != nil {
			return nil, nil, err
		}
		key := int(bi.Int64())
		strKey := strconv.Itoa(key)
		for _, k := range keys {
			if k == strKey {
				goto DUPLICATE
			}
		}

		err = kvStore.KVPut(tableName, strKey, []byte(strKey))
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
		bi, err := rand.Int(rand.Reader, big.NewInt(10000))
		if err != nil {
			return nil, nil, err
		}
		key := int(bi.Int64())
		strKey := strconv.Itoa(key)
		for _, k := range keys {
			if k == key {
				goto DUPLICATE
			}
		}

		err = kvStore.KVPut(tableName, strKey, []byte(strKey))
		if err != nil {
			return nil, nil, err
		}
		keys = append(keys, key)
		values = append(values, key)
	}
	return keys, values, nil
}
