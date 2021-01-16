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
	"encoding/json"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

type TestDocument struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int64  `json:"age"`
}

func TestDocumentStore(t *testing.T) {
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
	docStore := collection.NewDocumentStore(fd, ai, user, mockClient, logger)

	t.Run("create_document_db", func(t *testing.T) {
		// create a document DB
		createDocumentDBs(t, []string{"docdb_0"}, docStore, nil)

		// load the schem and check the count of simple indexes
		schema := loadSchemaAndCheckSimpleIndexCount(t, docStore, "docdb_0", 1)

		// check the default index
		checkIndex(t, schema.SimpleIndexes[0], collection.DefaultIndexFieldName, collection.StringIndex)
	})

	t.Run("delete_document_db", func(t *testing.T) {
		// create multiple document DB
		createDocumentDBs(t, []string{"docdb_1_1", "docdb_1_2", "docdb_1_3"}, docStore, nil)
		checkIfDBsExists(t, []string{"docdb_1_1", "docdb_1_2", "docdb_1_3"}, docStore)

		// delete the db in the middle
		err = docStore.DeleteDocumentDB("docdb_1_2")
		if err != nil {
			t.Fatal(err)
		}

		// check if other two db exists
		checkIfDBsExists(t, []string{"docdb_1_1", "docdb_1_3"}, docStore)
	})

	t.Run("create_document_db_with_multiple_simple_indexes", func(t *testing.T) {
		// create a document DB and add simple indexes
		si := make(map[string]collection.IndexType)
		si["field1"] = collection.StringIndex
		si["field2"] = collection.NumberIndex
		createDocumentDBs(t, []string{"docdb_2"}, docStore, si)

		// load the schem and check the count of simple indexes
		schema := loadSchemaAndCheckSimpleIndexCount(t, docStore, "docdb_2", 3)

		// first check the default index
		checkIndex(t, schema.SimpleIndexes[0], collection.DefaultIndexFieldName, collection.StringIndex)

		//second check the string index
		checkIndex(t, schema.SimpleIndexes[1], "field1", collection.StringIndex)

		//third check the string index
		checkIndex(t, schema.SimpleIndexes[2], "field2", collection.NumberIndex)
	})

	t.Run("create_and open_document_db", func(t *testing.T) {
		// create a document DB
		createDocumentDBs(t, []string{"docdb_3"}, docStore, nil)

		err := docStore.OpenDocumentDB("docdb_3")
		if err != nil {
			t.Fatal(err)
		}

		// check if the DB is opened properly
		if !docStore.IsDBOpened("docdb_3") {
			t.Fatalf("db not opened")
		}

	})

	t.Run("put_and_get", func(t *testing.T) {
		// create a document DB
		createDocumentDBs(t, []string{"docdb_4"}, docStore, nil)

		err := docStore.OpenDocumentDB("docdb_4")
		if err != nil {
			t.Fatal(err)
		}

		// create a json document
		document1 := &TestDocument{
			ID:        "1",
			FirstName: "John",
			LastName:  "Doe",
			Age:       25,
		}
		data, err := json.Marshal(document1)
		if err != nil {
			t.Fatal(err)
		}

		// insert the docment in the DB
		err = docStore.Put("docdb_4", data)
		if err != nil {
			t.Fatal(err)
		}

		// get the data and test if the retreived data is okay
		gotData, err := docStore.Get("docdb_4", "1")
		if err != nil {
			t.Fatal(err)
		}

		var doc TestDocument
		err = json.Unmarshal(gotData, &doc)
		if err != nil {
			t.Fatal(err)
		}
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
		createDocumentDBs(t, []string{"docdb_5"}, docStore, si)

		err := docStore.OpenDocumentDB("docdb_5")
		if err != nil {
			t.Fatal(err)
		}

		// Add documents
		createTestDocuments(t, docStore, "docdb_5")

		// get string index and check if the documents returned are okay
		docs, err := docStore.Get("docdb_5", "2")
		if err != nil {
			t.Fatal(err)
		}
		var gotDoc TestDocument
		err = json.Unmarshal(docs, &gotDoc)
		if err != nil {
			t.Fatal(err)
		}
		if gotDoc.ID != "2" ||
			gotDoc.FirstName != "John" ||
			gotDoc.LastName != "boy" ||
			gotDoc.Age != 25 {
			t.Fatalf("invalid json data received")
		}
	})

	t.Run("count_all", func(t *testing.T) {
		// create a document DB
		si := make(map[string]collection.IndexType)
		si["first_name"] = collection.StringIndex
		si["age"] = collection.NumberIndex
		createDocumentDBs(t, []string{"docdb_6"}, docStore, si)

		err := docStore.OpenDocumentDB("docdb_6")
		if err != nil {
			t.Fatal(err)
		}

		// Add documents
		createTestDocuments(t, docStore, "docdb_6")

		count1, err := docStore.Count("docdb_6", "")
		if err != nil {
			t.Fatal(err)
		}

		if count1 != 5 {
			t.Fatalf("expected count %d, got %d", 5, count1)
		}

	})

	t.Run("count_with_expr", func(t *testing.T) {
		// create a document DB
		si := make(map[string]collection.IndexType)
		si["first_name"] = collection.StringIndex
		si["age"] = collection.NumberIndex
		createDocumentDBs(t, []string{"docdb_7"}, docStore, si)

		err := docStore.OpenDocumentDB("docdb_7")
		if err != nil {
			t.Fatal(err)
		}

		// Add documents
		createTestDocuments(t, docStore, "docdb_7")

		// String count
		count1, err := docStore.Count("docdb_7", "first_name=John")
		if err != nil {
			t.Fatal(err)
		}
		if count1 != 2 {
			t.Fatalf("expected count %d, got %d", 2, count1)
		}

		// Number =
		count2, err := docStore.Count("docdb_7", "age=25")
		if err != nil {
			t.Fatal(err)
		}
		if count2 != 3 {
			t.Fatalf("expected count %d, got %d", 3, count2)
		}

		// Number =>
		count3, err := docStore.Count("docdb_7", "age=>30")
		if err != nil {
			t.Fatal(err)
		}
		if count3 != 2 {
			t.Fatalf("expected count %d, got %d", 2, count3)
		}

		// Number >
		count4, err := docStore.Count("docdb_7", "age>30")
		if err != nil {
			t.Fatal(err)
		}
		if count4 != 1 {
			t.Fatalf("expected count %d, got %d", 1, count4)
		}
	})

	t.Run("find", func(t *testing.T) {
		// create a document DB
		si := make(map[string]collection.IndexType)
		si["first_name"] = collection.StringIndex
		si["age"] = collection.NumberIndex
		createDocumentDBs(t, []string{"docdb_8"}, docStore, si)

		err := docStore.OpenDocumentDB("docdb_8")
		if err != nil {
			t.Fatal(err)
		}

		// Add documents
		createTestDocuments(t, docStore, "docdb_8")

		// String =
		docs, err := docStore.Find("docdb_8", "first_name=John", -1)
		if err != nil {
			t.Fatal(err)
		}
		if len(docs) != 2 {
			t.Fatalf("expected count %d, got %d", 2, len(docs))
		}
		var gotDoc1 TestDocument
		err = json.Unmarshal(docs[0], &gotDoc1)
		if err != nil {
			t.Fatal(err)
		}
		if gotDoc1.ID != "1" ||
			gotDoc1.FirstName != "John" ||
			gotDoc1.LastName != "Doe" ||
			gotDoc1.Age != 45 {
			t.Fatalf("invalid json data received")
		}
		var gotDoc2 TestDocument
		err = json.Unmarshal(docs[1], &gotDoc2)
		if err != nil {
			t.Fatal(err)
		}
		if gotDoc2.ID != "2" ||
			gotDoc2.FirstName != "John" ||
			gotDoc2.LastName != "boy" ||
			gotDoc2.Age != 25 {
			t.Fatalf("invalid json data received")
		}

		// Number =
		docs, err = docStore.Find("docdb_8", "age=25", -1)
		if err != nil {
			t.Fatal(err)
		}
		if len(docs) != 3 {
			t.Fatalf("expected count %d, got %d", 3, len(docs))
		}
		err = json.Unmarshal(docs[0], &gotDoc1)
		if err != nil {
			t.Fatal(err)
		}
		if gotDoc1.ID != "2" ||
			gotDoc1.FirstName != "John" ||
			gotDoc1.LastName != "boy" ||
			gotDoc1.Age != 25 {
			t.Fatalf("invalid json data received")
		}
		err = json.Unmarshal(docs[1], &gotDoc2)
		if err != nil {
			t.Fatal(err)
		}
		if gotDoc2.ID != "4" ||
			gotDoc2.FirstName != "Charlie" ||
			gotDoc2.LastName != "chaplin" ||
			gotDoc2.Age != 25 {
			t.Fatalf("invalid json data received")
		}
		var gotDoc3 TestDocument
		err = json.Unmarshal(docs[2], &gotDoc3)
		if err != nil {
			t.Fatal(err)
		}
		if gotDoc3.ID != "5" ||
			gotDoc3.FirstName != "Alice" ||
			gotDoc3.LastName != "wonderland" ||
			gotDoc3.Age != 25 {
			t.Fatalf("invalid json data received")
		}

		// Number = with limit
		docs, err = docStore.Find("docdb_8", "age=25", 2)
		if err != nil {
			t.Fatal(err)
		}
		if len(docs) != 2 {
			t.Fatalf("expected count %d, got %d", 2, len(docs))
		}
		err = json.Unmarshal(docs[0], &gotDoc1)
		if err != nil {
			t.Fatal(err)
		}
		if gotDoc1.ID != "2" ||
			gotDoc1.FirstName != "John" ||
			gotDoc1.LastName != "boy" ||
			gotDoc1.Age != 25 {
			t.Fatalf("invalid json data received")
		}
		err = json.Unmarshal(docs[1], &gotDoc2)
		if err != nil {
			t.Fatal(err)
		}
		if gotDoc2.ID != "4" ||
			gotDoc2.FirstName != "Charlie" ||
			gotDoc2.LastName != "chaplin" ||
			gotDoc2.Age != 25 {
			t.Fatalf("invalid json data received")
		}

		// Number =>
		docs, err = docStore.Find("docdb_8", "age=>30", -1)
		if err != nil {
			t.Fatal(err)
		}
		if len(docs) != 2 {
			t.Fatalf("expected count %d, got %d", 2, len(docs))
		}
		err = json.Unmarshal(docs[0], &gotDoc1)
		if err != nil {
			t.Fatal(err)
		}
		if gotDoc1.ID != "3" ||
			gotDoc1.FirstName != "Bob" ||
			gotDoc1.LastName != "michel" ||
			gotDoc1.Age != 30 {
			t.Fatalf("invalid json data received")
		}
		err = json.Unmarshal(docs[1], &gotDoc2)
		if err != nil {
			t.Fatal(err)
		}
		if gotDoc2.ID != "1" ||
			gotDoc2.FirstName != "John" ||
			gotDoc2.LastName != "Doe" ||
			gotDoc2.Age != 45 {
			t.Fatalf("invalid json data received")
		}

		// Number >
		docs, err = docStore.Find("docdb_8", "age>30", -1)
		if err != nil {
			t.Fatal(err)
		}
		if len(docs) != 1 {
			t.Fatalf("expected count %d, got %d", 1, len(docs))
		}
		err = json.Unmarshal(docs[0], &gotDoc1)
		if err != nil {
			t.Fatal(err)
		}
		if gotDoc1.ID != "1" ||
			gotDoc1.FirstName != "John" ||
			gotDoc1.LastName != "Doe" ||
			gotDoc1.Age != 45 {
			t.Fatalf("invalid json data received")
		}
	})

	t.Run("del", func(t *testing.T) {
		// create a document DB
		si := make(map[string]collection.IndexType)
		si["first_name"] = collection.StringIndex
		si["age"] = collection.NumberIndex
		createDocumentDBs(t, []string{"docdb_9"}, docStore, si)

		err := docStore.OpenDocumentDB("docdb_9")
		if err != nil {
			t.Fatal(err)
		}

		// Add document and get to see if it is added
		addDocument(t, docStore, "docdb_9", "1", "John", "Doe", 45)
		docs, err := docStore.Get("docdb_9", "1")
		if err != nil {
			t.Fatal(err)
		}
		var gotDoc TestDocument
		err = json.Unmarshal(docs, &gotDoc)
		if err != nil {
			t.Fatal(err)
		}
		if gotDoc.ID != "1" ||
			gotDoc.FirstName != "John" ||
			gotDoc.LastName != "Doe" ||
			gotDoc.Age != 45 {
			t.Fatalf("invalid json data received")
		}

		// del document
		err = docStore.Del("docdb_9", "1")
		if err != nil {
			t.Fatal(err)
		}
		_, err = docStore.Get("docdb_9", "1")
		if !errors.Is(err, collection.ErrEntryNotFound) {
			t.Fatal(err)
		}
	})

	t.Run("add_add", func(t *testing.T) {
		// create a document DB
		si := make(map[string]collection.IndexType)
		si["first_name"] = collection.StringIndex
		si["age"] = collection.NumberIndex
		createDocumentDBs(t, []string{"docdb_10"}, docStore, si)

		err := docStore.OpenDocumentDB("docdb_10")
		if err != nil {
			t.Fatal(err)
		}

		addDocument(t, docStore, "docdb_10", "1", "John", "Doe", 45)
		addDocument(t, docStore, "docdb_10", "1", "John", "Doe", 25)

		// count the total docs using id field
		count1, err := docStore.Count("docdb_10", "")
		if err != nil {
			t.Fatal(err)
		}
		if count1 != 1 {
			t.Fatalf("expected count %d, got %d", 1, count1)
		}

		// count the total docs using another index to make sure we dont have it any index
		docs, err := docStore.Find("docdb_10", "age=>20", -1)
		if err != nil {
			t.Fatal(err)
		}
		if len(docs) != 1 {
			t.Fatalf("expected count %d, got %d", 1, len(docs))
		}
	})

}

