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
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"io/ioutil"
	"testing"
)

type TestDocument struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       string `json:"age"`
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
		createDocumentDBs(t, []string{"doc_0"}, docStore)

		// load the schem and check the count of simple indexes
		schema := loadSchemaAndCheckSimpleIndexCount(t, docStore, "doc_0", 1)

		// check the default index
		checkIndex(t, schema.SimpleIndexs[0], collection.DefaultIndexFieldName, collection.StringIndex)
	})

	t.Run("delete_document_db", func(t *testing.T) {
		// create multiple document DB
		createDocumentDBs(t, []string{"doc_1_1", "doc_1_2", "doc_1_3"}, docStore)
		checkIfDBsExists(t, []string{"doc_1_1", "doc_1_2", "doc_1_3"}, docStore)

		// delete the db in the middle
		err = docStore.DeleteDocumentDB("doc_1_2")
		if err != nil {
			t.Fatal(err)
		}

		// check if other two db exists
		checkIfDBsExists(t, []string{"doc_1_1", "doc_1_3"}, docStore)
	})

	t.Run("create_document_db_with_ultiple_simple_indexes", func(t *testing.T) {
		// create a document DB
		createDocumentDBs(t, []string{"doc_2"}, docStore)

		// add a string index
		err = docStore.AddSimpleIndex("doc_2", "field1", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}

		// add a number index
		err = docStore.AddSimpleIndex("doc_2", "field2", collection.NumberIndex)
		if err != nil {
			t.Fatal(err)
		}

		// load the schem and check the count of simple indexes
		schema := loadSchemaAndCheckSimpleIndexCount(t, docStore, "doc_2", 3)

		// first check the default index
		checkIndex(t, schema.SimpleIndexs[0], collection.DefaultIndexFieldName, collection.StringIndex)

		//second check the string index
		checkIndex(t, schema.SimpleIndexs[1], "field1", collection.StringIndex)

		//third check the string index
		checkIndex(t, schema.SimpleIndexs[2], "field2", collection.NumberIndex)
	})

	t.Run("create_and open_document_db", func(t *testing.T) {
		// create a document DB
		createDocumentDBs(t, []string{"doc_3"}, docStore)

		err := docStore.OpenDocumentDB("doc_3")
		if err != nil {
			t.Fatal(err)
		}

		// check if the DB is opened properly
		if !docStore.IsDBOpened("doc_3") {
			t.Fatalf("db not opened")
		}

	})

	t.Run("put_and_get", func(t *testing.T) {
		// create a document DB
		createDocumentDBs(t, []string{"doc_4"}, docStore)

		err := docStore.OpenDocumentDB("doc_4")
		if err != nil {
			t.Fatal(err)
		}

		// create a json document
		document1 := &TestDocument{
			ID:        "1",
			FirstName: "John",
			LastName:  "Doe",
			Age:       "25",
		}
		data, err := json.Marshal(document1)
		if err != nil {
			t.Fatal(err)
		}

		// insert the docment in the DB
		err = docStore.Put("doc_4", data)
		if err != nil {
			t.Fatal(err)
		}

		// get the data and test if the retreived data is okay
		gotData, err := docStore.Get("doc_4", "id=1", 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(gotData) != 1 {
			t.Fatalf("got invalid data")
		}
		var doc TestDocument
		err = json.Unmarshal(gotData[0], &doc)
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
		createDocumentDBs(t, []string{"doc_5"}, docStore)

		// add a string index
		err = docStore.AddSimpleIndex("doc_5", "first_name", collection.StringIndex)
		if err != nil {
			t.Fatal(err)
		}

		// adda number index
		err = docStore.AddSimpleIndex("doc_5", "age", collection.NumberIndex)
		if err != nil {
			t.Fatal(err)
		}

		err := docStore.OpenDocumentDB("doc_5")
		if err != nil {
			t.Fatal(err)
		}

		// Add d0cuments
		createTestDocuments(t, docStore, "doc_5")

		// get string index and check if the documents returned are okay
		docs, err := docStore.Get("doc_5", "first_name=John", 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(docs) != 2 {
			t.Fatalf("got invalid data")
		}
		var doc1 TestDocument
		err = json.Unmarshal(docs[0], &doc1)
		if err != nil {
			t.Fatal(err)
		}
		if doc1.ID != "1" ||
			doc1.FirstName != "John" ||
			doc1.LastName != "Doe" ||
			doc1.Age != "45" {
			t.Fatalf("invalid json data received")
		}

		var doc2 TestDocument
		err = json.Unmarshal(docs[1], &doc2)
		if err != nil {
			t.Fatal(err)
		}
		if doc2.ID != "2" ||
			doc2.FirstName != "John" ||
			doc2.LastName != "boy" ||
			doc2.Age != "25" {
			t.Fatalf("invalid json data received")
		}

		// get number index with limit
		docs, err = docStore.Get("doc_5", "age=25", 2)
		if err != nil {
			t.Fatal(err)
		}
		if len(docs) != 2 {
			t.Fatalf("got invalid data")
		}
		err = json.Unmarshal(docs[0], &doc1)
		if err != nil {
			t.Fatal(err)
		}
		if doc1.ID != "2" ||
			doc1.FirstName != "John" ||
			doc1.LastName != "boy" ||
			doc1.Age != "25" {
			t.Fatalf("invalid json data received")
		}
		err = json.Unmarshal(docs[1], &doc2)
		if err != nil {
			t.Fatal(err)
		}
		if doc2.ID != "4" ||
			doc2.FirstName != "Charlie" ||
			doc2.LastName != "chaplin" ||
			doc2.Age != "25" {
			t.Fatalf("invalid json data received")
		}

		// get number => expression
		docs, err = docStore.Get("doc_5", "age=>20", 5)
		if err != nil {
			t.Fatal(err)
		}
		if len(docs) != 5 {
			t.Fatalf("got invalid data")
		}

	})
}

func createDocumentDBs(t *testing.T, dbNames []string, docStore *collection.Document) {
	for _, dbName := range dbNames {
		err := docStore.CreateDocumentDB(dbName)
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
	if len(schema.SimpleIndexs) != count {
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
	addDocument(t, docStore, dbName, "1", "John", "Doe", "45")
	addDocument(t, docStore, dbName, "2", "John", "boy", "25")
	addDocument(t, docStore, dbName, "3", "Bob", "michel", "30")
	addDocument(t, docStore, dbName, "4", "Charlie", "chaplin", "25")
	addDocument(t, docStore, dbName, "5", "Alice", "wonderland", "25")
}

func addDocument(t *testing.T, docStore *collection.Document, dbName string, id, fname, lname, age string) {
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
