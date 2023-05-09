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
	kvFile                = "key_value_tables"
	defaultCollectionName = "KV"
	// CSVHeaderKey is the key used to store the header of the csv file
	CSVHeaderKey = "__csv_header__"
)

// KeyValue is the main object used to do all operation on the key value tables.
type KeyValue struct {
	podName      string
	fd           *feed.API
	ai           *account.Info
	user         utils.Address
	client       blockstore.Client
	openKVTables map[string]*KVTable
	openKVTMu    sync.RWMutex
	iterator     *Iterator
	logger       logging.Logger
}

// KVTable is the object used to do all operation on a key value table.
type KVTable struct {
	index     *Index
	indexType IndexType
	columns   []string
}

// TableKeyCount is the object used to store the count of keys in a table.
type TableKeyCount struct {
	Count     uint64 `json:"count"`
	TableName string `json:"tableName"`
}

// NewKeyValueStore is the main object used to do all operation on the key value tables.
func NewKeyValueStore(podName string, fd *feed.API, ai *account.Info, user utils.Address, client blockstore.Client, logger logging.Logger) *KeyValue {
	return &KeyValue{
		podName:      podName,
		fd:           fd,
		ai:           ai,
		user:         user,
		client:       client,
		openKVTables: make(map[string]*KVTable),
		logger:       logger,
	}
}

// CreateKVTable creates the key value table  with a given index type.
func (kv *KeyValue) CreateKVTable(name, encryptionPassword string, indexType IndexType) error {
	if kv.fd.IsReadOnlyFeed() { // skipcq: TCV-001
		return ErrReadOnlyIndex
	}

	// load the existing db's and see if this name is already there
	kvtables, err := kv.LoadKVTables(encryptionPassword)
	if err != nil { // skipcq: TCV-001
		return err
	}
	if _, ok := kvtables[name]; ok {
		return ErrKvTableAlreadyPresent
	}

	//  since this tables is not present already, create the index required for this table
	err = CreateIndex(kv.podName, defaultCollectionName, name, encryptionPassword, indexType, kv.fd, kv.user, kv.client, true)
	if err != nil { // skipcq: TCV-001
		return err
	}

	// record the table as created
	kvtables[name] = []string{indexType.String()}
	return kv.storeKVTables(kvtables, encryptionPassword)
}

// DeleteKVTable deletes a given key value table with all it's index and data entries.
func (kv *KeyValue) DeleteKVTable(name, encryptionPassword string) error {
	if kv.fd.IsReadOnlyFeed() { // skipcq: TCV-001
		return ErrReadOnlyIndex
	}

	kvtables, err := kv.LoadKVTables(encryptionPassword)
	if err != nil { // skipcq: TCV-001
		return err
	}

	if _, ok := kvtables[name]; !ok {
		return ErrKVTableNotPresent
	}

	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if table, ok := kv.openKVTables[name]; ok {
		err = table.index.DeleteIndex(encryptionPassword)
		if err != nil { // skipcq: TCV-001
			return err
		}
		delete(kv.openKVTables, name)
	} else {
		idx, err := OpenIndex(kv.podName, defaultCollectionName, name, encryptionPassword, kv.fd, kv.ai, kv.user, kv.client, kv.logger)
		if err != nil { // skipcq: TCV-001
			return err
		}
		err = idx.DeleteIndex(encryptionPassword)
		if err != nil { // skipcq: TCV-001
			return err
		}
	}
	delete(kvtables, name)
	return kv.storeKVTables(kvtables, encryptionPassword)
}

// DeleteAllKVTables deletes all key value tables with all their index and data entries.
func (kv *KeyValue) DeleteAllKVTables(encryptionPassword string) error {
	if kv.fd.IsReadOnlyFeed() { // skipcq: TCV-001
		return ErrReadOnlyIndex
	}

	kvtables, err := kv.LoadKVTables(encryptionPassword)
	if err != nil { // skipcq: TCV-001
		return err
	}
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	for name := range kvtables {
		if _, ok := kvtables[name]; !ok {
			return ErrKVTableNotPresent
		}

		if table, ok := kv.openKVTables[name]; ok {
			err = table.index.DeleteIndex(encryptionPassword)
			if err != nil { // skipcq: TCV-001
				return err
			}
			delete(kv.openKVTables, name)
		} else {
			idx, err := OpenIndex(kv.podName, defaultCollectionName, name, encryptionPassword, kv.fd, kv.ai, kv.user, kv.client, kv.logger)
			if err != nil { // skipcq: TCV-001
				if err == ErrIndexNotPresent {
					continue
				}
				return err
			}
			err = idx.DeleteIndex(encryptionPassword)
			if err != nil { // skipcq: TCV-001
				return err
			}
		}
		delete(kvtables, name)
	}

	return kv.storeKVTables(kvtables, encryptionPassword)
}

