/*
Copyright © 2021 FairOS Authors

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
	"encoding/json"
	"errors"
	"io"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// Inode is the structure of the inode
type Inode struct {
	Meta           *MetaData `json:"meta"`
	FileOrDirNames []string  `json:"fileOrDirNames"`
}

var (
	// ErrResourceDeleted is returned when the resource is deleted
	ErrResourceDeleted = errors.New("resource was deleted")
)

// GetMeta returns iNode metadata
// skipcq: TCV-001
func (in *Inode) GetMeta() *MetaData {
	return in.Meta
}

// GetFileOrDirNames returns file and folder names in iNode
// skipcq: TCV-001
func (in *Inode) GetFileOrDirNames() []string {
	return in.FileOrDirNames
}

// SetFileOrDirNames sets file and folder names in iNode
// skipcq: TCV-001
func (in *Inode) SetFileOrDirNames(fileOrDirNames []string) {
	in.FileOrDirNames = fileOrDirNames
}

// Unmarshal unmarshals the data into iNode
func (in *Inode) Unmarshal(data []byte) error {
	if string(data) == utils.DeletedFeedMagicWord {
		return ErrResourceDeleted
	}
	return json.Unmarshal(data, in)
}

// GetInode returns the inode of the given directory
func (d *Directory) GetInode(podPassword, dirNameWithPath string) (*Inode, error) {
	node := d.GetDirFromDirectoryMap(dirNameWithPath)
	if node != nil {
		return node, nil
	}
	data := []byte{}
	r, _, err := d.file.Download(utils.CombinePathAndFile(dirNameWithPath, indexFileName), podPassword)
	if err != nil { // skipcq: TCV-001
		topic := utils.HashString(dirNameWithPath)
		_, data, err = d.fd.GetFeedData(topic, d.getAddress(), []byte(podPassword), false)
		if err != nil { // skipcq: TCV-001
			return nil, ErrDirectoryNotPresent
		}
		// TODO remove this and upload to indexfile
	} else {
		data, err = io.ReadAll(r)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
	}
	var inode Inode
	err = inode.Unmarshal(data)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	d.AddToDirectoryMap(dirNameWithPath, &inode)
	return &inode, nil
}
