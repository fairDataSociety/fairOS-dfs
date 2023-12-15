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
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	nameLength = 100
	// S_IFDIR is the mode for directory
	S_IFDIR     = 0040000
	defaultMode = 0700
)

// MkDir creates a directory in the given path
func (d *Directory) MkDir(dirToCreateWithPath, podPassword string, mode uint32) error {
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

	// check if parent path exists
	_, err := d.GetInode(podPassword, parentPath)
	if err != nil {
		return ErrDirectoryNotPresent
	}

	_, err = d.GetInode(podPassword, totalPath)
	if err == nil {
		return ErrDirectoryAlreadyPresent
	}

	if mode == 0 {
		mode = S_IFDIR | defaultMode
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
		Mode:             mode,
	}
	dirInode := &Inode{
		Meta: &meta,
	}

	err = d.SetInode(podPassword, dirInode)
	if err != nil { // skipcq: TCV-001
		return err
	}

	return d.AddEntryToDir(parentPath, podPassword, dirName, false)
}

// MkRootDir creates the root directory for the pod
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
	return d.SetInode(podPassword, parentDirInode)
}

// AddRootDir adds the root directory to the directory map
func (d *Directory) AddRootDir(podName, podPassword string, podAddress utils.Address, fd *feed.API) error {
	parentDirInode, err := d.GetInode(podPassword, utils.CombinePathAndFile(utils.PathSeparator, ""))
	if err != nil {
		return err
	}
	d.AddToDirectoryMap(utils.PathSeparator, parentDirInode)
	return nil
}
