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
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	documentFile = "document_dbs"

	// DefaultIndexFieldName is the default index identifier
	DefaultIndexFieldName = "id"
)

// Document
type Document struct {
	podName     string
	fd          *feed.API
	ai          *account.Info
	user        utils.Address
	file        *file.File
	client      blockstore.Client
	openDocDBs  map[string]*DocumentDB
	openDocDBMu sync.RWMutex
	logger      logging.Logger
	entryGetter taskmanager.TaskManagerGO
}

// DocumentDB
type DocumentDB struct {
	name          string
	mutable       bool
	simpleIndexes map[string]*Index
	mapIndexes    map[string]*Index
	listIndexes   map[string]*Index
}

// DBSchema
type DBSchema struct {
	Name            string   `json:"name"`
	Mutable         bool     `json:"mutable"`
	SimpleIndexes   []SIndex `json:"simple_indexes,omitempty"`
	MapIndexes      []SIndex `json:"map_indexes,omitempty"`
	ListIndexes     []SIndex `json:"list_indexes,omitempty"`
	CompoundIndexes []CIndex `json:"compound_indexes,omitempty"`
}

// SIndex
type SIndex struct {
	FieldName string    `json:"name"`
	FieldType IndexType `json:"type"`
}

// CIndex
type CIndex struct {
	SimpleIndexes []SIndex
}

// DocBatch
type DocBatch struct {
	db      *DocumentDB
	batches map[string]*Batch
}

// NewDocumentStore instantiates a document DB object through which all document DB are spawned.
func NewDocumentStore(podName string, fd *feed.API, ai *account.Info, user utils.Address, file *file.File,
	tm taskmanager.TaskManagerGO, client blockstore.Client, logger logging.Logger) *Document {
	return &Document{
		podName:     podName,
		fd:          fd,
		ai:          ai,
		user:        user,
		file:        file,
		client:      client,
		openDocDBs:  make(map[string]*DocumentDB),
		logger:      logger,
		entryGetter: tm,
	}
}

// CreateDocumentDB creates a new document database and its related indexes.
func (d *Document) CreateDocumentDB(dbName, encryptionPassword string, indexes map[string]IndexType, mutable bool) error {
	d.logger.Info("creating document db: ", dbName)
	if d.fd.IsReadOnlyFeed() {
		d.logger.Errorf("creating document db: %v", ErrReadOnlyIndex)
		return ErrReadOnlyIndex
	}

	// check if the db is already present and opened
	if d.IsDBOpened(dbName) {
		d.logger.Errorf("creating document db: %v", ErrDocumentDBAlreadyOpened)
		return ErrDocumentDBAlreadyOpened
	}

	// load the existing db's and see if this name is already there
	docTables, err := d.LoadDocumentDBSchemas(encryptionPassword)
	if err != nil { // skipcq: TCV-001
		return err
	}
	if _, ok := docTables[dbName]; ok {
		d.logger.Errorf("creating document db: %v", ErrDocumentDBAlreadyPresent)
		return ErrDocumentDBAlreadyPresent
	}

	// since this db is not present already, create the table
	d.logger.Info("creating simple index: ", DefaultIndexFieldName)
	err = CreateIndex(d.podName, dbName, DefaultIndexFieldName, encryptionPassword, StringIndex, d.fd, d.user, d.client, mutable)
	if err != nil { // skipcq: TCV-001
		return err
	}

	var simpleIndexes []SIndex
	var mapIndexes []SIndex
	var listIndexes []SIndex

	// create the default index
	defaultIndex := SIndex{
		FieldName: DefaultIndexFieldName,
		FieldType: StringIndex,
	}
	simpleIndexes = append(simpleIndexes, defaultIndex)

	// Now add the other indexes to simpleIndexes array
	for fieldName, fieldType := range indexes {
		// create the simple index
		err = CreateIndex(d.podName, dbName, fieldName, encryptionPassword, fieldType, d.fd, d.user, d.client, mutable)
		if err != nil { // skipcq: TCV-001
			return err
		}
		newIndex := SIndex{
			FieldName: fieldName,
			FieldType: fieldType,
		}
		if fieldType == MapIndex {
			d.logger.Info("created map index: ", dbName, fieldName, fieldType, mutable)
			mapIndexes = append(mapIndexes, newIndex)
		} else if fieldType == ListIndex {
			d.logger.Info("created list index: ", dbName, fieldName, fieldType, mutable)
			listIndexes = append(listIndexes, newIndex)
		} else {
			d.logger.Info("created simple index: ", dbName, fieldName, fieldType, mutable)
			simpleIndexes = append(simpleIndexes, newIndex)
		}
	}

	// add the simple indexes to the schema
	docTables[dbName] = DBSchema{
		Name:          dbName,
		Mutable:       mutable,
		SimpleIndexes: simpleIndexes,
		MapIndexes:    mapIndexes,
		ListIndexes:   listIndexes,
	}

	err = d.storeDocumentDBSchemas(encryptionPassword, docTables)
	if err != nil { // skipcq: TCV-001
		d.logger.Errorf("creating document db: %v", err.Error())
		return err
	}
	d.logger.Info("created document db: ", dbName)
	return nil
}

