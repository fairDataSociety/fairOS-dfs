/*
Copyright © 2020 FairOS Authors
push
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
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestDocument struct {
	ID        string            `json:"id"`
	FirstName string            `json:"first_name"`
	LastName  string            `json:"last_name"`
	Age       float64           `json:"age"`
	TagMap    map[string]string `json:"tag_map"`
	TagList   []string          `json:"tag_list"`
}

func TestDocumentStore(t *testing.T) {
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
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()

	file := f.NewFile("pod1", mockClient, fd, user, tm, logger)
	docStore := collection.NewDocumentStore("pod1", fd, ai, user, file, tm, mockClient, logger)
	podPassword, _ := utils.GetRandString(pod.PasswordLength)
	t.Run("create_document_db_errors", func(t *testing.T) {
		nilFd := feed.New(&account.Info{}, mockClient, logger)
		nilDocStore := collection.NewDocumentStore("pod1", nilFd, ai, user, file, tm, mockClient, logger)
		err := nilDocStore.CreateDocumentDB("docdb_err", podPassword, nil, true)
		if !errors.Is(err, collection.ErrReadOnlyIndex) {
			t.Fatal("should be readonly index")
		}

		// create a document DB
		createDocumentDBs(t, []string{"docdb_err"}, docStore, nil, podPassword)

		err = docStore.CreateDocumentDB("docdb_err", podPassword, nil, true)
		if !errors.Is(err, collection.ErrDocumentDBAlreadyPresent) {
			t.Fatal("db should be present already")
		}

		err = docStore.OpenDocumentDB("docdb_err", podPassword)
		require.NoError(t, err)
		err = docStore.CreateDocumentDB("docdb_err", podPassword, nil, true)
		if !errors.Is(err, collection.ErrDocumentDBAlreadyOpened) {
			t.Fatal("db should be opened already")
		}
	})

	t.Run("create_document_db", func(t *testing.T) {
		// create a document DB
		createDocumentDBs(t, []string{"docdb_0"}, docStore, nil, podPassword)

		// load the schem and check the count of simple indexes
		schema := loadSchemaAndCheckSimpleIndexCount(t, docStore, "docdb_0", podPassword, 1)

		// check the default index
		checkIndex(t, schema.SimpleIndexes[0], collection.DefaultIndexFieldName, collection.StringIndex)
	})

	t.Run("delete_document_db", func(t *testing.T) {
		// create multiple document DB
		createDocumentDBs(t, []string{"docdb_1_1", "docdb_1_2", "docdb_1_3"}, docStore, nil, podPassword)
		checkIfDBsExists(t, []string{"docdb_1_1", "docdb_1_2", "docdb_1_3"}, docStore, podPassword)

		// delete the db in the middle
		err = docStore.DeleteDocumentDB("docdb_1_2", podPassword)
		require.NoError(t, err)

		// check if other two db exists
		checkIfDBsExists(t, []string{"docdb_1_1", "docdb_1_3"}, docStore, podPassword)
		err = docStore.DeleteDocumentDB("docdb_1_1", podPassword)
		require.NoError(t, err)
		err = docStore.DeleteDocumentDB("docdb_1_3", podPassword)
		require.NoError(t, err)
		checkIfDBNotExists(t, "docdb_1_1", podPassword, docStore)
		checkIfDBNotExists(t, "docdb_1_3", podPassword, docStore)
	})

	t.Run("delete_all_document_db", func(t *testing.T) {
		// create multiple document DB
		createDocumentDBs(t, []string{"docdb_1_1", "docdb_1_2", "docdb_1_3"}, docStore, nil, podPassword)
		checkIfDBsExists(t, []string{"docdb_1_1", "docdb_1_2", "docdb_1_3"}, docStore, podPassword)

		// delete the db in the middle
		err = docStore.DeleteDocumentDB("docdb_1_2", podPassword)
		require.NoError(t, err)

		// check if other two db exists
		checkIfDBsExists(t, []string{"docdb_1_1", "docdb_1_3"}, docStore, podPassword)
		err = docStore.DeleteAllDocumentDBs(podPassword)
		require.NoError(t, err)

		checkIfDBNotExists(t, "docdb_1_1", podPassword, docStore)
		checkIfDBNotExists(t, "docdb_1_3", podPassword, docStore)
	})

	t.Run("create_document_db_with_multiple_indexes", func(t *testing.T) {
		// create a document DB and add simple indexes
		si := make(map[string]collection.IndexType)
		si["field1"] = collection.StringIndex
		si["field2"] = collection.NumberIndex
		si["field3"] = collection.MapIndex
		si["field4"] = collection.ListIndex
		createDocumentDBs(t, []string{"docdb_2"}, docStore, si, podPassword)

		// load the schem and check the count of simple indexes
		schema := loadSchemaAndCheckSimpleIndexCount(t, docStore, "docdb_2", podPassword, 3)

		// first check the default index
		checkIndex(t, schema.SimpleIndexes[0], collection.DefaultIndexFieldName, collection.StringIndex)

		checkIndex(t, schema.SimpleIndexes[0], "id", collection.StringIndex)

		// second check the field in index 1
		if schema.SimpleIndexes[1].FieldName == "field1" {
			checkIndex(t, schema.SimpleIndexes[1], "field1", collection.StringIndex)
		} else {
			checkIndex(t, schema.SimpleIndexes[1], "field2", collection.NumberIndex)
		}

		// third check the field in index 2
		if schema.SimpleIndexes[2].FieldName == "field2" {
			checkIndex(t, schema.SimpleIndexes[2], "field2", collection.NumberIndex)
		} else {
			checkIndex(t, schema.SimpleIndexes[2], "field1", collection.StringIndex)
		}

		if schema.MapIndexes[0].FieldName == "field3." {
			checkIndex(t, schema.MapIndexes[0], "field3.", collection.MapIndex)
		}

		if schema.ListIndexes[0].FieldName == "field4" {
			checkIndex(t, schema.ListIndexes[0], "field4", collection.ListIndex)
		}

	})

	t.Run("create_and open_document_db", func(t *testing.T) {
		// create a document DB
		createDocumentDBs(t, []string{"docdb_3"}, docStore, nil, podPassword)

		err := docStore.OpenDocumentDB("docdb_3", podPassword)
		require.NoError(t, err)

		// check if the DB is opened properly
		if !docStore.IsDBOpened("docdb_3") {
			t.Fatalf("db not opened")
		}
	})

	t.Run("put_immutable_error", func(t *testing.T) {
		// create a document DB
		err := docStore.CreateDocumentDB("doc_do_immutable", podPassword, nil, false)
		require.NoError(t, err)

		err = docStore.OpenDocumentDB("doc_do_immutable", podPassword)
		require.NoError(t, err)

		// create a json document
		document1 := &TestDocument{
			ID:        "1",
			FirstName: "John",
			LastName:  "Doe",
			Age:       25,
		}
		data, err := json.Marshal(document1)
		require.NoError(t, err)

		// insert the docment in the DB
		err = docStore.Put("doc_do_immutable", data)
		if !errors.Is(err, collection.ErrModifyingImmutableDocDB) {
			t.Fatal("db is immutable")
		}
	})

	t.Run("put_and_get", func(t *testing.T) {
		// create a document DB
		createDocumentDBs(t, []string{"docdb_4"}, docStore, nil, podPassword)

		err := docStore.OpenDocumentDB("docdb_4", podPassword)
		require.NoError(t, err)

		invalidType := struct {
			Time int64 `json:"created_at"`
		}{
			Time: time.Now().Unix(),
		}

		data, err := json.Marshal(invalidType)
		require.NoError(t, err)

		err = docStore.Put("docdb_4", data)
		if !errors.Is(err, collection.ErrDocumentDBIndexFieldNotPresent) {
			t.Fatal("index is not present")
		}

		document1 := &TestDocument{
			ID:        "",
			FirstName: "John",
			LastName:  "Doe",
			Age:       25,
		}
		data, err = json.Marshal(document1)
		require.NoError(t, err)

		// insert the docment in the DB
		err = docStore.Put("docdb_4", data)
		if !errors.Is(err, collection.ErrInvalidDocumentId) {
			t.Fatal("index is invalid")
		}

		// create a json document
		document1 = &TestDocument{
			ID:        "1",
			FirstName: "John",
			LastName:  "Doe",
			Age:       25,
		}
		data, err = json.Marshal(document1)
		require.NoError(t, err)

		// insert the docment in the DB
		err = docStore.Put("docdb_4", data)
		require.NoError(t, err)

		// get the data and test if the retreived data is okay
		gotData, err := docStore.Get("docdb_4", "1", podPassword)
		require.NoError(t, err)

		var doc TestDocument
		err = json.Unmarshal(gotData, &doc)
		require.NoError(t, err)

		if doc.ID != document1.ID ||
			doc.FirstName != document1.FirstName ||
			doc.LastName != document1.LastName ||
			doc.Age != document1.Age {
			t.Fatalf("invalid json data received")
		}
	})

	t.Run("put_and_get_multiple_index", func(t *testing.T) {
		// create a document DB
		si := make(map[string]collection.IndexType)
		si["first_name"] = collection.StringIndex
		si["age"] = collection.NumberIndex
		si["tag_map"] = collection.MapIndex
		si["tag_list"] = collection.ListIndex
		createDocumentDBs(t, []string{"docdb_5"}, docStore, si, podPassword)

		err := docStore.OpenDocumentDB("docdb_5", podPassword)
		require.NoError(t, err)

		// Add documents
		createTestDocuments(t, docStore, "docdb_5")

		// get string index and check if the documents returned are okay
		docs, err := docStore.Get("docdb_5", "2", podPassword)
		require.NoError(t, err)

		var gotDoc TestDocument
		err = json.Unmarshal(docs, &gotDoc)
		require.NoError(t, err)

		if gotDoc.ID != "2" ||
			gotDoc.FirstName != "John" ||
			gotDoc.LastName != "boy" ||
			gotDoc.Age != 25 ||
			gotDoc.TagMap["tgf21"] != "tgv21" ||
			gotDoc.TagMap["tgf22"] != "tgv22" {
			t.Fatalf("invalid json data received")
		}
	})

	t.Run("count_all", func(t *testing.T) {
		// create a document DB
		si := make(map[string]collection.IndexType)
		si["first_name"] = collection.StringIndex
		si["age"] = collection.NumberIndex
		si["tag_map"] = collection.MapIndex
		si["tag_list"] = collection.ListIndex
		createDocumentDBs(t, []string{"docdb_6"}, docStore, si, podPassword)

		err := docStore.OpenDocumentDB("docdb_6", podPassword)
		require.NoError(t, err)

		// Add documents
		createTestDocuments(t, docStore, "docdb_6")

		count1, err := docStore.Count("docdb_6", "")
		require.NoError(t, err)

		if count1 != 6 {
			t.Fatalf("expected count %d, got %d", 6, count1)
		}

	})

	t.Run("count_with_expr", func(t *testing.T) {
		// create a document DB
		si := make(map[string]collection.IndexType)
		si["first_name"] = collection.StringIndex
		si["age"] = collection.NumberIndex
		si["tag_map"] = collection.MapIndex
		si["tag_list"] = collection.ListIndex
		createDocumentDBs(t, []string{"docdb_7"}, docStore, si, podPassword)

		err := docStore.OpenDocumentDB("docdb_7", podPassword)
		require.NoError(t, err)

		// Add documents
		createTestDocuments(t, docStore, "docdb_7")

		// String count
		count1, err := docStore.Count("docdb_7", "first_name=>John")
		require.NoError(t, err)

		if count1 != 2 {
			t.Fatalf("expected count %d, got %d", 2, count1)
		}

		count1, err = docStore.Count("docdb_7", "tag_map=tgf11:tgv11")
		require.NoError(t, err)

		if count1 != 1 {
			t.Fatalf("expected count %d, got %d", 1, count1)
		}

		// Number =
		count2, err := docStore.Count("docdb_7", "age=25")
		require.NoError(t, err)

		if count2 != 3 {
			t.Fatalf("expected count %d, got %d", 3, count2)
		}

		// Number =>
		count3, err := docStore.Count("docdb_7", "age=>30")
		require.NoError(t, err)

		if count3 != 3 {
			t.Fatalf("expected count %d, got %d", 3, count3)
		}

		// Number >
		count4, err := docStore.Count("docdb_7", "age>30")
		require.NoError(t, err)

		if count4 != 2 {
			t.Fatalf("expected count %d, got %d", 2, count4)
		}
	})

	t.Run("find", func(t *testing.T) {
		// create a document DB
		si := make(map[string]collection.IndexType)
		si["first_name"] = collection.StringIndex
		si["age"] = collection.NumberIndex
		si["tag_map"] = collection.MapIndex
		si["tag_list"] = collection.ListIndex
		createDocumentDBs(t, []string{"docdb_8"}, docStore, si, podPassword)

		err := docStore.OpenDocumentDB("docdb_8", podPassword)
		require.NoError(t, err)

		// Add documents
		createTestDocuments(t, docStore, "docdb_8")

		// String =>
		docs, err := docStore.Find("docdb_8", "first_name=>John", podPassword, -1)
		require.NoError(t, err)

		if len(docs) != 2 {
			t.Fatalf("expected count %d, got %d", 2, len(docs))
		}
		var gotDoc1 TestDocument
		err = json.Unmarshal(docs[0], &gotDoc1)
		require.NoError(t, err)

		if gotDoc1.ID != "1" ||
			gotDoc1.FirstName != "John" ||
			gotDoc1.LastName != "Doe" ||
			gotDoc1.Age != 45.523793600000005 {
			t.Fatalf("invalid json data received")
		}
		var gotDoc2 TestDocument
		err = json.Unmarshal(docs[1], &gotDoc2)
		require.NoError(t, err)

		if gotDoc2.ID != "2" ||
			gotDoc2.FirstName != "John" ||
			gotDoc2.LastName != "boy" ||
			gotDoc2.Age != 25 {
			t.Fatalf("invalid json data received")
		}

		// tag
		docs, err = docStore.Find("docdb_8", "tag_map=tgf21:tgv21", podPassword, -1)
		require.NoError(t, err)

		if len(docs) != 1 {
			t.Fatalf("expected count %d, got %d", 1, len(docs))
		}
		err = json.Unmarshal(docs[0], &gotDoc2)
		require.NoError(t, err)

		if gotDoc2.ID != "2" ||
			gotDoc2.FirstName != "John" ||
			gotDoc2.LastName != "boy" ||
			gotDoc2.Age != 25 ||
			gotDoc2.TagMap["tgf21"] != "tgv21" {
			t.Fatalf("invalid json data received")
		}

		// Number =
		docs, err = docStore.Find("docdb_8", "age=25", podPassword, -1)
		require.NoError(t, err)

		if len(docs) != 3 {
			t.Fatalf("expected count %d, got %d", 3, len(docs))
		}
		err = json.Unmarshal(docs[0], &gotDoc1)
		require.NoError(t, err)

		if gotDoc1.ID != "2" ||
			gotDoc1.FirstName != "John" ||
			gotDoc1.LastName != "boy" ||
			gotDoc1.Age != 25 {
			t.Fatalf("invalid json data received")
		}
		err = json.Unmarshal(docs[1], &gotDoc2)
		require.NoError(t, err)

		if gotDoc2.ID != "4" ||
			gotDoc2.FirstName != "Charlie" ||
			gotDoc2.LastName != "chaplin" ||
			gotDoc2.Age != 25 {
			t.Fatalf("invalid json data received")
		}
		var gotDoc3 TestDocument
		err = json.Unmarshal(docs[2], &gotDoc3)
		require.NoError(t, err)

		if gotDoc3.ID != "5" ||
			gotDoc3.FirstName != "Alice" ||
			gotDoc3.LastName != "wonderland" ||
			gotDoc3.Age != 25 {
			t.Fatalf("invalid json data received")
		}

		// Number = with limit
		docs, err = docStore.Find("docdb_8", "age=25", podPassword, 2)
		require.NoError(t, err)

		if len(docs) != 2 {
			t.Fatalf("expected count %d, got %d", 2, len(docs))
		}
		err = json.Unmarshal(docs[0], &gotDoc1)
		require.NoError(t, err)
		if gotDoc1.ID != "2" ||
			gotDoc1.FirstName != "John" ||
			gotDoc1.LastName != "boy" ||
			gotDoc1.Age != 25 {
			t.Fatalf("invalid json data received")
		}
		err = json.Unmarshal(docs[1], &gotDoc2)
		require.NoError(t, err)
		if gotDoc2.ID != "4" ||
			gotDoc2.FirstName != "Charlie" ||
			gotDoc2.LastName != "chaplin" ||
			gotDoc2.Age != 25 {
			t.Fatalf("invalid json data received")
		}

		// Number =>
		docs, err = docStore.Find("docdb_8", "age=>30", podPassword, -1)
		require.NoError(t, err)
		if len(docs) != 3 {
			t.Fatalf("expected count %d, got %d", 3, len(docs))
		}
		err = json.Unmarshal(docs[0], &gotDoc1)
		require.NoError(t, err)
		if gotDoc1.ID != "3" ||
			gotDoc1.FirstName != "Bob" ||
			gotDoc1.LastName != "michel" ||
			gotDoc1.Age != 30 {
			t.Fatalf("invalid json data received")
		}
		err = json.Unmarshal(docs[2], &gotDoc2)
		require.NoError(t, err)
		if gotDoc2.ID != "1" ||
			gotDoc2.FirstName != "John" ||
			gotDoc2.LastName != "Doe" ||
			gotDoc2.Age != 45.523793600000005 {
			t.Fatalf("invalid json data received")
		}

		// Number >
		docs, err = docStore.Find("docdb_8", "age>30", podPassword, -1)
		require.NoError(t, err)
		if len(docs) != 2 {
			t.Fatalf("expected count %d, got %d", 2, len(docs))
		}
		err = json.Unmarshal(docs[1], &gotDoc1)
		require.NoError(t, err)
		if gotDoc1.ID != "1" ||
			gotDoc1.FirstName != "John" ||
			gotDoc1.LastName != "Doe" ||
			gotDoc1.Age != 45.523793600000005 {
			t.Fatalf("invalid json data received")
		}

		docs, err = docStore.Find("docdb_8", "tag_map=>tgf11:tgv11", podPassword, -1)
		require.NoError(t, err)

		assert.Equal(t, len(docs), 12)

		docs, err = docStore.Find("docdb_8", "tag_map>tgf11:tgv11", podPassword, -1)
		require.NoError(t, err)

		assert.Equal(t, len(docs), 11)

		docs, err = docStore.Find("docdb_8", "tag_map=>tgf41:tgv41", podPassword, -1)
		require.NoError(t, err)

		assert.Equal(t, len(docs), 6)

		docs, err = docStore.Find("docdb_8", "tag_map>tgf41:tgv41", podPassword, -1)
		require.NoError(t, err)

		assert.Equal(t, len(docs), 5)

		docs, err = docStore.Find("docdb_8", "age<=30", podPassword, -1)
		require.NoError(t, err)

		assert.Equal(t, len(docs), 4)

		docs, err = docStore.Find("docdb_8", "age<30", podPassword, -1)
		require.NoError(t, err)

		assert.Equal(t, len(docs), 3)

		// Number !=
		docs, err = docStore.Find("docdb_8", "age!=25", podPassword, -1)
		require.NoError(t, err)
		if len(docs) != 3 {
			t.Fatalf("expected count %d, got %d", 3, len(docs))
		}
		for _, v := range docs {
			var doc TestDocument
			err = json.Unmarshal(v, &doc)
			if err != nil {
				t.Fatal(err)
			}
			if doc.Age == 25 {
				t.Fatal("age should not be 25")
			}
		}

		// String !=
		_, err = docStore.Find("docdb_8", "first_name!=Bob", podPassword, -1)
		if err == nil {
			t.Fatal("should not be err ", err)
		}
	})

	t.Run("del", func(t *testing.T) {
		// create a document DB
		si := make(map[string]collection.IndexType)
		si["first_name"] = collection.StringIndex
		si["age"] = collection.NumberIndex
		createDocumentDBs(t, []string{"docdb_9"}, docStore, si, podPassword)

		err := docStore.OpenDocumentDB("docdb_9", podPassword)
		require.NoError(t, err)

		// Add document and get to see if it is added
		tag1 := make(map[string]string)
		tag1["tgf11"] = "tgv11"
		tag1["tgf12"] = "tgv12"
		var list1 []string
		list1 = append(list1, "lst11", "lst12")
		addDocument(t, docStore, "docdb_9", "1", "John", "Doe", 45, tag1, list1)
		docs, err := docStore.Get("docdb_9", "1", podPassword)
		require.NoError(t, err)
		var gotDoc TestDocument
		err = json.Unmarshal(docs, &gotDoc)
		require.NoError(t, err)
		if gotDoc.ID != "1" ||
			gotDoc.FirstName != "John" ||
			gotDoc.LastName != "Doe" ||
			gotDoc.Age != 45 {
			t.Fatalf("invalid json data received")
		}

		// del document
		err = docStore.Del("docdb_9", "1")
		require.NoError(t, err)
		_, err = docStore.Get("docdb_9", "1", podPassword)
		if !errors.Is(err, collection.ErrEntryNotFound) {
			t.Fatal(err)
		}
	})

	t.Run("del_different_indexes", func(t *testing.T) {
		// create a document DB
		si := make(map[string]collection.IndexType)
		si["first_name"] = collection.StringIndex
		si["age"] = collection.NumberIndex
		si["tag_map"] = collection.MapIndex
		si["tag_list"] = collection.ListIndex
		createDocumentDBs(t, []string{"docdb_99"}, docStore, si, podPassword)

		err := docStore.OpenDocumentDB("docdb_99", podPassword)
		require.NoError(t, err)

		// Add document and get to see if it is added
		tag1 := make(map[string]string)
		tag1["tgf11"] = "tgv11"
		tag1["tgf12"] = "tgv12"
		var list1 []string
		list1 = append(list1, "lst11", "lst12")
		addDocument(t, docStore, "docdb_99", "1", "John", "Doe", 45, tag1, list1)
		docs, err := docStore.Get("docdb_99", "1", podPassword)
		require.NoError(t, err)
		var gotDoc TestDocument
		err = json.Unmarshal(docs, &gotDoc)
		require.NoError(t, err)
		if gotDoc.ID != "1" ||
			gotDoc.FirstName != "John" ||
			gotDoc.LastName != "Doe" ||
			gotDoc.Age != 45 {
			t.Fatalf("invalid json data received")
		}

		// del document
		err = docStore.Del("docdb_99", "1")
		require.NoError(t, err)
		_, err = docStore.Get("docdb_99", "1", podPassword)
		if !errors.Is(err, collection.ErrEntryNotFound) {
			t.Fatal(err)
		}
	})

	t.Run("add_add", func(t *testing.T) {
		// create a document DB
		si := make(map[string]collection.IndexType)
		si["first_name"] = collection.StringIndex
		si["age"] = collection.NumberIndex
		createDocumentDBs(t, []string{"docdb_10"}, docStore, si, podPassword)

		err := docStore.OpenDocumentDB("docdb_10", podPassword)
		require.NoError(t, err)

		tag1 := make(map[string]string)
		tag1["tgf11"] = "tgv11"
		tag1["tgf12"] = "tgv12"
		var list1 []string
		list1 = append(list1, "lst11", "lst12")
		addDocument(t, docStore, "docdb_10", "1", "John", "Doe", 45, tag1, list1)
		addDocument(t, docStore, "docdb_10", "1", "John", "Doe", 25, tag1, list1)

		// count the total docs using id field
		count1, err := docStore.Count("docdb_10", "")
		require.NoError(t, err)
		if count1 != 1 {
			t.Fatalf("expected count %d, got %d", 1, count1)
		}

		// count the total docs using another index to make sure we don't have it any index
		docs, err := docStore.Find("docdb_10", "age=>20", podPassword, -1)
		require.NoError(t, err)
		if len(docs) != 1 {
			t.Fatalf("expected count %d, got %d", 1, len(docs))
		}
	})

	t.Run("batch-mutable", func(t *testing.T) {
		// create a document DB
		si := make(map[string]collection.IndexType)
		si["first_name"] = collection.StringIndex
		si["age"] = collection.NumberIndex
		si["tag_map"] = collection.MapIndex
		si["tag_list"] = collection.ListIndex
		createDocumentDBs(t, []string{"docdb_11"}, docStore, si, podPassword)

		err := docStore.OpenDocumentDB("docdb_11", podPassword)
		require.NoError(t, err)

		docBatch, err := docStore.CreateDocBatch("docdb_11", podPassword)
		require.NoError(t, err)

		tag1 := make(map[string]string)
		tag1["tgf11"] = "tgv11"
		tag1["tgf12"] = "tgv12"
		var list1 []string
		list1 = append(list1, "lst11", "lst12")
		addBatchDocument(t, docStore, docBatch, "1", "John", "Doe", 45, tag1, list1)
		tag2 := make(map[string]string)
		tag2["tgf21"] = "tgv21"
		tag2["tgf22"] = "tgv22"
		var list2 []string
		list2 = append(list2, "lst21", "lst22")
		addBatchDocument(t, docStore, docBatch, "2", "John", "boy", 25, tag2, list2)
		tag3 := make(map[string]string)
		tag3["tgf31"] = "tgv31"
		tag3["tgf32"] = "tgv32"
		var list3 []string
		list3 = append(list3, "lst31", "lst32")
		addBatchDocument(t, docStore, docBatch, "3", "Alice", "wonderland", 20, tag3, list3)
		tag4 := make(map[string]string)
		tag4["tgf41"] = "tgv41"
		tag4["tgf42"] = "tgv42"
		var list4 []string
		list4 = append(list4, "lst41", "lst42")
		addBatchDocument(t, docStore, docBatch, "4", "John", "Doe", 35, tag4, list4) // this tests the overwriting in batch

		err = docStore.DocBatchWrite(docBatch, "")
		require.NoError(t, err)

		// count the total docs using id field
		count1, err := docStore.Count("docdb_11", "")
		require.NoError(t, err)
		if count1 != 4 {
			t.Fatalf("expected count %d, got %d", 4, count1)
		}

		// count the total docs using another index to make sure we don't have it any index
		docs, err := docStore.Find("docdb_11", "age=>20", podPassword, -1)
		require.NoError(t, err)
		if len(docs) != 4 {
			t.Fatalf("expected count %d, got %d", 3, len(docs))
		}

		// tag
		docs, err = docStore.Find("docdb_11", "tag_map=tgf21:tgv21", podPassword, -1)
		require.NoError(t, err)
		if len(docs) != 1 {
			t.Fatalf("expected count %d, got %d", 1, len(docs))
		}
		err = docStore.DeleteDocumentDB("docdb_11", podPassword)
		require.NoError(t, err)
	})
	/*
		t.Run("batch-immutable", func(t *testing.T) {
			// create a document DB
			si := make(map[string]collection.IndexType)
			si["first_name"] = collection.StringIndex
			si["age"] = collection.NumberIndex
			si["tag_map"] = collection.MapIndex
			si["tag_list"] = collection.ListIndex
			// createDocumentDBs(t, []string{"docdb_12"}, docStore, si)
			err := docStore.CreateDocumentDB("docdb_12", si, false)
			if err != nil {
				t.Fatal(err)
			}

			err = docStore.OpenDocumentDB("docdb_12")
			if err != nil {
				t.Fatal(err)
			}

			docBatch, err := docStore.CreateDocBatch("docdb_12")
			if err != nil {
				t.Fatal(err)
			}

			tag1 := make(map[string]string)
			tag1["tgf11"] = "tgv11"
			tag1["tgf12"] = "tgv12"
			var list1 []string
			list1 = append(list1, "lst11")
			list1 = append(list1, "lst12")
			addBatchDocument(t, docStore, docBatch, "1", "John", "Doe", 45, tag1, list1)
			tag2 := make(map[string]string)
			tag2["tgf21"] = "tgv21"
			tag2["tgf22"] = "tgv22"
			var list2 []string
			list2 = append(list2, "lst21")
			list2 = append(list2, "lst22")
			addBatchDocument(t, docStore, docBatch, "2", "John", "boy", 25, tag2, list2)
			tag3 := make(map[string]string)
			tag3["tgf31"] = "tgv31"
			tag3["tgf32"] = "tgv32"
			var list3 []string
			list3 = append(list3, "lst31")
			list3 = append(list3, "lst32")
			addBatchDocument(t, docStore, docBatch, "3", "Alice", "wonderland", 20, tag3, list3)
			tag4 := make(map[string]string)
			tag4["tgf41"] = "tgv41"
			tag4["tgf42"] = "tgv42"
			var list4 []string
			list4 = append(list4, "lst41")
			list4 = append(list4, "lst42")
			addBatchDocument(t, docStore, docBatch, "4", "John", "Doe", 35, tag4, list4) // this tests the overwriting in batch

			err = docStore.DocBatchWrite(docBatch, "")
			if err != nil {
				t.Fatal(err)
			}

			// count the total docs using id field
			count1, err := docStore.Count("docdb_12", "")
			if err != nil {
				t.Fatal(err)
			}
			if count1 != 4 {
				t.Fatalf("expected count %d, got %d", 4, count1)
			}

			// count the total docs using another index to make sure we dont have it any index
			docs, err := docStore.Find("docdb_12", "age=>20", -1)
			if err != nil {
				t.Fatal(err)
			}
			if len(docs) != 4 {
				t.Fatalf("expected count %d, got %d", 4, len(docs))
			}

			// tag
			docs, err = docStore.Find("docdb_12", "tag_map=tgf21:tgv21", -1)
			if err != nil {
				t.Fatal(err)
			}
			if len(docs) != 1 {
				t.Fatalf("expected count %d, got %d", 1, len(docs))
			}
			err = docStore.DeleteDocumentDB("docdb_12")
			if err != nil {
				t.Fatal(err)
			}
		})
	*/
}

