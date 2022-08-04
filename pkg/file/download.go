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
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

var (
	// ErrFileNotPresent denotes file is not present
	ErrFileNotPresent = errors.New("file not present")

	// ErrFileAlreadyPresent denotes file is present
	ErrFileAlreadyPresent = errors.New("file already present in the destination dir")

	// ErrFileNotFound denotes file is not found in dfs
	ErrFileNotFound = errors.New("file not found in dfs")
)

// Download does all the validation for the existence of the file and creates a
// Reader to read the contents of the file from the pod.
func (f *File) Download(podFileWithPath string) (io.ReadCloser, uint64, error) {
	// check if file present
	totalFilePath := utils.CombinePathAndFile(podFileWithPath, "")
	if !f.IsFileAlreadyPresent(totalFilePath) {
		return nil, 0, ErrFileNotPresent
	}

	meta := f.GetFromFileMap(totalFilePath)
	if meta == nil { // skipcq: TCV-001
		return nil, 0, ErrFileNotFound
	}

	fileInodeBytes, _, err := f.getClient().DownloadBlob(meta.InodeAddress)
	if err != nil { // skipcq: TCV-001
		return nil, 0, err
	}
	var fileInode INode
	err = json.Unmarshal(fileInodeBytes, &fileInode)
	if err != nil { // skipcq: TCV-001
		return nil, 0, err
	}

	//need to change the access time for podFile if it is owned by user
	if !f.fd.IsReadOnlyFeed() {
		meta.AccessTime = time.Now().Unix()
		err = f.updateMeta(meta)
		if err != nil { // skipcq: TCV-001
			return nil, 0, err
		}
	}

	reader := NewReader(fileInode, f.getClient(), meta.Size, meta.BlockSize, meta.Compression, false)
	return reader, meta.Size, nil
}
