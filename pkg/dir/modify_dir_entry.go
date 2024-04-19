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
	"time"
)

const (
	indexFileName = "index.dfs"
)

// AddEntryToDir adds a new entry (directory/file) to a given directory.
// This is typically called when a new directory is created under the given directory or
// a new file is uploaded under the given directory.
func (d *Directory) AddEntryToDir(parentDir, podPassword, itemToAdd string, isFile bool) error {
	// validation checks of the arguments
	if parentDir == "" {
		return ErrInvalidDirectoryName
	}

	if itemToAdd == "" {
		return ErrInvalidFileOrDirectoryName
	}

	dirInode, err := d.GetInode(podPassword, parentDir)
	// check if parent directory present
	if err != nil {
		return ErrDirectoryNotPresent
	}
	fmt.Println(" =============== dirInode", dirInode.FileOrDirNames)
	// add file or directory entry
	if isFile {
		itemToAdd = "_F_" + itemToAdd
	} else { // skipcq: TCV-001
		itemToAdd = "_D_" + itemToAdd
	}
	dirInode.FileOrDirNames = append(dirInode.FileOrDirNames, itemToAdd)
	dirInode.Meta.ModificationTime = time.Now().Unix()

	return d.SetInode(podPassword, dirInode)
}

// RemoveEntryFromDir removes an entry (directory/file) under the given directory.
// This is typically called when a  directory is deleted under the given directory or
// a file is removed under the given directory.
func (d *Directory) RemoveEntryFromDir(parentDir, podPassword, itemToDelete string, isFile bool) error {
	// validation checks of the arguments
	if parentDir == "" { // skipcq: TCV-001
		return ErrInvalidDirectoryName
	}

	if itemToDelete == "" { // skipcq: TCV-001
		return ErrInvalidFileOrDirectoryName
	}
	parentDir = filepath.ToSlash(parentDir)
	parentDirInode, err := d.GetInode(podPassword, parentDir)
	// check if parent directory present
	if err != nil {
		d.logger.Errorf("remove entry from dir: parent directory not present %s\n", parentDir)
		return ErrDirectoryNotPresent
	}

	if isFile {
		itemToDelete = "_F_" + itemToDelete
	} else {
		itemToDelete = "_D_" + itemToDelete
	}
	var fileNames []string
	for _, fileOrDirName := range parentDirInode.FileOrDirNames {
		if fileOrDirName != itemToDelete {
			fileNames = append(fileNames, fileOrDirName)
		}
	}

	parentDirInode.FileOrDirNames = fileNames
	parentDirInode.Meta.ModificationTime = time.Now().Unix()

	return d.SetInode(podPassword, parentDirInode)
}
