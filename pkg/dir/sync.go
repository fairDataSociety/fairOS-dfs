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

func (d *Directory) SyncDirectory(dirNameWithPath string) error {
	topic := utils.HashString(dirNameWithPath)
	_, data, err := d.fd.GetFeedData(topic, d.userAddress)
	if err != nil {
		return fmt.Errorf("dir sync: %v", err)
	}

	var dirInode *Inode
	err = json.Unmarshal(data, &dirInode)
	if err != nil {
		return fmt.Errorf("dir sync: %v", err)
	}
	d.AddToDirectoryMap(dirNameWithPath, dirInode)

	for _, fileOrDirName := range dirInode.FileOrDirNames {
		if strings.HasPrefix(fileOrDirName, "_F_") {
			fileName := strings.TrimLeft(fileOrDirName, "_F_")
			filePath := utils.CombinePathAndFile(dirNameWithPath, fileName)
			err := d.file.LoadFileMeta(filePath)
			if err != nil {
				return err
			}

		} else if strings.HasPrefix(fileOrDirName, "_D_") {
			dirName := strings.TrimLeft(fileOrDirName, "_D_")
			path := utils.CombinePathAndFile(dirNameWithPath, dirName)
			d.logger.Infof(dirNameWithPath)

			err = d.SyncDirectory(path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
