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
	collectionFile = "collections"
)

type Collection struct {
	fd              *feed.API
	ai              *account.Info
	user            utils.Address
	client          blockstore.Client
	openCollections map[string]*Index
	openMu          sync.RWMutex
}

func NewCollection(fd *feed.API, ai *account.Info, user utils.Address, client blockstore.Client) *Collection {
	return &Collection{
		fd:              fd,
		ai:              ai,
		user:            user,
		client:          client,
		openCollections: make(map[string]*Index),
	}
}

func (c *Collection) CreateCollection(name, index string) error {
	// for now , it will be a single index collection
	err := CreateIndex(name, index, c.fd, c.user)
	if err != nil {
		return err
	}
	collections, err := c.LoadCollections()
	if err != nil {
		return err
	}
	if _, ok := collections[name]; ok {
		return fmt.Errorf("collection already present")
	}
	collections[name] = []string{index}
	return c.storeCollections(collections)
}

func (c *Collection) DeleteCollection(name, index string) error {
	collections, err := c.LoadCollections()
	if err != nil {
		return err
	}
	if _, ok := collections[name]; ok {
		if idx, ok := c.openCollections[name]; ok {
			err = idx.DeleteIndex()
			if err != nil {
				return err
			}
			collections[name] = []string{index}
			return c.storeCollections(collections)
		}
	}
	return fmt.Errorf("collection not present")
}

func (c *Collection) OpenCollection(name, index string) error {
	collections, err := c.LoadCollections()
	if err != nil {
		return err
	}
	if _, ok := collections[name]; !ok {
		return fmt.Errorf("collection not present")
	}
	idx, err := OpenIndex(name, index, c.fd, c.ai, c.user, c.client)
	if err != nil {
		return err
	}

	c.openMu.Lock()
	defer c.openMu.Unlock()
	c.openCollections[name] = idx
	return nil
}

func (c *Collection) Put(name, key string, value []byte) error {
	c.openMu.Lock()
	defer c.openMu.Unlock()
	if idx, ok := c.openCollections[name]; ok {
		return idx.Put(key, value)
	}
	return fmt.Errorf("collection not opened")
}

func (c *Collection) Get(name, key string) ([]byte, error) {
	c.openMu.Lock()
	defer c.openMu.Unlock()
	if idx, ok := c.openCollections[name]; ok {
		return idx.Get(key)
	}
	return nil, fmt.Errorf("collection not opened")
}

func (c *Collection) Delete(name, key string) ([]byte, error) {
	c.openMu.Lock()
	defer c.openMu.Unlock()
	if idx, ok := c.openCollections[name]; ok {
		return idx.Delete(key)
	}
	return nil, fmt.Errorf("collection not opened")
}

func (c *Collection) LoadCollections() (map[string][]string, error) {
	collections := make(map[string][]string)
	topic := utils.HashString(collectionFile)
	_, data, err := c.fd.GetFeedData(topic, c.user)
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

func (c *Collection) storeCollections(collections map[string][]string) error {
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
	topic := utils.HashString(collectionFile)
	_, err := c.fd.UpdateFeed(topic, c.user, buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}
