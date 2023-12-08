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
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"

	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
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
	var inode Inode
	var data []byte
	r, _, err := d.file.Download(utils.CombinePathAndFile(dirNameWithPath, indexFileName), podPassword)
	if err != nil { // skipcq: TCV-001
		topic := utils.HashString(dirNameWithPath)
		_, data, err = d.fd.GetFeedData(topic, d.getAddress(), []byte(podPassword), false)
		if err != nil { // skipcq: TCV-001
			return nil, ErrDirectoryNotPresent
		}
		err = inode.Unmarshal(data)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		err = d.SetInode(podPassword, &inode)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}

		// ignore delete error
		_ = d.fd.DeleteFeedFromTopic(topic, d.getAddress())
	} else {
		data, err = io.ReadAll(r)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		err = inode.Unmarshal(data)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
	}
	d.AddToDirectoryMap(dirNameWithPath, &inode)
	return &inode, nil
}

// SetInode saves the inode of the given directory
func (d *Directory) SetInode(podPassword string, iNode *Inode) error {
	totalPath := utils.CombinePathAndFile(iNode.Meta.Path, iNode.Meta.Name)
	data, err := json.Marshal(iNode)
	if err != nil { // skipcq: TCV-001
		return err
	}

	err = d.file.Upload(bufio.NewReader(bytes.NewBuffer(data)), indexFileName, int64(len(data)), file.MinBlockSize, 0, totalPath, "gzip", podPassword)
	if err != nil {
		return err
	}
	d.AddToDirectoryMap(totalPath, iNode)
	return nil
}

// RemoveInode removes the inode of the given directory
func (d *Directory) RemoveInode(podPassword, dirNameWithPath string) error {
	parentPath := filepath.ToSlash(filepath.Dir(dirNameWithPath))
	dirToDelete := filepath.Base(dirNameWithPath)
	var totalPath string
	if parentPath == utils.PathSeparator && filepath.ToSlash(dirToDelete) == utils.PathSeparator {
		totalPath = utils.CombinePathAndFile(parentPath, "")
	} else {
		totalPath = utils.CombinePathAndFile(parentPath, dirToDelete)
	}
	err := d.file.RmFile(utils.CombinePathAndFile(totalPath, indexFileName), podPassword)
	if err != nil {
		return err
	}
	d.RemoveFromDirectoryMap(totalPath)
	// return if root directory
	if parentPath == "" || (parentPath == utils.PathSeparator && filepath.ToSlash(totalPath) == utils.PathSeparator) {
		return nil
	}
	// remove the directory entry from the parent dir
	return d.RemoveEntryFromDir(filepath.ToSlash(filepath.Dir(parentPath)), podPassword, filepath.Base(totalPath), false)
}
