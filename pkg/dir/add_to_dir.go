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

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (d *Directory) AddToDir(podName, dirNameWithPath, itemToAdd string, isFile bool) error {
	// validation checks of the arguments
	if podName == "" {
		return ErrInvalidPodName
	}

	if dirNameWithPath == "" {
		return ErrInvalidDirectoryName
	}

	if itemToAdd == "" {
		return ErrInvalidFileOrDirectoryName
	}

	// check if directory present
	totalPath := podName + dirNameWithPath
	if d.GetDirFromDirectoryMap(totalPath) == nil {
		return ErrDirectoryNotPresent
	}

	// get the latest meta from swarm
	topic := utils.HashString(totalPath)
	_, data, err := d.fd.GetFeedData(topic, d.userAddress)
	if err != nil {
		return fmt.Errorf("add entry to dir : %v", err)
	}

	var dirInode Inode
	err = json.Unmarshal(data, &dirInode)
	if err != nil {
		return fmt.Errorf("add entry to dir : %v", err)
	}

	// add file or directory entry
	if isFile {
		itemToAdd = "_F_" + itemToAdd
	} else {
		itemToAdd = "_D_" + itemToAdd
	}
	dirInode.fileOrDirNames = append(dirInode.fileOrDirNames, itemToAdd)

	// update the feed of the dir and the data structure with latest info
	data, err = json.Marshal(dirInode)
	if err != nil {
		return fmt.Errorf("add entry to dir : %v", err)
	}
	_, err = d.fd.CreateFeed(topic, d.userAddress, data)
	if err != nil {
		return fmt.Errorf("add entry to dir : %v", err)
	}
	d.AddToDirectoryMap(totalPath, &dirInode)
	return nil
}
