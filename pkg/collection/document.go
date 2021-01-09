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

package collection

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	DocumentFile          = "document_dbs"
	DefaultIndexFieldName = "id"
)

type Document struct {
	fd          *feed.API
	ai          *account.Info
	user        utils.Address
	client      blockstore.Client
	openDocDBs  map[string]*DocumentDB
	openDOcDBMu sync.RWMutex
	iterator    *Iterator
	logger      logging.Logger
}

type DocumentDB struct {
	name            string
	simpleIndexes   map[string]*Index
	compoundIndexes map[string]*CompoundIndex
}

type DBSchema struct {
	Name            string   `json:"name"`
	SimpleIndexs    []SIndex `json:"simple_indexes"`
	CompoundIndexes []CIndex `json:"compound_indexes,omitempty"`
}

type SIndex struct {
	FieldName string    `json:"name"`
	FieldType IndexType `json:"type"`
}

type CIndex struct {
	SimpleIndexs []SIndex
}

func NewDocumentStore(fd *feed.API, ai *account.Info, user utils.Address, client blockstore.Client, logger logging.Logger) *Document {
	return &Document{
		fd:         fd,
		ai:         ai,
		user:       user,
		client:     client,
		openDocDBs: make(map[string]*DocumentDB),
		logger:     logger,
	}
}

func (d *Document) CreateDocumentDB(dbName string) error {
	// load the existing db's and see if this name is already there
	docTables, err := d.LoadDocumentDBSchemas()
	if err != nil {
		return err
	}
	if _, ok := docTables[dbName]; ok {
		return ErrDocumentDBAlreadyPresent
	}

	///  since this db is not present already, create the default index required for this table
	err = CreateIndex(dbName, DefaultIndexFieldName, StringIndex, d.fd, d.user, d.client)
	if err != nil {
		return err
	}

	// record the table as created
	si := SIndex{
		FieldName: DefaultIndexFieldName,
		FieldType: StringIndex,
	}
	docTables[dbName] = DBSchema{
		Name:         dbName,
		SimpleIndexs: []SIndex{si},
	}
	return d.storeDocumentDBSchemas(docTables)
}

func (d *Document) AddSimpleIndex(dbName string, fieldName string, indexType IndexType) error {
	// load the existing db's and see if this name is present
	docTables, err := d.LoadDocumentDBSchemas()
	if err != nil {
		return err
	}
	schema, ok := docTables[dbName]
	if !ok {
		return ErrDocumentDBNotPresent
	}

	// check if this index is already present
	for _, idx := range schema.SimpleIndexs {
		if idx.FieldName == fieldName {
			return ErrDocumentDBIndexAlreadyPresent
		}
	}

	// create the index
	err = CreateIndex(dbName, fieldName, indexType, d.fd, d.user, d.client)
	if err != nil {
		return err
	}

	// Now add the index to schema
	newIndex := SIndex{
		FieldName: fieldName,
		FieldType: indexType,
	}
	schema.SimpleIndexs = append(schema.SimpleIndexs, newIndex)

	// store the modified schema
	docTables[dbName] = schema
	return d.storeDocumentDBSchemas(docTables)
}

func (d *Document) AddCompoundIndex(fieldNames []string, indexTypes []IndexType) {
	// TODO: creation of compound indexes
}

func (d *Document) OpenDocumentDB(dbName string) error {
	// load the existing db's and see if this name is present
	docTables, err := d.LoadDocumentDBSchemas()
	if err != nil {
		return err
	}
	schema, ok := docTables[dbName]
	if !ok {
		return ErrDocumentDBNotPresent
	}

	// open the simple indexes
	simpleIndexs := make(map[string]*Index)
	for _, si := range schema.SimpleIndexs {
		idx, err := OpenIndex(dbName, si.FieldName, d.fd, d.ai, d.user, d.client, d.logger)
		if err != nil {
			return err
		}
		simpleIndexs[si.FieldName] = idx
	}

	// TODO:  open the compound indexes

	// create the document DB index map
	docDB := &DocumentDB{
		name:            dbName,
		simpleIndexes:   simpleIndexs,
		compoundIndexes: nil,
	}

	// add to the open DB map
	d.openDOcDBMu.Lock()
	defer d.openDOcDBMu.Unlock()
	d.openDocDBs[dbName] = docDB

	return nil
}

func (d *Document) DeleteDocumentDB(dbName string) error {
	// load the existing db's and see if this name is already there
	docTables, err := d.LoadDocumentDBSchemas()
	if err != nil {
		return err
	}

	// check if the table exists before deleting
	if _, found := docTables[dbName]; !found {
		return ErrDocumentDBNotPresent
	}

	// delete the document db with the given name
	delete(docTables, dbName)

	// store the rest of the document db
	return d.storeDocumentDBSchemas(docTables)
}