// OpenDocumentDB open a document database and its related indexes.
func (d *Document) OpenDocumentDB(dbName, encryptionPassword string) error {
	d.logger.Info("opening document db: ", dbName)
	// check if the db is already present and opened
	if d.IsDBOpened(dbName) { // skipcq: TCV-001
		d.logger.Errorf("opening document db: %v", ErrDocumentDBAlreadyOpened)
		return ErrDocumentDBAlreadyOpened
	}

	// load the existing db's and see if this name is present
	docTables, err := d.LoadDocumentDBSchemas(encryptionPassword)
	if err != nil { // skipcq: TCV-001
		d.logger.Errorf("opening document db: %v", err.Error())
		return err
	}
	schema, ok := docTables[dbName]
	if !ok { // skipcq: TCV-001
		d.logger.Errorf("opening document db: %v", ErrDocumentDBNotPresent)
		return ErrDocumentDBNotPresent
	}

	// open the simple indexes
	simpleIndexs := make(map[string]*Index)
	for _, si := range schema.SimpleIndexes {
		d.logger.Info("opening simple index: ", si.FieldName)
		idx, err := OpenIndex(d.podName, dbName, si.FieldName, encryptionPassword, d.fd, d.ai, d.user, d.client, d.logger)
		if err != nil { // skipcq: TCV-001
			d.logger.Errorf("opening simple index: %v", err.Error())
			return err
		}
		simpleIndexs[si.FieldName] = idx
	}

	// open the map indexes
	mapIndexs := make(map[string]*Index)
	for _, mi := range schema.MapIndexes {
		d.logger.Info("opening map index: ", mi.FieldName)
		idx, err := OpenIndex(d.podName, dbName, mi.FieldName, encryptionPassword, d.fd, d.ai, d.user, d.client, d.logger)
		if err != nil { // skipcq: TCV-001
			d.logger.Errorf("opening map index: %v", err.Error())
			return err
		}
		mapIndexs[mi.FieldName] = idx
	}

	// open the list indexes
	listIndexes := make(map[string]*Index)
	for _, li := range schema.ListIndexes {
		d.logger.Info("opening list index: ", li.FieldName)
		idx, err := OpenIndex(d.podName, dbName, li.FieldName, encryptionPassword, d.fd, d.ai, d.user, d.client, d.logger)
		if err != nil { // skipcq: TCV-001
			d.logger.Errorf("opening list index: %v", err.Error())
			return err
		}
		listIndexes[li.FieldName] = idx
	}

	// create the document DB index map
	docDB := &DocumentDB{
		name:          schema.Name,
		mutable:       schema.Mutable,
		simpleIndexes: simpleIndexs,
		mapIndexes:    mapIndexs,
		listIndexes:   listIndexes,
	}

	// add to the open DB map
	d.addToOpenedDb(dbName, docDB)
	d.logger.Info("document db opened: ", schema.Name)
	return nil
}

// DeleteDocumentDB a document DB, all its data and its related indxes.
func (d *Document) DeleteDocumentDB(dbName, encryptionPassword string) error {
	d.logger.Info("deleting document db: ", dbName)
	if d.fd.IsReadOnlyFeed() { // skipcq: TCV-001
		d.logger.Errorf("deleting document db: %v", ErrReadOnlyIndex)
		return ErrReadOnlyIndex
	}

	// load the existing db's and see if this name is already there
	docTables, err := d.LoadDocumentDBSchemas(encryptionPassword)
	if err != nil { // skipcq: TCV-001
		d.logger.Errorf("deleting document db: %v", err.Error())
		return err
	}

	// check if the table exists before deleting
	_, found := docTables[dbName]
	if !found { // skipcq: TCV-001
		d.logger.Errorf("deleting document db: %v", ErrDocumentDBNotPresent)
		return ErrDocumentDBNotPresent
	}

	// open and delete the indexes
	if !d.IsDBOpened(dbName) {
		err = d.OpenDocumentDB(dbName, encryptionPassword)
		if err != nil { // skipcq: TCV-001
			d.logger.Errorf("deleting document db: %v", err.Error())
			return err
		}
	}
	defer d.removeFromOpenedDB(dbName)

	docDB := d.getOpenedDb(dbName)
	//TODO: before deleting the indexes, unpin all the documents referenced in the ID index
	for _, si := range docDB.simpleIndexes {
		d.logger.Info("deleting simple index: ", si.name, si.indexType)
		err = si.DeleteIndex(encryptionPassword)
		if err != nil { // skipcq: TCV-001
			d.logger.Errorf("deleting simple index: %v", err.Error())
			return err
		}
	}
	for _, mi := range docDB.mapIndexes {
		d.logger.Info("deleting map index: ", mi.name, mi.indexType)
		err = mi.DeleteIndex(encryptionPassword)
		if err != nil { // skipcq: TCV-001
			d.logger.Errorf("deleting map index: %v", err.Error())
			return err
		}
	}
	for _, li := range docDB.listIndexes {
		d.logger.Info("deleting list index: ", li.name, li.indexType)
		err = li.DeleteIndex(encryptionPassword)
		if err != nil { // skipcq: TCV-001
			d.logger.Errorf("deleting map index: %v", err.Error())
			return err
		}
	}

	// delete the document db from the DB file
	delete(docTables, dbName)

	if len(docTables) == 0 {
		docTables = map[string]DBSchema{}
	}

	// store the rest of the document db
	err = d.storeDocumentDBSchemas(encryptionPassword, docTables)
	if err != nil { // skipcq: TCV-001
		d.logger.Errorf("deleting document db: ", err.Error())
		return err
	}

	d.logger.Info("deleted document db: ", dbName)
	return nil
}

// DeleteAllDocumentDBs deletes all document DBs, all their data and related indxes.
func (d *Document) DeleteAllDocumentDBs(encryptionPassword string) error {
	if d.fd.IsReadOnlyFeed() { // skipcq: TCV-001
		d.logger.Errorf("deleting document db: %v", ErrReadOnlyIndex)
		return ErrReadOnlyIndex
	}

	// load the existing db's and see if this name is already there
	docTables, err := d.LoadDocumentDBSchemas(encryptionPassword)
	if err != nil { // skipcq: TCV-001
		d.logger.Errorf("deleting document db: %v", err.Error())
		return err
	}

	for dbName := range docTables {
		// open and delete the indexes
		if !d.IsDBOpened(dbName) {
			err = d.OpenDocumentDB(dbName, encryptionPassword)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("deleting document db: %v", err.Error())
				return err
			}
		}
		defer d.removeFromOpenedDB(dbName)

		docDB := d.getOpenedDb(dbName)
		//TODO: before deleting the indexes, unpin all the documents referenced in the ID index
		for _, si := range docDB.simpleIndexes {
			d.logger.Info("deleting simple index: ", si.name, si.indexType)
			err = si.DeleteIndex(encryptionPassword)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("deleting simple index: %v", err.Error())
				return err
			}
		}
		for _, mi := range docDB.mapIndexes {
			d.logger.Info("deleting map index: ", mi.name, mi.indexType)
			err = mi.DeleteIndex(encryptionPassword)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("deleting map index: %v", err.Error())
				return err
			}
		}
		for _, li := range docDB.listIndexes {
			d.logger.Info("deleting list index: ", li.name, li.indexType)
			err = li.DeleteIndex(encryptionPassword)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("deleting map index: %v", err.Error())
				return err
			}
		}

		// delete the document db from the DB file
		delete(docTables, dbName)

		d.logger.Info("deleted document db: ", dbName)
	}
	docTables = map[string]DBSchema{}
	err = d.storeDocumentDBSchemas(encryptionPassword, docTables)
	if err != nil { // skipcq: TCV-001
		d.logger.Errorf("deleting document db: ", err.Error())
		return err
	}
	return nil
}

