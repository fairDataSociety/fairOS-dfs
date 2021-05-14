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

package file

import (
	"encoding/json"
	"strconv"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
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

// ListFiles given a list of files, list files gives back the information related to each file.
func (f *File) ListFiles(files []string) ([]Entry, error) {
	var fileEntries []Entry
	for _, filePath := range files {
		fileTopic := utils.HashString(filePath)
		_, data, err := f.fd.GetFeedData(fileTopic, f.userAddress)
		if err != nil {
			continue
		}
		var meta *MetaData
		err = json.Unmarshal(data, &meta)
		if err != nil {
			continue
		}
		entry := Entry{
			Name:             meta.Name,
			ContentType:      meta.ContentType,
			Size:             strconv.FormatUint(meta.Size, 10),
			BlockSize:        strconv.FormatInt(int64(uint64(meta.BlockSize)), 10),
			CreationTime:     strconv.FormatInt(meta.CreationTime, 10),
			AccessTime:       strconv.FormatInt(meta.AccessTime, 10),
			ModificationTime: strconv.FormatInt(meta.ModificationTime, 10),
		}
		fileEntries = append(fileEntries, entry)
	}
	return fileEntries, nil
}
