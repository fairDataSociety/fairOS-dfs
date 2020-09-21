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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
)

type FileStats struct {
	Account          string `json:"account"`
	PodName          string `json:"pod_name"`
	FilePath         string `json:"file_path"`
	FileName         string `json:"file_name"`
	FileSize         string `json:"file_size"`
	BlockSize        string `json:"block_size"`
	Compression      string `json:"compression"`
	ContentType      string `json:"content_type"`
	CreationTime     string `json:"creation_time"`
	ModificationTime string `json:"modification_time"`
	AccessTime       string `json:"access_time"`
	Blocks           []Blocks
}

type Blocks struct {
	Name           string `json:"name"`
	Reference      string `json:"reference"`
	Size           string `json:"size"`
	CompressedSize string `json:"compressed_size"`
}

func (f *File) FileStat(podName, fileName, account string) (*FileStats, error) {
	meta := f.GetFromFileMap(fileName)
	if meta == nil {
		return nil, fmt.Errorf("file not found")
	}

	fileInodeBytes, _, err := f.getClient().DownloadBlob(meta.InodeAddress)
	if err != nil {
		return nil, err
	}
	var fileInode FileINode
	err = json.Unmarshal(fileInodeBytes, &fileInode)
	if err != nil {
		return nil, err
	}

	var fileBlocks []Blocks
	for _, b := range fileInode.FileBlocks {
		fb := Blocks{
			Name:           b.Name,
			Reference:      hex.EncodeToString(b.Address),
			Size:           strconv.Itoa(int(b.Size)),
			CompressedSize: strconv.Itoa(int(b.CompressedSize)),
		}
		fileBlocks = append(fileBlocks, fb)
	}
	return &FileStats{
		Account:          account,
		PodName:          podName,
		FilePath:         meta.Path,
		FileName:         meta.Name,
		FileSize:         strconv.FormatUint(meta.FileSize, 10),
		BlockSize:        strconv.Itoa(int(meta.BlockSize)),
		Compression:      meta.Compression,
		ContentType:      meta.ContentType,
		CreationTime:     strconv.FormatInt(meta.CreationTime, 10),
		ModificationTime: strconv.FormatInt(meta.ModificationTime, 10),
		AccessTime:       strconv.FormatInt(meta.AccessTime, 10),
		Blocks:           fileBlocks,
	}, nil
}