// Count counts the number of document in a document DB which matches a given expression
func (d *Document) Count(dbName, expr string) (uint64, error) {
	d.logger.Info("counting document db: ", dbName, expr)
	db := d.getOpenedDb(dbName)
	if db == nil { // skipcq: TCV-001
		d.logger.Errorf("counting document db: %v", ErrDocumentDBNotOpened)
		return 0, ErrDocumentDBNotOpened
	}

	// count all documents
	if expr == "" {
		idx, found := db.simpleIndexes[DefaultIndexFieldName]
		if !found { // skipcq: TCV-001
			d.logger.Errorf("counting document db: %v", ErrIndexNotPresent)
			return 0, ErrIndexNotPresent
		}
		return idx.CountIndex(idx.encryptionPassword)
	}

	// count documents based on expression
	fieldName, operator, fieldValue, err := d.resolveExpression(expr)
	if err != nil { // skipcq: TCV-001
		d.logger.Errorf("counting document db: %v", err.Error())
		return 0, err
	}
	idx, found := db.simpleIndexes[fieldName]
	if !found {
		idx, found = db.mapIndexes[fieldName]
		if !found {
			idx, found = db.listIndexes[fieldName]
			if !found { // skipcq: TCV-001
				d.logger.Errorf("counting document db: %v", ErrIndexNotPresent)
				return 0, ErrIndexNotPresent
			}
		} else {
			fieldValue = strings.ReplaceAll(fieldValue, ":", "")
		}
	}

	switch idx.indexType {
	case StringIndex:
		itr, err := idx.NewStringIterator(fieldValue, "", -1)
		if err != nil { // skipcq: TCV-001
			d.logger.Errorf("counting document db: ", err.Error())
			return 0, err
		}
		switch operator {
		case "=":
			itr.Next()
			re, err := regexp.Compile(fieldValue)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("counting document db: %v", err.Error())
				return 0, err
			}
			matched := re.Match([]byte(itr.StringKey()))
			if matched {
				refs := itr.ValueAll()
				return uint64(len(refs)), nil
			}
		case "=>", ">": // skipcq: TCV-001
			var count uint64
			re, err := regexp.Compile(fieldValue)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("counting document db: %v", err.Error())
				return 0, err
			}

			for itr.Next() {
				matched := re.Match([]byte(itr.StringKey()))
				if matched {
					refs := itr.ValueAll()
					count = count + uint64(len(refs))
				}
			}
			d.logger.Info("counting document db: ", dbName, expr, count)
			return count, nil
		}
	case MapIndex, ListIndex:
		itr, err := idx.NewStringIterator(fieldValue, "", -1)
		if err != nil { // skipcq: TCV-001
			d.logger.Errorf("counting document db: ", err.Error())
			return 0, err
		}
		switch operator {
		case "=":
			itr.Next()
			refs := itr.ValueAll()
			count := uint64(len(refs))
			d.logger.Info("counting document db: ", dbName, expr, count)
			return count, nil
		case "=>": // skipcq: TCV-001
			var count uint64
			for itr.Next() {
				refs := itr.ValueAll()
				count = count + uint64(len(refs))
			}
			d.logger.Info("counting document db: ", dbName, expr, count)
			return count, nil
		case ">": // skipcq: TCV-001
			var count uint64
			for itr.Next() {
				if itr.StringKey() == fieldValue {
					continue
				}
				refs := itr.ValueAll()
				count = count + uint64(len(refs))
			}
			d.logger.Info("counting document db: ", dbName, expr, count)
			return count, nil
		}
	case NumberIndex:
		start, err := strconv.ParseInt(fieldValue, 10, 64)
		if err != nil { // skipcq: TCV-001
			d.logger.Errorf("counting document db: ", err.Error())
			return 0, err
		}
		itr, err := idx.NewIntIterator(start, -1, -1)
		if err != nil { // skipcq: TCV-001
			d.logger.Errorf("counting document db: ", err.Error())
			return 0, err
		}
		switch operator {
		case "=":
			itr.Next()
			refs := itr.ValueAll()
			count := uint64(len(refs))
			d.logger.Info("counting document db: ", dbName, expr, count)
			return count, nil
		case "=>":
			var count uint64
			for itr.Next() {
				refs := itr.ValueAll()
				count = count + uint64(len(refs))
			}
			d.logger.Info("counting document db: ", dbName, expr, count)
			return count, nil
		case ">":
			var count uint64
			for itr.Next() {
				if itr.IntegerKey() == start {
					continue
				}
				refs := itr.ValueAll()
				count = count + uint64(len(refs))
			}
			d.logger.Info("counting document db: ", dbName, expr, count)
			return count, nil
		}
	case BytesIndex: // skipcq: TCV-001
		d.logger.Errorf("counting document db: ", ErrIndexNotSupported)
		return 0, ErrIndexNotSupported
	default: // skipcq: TCV-001
		d.logger.Errorf("counting document db: ", ErrInvalidIndexType)
		return 0, ErrInvalidIndexType
	}
	return 0, nil // skipcq: TCV-001
}

