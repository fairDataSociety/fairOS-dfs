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

package file

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// File represents a file in a pod
type File struct {
	podName     string
	userAddress utils.Address
	client      blockstore.Client
	fd          *feed.API
	fileMap     map[string]*MetaData
	tagMap      sync.Map
	fileMu      *sync.RWMutex
	logger      logging.Logger
	syncManager taskmanager.TaskManagerGO
}

// NewFile creates the base file object which has all the methods related to file manipulation.
func NewFile(podName string, client blockstore.Client, fd *feed.API, user utils.Address,
	m taskmanager.TaskManagerGO, logger logging.Logger) *File {
	return &File{
		podName:     podName,
		userAddress: user,
		client:      client,
		fd:          fd,
		fileMap:     make(map[string]*MetaData),
		fileMu:      &sync.RWMutex{},
		logger:      logger,
		syncManager: m,
	}
}

func (f *File) getClient() blockstore.Client {
	return f.client
}

// AddToFileMap adds a file metadata into fileMap
func (f *File) AddToFileMap(filePath string, meta *MetaData) {
	f.fileMu.Lock()
	defer f.fileMu.Unlock()
	f.fileMap[filePath] = meta
}

// RemoveFromFileMap removes a file metadata from fileMap
func (f *File) RemoveFromFileMap(filePath string) {
	f.fileMu.Lock()
	defer f.fileMu.Unlock()
	delete(f.fileMap, filePath)
}

// GetFromFileMap gets file metadata from the fileMap
func (f *File) GetFromFileMap(filePath string) *MetaData {
	f.fileMu.Lock()
	defer f.fileMu.Unlock()
	if meta, ok := f.fileMap[filePath]; ok {
		return meta
	}
	return nil
}

// IsFileAlreadyPresent checks if a file is present in the fileMap
func (f *File) IsFileAlreadyPresent(fileWithPath string) bool {
	f.fileMu.Lock()
	defer f.fileMu.Unlock()
	if _, ok := f.fileMap[fileWithPath]; ok {
		return true
	}
	return false
}

// RemoveAllFromFileMap resets the fileMap
func (f *File) RemoveAllFromFileMap() {
	f.fileMu.Lock()
	defer f.fileMu.Unlock()
	f.fileMap = make(map[string]*MetaData)
}

// AddToTagMap adds a mapping filename and tag into tagMap
func (f *File) AddToTagMap(filePath string, tag uint32) {
	f.tagMap.Store(filePath, tag)
}

// LoadFromTagMap gets a tag from tagMap
func (f *File) LoadFromTagMap(filePath string) uint32 {
	tag, ok := f.tagMap.Load(filePath)
	if ok {
		formattedTag, ok := tag.(uint32)
		if !ok { // skipcq: TCV-001
			return 0
		}
		return formattedTag
	}
	return 0 // skipcq: TCV-001
}

// DeleteFromTagMap deletes a tag from tagMap
func (f *File) DeleteFromTagMap(filePath string) { // skipcq: TCV-001
	f.tagMap.Delete(filePath)
}

type lsTask struct {
	f           *File
	topic       []byte
	path        string
	podPassword string
	entries     *[]Entry
	mtx         sync.Locker
	wg          *sync.WaitGroup
}

func newLsTask(f *File, topic []byte, path, podPassword string, l *[]Entry, mtx sync.Locker, wg *sync.WaitGroup) *lsTask {
	return &lsTask{
		f:           f,
		topic:       topic,
		path:        path,
		entries:     l,
		mtx:         mtx,
		wg:          wg,
		podPassword: podPassword,
	}
}

// Execute
func (lt *lsTask) Execute(context.Context) error {
	defer lt.wg.Done()
	_, data, err := lt.f.fd.GetFeedData(lt.topic, lt.f.userAddress, []byte(lt.podPassword))
	if err != nil { // skipcq: TCV-001
		return fmt.Errorf("file mtdt : %v", err)
	}
	if string(data) == utils.DeletedFeedMagicWord { // skipcq: TCV-001
		return nil
	}
	var meta *MetaData
	err = json.Unmarshal(data, &meta)
	if err != nil { // skipcq: TCV-001
		return fmt.Errorf("file mtdt : %v", err)
	}
	entry := Entry{
		Name:             meta.Name,
		ContentType:      meta.ContentType,
		Size:             strconv.FormatUint(meta.Size, 10),
		BlockSize:        strconv.FormatInt(int64(meta.BlockSize), 10),
		CreationTime:     strconv.FormatInt(meta.CreationTime, 10),
		AccessTime:       strconv.FormatInt(meta.AccessTime, 10),
		ModificationTime: strconv.FormatInt(meta.ModificationTime, 10),
		Mode:             meta.Mode,
	}
	lt.f.AddToFileMap(utils.CombinePathAndFile(meta.Path, meta.Name), meta)
	lt.mtx.Lock()
	defer lt.mtx.Unlock()
	*lt.entries = append(*lt.entries, entry)
	return nil
}

// Name
func (lt *lsTask) Name() string {
	return lt.f.userAddress.String() + lt.f.podName + lt.path
}
