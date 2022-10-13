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
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// RmDir removes a given directory and all the entries (file/directory) under that.
func (d *Directory) RmDir(directoryNameWithPath string) error {
	if directoryNameWithPath == "" {
		return ErrInvalidDirectoryName
	}
	directoryNameWithPath = filepath.ToSlash(directoryNameWithPath)
	parentPath := filepath.ToSlash(filepath.Dir(directoryNameWithPath))
	dirToDelete := filepath.Base(directoryNameWithPath)
	// validation checks of the arguments
	if parentPath == "." { // skipcq: TCV-001
		return ErrInvalidDirectoryName
	}
	if dirToDelete == "." { // skipcq: TCV-001
		return ErrInvalidDirectoryName
	}

	// check if directory present
	var totalPath string
	if parentPath == utils.PathSeparator && filepath.ToSlash(dirToDelete) == utils.PathSeparator {
		totalPath = utils.CombinePathAndFile(parentPath, "")
	} else {
		totalPath = utils.CombinePathAndFile(parentPath, dirToDelete)
	}
	if d.GetDirFromDirectoryMap(totalPath) == nil {
		return ErrDirectoryNotPresent
	}

	// recursive delete
	dirInode := d.GetDirFromDirectoryMap(totalPath)
	if dirInode.FileOrDirNames != nil && len(dirInode.FileOrDirNames) > 0 {
		for _, fileOrDirName := range dirInode.FileOrDirNames {
			if strings.HasPrefix(fileOrDirName, "_F_") {
				fileName := strings.TrimPrefix(fileOrDirName, "_F_")
				filePath := utils.CombinePathAndFile(directoryNameWithPath, fileName)
				err := d.file.RmFile(filePath)
				if err != nil { // skipcq: TCV-001
					return err
				}
				err = d.RemoveEntryFromDir(directoryNameWithPath, fileName, true)
				if err != nil { // skipcq: TCV-001
					return err
				}
			} else if strings.HasPrefix(fileOrDirName, "_D_") {
				dirName := strings.TrimPrefix(fileOrDirName, "_D_")
				path := utils.CombinePathAndFile(directoryNameWithPath, dirName)
				d.logger.Infof(directoryNameWithPath)

				err := d.RmDir(path)
				if err != nil { // skipcq: TCV-001
					return err
				}
			}
		}
	}

	// remove the feed and clear the data structure
	topic := utils.HashString(totalPath)
	_, err := d.fd.UpdateFeed(topic, d.userAddress, []byte(utils.DeletedFeedMagicWord))
	if err != nil { // skipcq: TCV-001
		return err
	}
	d.RemoveFromDirectoryMap(totalPath)
	// return if root directory
	if parentPath == utils.PathSeparator && filepath.ToSlash(dirToDelete) == utils.PathSeparator {
		return nil
	}
	// remove the directory entry from the parent dir
	return d.RemoveEntryFromDir(parentPath, dirToDelete, false)
}

// RmRootDir removes root directory and all the entries (file/directory) under that.
func (d *Directory) RmRootDir() error {
	dirToDelete := utils.PathSeparator

	// check if directory present
	var totalPath = utils.CombinePathAndFile(dirToDelete, "")

	if d.GetDirFromDirectoryMap(totalPath) == nil { // skipcq: TCV-001
		return ErrDirectoryNotPresent
	}

	// recursive delete
	dirInode := d.GetDirFromDirectoryMap(totalPath)
	if dirInode.FileOrDirNames != nil && len(dirInode.FileOrDirNames) > 0 {
		for _, fileOrDirName := range dirInode.FileOrDirNames {
			if strings.HasPrefix(fileOrDirName, "_F_") {
				fileName := strings.TrimPrefix(fileOrDirName, "_F_")
				filePath := utils.CombinePathAndFile(dirToDelete, fileName)
				err := d.file.RmFile(filePath)
				if err != nil { // skipcq: TCV-001
					return err
				}
				err = d.RemoveEntryFromDir(dirToDelete, fileName, true)
				if err != nil { // skipcq: TCV-001
					return err
				}
			} else if strings.HasPrefix(fileOrDirName, "_D_") {
				dirName := strings.TrimPrefix(fileOrDirName, "_D_")
				path := utils.CombinePathAndFile(dirToDelete, dirName)
				d.logger.Infof(dirToDelete)

				err := d.RmDir(path)
				if err != nil { // skipcq: TCV-001
					return err
				}
			}
		}
	}

	// remove the feed and clear the data structure
	topic := utils.HashString(totalPath)
	_, err := d.fd.UpdateFeed(topic, d.userAddress, []byte(utils.DeletedFeedMagicWord))
	if err != nil { // skipcq: TCV-001
		return err
	}
	d.RemoveFromDirectoryMap(totalPath)

	return nil
}