// Put inserts a document in to a document database.
func (d *Document) Put(dbName string, doc []byte) error {
	d.logger.Info("inserting in to document db: ", dbName, len(doc))
	if d.fd.IsReadOnlyFeed() { // skipcq: TCV-001
		d.logger.Errorf("inserting in to document db: ", ErrReadOnlyIndex)
		return ErrReadOnlyIndex
	}

	db := d.getOpenedDb(dbName)
	if db == nil { // skipcq: TCV-001
		d.logger.Errorf("inserting in to document db: ", ErrDocumentDBNotOpened)
		return ErrDocumentDBNotOpened
	}

	if !db.mutable {
		d.logger.Errorf("inserting in to document db: ", ErrModifyingImmutableDocDB)
		return ErrModifyingImmutableDocDB
	}

	var t interface{}
	err := json.Unmarshal(doc, &t)
	if err != nil { // skipcq: TCV-001
		d.logger.Errorf("inserting in to document db: ", err.Error())
		return err
	}
	docMap := t.(map[string]interface{})

	// check if docMap has all the fields in the simpleIndex
	for field := range db.simpleIndexes {
		if _, found := docMap[field]; !found {
			d.logger.Errorf("inserting in to document db: ", ErrDocumentDBIndexFieldNotPresent)
			return ErrDocumentDBIndexFieldNotPresent
		}
	}

	// check if the id is already present
	// and remove it if it is present
	idValue := docMap[DefaultIndexFieldName]
	switch v := idValue.(type) {
	case string:
		if v == "" {
			d.logger.Errorf("inserting in to document db: ", ErrInvalidDocumentId)
			return ErrInvalidDocumentId
		} else {
			idIndex := db.simpleIndexes[DefaultIndexFieldName]
			refs, err := idIndex.Get(v)
			if err != nil {
				break
			}
			if len(refs) > 0 {
				err = d.Del(dbName, v)
				if err != nil { // skipcq: TCV-001
					d.logger.Errorf("inserting in to document db: ", err.Error())
					return err
				}
			}
			d.logger.Info("removed already existing doc of the same id: ", v)
		}
	default: // skipcq: TCV-001
		d.logger.Errorf("inserting in to document db: ", ErrInvalidIndexType)
		return ErrInvalidIndexType
	}

	// upload the document
	ref, err := d.client.UploadBlob(doc, 0, true)
	if err != nil { // skipcq: TCV-001
		d.logger.Errorf("inserting in to document db: ", err.Error())
		return err
	}
	d.logger.Info("upload the document in document db: ", dbName, len(doc))

	// update the indexes
	indexes := make(map[string]*Index)
	for field, index := range db.simpleIndexes {
		indexes[field] = index
	}
	for field, index := range db.mapIndexes {
		indexes[field] = index
	}
	for field, index := range db.listIndexes {
		indexes[field] = index
	}
	for field, index := range indexes {
		v := docMap[field] // it is already checked to be present
		switch index.indexType {
		case StringIndex:
			apnd := true
			if field == DefaultIndexFieldName {
				apnd = false
			}
			err := index.Put(v.(string), ref, StringIndex, apnd)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("inserting in to document db: ", err.Error())
				return err
			}
			d.logger.Info("updating in to simple index: ", dbName, v.(string))
		case MapIndex:
			valMap := v.(map[string]interface{})
			for keyField, vf := range valMap {
				valueField := vf.(string)
				mapField := keyField + valueField
				err := index.Put(mapField, ref, StringIndex, true)
				if err != nil { // skipcq: TCV-001
					d.logger.Errorf("inserting in to document db: ", err.Error())
					return err
				}
				d.logger.Info("updating map index: ", dbName, keyField, valueField)
			}
		case ListIndex:
			valList := v.([]interface{})
			for _, listVal := range valList {
				err := index.Put(listVal.(string), ref, StringIndex, true)
				if err != nil {
					d.logger.Errorf("inserting in to document db: ", err.Error())
					return err
				}
				d.logger.Info("updating list index: ", dbName, listVal)
			}
		case NumberIndex:
			val := v.(float64)
			// valStr := strconv.FormatFloat(val, 'f', 6, 64)
			err := index.PutNumber(val, ref, NumberIndex, true)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("inserting in to document db: ", err.Error())
				return err
			}
			d.logger.Info("updating number index: ", dbName, val)
		case BytesIndex: // skipcq: TCV-001
			d.logger.Errorf("inserting in to document db: ", ErrIndexNotSupported)
			return ErrIndexNotSupported
		default: // skipcq: TCV-001
			d.logger.Errorf("inserting in to document db: ", ErrInvalidIndexType)
			return ErrInvalidIndexType
		}
	}
	return nil
}

// Get retrieves a specific document from a document database matching the dcument id.
func (d *Document) Get(dbName, id, podPassword string) ([]byte, error) {
	d.logger.Info("getting from document db: ", dbName, id)
	db := d.getOpenedDb(dbName)
	if db == nil { // skipcq: TCV-001
		d.logger.Errorf("getting from document db: ", ErrDocumentDBNotOpened)
		return nil, ErrDocumentDBNotOpened
	}

	idIndex := db.simpleIndexes[DefaultIndexFieldName]
	reference, err := idIndex.Get(id)
	if err != nil {
		d.logger.Errorf("getting from document db: ", err.Error())
		return nil, err
	}

	if len(reference) == 0 { // skipcq: TCV-001
		d.logger.Errorf("getting from document db: ", ErrDocumentNotPresent)
		return nil, ErrDocumentNotPresent
	}

	if idIndex.mutable {
		data, _, err := d.client.DownloadBlob(reference[0])
		if err != nil { // skipcq: TCV-001
			d.logger.Errorf("getting from document db: ", err.Error())
			return nil, err
		}
		d.logger.Info("getting from document db: ", dbName, id, len(data))
		return data, nil
	} else { // skipcq: TCV-001
		b := bytes.NewBuffer(reference[0])
		seekOffset, err := binary.ReadUvarint(b)
		if err != nil {
			d.logger.Errorf("getting from document db: ", err.Error())
			return nil, err
		}

		data, err := d.getLineFromFile(idIndex.podFile, podPassword, seekOffset)
		if err != nil {
			d.logger.Errorf("getting from document db: ", err.Error())
			return nil, err
		}
		d.logger.Info("getting from document db: ", dbName, id, len(data))
		return data, nil
	}

}

