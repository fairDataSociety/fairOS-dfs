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
	"strconv"
)

// Stats
type Stats struct {
	PodName          string   `json:"podName"`
	Mode             uint32   `json:"mode"`
	FilePath         string   `json:"filePath"`
	FileName         string   `json:"fileName"`
	FileSize         string   `json:"fileSize"`
	BlockSize        string   `json:"blockSize"`
	Compression      string   `json:"compression"`
	ContentType      string   `json:"contentType"`
	CreationTime     string   `json:"creationTime"`
	ModificationTime string   `json:"modificationTime"`
	AccessTime       string   `json:"accessTime"`
	Blocks           []Blocks `json:"blocks"`
}

// Blocks
type Blocks struct {
	Reference      string `json:"reference"`
	Size           string `json:"size"`
	CompressedSize string `json:"compressedSize"`
}

// GetStats given a filename this function returns all the information about the file
// including the block information.
func (f *File) GetStats(podName, podFileWithPath, podPassword string) (*Stats, error) {
	meta := f.GetInode(podPassword, podFileWithPath)
	if meta == nil { // skipcq: TCV-001
		return nil, ErrFileNotFound
	}

	fileInodeBytes, _, err := f.getClient().DownloadBlob(meta.InodeAddress)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	var fileInode INode
	err = json.Unmarshal(fileInodeBytes, &fileInode)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	var fileBlocks []Blocks
	for _, b := range fileInode.Blocks {
		fb := Blocks{
			Reference:      hex.EncodeToString(b.Reference.Bytes()),
			Size:           strconv.Itoa(int(b.Size)),
			CompressedSize: strconv.Itoa(int(b.CompressedSize)),
		}
		fileBlocks = append(fileBlocks, fb)
	}
	f.AddToFileMap(podFileWithPath, meta)
	return &Stats{
		PodName:          podName,
		FilePath:         meta.Path,
		FileName:         meta.Name,
		Mode:             meta.Mode,
		FileSize:         strconv.FormatUint(meta.Size, 10),
		BlockSize:        strconv.Itoa(int(meta.BlockSize)),
		Compression:      meta.Compression,
		ContentType:      meta.ContentType,
		CreationTime:     strconv.FormatInt(meta.CreationTime, 10),
		ModificationTime: strconv.FormatInt(meta.ModificationTime, 10),
		AccessTime:       strconv.FormatInt(meta.AccessTime, 10),
		Blocks:           fileBlocks,
	}, nil
}
