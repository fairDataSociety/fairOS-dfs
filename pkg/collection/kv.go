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
	CSVHeaderKey          = "__csv_header__"
)

type KeyValue struct {
	fd           *feed.API
	ai           *account.Info
	user         utils.Address
	client       blockstore.Client
	openKVTables map[string]*KVTable
	openKVTMu    sync.RWMutex
	iterator     *Iterator
	logger       logging.Logger
}

type KVTable struct {
	index     *Index
	indexType IndexType
	columns   []string
}

func NewKeyValueStore(fd *feed.API, ai *account.Info, user utils.Address, client blockstore.Client, logger logging.Logger) *KeyValue {
	return &KeyValue{
		fd:           fd,
		ai:           ai,
		user:         user,
		client:       client,
		openKVTables: make(map[string]*KVTable),
		logger:       logger,
	}
}

func (kv *KeyValue) CreateKVTable(name string, indexType IndexType) error {
	if kv.fd.IsReadOnlyFeed() {
		return ErrReadOnlyIndex
	}

	// load the existing db's and see if this name is already there
	kvtables, err := kv.LoadKVTables()
	if err != nil {
		return err
	}
	if _, ok := kvtables[name]; ok {
		return ErrKvTableAlreadyPresent
	}

	//  since this tables is not present already, create the index required for this table
	err = CreateIndex(defaultCollectionName, name, indexType, kv.fd, kv.user, kv.client, true)
	if err != nil {
		return err
	}

	// record the table as created
	kvtables[name] = []string{indexType.String()}
	return kv.storeKVTables(kvtables)
}

func (kv *KeyValue) DeleteKVTable(name string) error {
	if kv.fd.IsReadOnlyFeed() {
		return ErrReadOnlyIndex
	}

	kvtables, err := kv.LoadKVTables()
	if err != nil {
		return err
	}

	if _, ok := kvtables[name]; !ok {
		return ErrKVTableNotPresent
	}

	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if table, ok := kv.openKVTables[name]; ok {
		err = table.index.DeleteIndex()
		if err != nil {
			return err
		}
		delete(kv.openKVTables, name)
	} else {
		idx, err := OpenIndex(defaultCollectionName, name, kv.fd, kv.ai, kv.user, kv.client, kv.logger)
		if err != nil {
			return err
		}
		err = idx.DeleteIndex()
		if err != nil {
			return err
		}
	}
	delete(kvtables, name)
	return kv.storeKVTables(kvtables)

}

func (kv *KeyValue) OpenKVTable(name string) error {
	kvtables, err := kv.LoadKVTables()
	if err != nil {
		return err
	}
	values, ok := kvtables[name]
	if !ok {
		return ErrKVTableNotPresent
	}
	idxType := toIndexTypeEnum(values[0])

	idx, err := OpenIndex(defaultCollectionName, name, kv.fd, kv.ai, kv.user, kv.client, kv.logger)
	if err != nil {
		return err
	}

	hdr, err := idx.Get(CSVHeaderKey)
	var columns []string
	if err == nil && len(hdr) >= 1 {
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

func (kv *KeyValue) KVCount(name string) (uint64, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if table, ok := kv.openKVTables[name]; ok {
		return table.index.CountIndex()
	} else {
		idx, err := OpenIndex(defaultCollectionName, name, kv.fd, kv.ai, kv.user, kv.client, kv.logger)
		if err != nil {
			return 0, err
		}
		return idx.CountIndex()
	}
}

func (kv *KeyValue) KVPut(name, key string, value []byte) error {
	if kv.fd.IsReadOnlyFeed() {
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
			ref, err := kv.client.UploadBlob(value, true, true)
			if err != nil {
				return err
			}
			return table.index.Put(key, ref, StringIndex, false)
		default:
			return ErrKVInvalidIndexType
		}
	}
	return ErrKVTableNotOpened
}

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
			if err != nil {
				return nil, nil, err
			}
			value[0] = data
		}
		return table.columns, value[0], nil
	}
	return nil, nil, ErrKVTableNotOpened
}

func (kv *KeyValue) KVDelete(name, key string) ([]byte, error) {
	if kv.fd.IsReadOnlyFeed() {
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
	return nil, ErrKVTableNotOpened
}

func (kv *KeyValue) KVBatch(name string, columns []string) (*Batch, error) {
	if kv.fd.IsReadOnlyFeed() {
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

func (kv *KeyValue) KVBatchPut(batch *Batch, key string, value []byte) error {
	if kv.fd.IsReadOnlyFeed() {
		return ErrReadOnlyIndex
	}

	if key == CSVHeaderKey {
		kv.openKVTMu.Lock()
		defer kv.openKVTMu.Unlock()
		if table, ok := kv.openKVTables[batch.idx.name]; ok {
			table.columns = strings.Split(string(value), ",")
		}
	}
	return batch.Put(key, value, false, false)
}

func (kv *KeyValue) KVBatchWrite(batch *Batch) error {
	if kv.fd.IsReadOnlyFeed() {
		return ErrReadOnlyIndex
	}
	_, err := batch.Write("")
	return err
}

func (kv *KeyValue) KVSeek(name, start, end string, limit int64) (*Iterator, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if table, ok := kv.openKVTables[name]; ok {
		switch table.indexType {
		case StringIndex:
			itr, err := table.index.NewStringIterator(start, end, limit)
			if err != nil {
				return nil, err
			}
			kv.iterator = itr
			return itr, nil
		case NumberIndex:
			startInt, err := strconv.ParseInt(start, 10, 64)
			if err != nil {
				return nil, err
			}
			endInt, err := strconv.ParseInt(end, 10, 64)
			if err != nil {
				return nil, err
			}
			itr, err := table.index.NewIntIterator(startInt, endInt, limit)
			if err != nil {
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

func (kv *KeyValue) KVGetNext(name string) ([]string, string, []byte, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if table, ok := kv.openKVTables[name]; ok {
		if kv.iterator != nil {
			ok := kv.iterator.Next()
			if !ok {
				return nil, "", nil, ErrNoNextElement
			}
			return table.columns, kv.iterator.StringKey(), kv.iterator.Value(), nil
		}
	}
	return nil, "", nil, ErrKVTableNotOpened
}

func (kv *KeyValue) LoadKVTables() (map[string][]string, error) {
	collections := make(map[string][]string)
	topic := utils.HashString(kvFile)
	_, data, err := kv.fd.GetFeedData(topic, kv.user)
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
		lines := strings.Split(line, ",")
		collections[lines[0]] = lines[1:]
	}
	return collections, nil
}

func (kv *KeyValue) storeKVTables(collections map[string][]string) error {
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
	_, err := kv.fd.UpdateFeed(topic, kv.user, buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}