// Del deletes a specific document from a document database matching a document id.
func (d *Document) Del(dbName, id string) error {
	d.logger.Info("deleting from document db: ", dbName, id)
	if d.fd.IsReadOnlyFeed() { // skipcq: TCV-001
		d.logger.Errorf("deleting from document db: ", ErrReadOnlyIndex)
		return ErrReadOnlyIndex
	}

	db := d.getOpenedDb(dbName)
	if db == nil { // skipcq: TCV-001
		d.logger.Errorf("deleting from document db: ", ErrDocumentDBNotOpened)
		return ErrDocumentDBNotOpened
	}

	if !db.mutable { // skipcq: TCV-001
		d.logger.Errorf("deleting from document db: ", ErrModifyingImmutableDocDB)
		return ErrModifyingImmutableDocDB
	}

	// get the "id" index and retrieve the original document
	idx := db.simpleIndexes[DefaultIndexFieldName]
	refs, err := idx.Get(id)
	if err != nil { // skipcq: TCV-001
		if errors.Is(err, ErrEntryNotFound) {
			return nil
		}
		return err
	}
	if len(refs) <= 0 {
		return nil
	}

	data, _, err := d.client.DownloadBlob(refs[0])
	if err != nil { // skipcq: TCV-001
		d.logger.Errorf("deleting from document db: ", err.Error())
		return err
	}

	var t interface{}
	err = json.Unmarshal(data, &t)
	if err != nil { // skipcq: TCV-001
		d.logger.Errorf("deleting from document db: ", err.Error())
		return err
	}
	docMap := t.(map[string]interface{})

	// delete all the indexes of the doc
	for field, index := range db.simpleIndexes {
		v := docMap[field] // it is already checked to be present
		switch index.indexType {
		case StringIndex:
			_, err := index.Delete(v.(string))
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("deleting from document db: ", err.Error())
				return err
			}
			d.logger.Info("deleting from simple index: ", dbName, id, v.(string))
		case MapIndex: // skipcq: TCV-001
			valMap := v.(map[string]interface{})
			for keyField, valueField := range valMap {
				vf := valueField.(string)
				mapField := keyField + vf
				_, err := index.Delete(mapField)
				if err != nil {
					d.logger.Errorf("deleting from document db: ", err.Error())
					return err
				}
				d.logger.Info("deleting from map index: ", dbName, id, keyField, vf)
			}
		case ListIndex: // skipcq: TCV-001
			valList := v.([]interface{})
			for _, listVal := range valList {
				_, err := index.Delete(listVal.(string))
				if err != nil {
					d.logger.Errorf("deleting from document db: ", err.Error())
					return err
				}
				d.logger.Info("deleting from list index: ", dbName, id, listVal)
			}
		case NumberIndex:
			val := v.(float64)
			// valStr := strconv.FormatFloat(val, 'f', 6, 64)
			_, err := index.DeleteNumber(val)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("deleting from document db: ", err.Error())
				return err
			}
			d.logger.Info("deleting from number index: ", dbName, id, val)
		case BytesIndex: // skipcq: TCV-001
			d.logger.Errorf("deleting from document db: ", ErrIndexNotSupported)
			return ErrIndexNotSupported
		default: // skipcq: TCV-001
			d.logger.Errorf("deleting from document db: ", ErrInvalidIndexType)
			return ErrInvalidIndexType
		}
	}

	// delete the original data (unpin)
	err = d.client.DeleteReference(refs[0])
	if err != nil { // skipcq: TCV-001
		d.logger.Errorf("deleting from document db: ", err.Error())
		return err
	}

	d.logger.Info("deleted document from document db: ", dbName, id, utils.NewReference(refs[0]).String())
	return nil
}

// Find selects a number of rows from a document database matching an expression.
func (d *Document) Find(dbName, expr, podPassword string, limit int) ([][]byte, error) {
	d.logger.Info("finding from document db: ", dbName, expr, limit)
	db := d.getOpenedDb(dbName)
	if db == nil { // skipcq: TCV-001
		d.logger.Errorf("finding from document db: ", ErrDocumentDBNotOpened)
		return nil, ErrDocumentDBNotOpened
	}

	// find all documents
	if expr == "" {
		idx, found := db.simpleIndexes[DefaultIndexFieldName]
		if !found { // skipcq: TCV-001
			d.logger.Errorf("finding from document db: ", ErrIndexNotPresent)
			return nil, ErrIndexNotPresent
		}
		return idx.Get("")
	}

	fieldName, operator, fieldValue, err := d.resolveExpression(expr)
	if err != nil { // skipcq: TCV-001
		d.logger.Errorf("finding from document db: ", err.Error())
		return nil, err
	}

	idx, found := db.simpleIndexes[fieldName]
	if !found {
		idx, found = db.mapIndexes[fieldName]
		if !found { // skipcq: TCV-001
			idx, found = db.listIndexes[fieldName]
			if !found {
				d.logger.Errorf("finding from document db: ", ErrIndexNotPresent)
				return nil, ErrIndexNotPresent
			}
		} else {
			fieldValue = strings.ReplaceAll(fieldValue, ":", "")
		}
	}

	var references [][]byte
	switch idx.indexType {
	case StringIndex:
		itr, err := idx.NewStringIterator(fieldValue, "", -1)
		if err != nil { // skipcq: TCV-001
			d.logger.Errorf("finding from document db: ", err.Error())
			return nil, err
		}
		switch operator {
		case "=":
			itr.Next()
			re, err := regexp.Compile(fieldValue)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("finding from document db: %v", err.Error())
				return nil, err
			}
			matched := re.Match([]byte(itr.StringKey()))
			if matched {
				references = itr.ValueAll()
			}
		case "=>", ">": // skipcq: TCV-001
			re, err := regexp.Compile(fieldValue)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("finding from document db: %v", err.Error())
				return nil, err
			}

			for itr.Next() {
				matched := re.Match([]byte(itr.StringKey()))
				if matched {
					refs := itr.ValueAll()
					references = append(references, refs...)
				}
			}
		default:
			return nil, fmt.Errorf("operator is not available: %s", operator)
		}
	case MapIndex, ListIndex:
		itr, err := idx.NewStringIterator(fieldValue, "", int64(limit))
		if err != nil { // skipcq: TCV-001
			d.logger.Errorf("finding from document db: ", err.Error())
			return nil, err
		}
		switch operator {
		case "=":
			itr.Next()
			references = itr.ValueAll()
		case "=>": // skipcq: TCV-001
			for itr.Next() {
				if limit > 0 && references != nil && len(references) > limit {
					break
				}
				refs := itr.ValueAll()
				references = append(references, refs...)
			}
		case ">": // skipcq: TCV-001
			for itr.Next() {
				if limit > 0 && references != nil && len(references) > limit {
					break
				}
				if itr.StringKey() == fieldValue {
					continue
				}
				refs := itr.ValueAll()
				references = append(references, refs...)
			}
		default:
			return nil, fmt.Errorf("operator is not available: %s", operator)
		}
	case NumberIndex:
		var start int64 = 0
		if operator != "<" && operator != "<=" {
			start, err = strconv.ParseInt(fieldValue, 10, 64)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("finding from document db: ", err.Error())
				return nil, err
			}
		} else if operator == "!=" {
			start = -1
		}
		itr, err := idx.NewIntIterator(start, -1, int64(limit))
		if err != nil { // skipcq: TCV-001
			d.logger.Errorf("finding from document db: ", err.Error())
			return nil, err
		}
		switch operator {
		case "=":
			itr.Next()
			references = itr.ValueAll()
		case "!=":
			for itr.Next() {
				if limit > 0 && references != nil && len(references) > limit { // skipcq: TCV-001
					break
				}
				val, err := strconv.ParseFloat(itr.StringKey(), 64)
				if err != nil { // skipcq: TCV-001
					break
				}
				end, err := strconv.ParseFloat(fieldValue, 64)
				if err != nil { // skipcq: TCV-001
					break
				}

				if val == end {
					continue
				}
				refs := itr.ValueAll()
				references = append(references, refs...)
			}
		case "=>":
			for itr.Next() {
				if limit > 0 && references != nil && len(references) > limit { // skipcq: TCV-001
					break
				}
				refs := itr.ValueAll()
				references = append(references, refs...)
			}
		case "<":
			end, err := strconv.ParseInt(fieldValue, 10, 64)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("finding from document db: ", err.Error())
				break
			}
			for itr.Next() {
				if limit > 0 && references != nil && len(references) > limit { // skipcq: TCV-001
					break
				}
				val, err := strconv.ParseInt(itr.StringKey(), 10, 64)
				if err != nil { // skipcq: TCV-001
					continue
				}
				if val >= end {
					break
				}
				refs := itr.ValueAll()
				references = append(references, refs...)
			}
		case "<=":
			end, err := strconv.ParseInt(fieldValue, 10, 64)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("finding from document db: ", err.Error())
				break
			}
			for itr.Next() {
				if limit > 0 && references != nil && len(references) > limit { // skipcq: TCV-001
					break
				}
				val, err := strconv.ParseInt(itr.StringKey(), 10, 64)
				if err != nil { // skipcq: TCV-001
					continue
				}
				if val > end {
					break
				}
				refs := itr.ValueAll()
				references = append(references, refs...)
			}
		case ">":
			for itr.Next() {
				if limit > 0 && references != nil && len(references) > limit { // skipcq: TCV-001
					break
				}
				if itr.IntegerKey() == start {
					continue
				}
				refs := itr.ValueAll()
				references = append(references, refs...)
			}
		default:
			return nil, fmt.Errorf("operator is not available: %s", operator)
		}
	case BytesIndex: // skipcq: TCV-001
		d.logger.Errorf("finding from document db: ", ErrIndexNotSupported)
		return nil, ErrIndexNotSupported
	default: // skipcq: TCV-001
		d.logger.Errorf("finding from document db: ", ErrInvalidIndexType)
		return nil, ErrInvalidIndexType
	}
	docs := [][]byte{}
	if idx.mutable {
		wg := new(sync.WaitGroup)
		mtx := &sync.Mutex{}
		for _, ref := range references {
			if limit > 0 && len(docs) >= limit {
				break
			}
			wg.Add(1)
			et := newEntryTask(d.client, &docs, ref, mtx)
			err := et.Execute(context.TODO())
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("finding from document db: ", err.Error())
			}
			wg.Done()
		}
		wg.Wait()
		d.logger.Info("found document from document db: ", dbName, expr, len(docs))
		return docs, nil
	} else { // skipcq: TCV-001
		for _, ref := range references {
			if limit > 0 && len(docs) >= limit {
				break
			}
			b := bytes.NewBuffer(ref)
			seekOffset, err := binary.ReadUvarint(b)
			if err != nil {
				d.logger.Errorf("getting from document db: ", err.Error())
				return nil, err
			}
			data, err := d.getLineFromFile(idx.podFile, podPassword, seekOffset)
			if err != nil {
				d.logger.Errorf("finding from document db: ", err.Error())
				return nil, err
			}
			docs = append(docs, data)
		}
		d.logger.Info("found document from document db: ", dbName, expr, len(docs))
		return docs, nil
	}
}

