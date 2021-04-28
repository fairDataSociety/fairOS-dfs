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
	"encoding/json"
	"fmt"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (d *Directory) AddFileToDirectory(dirNameWithPath, fileToAdd string) error {
	_, dirInode, err := d.GetDirNode(dirNameWithPath, d.fd, d.userAddress)
	if err != nil {
		return err
	}

	// check if file already present
	for _, fname := range dirInode.GetFileOrDirNames() {
		if fname == fileToAdd {
			return fmt.Errorf("file already present")
		}
	}

	// add file to inode
	dirInode.fileOrDirNames = append(dirInode.fileOrDirNames, fileToAdd)
	data, err := json.Marshal(dirInode)
	if err != nil {
		return err
	}
	dirInode.Meta.ModificationTime = time.Now().Unix()

	// update the dir inode feed
	topic := utils.HashString(dirNameWithPath)
	_, err = d.getFeed().UpdateFeed(topic, d.getAddress(), data)
	if err != nil {
		return err
	}

	d.AddToDirectoryMap(dirNameWithPath, dirInode)
	return nil
}

func (d *Directory) RemoveFileFromDirectory(dirNameWithPath, fileToRemove string) error {
	_, dirInode, err := d.GetDirNode(dirNameWithPath, d.fd, d.userAddress)
	if err != nil {
		return err
	}

	// check if file already present
	var newFiles []string
	fileNotPresent := false
	for _, fname := range dirInode.GetFileOrDirNames() {
		if fname != fileToRemove {
			newFiles = append(newFiles, fname)
		} else {
			fileNotPresent = true
		}
	}

	if fileNotPresent {
		return fmt.Errorf("file not present")
	}

	// update the dirInode
	dirInode.fileOrDirNames = newFiles

	// update the dir inode feed
	data, err := json.Marshal(dirInode)
	if err != nil {
		return err
	}
	dirInode.Meta.ModificationTime = time.Now().Unix()
	topic := utils.HashString(dirNameWithPath)
	_, err = d.getFeed().UpdateFeed(topic, d.getAddress(), data)
	if err != nil {
		return err
	}

	d.AddToDirectoryMap(dirNameWithPath, dirInode)
	return nil
}
