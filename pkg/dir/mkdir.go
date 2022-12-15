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
	"path/filepath"
	"strings"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	nameLength = 100
)

// MkDir
func (d *Directory) MkDir(dirToCreateWithPath, podPassword string) error {
	parentPath := filepath.ToSlash(filepath.Dir(dirToCreateWithPath))
	dirName := filepath.Base(dirToCreateWithPath)

	// validation checks of the arguments
	if dirName == "" || strings.HasPrefix(filepath.ToSlash(dirName), utils.PathSeparator) {
		return ErrInvalidDirectoryName
	}

	if len(dirName) > nameLength {
		return ErrTooLongDirectoryName
	}

	// check if directory already present
	totalPath := utils.CombinePathAndFile(parentPath, dirName)
	topic := utils.HashString(totalPath)

	// check if parent path exists
	if d.GetDirFromDirectoryMap(parentPath) == nil {
		return ErrDirectoryNotPresent
	}

	if d.GetDirFromDirectoryMap(totalPath) != nil {
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
	data, err := json.Marshal(dirInode)
	if err != nil { // skipcq: TCV-001
		return err
	}

	// upload the metadata as blob
	err = d.fd.UpdateFeed(topic, d.userAddress, data, []byte(podPassword))
	if err != nil { // skipcq: TCV-001
		fmt.Println("feed upload failed ", err)
		return err
	}

	d.AddToDirectoryMap(totalPath, dirInode)

	// get the parent directory entry and add this new directory to its list of children
	parentHash := utils.HashString(utils.CombinePathAndFile(parentPath, ""))
	dirName = "_D_" + dirName
	parentData, err := d.fd.GetFeedData(parentHash, d.userAddress, []byte(podPassword))
	if err != nil {
		return err
	}

	// unmarshall the data and add the directory entry to the parent
	var parentDirInode *Inode
	err = json.Unmarshal(parentData, &parentDirInode)
	if err != nil { // skipcq: TCV-001
		return err
	}
	parentDirInode.FileOrDirNames = append(parentDirInode.FileOrDirNames, dirName)

	// marshall it back and update the parent feed
	parentData, err = json.Marshal(parentDirInode)
	if err != nil { // skipcq: TCV-001
		return err
	}

	err = d.fd.UpdateFeed(parentHash, d.userAddress, parentData, []byte(podPassword))
	if err != nil { // skipcq: TCV-001
		fmt.Println("parent feed upload failed ", err)
		return err
	}
	d.AddToDirectoryMap(parentPath, parentDirInode)
	return nil
}

// MkRootDir
func (d *Directory) MkRootDir(podName, podPassword string, podAddress utils.Address, fd *feed.API) error {
	// create the root parent dir
	now := time.Now().Unix()
	meta := MetaData{
		Version:          MetaVersion,
		Path:             "",
		Name:             utils.PathSeparator,
		CreationTime:     now,
		ModificationTime: now,
		AccessTime:       now,
	}
	parentDirInode := &Inode{
		Meta: &meta,
	}

	parentData, err := json.Marshal(&parentDirInode)
	if err != nil { // skipcq: TCV-001
		return err
	}
	parentPath := utils.CombinePathAndFile(utils.PathSeparator, "")
	parentHash := utils.HashString(parentPath)

	err = fd.UpdateFeed(parentHash, podAddress, parentData, []byte(podPassword))
	if err != nil { // skipcq: TCV-001
		return err
	}
	d.AddToDirectoryMap(utils.PathSeparator, parentDirInode)
	return nil
}

// AddRootDir
func (d *Directory) AddRootDir(podName, podPassword string, podAddress utils.Address, fd *feed.API) error {
	parentPath := utils.CombinePathAndFile(utils.PathSeparator, "")
	parentHash := utils.HashString(parentPath)
	parentDataBytes, err := fd.GetFeedData(parentHash, podAddress, []byte(podPassword))
	if err != nil {
		return err
	}
	var parentDirInode Inode
	err = parentDirInode.Unmarshal(parentDataBytes)
	if err != nil {
		return err
	}
	d.AddToDirectoryMap(utils.PathSeparator, &parentDirInode)
	return nil
}
