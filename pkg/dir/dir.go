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

package dir

import (
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"strings"
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

type Directory struct {
	podName     string
	client      blockstore.Client
	fd          *feed.API
	userAddress utils.Address
	file        *f.File
	dirMap      map[string]*Inode // path to dirInode cache
	dirMu       *sync.RWMutex
	logger      logging.Logger
}

func NewDirectory(podName string, client blockstore.Client, fd *feed.API, user utils.Address, file *f.File, logger logging.Logger) *Directory {
	return &Directory{
		podName:     podName,
		client:      client,
		fd:          fd,
		userAddress: user,
		file:        file,
		dirMap:      make(map[string]*Inode),
		dirMu:       &sync.RWMutex{},
		logger:      logger,
	}
}

func (d *Directory) getFeed() *feed.API {
	return d.fd
}

func (d *Directory) getAddress() utils.Address {
	return d.userAddress
}

func (d *Directory) getClient() blockstore.Client {
	return d.client
}

func (d *Directory) AddToDirectoryMap(path string, dirInode *Inode) {
	d.dirMu.Lock()
	defer d.dirMu.Unlock()
	d.dirMap[path] = dirInode
}

func (d *Directory) RemoveFromDirectoryMap(path string) {
	d.dirMu.Lock()
	defer d.dirMu.Unlock()
	delete(d.dirMap, path)
}

func (d *Directory) GetDirFromDirectoryMap(path string) *Inode {
	d.dirMu.Lock()
	defer d.dirMu.Unlock()
	for k := range d.dirMap {
		if k == path {
			return d.dirMap[path]
		}
	}
	return nil
}

func (d *Directory) GetPrefixPodFromPathMap(prefix string) *Inode {
	d.dirMu.Lock()
	defer d.dirMu.Unlock()
	for k := range d.dirMap {
		if strings.HasPrefix(k, prefix) {
			delete(d.dirMap, k)
		}
	}
	return nil
}
