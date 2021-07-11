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
	"path/filepath"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// RmDir removes a given directory and all the entries (file/directory) under that.
func (d *Directory) RmDir(directoryNameWithPath string) error {
	parentPath := filepath.Dir(directoryNameWithPath)
	dirToDelete := filepath.Base(directoryNameWithPath)
	// validation checks of the arguments
	if parentPath == "" {
		return ErrInvalidDirectoryName
	}
	if dirToDelete == "" {
		return ErrInvalidDirectoryName
	}

	// check if directory present
	totalPath := utils.CombinePathAndFile(d.podName, parentPath, dirToDelete)
	if d.GetDirFromDirectoryMap(totalPath) == nil {
		return ErrDirectoryNotPresent
	}

	// return if the directory is not empty
	// TODO: in future do a recursive delete
	dirInode := d.GetDirFromDirectoryMap(totalPath)
	if dirInode.FileOrDirNames != nil && len(dirInode.FileOrDirNames) > 0 {
		return ErrDirectoryNotEmpty
	}

	// remove the feed and clear the data structure
	topic := utils.HashString(totalPath)
	err := d.fd.DeleteFeed(topic, d.userAddress)
	if err != nil {
		return err
	}
	d.RemoveFromDirectoryMap(totalPath)

	// remove the directory entry from the parent dir
	return d.RemoveEntryFromDir(parentPath, dirToDelete, false)
}