func createDocumentDBs(t *testing.T, dbNames []string, docStore *collection.Document, si map[string]collection.IndexType) {
	for _, dbName := range dbNames {
		err := docStore.CreateDocumentDB(dbName, si)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func checkIfDBsExists(t *testing.T, dbNames []string, docStore *collection.Document) {
	tables, err := docStore.LoadDocumentDBSchemas()
	if err != nil {
		t.Fatal(err)
	}
	for _, tableName := range dbNames {
		if _, found := tables[tableName]; !found {
			t.Fatalf("document db not found")
		}
	}
}

func loadSchemaAndCheckSimpleIndexCount(t *testing.T, docStore *collection.Document, dbName string, count int) collection.DBSchema {
	tables, err := docStore.LoadDocumentDBSchemas()
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
	if si.FieldName != filedName {
		t.Fatalf("index field not found")
	}
	if si.FieldType != idxType {
		t.Fatalf("index field type is not correct")
	}
}

func createTestDocuments(t *testing.T, docStore *collection.Document, dbName string) {
	addDocument(t, docStore, dbName, "1", "John", "Doe", 45)
	addDocument(t, docStore, dbName, "2", "John", "boy", 25)
	addDocument(t, docStore, dbName, "3", "Bob", "michel", 30)
	addDocument(t, docStore, dbName, "4", "Charlie", "chaplin", 25)
	addDocument(t, docStore, dbName, "5", "Alice", "wonderland", 25)
}

func addDocument(t *testing.T, docStore *collection.Document, dbName, id, fname, lname string, age int64) {
	// create the doc
	doc := &TestDocument{
		ID:        id,
		FirstName: fname,
		LastName:  lname,
		Age:       age,
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
