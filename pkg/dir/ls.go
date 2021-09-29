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
	"strconv"
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	MineTypeDirectory = "inode/directory"
)

type Entry struct {
	Name             string `json:"name"`
	ContentType      string `json:"content_type"`
	Size             string `json:"size,omitempty"`
	BlockSize        string `json:"block_size,omitempty"`
	CreationTime     string `json:"creation_time"`
	ModificationTime string `json:"modification_time"`
	AccessTime       string `json:"access_time"`
}

// ListDir given a directory, this function lists all the children (directory) inside the given directory.
// it also creates a list of files inside the directory and gives it back, so that the file listing
// function can give information about those files.
func (d *Directory) ListDir(dirNameWithPath string) ([]Entry, []string, error) {
	topic := utils.HashString(utils.CombinePathAndFile(d.podName, dirNameWithPath, ""))
	_, data, err := d.fd.GetFeedData(topic, d.getAddress())
	if err != nil {
		if dirNameWithPath == utils.PathSeparator {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("list dir : %v", err)
	}

	var dirInode Inode
	err = json.Unmarshal(data, &dirInode)
	if err != nil {
		return nil, nil, fmt.Errorf("list dir : %v", err)
	}

	var listEntries []Entry
	var files []string
	for _, fileOrDirName := range dirInode.FileOrDirNames {
		if strings.HasPrefix(fileOrDirName, "_D_") {
			dirName := strings.TrimPrefix(fileOrDirName, "_D_")
			dirPath := utils.CombinePathAndFile(d.podName, dirNameWithPath, dirName)
			dirTopic := utils.HashString(dirPath)
			_, data, err := d.fd.GetFeedData(dirTopic, d.getAddress())
			if err != nil {
				return nil, nil, fmt.Errorf("list dir : %v", err)
			}

			var dirInode *Inode
			err = json.Unmarshal(data, &dirInode)
			if err != nil {
				continue
			}
			entry := Entry{
				Name:             dirInode.Meta.Name,
				ContentType:      MineTypeDirectory, // per RFC2425
				CreationTime:     strconv.FormatInt(dirInode.Meta.CreationTime, 10),
				AccessTime:       strconv.FormatInt(dirInode.Meta.AccessTime, 10),
				ModificationTime: strconv.FormatInt(dirInode.Meta.ModificationTime, 10),
			}
			listEntries = append(listEntries, entry)
		} else if strings.HasPrefix(fileOrDirName, "_F_") {
			fileName := strings.TrimPrefix(fileOrDirName, "_F_")
			filePath := utils.CombinePathAndFile(d.podName, dirNameWithPath, fileName)
			files = append(files, filePath)
		}
	}
	return listEntries, files, nil
}