func createDocumentDBs(t *testing.T, dbNames []string, docStore *collection.Document, si map[string]collection.IndexType, podPassword string) {
	t.Helper()
	for _, dbName := range dbNames {
		err := docStore.CreateDocumentDB(dbName, podPassword, si, true)
		require.NoError(t, err)
	}
}

func checkIfDBsExists(t *testing.T, dbNames []string, docStore *collection.Document, podPassword string) {
	t.Helper()
	tables, err := docStore.LoadDocumentDBSchemas(podPassword)
	if err != nil {
		t.Fatal(err)
	}
	for _, tableName := range dbNames {
		if _, found := tables[tableName]; !found {
			t.Fatalf("document db not found")
		}
	}
}

func checkIfDBNotExists(t *testing.T, tableName, podPassword string, docStore *collection.Document) {
	t.Helper()
	tables, err := docStore.LoadDocumentDBSchemas(podPassword)
	if err != nil {
		t.Fatal(err)
	}
	if _, found := tables[tableName]; found {
		t.Fatalf("document db found")
	}
}

func loadSchemaAndCheckSimpleIndexCount(t *testing.T, docStore *collection.Document, dbName, podPassword string, count int) collection.DBSchema {
	t.Helper()
	tables, err := docStore.LoadDocumentDBSchemas(podPassword)
	if err != nil {
		t.Fatal(err)
	}
	schema, found := tables[dbName]
	if !found {
		t.Fatalf("document db not found in schema")
	}
	if len(schema.SimpleIndexes) != count {
		t.Fatalf("index count mismatch")
	}
	return schema
}

