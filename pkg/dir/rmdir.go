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
	var totalPath string
	if parentPath == "/" && dirToDelete == "/" {
		totalPath = utils.CombinePathAndFile(d.podName, parentPath, "")
	} else {
		totalPath = utils.CombinePathAndFile(d.podName, parentPath, dirToDelete)

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
				filePath := utils.CombinePathAndFile(d.podName, directoryNameWithPath, fileName)
				err := d.file.RmFile(filePath)
				if err != nil {
					return err
				}
				err = d.RemoveEntryFromDir(directoryNameWithPath, fileName, true)
				if err != nil {
					return err
				}
			} else if strings.HasPrefix(fileOrDirName, "_D_") {
				dirName := strings.TrimPrefix(fileOrDirName, "_D_")
				path := utils.CombinePathAndFile(d.podName, directoryNameWithPath, dirName)
				d.logger.Infof(directoryNameWithPath)

				err := d.RmDir(path)
				if err != nil {
					return err
				}
			}
		}
	}

	// remove the feed and clear the data structure
	topic := utils.HashString(totalPath)
	_, err := d.fd.UpdateFeed(topic, d.userAddress, []byte(utils.DeletedFeedMagicWord))
	if err != nil {
		return err
	}
	d.RemoveFromDirectoryMap(totalPath)

	// return if root directory
	if parentPath == "/" && dirToDelete == "/" {
		return nil
	}
	// remove the directory entry from the parent dir
	return d.RemoveEntryFromDir(parentPath, dirToDelete, false)
}

// RmRootDir removes root directory and all the entries (file/directory) under that.
func (d *Directory) RmRootDir() error {
	dirToDelete := filepath.Base("/")

	// check if directory present
	var totalPath = utils.CombinePathAndFile(d.podName, dirToDelete, "")

	if d.GetDirFromDirectoryMap(totalPath) == nil {
		return ErrDirectoryNotPresent
	}

	// recursive delete
	dirInode := d.GetDirFromDirectoryMap(totalPath)
	if dirInode.FileOrDirNames != nil && len(dirInode.FileOrDirNames) > 0 {
		for _, fileOrDirName := range dirInode.FileOrDirNames {
			if strings.HasPrefix(fileOrDirName, "_F_") {
				fileName := strings.TrimPrefix(fileOrDirName, "_F_")
				filePath := utils.CombinePathAndFile(d.podName, dirToDelete, fileName)
				err := d.file.RmFile(filePath)
				if err != nil {
					return err
				}
				err = d.RemoveEntryFromDir(dirToDelete, fileName, true)
				if err != nil {
					return err
				}
			} else if strings.HasPrefix(fileOrDirName, "_D_") {
				dirName := strings.TrimPrefix(fileOrDirName, "_D_")
				path := utils.CombinePathAndFile(d.podName, dirToDelete, dirName)
				d.logger.Infof(dirToDelete)

				err := d.RmDir(path)
				if err != nil {
					return err
				}
			}
		}
	}

	// remove the feed and clear the data structure
	topic := utils.HashString(totalPath)
	_, err := d.fd.UpdateFeed(topic, d.userAddress, []byte(utils.DeletedFeedMagicWord))
	if err != nil {
		return err
	}
	d.RemoveFromDirectoryMap(totalPath)

	return nil
}