// OpenKVTable open a given key value table and loads the index.
func (kv *KeyValue) OpenKVTable(name, encryptionPassword string) error {
	kvtables, err := kv.LoadKVTables(encryptionPassword)
	if err != nil { // skipcq: TCV-001
		return err
	}
	values, ok := kvtables[name]
	if !ok {
		return ErrKVTableNotPresent
	}
	idxType := toIndexTypeEnum(values[0])

	idx, err := OpenIndex(kv.podName, defaultCollectionName, name, encryptionPassword, kv.fd, kv.ai, kv.user, kv.client, kv.logger)
	if err != nil { // skipcq: TCV-001
		return err
	}

	hdr, err := idx.Get(CSVHeaderKey)
	var columns []string
	if err == nil && len(hdr) >= 1 { // skipcq: TCV-001
		columns = strings.Split(string(hdr[0]), ",")
	}

	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	kvTable := &KVTable{
		index:     idx,
		indexType: idxType,
		columns:   columns,
	}
	kv.openKVTables[name] = kvTable

	return nil
}

// KVCount counts the number of entries in the given key value table.
func (kv *KeyValue) KVCount(name string) (*TableKeyCount, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if table, ok := kv.openKVTables[name]; ok {
		return &TableKeyCount{
			Count:     table.index.count,
			TableName: name,
		}, nil
	}
	return nil, ErrKVTableNotOpened
}

// IsEmpty checks if the given key value table is empty.
func (kv *KeyValue) IsEmpty(name string) (bool, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if table, ok := kv.openKVTables[name]; ok {
		return table.index.IsEmpty(table.index.encryptionPassword)
	}
	return false, ErrKVTableNotOpened
}

// IsEmpty checks if the given key value table is empty.
func (kv *KeyValue) IsEmpty(name, encryptionPassword string) (bool, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if table, ok := kv.openKVTables[name]; ok {
		return table.index.IsEmpty(table.index.encryptionPassword)
	} else {
		idx, err := OpenIndex(kv.podName, defaultCollectionName, name, encryptionPassword, kv.fd, kv.ai, kv.user, kv.client, kv.logger)
		if err != nil {
			return true, err
		}
		return idx.IsEmpty(idx.encryptionPassword)
	}
}

// KVPut inserts a given key and value in to the KV table.
func (kv *KeyValue) KVPut(name, key string, value []byte) error {
	if kv.fd.IsReadOnlyFeed() { // skipcq: TCV-001
		return ErrReadOnlyIndex
	}

	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if table, ok := kv.openKVTables[name]; ok {
		switch table.indexType {
		case StringIndex:
			return table.index.Put(key, value, StringIndex, false)
		case NumberIndex:
			fkey, err := strconv.ParseFloat(key, 64)
			if err != nil {
				return ErrKVKeyNotANumber
			}
			return table.index.PutNumber(fkey, value, NumberIndex, false)
		case BytesIndex:
			ref, err := kv.client.UploadBlob(value, 0, true)
			if err != nil { // skipcq: TCV-001
				return err
			}
			return table.index.Put(key, ref, StringIndex, false)
		default: // skipcq: TCV-001
			return ErrKVInvalidIndexType
		}
	}
	return ErrKVTableNotOpened
}

// KVGet retrieves a value from the KV table given a key.
func (kv *KeyValue) KVGet(name, key string) ([]string, []byte, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if table, ok := kv.openKVTables[name]; ok {
		value, err := table.index.Get(key)
		if err != nil {
			return nil, nil, err
		}
		if table.indexType == BytesIndex {
			data, _, err := kv.client.DownloadBlob(value[0])
			if err != nil { // skipcq: TCV-001
				return nil, nil, err
			}
			value[0] = data
		}
		return table.columns, value[0], nil
	}
	return nil, nil, ErrKVTableNotOpened
}

// KVDelete removed a key value entry from the KV table given a key.
func (kv *KeyValue) KVDelete(name, key string) ([]byte, error) {
	if kv.fd.IsReadOnlyFeed() { // skipcq: TCV-001
		return nil, ErrReadOnlyIndex
	}

	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if table, ok := kv.openKVTables[name]; ok {
		refs, err := table.index.Delete(key)
		if err != nil {
			return nil, err
		}
		return refs[0], err
	}
	return nil, ErrKVTableNotOpened // skipcq: TCV-001
}

