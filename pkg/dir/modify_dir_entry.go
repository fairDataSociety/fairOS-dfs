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

func (d *Directory) AddEntryToDir(parentDir, itemToAdd string, isFile bool) error {
	// validation checks of the arguments
	if parentDir == "" {
		return ErrInvalidDirectoryName
	}

	if itemToAdd == "" {
		return ErrInvalidFileOrDirectoryName
	}

	// check if parent directory present
	if d.GetDirFromDirectoryMap(parentDir) == nil {
		return ErrDirectoryNotPresent
	}

	// get the latest meta from swarm
	topic := utils.HashString(parentDir)
	_, data, err := d.fd.GetFeedData(topic, d.userAddress)
	if err != nil {
		return fmt.Errorf("modify dir entry: %v", err)
	}

	var dirInode Inode
	err = json.Unmarshal(data, &dirInode)
	if err != nil {
		return fmt.Errorf("modify dir entry : %v", err)
	}

	// add file or directory entry
	if isFile {
		itemToAdd = "_F_" + itemToAdd
	} else {
		itemToAdd = "_D_" + itemToAdd
	}
	dirInode.FileOrDirNames = append(dirInode.FileOrDirNames, itemToAdd)
	dirInode.Meta.ModificationTime = time.Now().Unix()

	// update the feed of the dir and the data structure with latest info
	data, err = json.Marshal(dirInode)
	if err != nil {
		return fmt.Errorf("modify dir entry : %v", err)
	}
	_, err = d.fd.UpdateFeed(topic, d.userAddress, data)
	if err != nil {
		return fmt.Errorf("modify dir entry : %v", err)
	}
	d.AddToDirectoryMap(parentDir, &dirInode)
	return nil
}


func (d *Directory) RemoveEntryFromDir(parentDir, itemToDelete string, isFile bool) error {
	// validation checks of the arguments
	if parentDir == "" {
		return ErrInvalidDirectoryName
	}

	if itemToDelete == "" {
		return ErrInvalidFileOrDirectoryName
	}

	parentHash := utils.HashString(parentDir)
	_, parentData, err := d.fd.GetFeedData(parentHash, d.userAddress)
	if err != nil {
		return err
	}

	var parentDirInode *Inode
	err = json.Unmarshal(parentData, &parentDirInode)
	if err != nil {
		return err
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

	parentData, err = json.Marshal(parentDirInode)
	if err != nil {
		return err
	}
	_, err = d.fd.UpdateFeed(parentHash, d.userAddress, parentData)
	if err != nil {
		return err
	}
	d.AddToDirectoryMap(parentDir, parentDirInode)
	return nil
}
