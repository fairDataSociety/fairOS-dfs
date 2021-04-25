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
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"strconv"
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	MineTypeDirectory = "inode/directory"
)

type DirOrFileEntry struct {
	Name             string `json:"name"`
	ContentType      string `json:"content_type"`
	Size             string `json:"size,omitempty"`
	BlockSize        string `json:"block_size,omitempty"`
	CreationTime     string `json:"creation_time"`
	ModificationTime string `json:"modification_time"`
	AccessTime       string `json:"access_time"`
}

func (d *Directory) ListDir(podName, dirNameWithPath string) ([]DirOrFileEntry, error) {
	totalPath := podName + dirNameWithPath
	topic := utils.HashString(totalPath)
	_, data, err := d.fd.GetFeedData(topic, d.getAddress())
	if err != nil {
		return nil, fmt.Errorf("list dir : %v", err)
	}

	var dirInode Inode
	err = json.Unmarshal(data, &dirInode)
	if err != nil {
		return nil, fmt.Errorf("list dir : %v", err)
	}

	var listEntries []DirOrFileEntry
	for _, fileOrDirName := range dirInode.fileOrDirNames {
		if strings.HasPrefix(fileOrDirName, "_D_") {
			dirName := strings.TrimLeft(fileOrDirName, "_D_")
			dirPath := totalPath + utils.PathSeperator + dirName
			dirTopic := utils.HashString(dirPath)
			_, data, err := d.fd.GetFeedData(dirTopic, d.getAddress())
			if err != nil {
				return nil, fmt.Errorf("list dir : %v", err)
			}

			var dirInode *Inode
			err = json.Unmarshal(data, &dirInode)
			if err != nil {
				continue
			}
			entry := DirOrFileEntry{
				Name:             dirInode.Meta.Name,
				ContentType:      MineTypeDirectory, // per RFC2425
				CreationTime:     strconv.FormatInt(dirInode.Meta.CreationTime, 10),
				AccessTime:       strconv.FormatInt(dirInode.Meta.AccessTime, 10),
				ModificationTime: strconv.FormatInt(dirInode.Meta.ModificationTime, 10),
			}
			listEntries = append(listEntries, entry)
		} else if strings.HasPrefix(fileOrDirName, "_F_") {
			fileName := strings.TrimLeft(fileOrDirName, "_F_")
			filePath := totalPath + utils.PathSeperator + fileName
			fileTopic := utils.HashString(filePath)
			_, data, err := d.getFeed().GetFeedData(fileTopic, d.getAddress())
			var meta *file.MetaData
			err = json.Unmarshal(data, &meta)
			if err != nil {
				continue
			}
			entry := DirOrFileEntry{
				Name:             meta.Name,
				ContentType:      meta.ContentType,
				Size:             strconv.FormatUint(meta.Size, 10),
				BlockSize:        strconv.FormatInt(int64(uint64(meta.BlockSize)), 10),
				CreationTime:     strconv.FormatInt(meta.CreationTime, 10),
				AccessTime:       strconv.FormatInt(meta.AccessTime, 10),
				ModificationTime: strconv.FormatInt(meta.ModificationTime, 10),
			}

			listEntries = append(listEntries, entry)
		}
	}
	return listEntries, nil
}