// LoadDocumentDBSchemas loads the schema of all documents belonging to a pod.
func (d *Document) LoadDocumentDBSchemas(encryptionPassword string) (map[string]DBSchema, error) {
	collections := make(map[string]DBSchema)
	topic := utils.HashString(documentFile)
	_, data, err := d.fd.GetFeedData(topic, d.user, []byte(encryptionPassword))
	if err != nil {
		if err.Error() != "feed does not exist or was not updated yet" { // skipcq: TCV-001
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
		if err != nil { // skipcq: TCV-001
			return nil, fmt.Errorf("loading collections: %v", err.Error())
		}
		line = strings.Trim(line, "\n")

		var schema DBSchema
		err = json.Unmarshal([]byte(line), &schema)
		if err != nil { // skipcq: TCV-001
			return nil, ErrUnmarshallingDBSchema
		}
		collections[schema.Name] = schema
	}
	return collections, nil
}

// IsDBOpened is used to check if a document DB is opened or not.
func (d *Document) IsDBOpened(dbName string) bool {
	d.openDocDBMu.Lock()
	defer d.openDocDBMu.Unlock()
	if _, found := d.openDocDBs[dbName]; found {
		return true
	}
	return false
}

func (d *Document) storeDocumentDBSchemas(encryptionPassword string, collections map[string]DBSchema) error {
	buf := bytes.NewBuffer(nil)
	collectionLen := len(collections)
	if collectionLen > 0 {
		for _, schema := range collections {
			line, err := json.Marshal(schema)
			if err != nil { // skipcq: TCV-001
				return ErrMarshallingDBSchema
			}
			buf.WriteString(string(line) + "\n")
		}
	}
	topic := utils.HashString(documentFile)
	_, err := d.fd.UpdateFeed(topic, d.user, buf.Bytes(), []byte(encryptionPassword))
	if err != nil { // skipcq: TCV-001
		return err
	}
	return nil
}

func (d *Document) getOpenedDb(dbName string) *DocumentDB {
	d.openDocDBMu.Lock()
	defer d.openDocDBMu.Unlock()
	db, found := d.openDocDBs[dbName]
	if !found { // skipcq: TCV-001
		return nil
	}
	return db
}

func (d *Document) addToOpenedDb(dbName string, docDB *DocumentDB) {
	d.openDocDBMu.Lock()
	defer d.openDocDBMu.Unlock()
	d.openDocDBs[dbName] = docDB
}

func (d *Document) removeFromOpenedDB(dbName string) {
	d.openDocDBMu.Lock()
	defer d.openDocDBMu.Unlock()
	delete(d.openDocDBs, dbName)
}

func (*Document) resolveExpression(expr string) (string, string, string, error) {
	var operator string
	if strings.Contains(expr, "=>") {
		operator = "=>"
	} else if strings.Contains(expr, "<=") { // skipcq: TCV-001
		operator = "<="
	} else if strings.Contains(expr, "<") { // skipcq: TCV-001
		operator = "<"
	} else if strings.Contains(expr, ">") {
		operator = ">"
	} else if strings.Contains(expr, "!=") {
		operator = "!="
	} else if strings.Contains(expr, "=") {
		operator = "="
	} else { // skipcq: TCV-001
		return "", "", "", ErrInvalidOperator
	}
	f := strings.Split(expr, operator)
	fieldName := f[0]
	fieldValue := f[1]

	return fieldName, operator, fieldValue, nil
}

// CreateDocBatch creates a batch index instead of normal index. This is used when doing a bulk insert.
func (d *Document) CreateDocBatch(dbName, podPassword string) (*DocBatch, error) {
	d.logger.Info("creating batch for inserting in document db: ", dbName)
	if d.fd.IsReadOnlyFeed() { // skipcq: TCV-001
		d.logger.Errorf("creating batch: ", ErrReadOnlyIndex)
		return nil, ErrReadOnlyIndex
	}

	// see if the document db is empty
	data, err := d.Find(dbName, "", podPassword, 1)
	if err != nil {
		if !errors.Is(err, ErrEntryNotFound) { // skipcq: TCV-001
			d.logger.Errorf("creating simple batch index: ", err.Error())
			return nil, err
		}
	}
	if data != nil { // skipcq: TCV-001
		d.logger.Errorf("creating simple batch index: ", ErrModifyingImmutableDocDB)
		return nil, ErrModifyingImmutableDocDB
	}

	d.openDocDBMu.Lock()
	defer d.openDocDBMu.Unlock()
	if db, ok := d.openDocDBs[dbName]; ok {
		var docBatch DocBatch
		docBatch.db = db
		docBatch.batches = make(map[string]*Batch)

		for fieldName, idx := range db.simpleIndexes {
			batch, err := NewBatch(idx)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("creating simple batch index: ", err.Error())
				return nil, err
			}
			docBatch.batches[fieldName] = batch
			d.logger.Info("created simple batch index: ", fieldName)
		}
		for fieldName, idx := range db.mapIndexes {
			batch, err := NewBatch(idx)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("creating map batch index: ", err.Error())
				return nil, err
			}
			docBatch.batches[fieldName] = batch
			d.logger.Info("created map batch index: ", fieldName)
		}
		for fieldName, idx := range db.listIndexes {
			batch, err := NewBatch(idx)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("creating list batch index: ", err.Error())
				return nil, err
			}
			docBatch.batches[fieldName] = batch
			d.logger.Info("created list batch index: ", fieldName)
		}

		d.logger.Info("created batch for inserting in document db: ", dbName)
		return &docBatch, nil
	}
	d.logger.Errorf("creating batch: ", ErrDocumentDBNotOpened) // skipcq: TCV-001
	return nil, ErrDocumentDBNotOpened                          // skipcq: TCV-001
}

