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
	"fmt"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

var (
	MetaVersion uint8 = 1
)

type MetaData struct {
	Version          uint8         `json:"version"`
	UserAddress      utils.Address `json:"user_address"`
	PodName          string        `json:"pod_name"`
	Path             string        `json:"file_path"`
	Name             string        `json:"file_name"`
	Size             uint64        `json:"file_size"`
	BlockSize        uint32        `json:"block_size"`
	ContentType      string        `json:"content_type"`
	Compression      string        `json:"compression"`
	CreationTime     int64         `json:"creation_time"`
	AccessTime       int64         `json:"access_time"`
	ModificationTime int64         `json:"modification_time"`
	InodeAddress     []byte        `json:"file_inode_reference"`
}

// used in syncing
func (f *File) LoadFileMeta(fileNameWithPath string) error {
	meta, err := f.GetMetaFromFileName(fileNameWithPath)
	if err != nil {
		return err
	}
	f.AddToFileMap(fileNameWithPath, meta)
	f.logger.Infof(fileNameWithPath)
	return nil
}

func (f *File) uploadMeta(meta *MetaData) error {
	// marshall the meta structure
	fileMetaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	// put the file meta as a feed
	totalPath := utils.CombinePathAndFile(f.podName, meta.Path, meta.Name)
	topic := utils.HashString(totalPath)
	fmt.Println("uploadMeta: topic       = ", topic)
	fmt.Println("uploadMeta: totalPath   = ", totalPath)
	fmt.Println("uploadMeta: userAddress = ", meta.UserAddress)
	_, err = f.fd.CreateFeed(topic, meta.UserAddress, fileMetaBytes)
	if err != nil {
		return err
	}

	return nil
}

func (f *File) updateMeta(meta *MetaData) error {
	// marshall the meta structure
	fileMetaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	// put the file meta as a feed
	totalPath := utils.CombinePathAndFile(f.podName, meta.Path, meta.Name)
	topic := utils.HashString(totalPath)
	_, err = f.fd.UpdateFeed(topic, meta.UserAddress, fileMetaBytes)
	if err != nil {
		return err
	}

	return nil
}

func (f *File) GetMetaFromFileName(fileNameWithPath string) (*MetaData, error) {
	topic := utils.HashString(fileNameWithPath)
	_, metaBytes, err := f.fd.GetFeedData(topic, f.userAddress)
	if err != nil {
		return nil, err
	}

	var meta *MetaData
	err = json.Unmarshal(metaBytes, &meta)
	if err != nil {
		return nil, err
	}

	return meta, nil
}
