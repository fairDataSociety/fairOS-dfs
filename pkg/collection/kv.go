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
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	kvFile           = "key_value_tables"
	defaultIndexName = "key"
)

type KeyValue struct {
	fd           *feed.API
	ai           *account.Info
	user         utils.Address
	client       blockstore.Client
	openKVTables map[string]*Index
	openKVTMu    sync.RWMutex
}

func NewKeyValueStore(fd *feed.API, ai *account.Info, user utils.Address, client blockstore.Client) *KeyValue {
	return &KeyValue{
		fd:           fd,
		ai:           ai,
		user:         user,
		client:       client,
		openKVTables: make(map[string]*Index),
	}
}

func (kv *KeyValue) CreateKVTable(name string) error {
	// for now , it will be a single index collection
	err := CreateIndex(name, defaultIndexName, kv.fd, kv.user, kv.client)
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
	kvtables[name] = []string{name}
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
			kvtables[name] = []string{name}
			return kv.storeKVTables(kvtables)
		}
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
	idx, err := OpenIndex(name, defaultIndexName, kv.fd, kv.ai, kv.user, kv.client)
	if err != nil {
		return err
	}

	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	kv.openKVTables[name] = idx
	return nil
}

func (kv *KeyValue) KVPut(name, key string, value []byte) error {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if idx, ok := kv.openKVTables[name]; ok {
		return idx.Put(key, value)
	}
	return fmt.Errorf("kv table not opened")
}

func (kv *KeyValue) KVGet(name, key string) ([]byte, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if idx, ok := kv.openKVTables[name]; ok {
		return idx.Get(key)
	}
	return nil, fmt.Errorf("kv table not opened")
}

func (kv *KeyValue) KVDelete(name, key string) ([]byte, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if idx, ok := kv.openKVTables[name]; ok {
		return idx.Delete(key)
	}
	return nil, fmt.Errorf("kv table not opened")
}

func (kv *KeyValue) KVBatch(name string) (*Batch, error) {
	kv.openKVTMu.Lock()
	defer kv.openKVTMu.Unlock()
	if idx, ok := kv.openKVTables[name]; ok {
		return idx.Batch()
	}
	return nil, fmt.Errorf("kv table not opened")
}

func (kv *KeyValue) KVBatchPut(batch *Batch, key string, value []byte) error {
	return batch.Put(key, value)
}

func (kv *KeyValue) KVBatchWrite(batch *Batch) error {
	return batch.Write()
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
			line := fmt.Sprintf("%s,", k)
			for _, val := range v {
				val := strings.Trim(val, "\n")
				line = fmt.Sprintf("%s,%s", line, val)
			}
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