// DocBatchPut is used to insert a single document to the batch index.
func (d *Document) DocBatchPut(docBatch *DocBatch, doc []byte, index int64) error {
	if d.fd.IsReadOnlyFeed() { // skipcq: TCV-001
		d.logger.Errorf("inserting in batch: ", ErrReadOnlyIndex)
		return ErrReadOnlyIndex
	}

	d.openDocDBMu.Lock()
	defer d.openDocDBMu.Unlock()

	var t interface{}
	err := json.Unmarshal(doc, &t)
	if err != nil { // skipcq: TCV-001
		d.logger.Errorf("inserting in batch: ", err.Error())
		return err
	}
	switch t.(type) {
	case map[string]interface{}:
		// it's an object
		docMap := t.(map[string]interface{})

		// check if docMap has all the fields in the simpleIndex
		for field := range docBatch.db.simpleIndexes {
			if _, found := docMap[field]; !found { // skipcq: TCV-001
				d.logger.Errorf("inserting in batch: ", ErrDocumentDBIndexFieldNotPresent)
				return ErrDocumentDBIndexFieldNotPresent
			}
		}

		var ref []byte
		if docBatch.db.mutable {

			// check if the id is already present
			// and remove it if it is present
			var valStr string
			idValue := docMap[DefaultIndexFieldName]
			switch v := idValue.(type) {
			case float64: // skipcq: TCV-001
				valStr = strconv.FormatFloat(v, 'f', 6, 64)
			case string:
				valStr = v
			default: // skipcq: TCV-001
				return ErrInvalidIndexType
			}

			if valStr == "" { // skipcq: TCV-001
				d.logger.Errorf("inserting in batch: ", ErrInvalidDocumentId)
				return ErrInvalidDocumentId
			} else {
				idBatchIndex := docBatch.batches[DefaultIndexFieldName]
				refs, err := idBatchIndex.Get(valStr)
				if err == nil { // skipcq: TCV-001
					// found a doc with the same id, so remove it and all the indexes
					if len(refs) > 0 {
						data, _, err := d.client.DownloadBlob(refs[0])
						if err != nil {
							d.logger.Errorf("inserting in batch: ", err.Error())
							return err
						}

						var t interface{}
						err = json.Unmarshal(data, &t)
						if err != nil {
							d.logger.Errorf("inserting in batch: ", err.Error())
							return err
						}
						oldDocMap := t.(map[string]interface{})

						for field, batchIndex := range docBatch.batches {
							v1 := oldDocMap[field] // it is already checked to be present
							switch batchIndex.idx.indexType {
							case StringIndex:
								_, err := batchIndex.Del(v1.(string))
								if err != nil {
									d.logger.Errorf("inserting in batch: ", err.Error())
									return err
								}
							case MapIndex:
								valMap := v1.(map[string]interface{})
								for keyField, valueField := range valMap {
									vf := valueField.(string)
									mapField := keyField + vf
									_, err := batchIndex.Del(mapField)
									if err != nil {
										d.logger.Errorf("inserting in batch: ", err.Error())
										return err
									}
								}
							case ListIndex:
								valList := v1.([]interface{})
								for _, listVal := range valList {
									_, err := batchIndex.Del(listVal.(string))
									if err != nil {
										d.logger.Errorf("inserting in batch: ", err.Error())
										return err
									}
								}
							case NumberIndex:
								val := v1.(float64)
								// valStr = strconv.FormatFloat(val, 'f', 6, 64)
								_, err := batchIndex.DelNumber(val)
								if err != nil {
									d.logger.Errorf("inserting in batch: ", err.Error())
									return err
								}
							case BytesIndex:
								d.logger.Errorf("inserting in batch: ", ErrIndexNotSupported)
								return ErrIndexNotSupported
							default:
								d.logger.Errorf("inserting in batch: ", ErrInvalidIndexType)
								return ErrInvalidIndexType
							}
						}

						err = d.client.DeleteReference(refs[0])
						if err != nil {
							d.logger.Errorf("inserting in batch: ", err.Error())
							return err
						}

					}
				}
			}

			// upload the document
			ref, err = d.client.UploadBlob(doc, 0, true)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("inserting in batch: ", err.Error())
				return err
			}
		} else { // skipcq: TCV-001
			// store the seek index of the document instead of its reference
			b := make([]byte, binary.MaxVarintLen64)
			n := binary.PutUvarint(b, uint64(index))
			ref = b[:n]
		}

		// update the indexes
		memory := !docBatch.db.mutable
		for field, batchIndex := range docBatch.batches {
			if v, found := docMap[field]; found { // it is already checked to be present
				switch batchIndex.idx.indexType {
				case StringIndex:
					var valStr1 string
					switch v := v.(type) {
					case float64: // skipcq: TCV-001
						if field == DefaultIndexFieldName {
							valStr1 = fmt.Sprintf("%d", int64(v))
						} else {
							valStr1 = fmt.Sprintf("%020.20g", v)
						}
					case string:
						valStr1 = v
					default: // skipcq: TCV-001
						return ErrInvalidIndexType
					}

					apnd := true
					if field == DefaultIndexFieldName {
						apnd = false
					}
					err := batchIndex.Put(valStr1, ref, apnd, memory)
					if err != nil { // skipcq: TCV-001
						d.logger.Errorf("inserting in batch: ", err.Error())
						return err
					}
				case MapIndex:
					valMap := v.(map[string]interface{})
					for keyField, valueField := range valMap {
						vf := valueField.(string)
						mapField := keyField + vf
						err := batchIndex.Put(mapField, ref, true, memory)
						if err != nil { // skipcq: TCV-001
							d.logger.Errorf("inserting in batch: ", err.Error())
							return err
						}
					}
				case ListIndex:
					valList := v.([]interface{})
					for _, listVal := range valList {
						listField := listVal.(string)
						err := batchIndex.Put(listField, ref, true, memory)
						if err != nil { // skipcq: TCV-001
							d.logger.Errorf("inserting in batch: ", err.Error())
							return err
						}
					}
				case NumberIndex:
					switch v1 := v.(type) {
					case string: // skipcq: TCV-001
						err := batchIndex.Put(v1, ref, true, memory)
						if err != nil {
							d.logger.Errorf("inserting in batch: ", err.Error())
							return err
						}
					case float64:
						err := batchIndex.PutNumber(v1, ref, true, memory)
						if err != nil { // skipcq: TCV-001
							d.logger.Errorf("inserting in batch: ", err.Error())
							return err
						}
					default: // skipcq: TCV-001
						return ErrIndexNotSupported
					}

				case BytesIndex:
					return ErrIndexNotSupported // skipcq: TCV-001
				default:
					return ErrInvalidIndexType // skipcq: TCV-001
				}
			}
		}
	default:
		// it's something else
		d.logger.Errorf("inserting in batch: unknown json format")
		return ErrUnknownJsonFormat
	}

	return nil
}

