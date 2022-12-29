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
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

type Entry struct {
	Name             string `json:"name"`
	Tag              uint32 `json:"tag,omitempty"`
	ContentType      string `json:"contentType"`
	Size             string `json:"size,omitempty"`
	BlockSize        string `json:"blockSize,omitempty"`
	CreationTime     string `json:"creationTime"`
	ModificationTime string `json:"modificationTime"`
	AccessTime       string `json:"accessTime"`
	Mode             uint32 `json:"mode"`
}

// ListFiles given a list of files, list files gives back the information related to each file.
func (f *File) ListFiles(files []string, podPassword string) ([]Entry, error) {
	fileEntries := &[]Entry{}
	wg := new(sync.WaitGroup)
	mtx := &sync.Mutex{}
	for _, filePath := range files {
		fileTopic := utils.HashString(utils.CombinePathAndFile(filePath, ""))
		lsTask := newLsTask(f, fileTopic, filePath, podPassword, fileEntries, mtx, wg)
		_, err := f.syncManager.Go(lsTask)
		if err != nil { // skipcq: TCV-001
			f.logger.Warningf("list files : %v", err)
			continue
		}
		wg.Add(1)
	}
	wg.Wait()
	return *fileEntries, nil
}
