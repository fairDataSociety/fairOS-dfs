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
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	MineTypeDirectory = "inode/directory"
)

type Entry struct {
	Name             string `json:"name"`
	ContentType      string `json:"contentType"`
	Size             string `json:"size,omitempty"`
	BlockSize        string `json:"blockSize,omitempty"`
	CreationTime     string `json:"creationTime"`
	ModificationTime string `json:"modificationTime"`
	AccessTime       string `json:"accessTime"`
}

// ListDir given a directory, this function lists all the children (directory) inside the given directory.
// it also creates a list of files inside the directory and gives it back, so that the file listing
// function can give information about those files.
func (d *Directory) ListDir(dirNameWithPath, podPassword string) ([]Entry, []string, error) {
	dirNameWithPath = filepath.ToSlash(dirNameWithPath)
	topic := utils.HashString(dirNameWithPath)
	data, err := d.fd.GetFeedData(topic, d.getAddress(), []byte(podPassword))
	if err != nil { // skipcq: TCV-001
		if dirNameWithPath == utils.PathSeparator {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("list dir : %v", err) // skipcq: TCV-001
	}

	dirInode := &Inode{}
	err = dirInode.Unmarshal(data)
	if err != nil {
		return nil, nil, fmt.Errorf("list dir : %v", err)
	}

	wg := new(sync.WaitGroup)
	mtx := &sync.Mutex{}
	listEntries := &[]Entry{}
	var files []string
	for _, fileOrDirName := range dirInode.FileOrDirNames {
		if strings.HasPrefix(fileOrDirName, "_D_") {
			dirName := strings.TrimPrefix(fileOrDirName, "_D_")
			dirPath := utils.CombinePathAndFile(dirNameWithPath, dirName)
			dirTopic := utils.HashString(dirPath)
			wg.Add(1)
			lsTask := newLsTask(d, dirTopic, dirPath, podPassword, listEntries, mtx, wg)
			_, err := d.syncManager.Go(lsTask)
			if err != nil {
				return nil, nil, fmt.Errorf("list dir : %v", err)
			}
		} else if strings.HasPrefix(fileOrDirName, "_F_") {
			fileName := strings.TrimPrefix(fileOrDirName, "_F_")
			filePath := utils.CombinePathAndFile(dirNameWithPath, fileName)
			files = append(files, filePath)
		}
	}
	wg.Wait()
	return *listEntries, files, nil
}