// DocBatchWrite commits the batch index into the Swarm network.
func (d *Document) DocBatchWrite(docBatch *DocBatch, podFile string) error {
	d.logger.Info("writing batch: ", docBatch.db.name)
	if d.fd.IsReadOnlyFeed() { // skipcq: TCV-001
		d.logger.Errorf("writing batch: ", ErrReadOnlyIndex)
		return ErrReadOnlyIndex
	}
	for _, batch := range docBatch.batches {
		man, err := batch.Write(podFile)
		if err != nil { // skipcq: TCV-001
			d.logger.Errorf("writing batch: ", err.Error())
			return err
		}
		batch.memDb = man
		batch.idx.memDB = man
		batch.idx.podFile = man.PodFile
	}
	d.logger.Info("written batch: ", docBatch.db.name)
	return nil
}

// DocFileIndex indexes an existing json file in the pod with the document DB.
// skipcq: TCV-001
func (d *Document) DocFileIndex(dbName, podFile, podPassword string) error {
	d.logger.Info("Indexing file to db: ", podFile, dbName)
	reader, err := d.file.OpenFileForIndex(podFile, podPassword)
	if err != nil {
		d.logger.Errorf("Indexing file: ", err.Error())
		return err
	}
	_, err = reader.Seek(0, 0)
	if err != nil {
		d.logger.Errorf("Indexing file: ", err.Error())
		return err
	}

	batch, err := d.CreateDocBatch(dbName, podPassword)
	if err != nil {
		d.logger.Errorf("Indexing file: ", err.Error())
		return err
	}

	seekIndex := int64(0)
	lineCount := 0
	for {
		data, err := reader.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			d.logger.Errorf("Indexing file: ", err.Error())
			return err
		}

		err = d.DocBatchPut(batch, data, seekIndex)
		if err != nil {
			d.logger.Errorf("Indexing file: ", err.Error())
			return err
		}
		seekIndex += int64(len(data))
		lineCount += 1

		if lineCount%10000 == 0 {
			d.logger.Info("indexed lines: ", lineCount)
		}
	}

	err = d.DocBatchWrite(batch, podFile)
	if err != nil {
		d.logger.Errorf("Indexing file: ", err.Error())
		return err
	}
	d.logger.Info("indexed file to db successfully: ", dbName, podFile, lineCount)
	return nil
}

// skipcq: TCV-001
func (d *Document) getLineFromFile(podFile, podPassword string, seekOffset uint64) ([]byte, error) {
	reader, err := d.file.OpenFileForIndex(podFile, podPassword)
	if err != nil {
		d.logger.Errorf("getting  line: ", err.Error())
		return nil, err
	}
	_, err = reader.Seek(int64(seekOffset), 0)
	if err != nil {
		d.logger.Errorf("getting  line: ", err.Error())
		return nil, err
	}
	data, err := reader.ReadLine()
	if err != nil {
		d.logger.Errorf("getting  line: ", err.Error())
		return nil, err
	}
	return data, nil
}

type entryTask struct {
	c       blockstore.Client
	ref     []byte
	allData *[][]byte
	mtx     sync.Locker
}

func newEntryTask(c blockstore.Client, allData *[][]byte, ref []byte, mtx sync.Locker) *entryTask {
	return &entryTask{
		c:       c,
		ref:     ref,
		allData: allData,
		mtx:     mtx,
	}
}

// Execute
func (et *entryTask) Execute(context.Context) error {
	data, _, err := et.c.DownloadBlob(et.ref)
	if err != nil {
		return err
	}
	et.mtx.Lock()
	defer et.mtx.Unlock()
	*et.allData = append(*et.allData, data)
	return nil
}

// Name
func (et *entryTask) Name() string {
	return string(et.ref)
}
