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
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (d *Directory) SyncDirectory(podName, dirNameWithPath string) error {
	totalPath := podName + dirNameWithPath
	topic := utils.HashString(totalPath)
	_, data, err := d.fd.GetFeedData(topic, d.userAddress)
	if err != nil {
		return fmt.Errorf("dir sync: %v", err)
	}

	var dirInode Inode
	err = json.Unmarshal(data, &dirInode)
	if err != nil {
		return fmt.Errorf("dir sync: %v", err)
	}

	for _, fileOrDirName := range dirInode.FileOrDirNames {
		if strings.HasPrefix(fileOrDirName, "_F_") {
			fileName := strings.TrimLeft(fileOrDirName, "_F_")
			filePath := totalPath + utils.PathSeperator + fileName
			err := d.file.LoadFileMeta(filePath)
			if err != nil {
				return err
			}

		} else if strings.HasPrefix(fileOrDirName, "_D_") {
			var dirInode *Inode
			err = json.Unmarshal(data, &dirInode)
			if err != nil {
				return err
			}

			path := dirInode.Meta.Path + utils.PathSeperator + dirInode.Meta.Name
			d.AddToDirectoryMap(path, dirInode)
			d.logger.Infof(path)

			err = d.SyncDirectory(podName, path,)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