func (d *Document) Put(dbName string, doc []byte) error {
	db := d.getOpenedDb(dbName)
	if db == nil {
		return ErrDocumentDBNoOpened
	}

	var t interface{}
	err := json.Unmarshal(doc, &t)
	if err != nil {
		return err
	}
	docMap := t.(map[string]interface{})

	// check if docMap has all the fields in the simpleIndex
	for field := range db.simpleIndexes {
		if _, found := docMap[field]; !found {
			return ErrDocumentDBIndexFieldNotPresent
		}
	}

	// upload the document
	ref, err := d.client.UploadBlob(doc, true, true)
	if err != nil {
		return err
	}

	// update the indexes
	for field, index := range db.simpleIndexes {
		v, _ := docMap[field] // it it already checked to be present
		switch index.indexType {
		case StringIndex:
			err := index.Put(v.(string), ref, StringIndex, true)
			if err != nil {
				return err
			}
		case NumberIndex:
			err := index.Put(v.(string), ref, NumberIndex, true)
			if err != nil {
				return err
			}
		case BytesIndex:
			return ErrKVIndexTypeNotSupported
		default:
			return ErrKVInvalidIndexType
		}
	}
	return nil
}

func (d *Document) Get(dbName, expr string, limit int) ([][]byte, error) {
	db := d.getOpenedDb(dbName)
	if db == nil {
		return nil, ErrDocumentDBNoOpened
	}

	var operator string
	if strings.Contains(expr, "=>") {
		operator = "=>"
	} else if strings.Contains(expr, "<=") {
		operator = "<="
	} else if strings.Contains(expr, "=") {
		operator = "="
	} else {
		return nil, ErrInvalidOperator
	}

	f := strings.Split(expr, operator)
	fieldName := f[0]
	fieldValue := f[1]

	idx, found := db.simpleIndexes[fieldName]
	if !found {
		return nil, ErrIndexNotPresent
	}

	var references [][]byte

	switch operator {
	case "=":
		refs, err := idx.Get(fieldValue)
		if err != nil {
			return nil, err
		}
		references = refs
	case "=>":
		if idx.indexType == NumberIndex {
			val, err := strconv.ParseInt(fieldValue, 10, 64)
			if err != nil {
				return nil, err
			}
			itr, err := idx.NewIntIterator(val, -1, int64(limit))
			if err != nil {
				return nil, err
			}

			for itr.Next() {
				if len(references) < limit {
					valueAll := itr.ValueAll()
					totalLen := len(references) + len(valueAll)
					if totalLen > limit {
						diff := totalLen - limit
						references = append(references, valueAll[diff:]...)
					} else {
						references = append(references, valueAll...)
					}
				} else {
					break
				}
			}
		} else {
			return nil, ErrInvalidOperator
		}
	case "<=":
		return nil, ErrNotImplemented
	default:
		return nil, ErrInvalidOperator
	}

	var docs [][]byte
	for _, ref := range references {
		if len(docs) >= limit {
			break
		}
		data, _, err := d.client.DownloadBlob(ref)
		if err != nil {
			return nil, err
		}
		docs = append(docs, data)
	}
	return docs, nil
}

//
//func (d *Document) Delete(docId string) error {
//
//}

func (d *Document) LoadDocumentDBSchemas() (map[string]DBSchema, error) {
	collections := make(map[string]DBSchema)
	topic := utils.HashString(DocumentFile)
	_, data, err := d.fd.GetFeedData(topic, d.user)
	if err != nil {
		if err.Error() != "no feed updates found" {
			return collections, err
		}
	}

	buf := bytes.NewBuffer(data)
	rd := bufio.NewReader(buf)
	for {
		line, err := rd.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("loading collections: %w", err)
		}
		line = strings.Trim(line, "\n")

		var schema DBSchema
		err = json.Unmarshal([]byte(line), &schema)
		if err != nil {
			return nil, ErrUnmarshallingDBSchema
		}
		collections[schema.Name] = schema
	}
	return collections, nil
}

func (d *Document) storeDocumentDBSchemas(collections map[string]DBSchema) error {
	buf := bytes.NewBuffer(nil)
	collectionLen := len(collections)
	if collectionLen > 0 {
		for _, schema := range collections {
			line, err := json.Marshal(schema)
			if err != nil {
				return ErrMarshallingDBSchema
			}
			buf.WriteString(string(line) + "\n")
		}
	}
	topic := utils.HashString(DocumentFile)
	_, err := d.fd.UpdateFeed(topic, d.user, buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (d *Document) IsDBOpened(dbName string) bool {
	d.openDOcDBMu.Lock()
	defer d.openDOcDBMu.Unlock()
	if _, found := d.openDocDBs[dbName]; found {
		return true
	}
	return false
}

func (d *Document) getOpenedDb(dbName string) *DocumentDB {
	d.openDOcDBMu.Lock()
	defer d.openDOcDBMu.Unlock()
	db, found := d.openDocDBs[dbName]
	if !found {
		return nil
	}
	return db
}

func (d *Document) getFieldIndex(db *DocumentDB, fieldName string) *Index {
	if index, found := db.simpleIndexes[fieldName]; found {
		return index
	}
	return nil
}
