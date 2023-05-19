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

	// ErrFileAlreadyPresent denotes file is present
	ErrFileAlreadyPresent = errors.New("file already present in the destination dir")

	// ErrFileNotFound denotes file is not found in dfs
	ErrFileNotFound = errors.New("file not found in dfs")

	// ErrFileTagPresent denotes file status is not available
	ErrFileTagPresent = errors.New("file status is not available")
)

// Download does all the validation for the existence of the file and creates a
// Reader to read the contents of the file from the pod.
func (f *File) Download(podFileWithPath, podPassword string) (io.ReadCloser, uint64, error) {
	return f.ReadSeeker(podFileWithPath, podPassword)
}

// ReadSeeker does all the validation for the existence of the file and creates a
// ReadSeekCloser to read the contents of the file from the pod.
func (f *File) ReadSeeker(podFileWithPath, podPassword string) (io.ReadSeekCloser, uint64, error) {
	// check if file present
	totalFilePath := utils.CombinePathAndFile(podFileWithPath, "")
	if !f.IsFileAlreadyPresent(podPassword, totalFilePath) {
		return nil, 0, ErrFileNotFound
	}

	meta := f.GetInode(podPassword, totalFilePath)
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

	// need to change the access time for podFile if it is owned by user
	// update accessTime in a go routine, to reduce the latency
	if !f.fd.IsReadOnlyFeed() {
		meta.AccessTime = time.Now().Unix()
		go func() {
			err = f.updateMeta(meta, podPassword)
			if err != nil { // skipcq: TCV-001
				f.logger.Errorf("error updating meta for file %s: %s", totalFilePath, err.Error())
			}
		}()
	}

	reader := NewReader(fileInode, f.getClient(), meta.Size, meta.BlockSize, meta.Compression, false)
	return reader, meta.Size, nil
}
