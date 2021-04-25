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

func (d *Directory) MkDir(podName, parentPath, dirName string) error {
	// validation checks of the arguments
	if podName == "" {
		return ErrInvalidPodName
	}

	if dirName == "" || strings.HasPrefix(dirName, utils.PathSeperator) {
		return ErrInvalidDirectoryName
	}

	if len(dirName) > nameLength {
		return ErrTooLongDirectoryName
	}


	// check if directory already present
	totalPath := podName + parentPath + utils.PathSeperator + dirName
	topic := utils.HashString(totalPath)
	addr, data, err := d.fd.GetFeedData(topic, d.acc.GetAddress())
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
	_, err = d.fd.CreateFeed(topic, d.acc.GetAddress(), data)
	if err != nil {
		return err
	}

	d.AddToDirectoryMap(totalPath, dirInode)
	return nil
}