func checkIndex(t *testing.T, si collection.SIndex, filedName string, idxType collection.IndexType) {
	t.Helper()
	if si.FieldName != filedName {
		t.Fatalf("index field not found: %s, %s", si.FieldName, filedName)
	}
	if si.FieldType != idxType {
		t.Fatalf("index field type is not correct: %s, %s", si.FieldType, idxType)
	}
}

func createTestDocuments(t *testing.T, docStore *collection.Document, dbName string) {
	t.Helper()
	tag1 := make(map[string]string)
	tag1["tgf11"] = "tgv11"
	tag1["tgf12"] = "tgv12"
	var list1 []string
	list1 = append(list1, "lst11", "lst12")
	addDocument(t, docStore, dbName, "1", "John", "Doe", 45.523793600000005, tag1, list1)
	tag2 := make(map[string]string)
	tag2["tgf21"] = "tgv21"
	tag2["tgf22"] = "tgv22"
	var list2 []string
	list2 = append(list2, "lst21", "lst22")
	addDocument(t, docStore, dbName, "2", "John", "boy", 25, tag2, list2)
	tag3 := make(map[string]string)
	tag3["tgf31"] = "tgv31"
	tag3["tgf32"] = "tgv32"
	var list3 []string
	list3 = append(list3, "lst31", "lst32")
	addDocument(t, docStore, dbName, "3", "Bob", "michel", 30, tag3, list3)
	tag4 := make(map[string]string)
	tag4["tgf41"] = "tgv41"
	tag4["tgf42"] = "tgv42"
	var list4 []string
	list4 = append(list4, "lst41", "lst42")
	addDocument(t, docStore, dbName, "4", "Charlie", "chaplin", 25, tag4, list4)
	tag5 := make(map[string]string)
	tag5["tgf51"] = "tgv51"
	tag5["tgf52"] = "tgv52"
	var list5 []string
	list5 = append(list5, "lst51", "lst52")
	addDocument(t, docStore, dbName, "5", "Alice", "wonderland", 25, tag5, list5)
	tag6 := make(map[string]string)
	tag6["tgf61"] = "tgv61"
	tag6["tgf62"] = "tgv62"
	var list6 []string
	list6 = append(list6, "lst61", "lst62")
	addDocument(t, docStore, dbName, "6", "Zuri", "wonder", 52, tag6, list6)
}

