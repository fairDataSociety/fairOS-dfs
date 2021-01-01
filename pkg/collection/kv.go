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
	"strings"
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	kvFile           = "key_value_tables"
	defaultIndexName = "key"
	CSVHeaderKey     = "__csv_header__"
)

type KeyValue struct {
	fd           *feed.API
	ai           *account.Info
	user         utils.Address
	client       blockstore.Client
	openKVTables map[string]*Index
	openKVTMu    sync.RWMutex
	iterator     *Iterator
	logger       logging.Logger
	columns      []string
}

func NewKeyValueStore(fd *feed.API, ai *account.Info, user utils.Address, client blockstore.Client, logger logging.Logger) *KeyValue {
	return &KeyValue{
		fd:           fd,
		ai:           ai,
		user:         user,
		client:       client,
		openKVTables: make(map[string]*Index),
		logger:       logger,
	}
}

func (kv *KeyValue) CreateKVTable(name string) error {
	// for now , it will be a single index collection
	err := CreateIndex(name, defaultIndexName, StringIndex, kv.fd, kv.user, kv.client)
	if err != nil {
		return err
	}
	kvtables, err := kv.LoadKVTables()
	if err != nil {
		return err
	}
	if _, ok := kvtables[name]; ok {
		return fmt.Errorf("kv table already present")
	}
	kvtables[name] = []string{defaultIndexName}
	return kv.storeKVTables(kvtables)
}

func (kv *KeyValue) DeleteKVTable(name string) error {
	kvtables, err := kv.LoadKVTables()
	if err != nil {
		return err
	}
	if _, ok := kvtables[name]; ok {
		if idx, ok := kv.openKVTables[name]; ok {
			err = idx.DeleteIndex()
			if err != nil {
				return err
			}
			delete(kv.openKVTables, name)
		} else {
			idx, err := OpenIndex(name, defaultIndexName, kv.fd, kv.ai, kv.user, kv.client, kv.logger)
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
	return fmt.Errorf("kv table not present")
}

func (kv *KeyValue) OpenKVTable(name string) error {
	kvtables, err := kv.LoadKVTables()
	if err != nil {
		return err
	}
	if _, ok := kvtables[name]; !ok {
		return fmt.Errorf("kv table not present")
	}
	idx, err := OpenIndex(name, defaultIndexName, kv.fd, kv.ai, kv.user, kv.client, kv.logger)
	if err != nil {
		return err
	}

	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	kv.openKVTables[name] = idx

	hdr, err := idx.Get(CSVHeaderKey)
	if err == nil {
		kv.columns = strings.Split(string(hdr), ",")
	}

	return nil
}

func (kv *KeyValue) KVCount(name string) (uint64, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if idx, ok := kv.openKVTables[name]; ok {
		return idx.CountIndex()
	} else {
		idx, err := OpenIndex(name, defaultIndexName, kv.fd, kv.ai, kv.user, kv.client, kv.logger)
		if err != nil {
			return 0, err
		}
		return idx.CountIndex()
	}
}

func (kv *KeyValue) KVPut(name, key string, value []byte) error {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if idx, ok := kv.openKVTables[name]; ok {
		return idx.Put(key, value, StringIndex)
	}
	return fmt.Errorf("kv table not opened")
}

func (kv *KeyValue) KVGet(name, key string) ([]string, []byte, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if idx, ok := kv.openKVTables[name]; ok {
		value, err := idx.Get(key)
		if err != nil {
			return nil, nil, err
		}
		return kv.columns, value, nil
	}
	return nil, nil, fmt.Errorf("kv table not opened")
}

func (kv *KeyValue) KVDelete(name, key string) ([]byte, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if idx, ok := kv.openKVTables[name]; ok {
		return idx.Delete(key)
	}
	return nil, fmt.Errorf("kv table not opened")
}

func (kv *KeyValue) KVBatch(name string, columns []string) (*Batch, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if idx, ok := kv.openKVTables[name]; ok {
		kv.columns = columns
		return idx.Batch()
	}
	return nil, fmt.Errorf("kv table not opened")
}

func (kv *KeyValue) KVBatchPut(batch *Batch, key string, value []byte) error {
	if key == CSVHeaderKey {
		kv.columns = strings.Split(string(value), ",")
		return nil
	}
	return batch.Put(key, value, StringIndex)
}

func (kv *KeyValue) KVBatchWrite(batch *Batch) error {
	return batch.Write()
}

func (kv *KeyValue) KVSeek(name, start, end string, limit int64) (*Iterator, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if idx, ok := kv.openKVTables[name]; ok {
		itr, err := idx.NewStringIterator(start, end, limit)
		if err != nil {
			return nil, err
		}
		kv.iterator = itr
		return itr, nil
	}
	return nil, fmt.Errorf("kv table not opened")
}

func (kv *KeyValue) KVGetNext(name string) ([]string, string, []byte, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if _, ok := kv.openKVTables[name]; ok {
		if kv.iterator != nil {
			ok := kv.iterator.Next()
			if !ok {
				return nil, "", nil, ErrNoNextElement
			}
			return kv.columns, kv.iterator.StringKey(), kv.iterator.Value(), nil
		}
	}
	return nil, "", nil, fmt.Errorf("kv table not opened")
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
