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
	"strings"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	nameLength = 100
)

func (d *Directory) MkDir(parentPath, dirName string) error {
	// validation checks of the arguments
	if dirName == "" || strings.HasPrefix(dirName, utils.PathSeperator) {
		return ErrInvalidDirectoryName
	}

	if len(dirName) > nameLength {
		return ErrTooLongDirectoryName
	}

	// check if directory already present
	totalPath := utils.CombinePathAndFile(parentPath, dirName)
	topic := utils.HashString(totalPath)
	addr, data, err := d.fd.GetFeedData(topic, d.userAddress)
	if err == nil && addr != nil && data != nil {
		return ErrDirectoryAlreadyPresent
	}

	// create the meta data
	now := time.Now().Unix()
	meta := MetaData{
		Version:          MetaVersion,
		Path:             parentPath,
		Name:             dirName,
		CreationTime:     now,
		ModificationTime: now,
		AccessTime:       now,
	}
	dirInode := &Inode{
		Meta: &meta,
	}
	data, err = json.Marshal(dirInode)
	if err != nil {
		return err
	}

	// upload the metadata as blob
	_, err = d.fd.CreateFeed(topic, d.userAddress, data)
	if err != nil {
		return err
	}
	d.AddToDirectoryMap(totalPath, dirInode)

	// if parent is root, then create the root entry
	parentHash := utils.HashString(parentPath)
	var parentData []byte
	var parentDirInode *Inode
	if parentPath == utils.PathSeperator {
		//create the root directory entry
		_, parentData, err = d.fd.GetFeedData(parentHash, d.userAddress)
		if err != nil {
			// create the root("/") dir
			meta := MetaData{
				Version:          MetaVersion,
				Path:             "",
				Name:             parentPath,
				CreationTime:     now,
				ModificationTime: now,
				AccessTime:       now,
			}
			parentDirInode = &Inode{
				Meta: &meta,
			}
		}
	} else {
		// update the root directory entry
		_, parentData, err = d.fd.GetFeedData(parentHash, d.userAddress)
		if err != nil {
			return err
		}

		err = json.Unmarshal(parentData, &parentDirInode)
		if err != nil {
			return err
		}
	}

	// add the directory entry to the parent
	dirName = "_D_" + dirName
	parentDirInode.FileOrDirNames = append(parentDirInode.FileOrDirNames, dirName)
	parentData, err = json.Marshal(parentDirInode)
	if err != nil {
		return err
	}
	_, err = d.fd.CreateFeed(parentHash, d.userAddress, parentData)
	if err != nil {
		return err
	}
	d.AddToDirectoryMap(parentPath, parentDirInode)
	return nil
}