// KVBatch prepares the index to do a batch insert if keys and values.
func (kv *KeyValue) KVBatch(name string, columns []string) (*Batch, error) {
	if kv.fd.IsReadOnlyFeed() { // skipcq: TCV-001
		return nil, ErrReadOnlyIndex
	}
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if table, ok := kv.openKVTables[name]; ok {
		table.columns = columns
		return NewBatch(table.index)
	}
	return nil, ErrKVTableNotOpened
}

// KVBatchPut inserts a key and value in to the memory for batch.
func (kv *KeyValue) KVBatchPut(batch *Batch, key string, value []byte) error {
	if kv.fd.IsReadOnlyFeed() { // skipcq: TCV-001
		return ErrReadOnlyIndex
	}

	if key == CSVHeaderKey { // skipcq: TCV-001
		kv.openKVTMu.Lock()
		defer kv.openKVTMu.Unlock()
		if table, ok := kv.openKVTables[batch.idx.name]; ok {
			table.columns = strings.Split(string(value), ",")
		}
	}
	return batch.Put(key, value, false, false)
}

// KVBatchWrite commits all the batch entries in to the key value table.
func (kv *KeyValue) KVBatchWrite(batch *Batch) error {
	if kv.fd.IsReadOnlyFeed() { // skipcq: TCV-001
		return ErrReadOnlyIndex
	}
	_, err := batch.Write("")
	return err
}

// KVSeek seek to given key with start prefix and prepare for iterating the table.
func (kv *KeyValue) KVSeek(name, start, end string, limit int64) (*Iterator, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if table, ok := kv.openKVTables[name]; ok {
		switch table.indexType {
		case StringIndex:
			itr, err := table.index.NewStringIterator(start, end, limit)
			if err != nil { // skipcq: TCV-001
				return nil, err
			}
			kv.iterator = itr
			return itr, nil
		case NumberIndex:
			startInt, err := strconv.ParseInt(start, 10, 64)
			if err != nil { // skipcq: TCV-001
				return nil, err
			}
			endInt, err := strconv.ParseInt(end, 10, 64)
			if err != nil { // skipcq: TCV-001
				return nil, err
			}
			itr, err := table.index.NewIntIterator(startInt, endInt, limit)
			if err != nil { // skipcq: TCV-001
				return nil, err
			}
			kv.iterator = itr
			return itr, nil
		case BytesIndex:
			return nil, ErrKVIndexTypeNotSupported
		default:
			return nil, ErrKVInvalidIndexType
		}

	}
	return nil, ErrKVTableNotOpened
}

// KVGetNext retrieve the next key value pair in the iteration.
func (kv *KeyValue) KVGetNext(name string) ([]string, string, []byte, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if table, ok := kv.openKVTables[name]; ok {
		if kv.iterator == nil {
			return nil, "", nil, ErrKVNilIterator
		}
		ok := kv.iterator.Next()
		if !ok {
			return nil, "", nil, ErrNoNextElement
		}
		return table.columns, kv.iterator.StringKey(), kv.iterator.Value(), nil
	}
	return nil, "", nil, ErrKVTableNotOpened
}

// LoadKVTables Loads the list of KV tables.
func (kv *KeyValue) LoadKVTables(encryptionPassword string) (map[string][]string, error) {
	collections := make(map[string][]string)
	topic := utils.HashString(kvFile)
	_, data, err := kv.fd.GetFeedData(topic, kv.user, []byte(encryptionPassword), false)
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
			return nil, fmt.Errorf("loading collections: %w", err)
		}
		line = strings.Trim(line, "\n")
		lines := strings.Split(line, ",")
		collections[lines[0]] = lines[1:]
	}
	return collections, nil
}

func (kv *KeyValue) storeKVTables(collections map[string][]string, encryptionPassword string) error {
	buf := bytes.NewBuffer(nil)
	collectionLen := len(collections)
	if collectionLen > 0 {
		for k, v := range collections {
			indexes := strings.Join(v, ",")
			line := fmt.Sprintf("%s,%s", k, indexes)
			buf.WriteString(line + "\n")
		}
	}
	topic := utils.HashString(kvFile)
	data := buf.Bytes()
	if buf.Len() == 0 {
		data = []byte(utils.DeletedFeedMagicWord)
	}
	_, err := kv.fd.UpdateFeed(kv.user, topic, data, []byte(encryptionPassword), false)
	if err != nil { // skipcq: TCV-001
		return err
	}
	return nil
}