func addDocument(t *testing.T, docStore *collection.Document, dbName, id, fname, lname string, age float64, tagMap map[string]string, tagList []string) {
	t.Helper()
	// create the doc
	doc := &TestDocument{
		ID:        id,
		FirstName: fname,
		LastName:  lname,
		Age:       age,
		TagMap:    tagMap,
		TagList:   tagList,
	}

	// marshall the doc
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}

	// insert the docment in the DB
	err = docStore.Put(dbName, data)
	if err != nil {
		t.Fatal(err)
	}
}

func addBatchDocument(t *testing.T, docStore *collection.Document, docBatch *collection.DocBatch, id, fname, lname string, age float64, tagMap map[string]string, tagList []string) {
	t.Helper()
	t.Run("valid-json", func(t *testing.T) {
		// create the doc
		doc := &TestDocument{
			ID:        id,
			FirstName: fname,
			LastName:  lname,
			Age:       age,
			TagMap:    tagMap,
			TagList:   tagList,
		}

		// marshall the doc
		data, err := json.Marshal(doc)
		require.NoError(t, err)

		// insert the document in the batch
		err = docStore.DocBatchPut(docBatch, data, 0)
		require.NoError(t, err)
	})
	t.Run("invalid-json", func(t *testing.T) {
		// create the doc
		doc := TestDocument{
			ID:        id,
			FirstName: fname,
			LastName:  lname,
			Age:       age,
			TagMap:    tagMap,
			TagList:   tagList,
		}

		// marshall the doc
		data, err := json.Marshal([]TestDocument{doc})
		require.NoError(t, err)

		// insert the document in the batch
		err = docStore.DocBatchPut(docBatch, data, 0)
		if err != collection.ErrUnknownJsonFormat {
			t.Fatal(err)
		}
	})

}
