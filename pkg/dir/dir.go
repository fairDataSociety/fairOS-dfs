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
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// Directory is the type used to define a directory in a pod
type Directory struct {
	podName     string
	client      blockstore.Client
	fd          *feed.API
	userAddress utils.Address
	file        f.IFile
	dirMap      map[string]*Inode // path to dirInode cache
	dirMu       *sync.RWMutex
	logger      logging.Logger
	syncManager taskmanager.TaskManagerGO
}

// NewDirectory the main directory object that handles all the directory related functions.
func NewDirectory(podName string, client blockstore.Client, fd *feed.API, user utils.Address,
	file f.IFile, m taskmanager.TaskManagerGO, logger logging.Logger) *Directory {
	return &Directory{
		podName:     podName,
		client:      client,
		fd:          fd,
		userAddress: user,
		file:        file,
		dirMap:      make(map[string]*Inode),
		dirMu:       &sync.RWMutex{},
		logger:      logger,
		syncManager: m,
	}
}

func (d *Directory) getAddress() utils.Address {
	return d.userAddress
}

// AddToDirectoryMap adds a directory in the path
func (d *Directory) AddToDirectoryMap(path string, dirInode *Inode) {
	d.dirMu.Lock()
	defer d.dirMu.Unlock()
	d.dirMap[path] = dirInode
}

// RemoveFromDirectoryMap removes a directory from the path
func (d *Directory) RemoveFromDirectoryMap(path string) {
	d.dirMu.Lock()
	defer d.dirMu.Unlock()
	delete(d.dirMap, path)
}

// GetDirFromDirectoryMap returns the directory Inode of the given path
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

// RemoveAllFromDirectoryMap resets user dirMap
func (d *Directory) RemoveAllFromDirectoryMap() {
	d.dirMu.Lock()
	defer d.dirMu.Unlock()
	d.dirMap = make(map[string]*Inode)
}

type syncTask struct {
	d    *Directory
	path string
	wg   *sync.WaitGroup
}

func newSyncTask(d *Directory, path string, wg *sync.WaitGroup) *syncTask {
	return &syncTask{
		d:    d,
		path: path,
		wg:   wg,
	}
}

func (st *syncTask) Execute(context.Context) error {
	defer st.wg.Done()
	return st.d.file.LoadFileMeta(st.path)
}

func (st *syncTask) Name() string {
	return st.path
}

type lsTask struct {
	d       *Directory
	topic   []byte
	path    string
	entries *[]Entry
	mtx     sync.Locker
	wg      *sync.WaitGroup
}

func newLsTask(d *Directory, topic []byte, path string, l *[]Entry, mtx sync.Locker, wg *sync.WaitGroup) *lsTask {
	return &lsTask{
		d:       d,
		topic:   topic,
		path:    path,
		entries: l,
		mtx:     mtx,
		wg:      wg,
	}
}

func (lt *lsTask) Execute(context.Context) error {
	defer lt.wg.Done()
	_, data, err := lt.d.fd.GetFeedData(lt.topic, lt.d.getAddress())
	if err != nil {
		return fmt.Errorf("list dir : %v", err)
	}
	var dirInode *Inode
	err = json.Unmarshal(data, &dirInode)
	if err != nil { // skipcq: TCV-001
		return fmt.Errorf("list dir : %v", err)
	}
	entry := Entry{
		Name:             dirInode.Meta.Name,
		ContentType:      MineTypeDirectory, // per RFC2425
		CreationTime:     strconv.FormatInt(dirInode.Meta.CreationTime, 10),
		AccessTime:       strconv.FormatInt(dirInode.Meta.AccessTime, 10),
		ModificationTime: strconv.FormatInt(dirInode.Meta.ModificationTime, 10),
	}
	lt.mtx.Lock()
	defer lt.mtx.Unlock()
	*lt.entries = append(*lt.entries, entry)
	return nil
}

func (lt *lsTask) Name() string {
	return lt.path
}
