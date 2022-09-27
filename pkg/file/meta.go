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
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

var (
	MetaVersion uint8 = 1

	ErrDeletedFeed = errors.New("deleted feed")
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

// LoadFileMeta is used in syncing
func (f *File) LoadFileMeta(fileNameWithPath string) error {
	meta, err := f.GetMetaFromFileName(fileNameWithPath, f.userAddress)
	if err != nil { // skipcq: TCV-001
		if err == ErrDeletedFeed {
			return nil
		}
		return err
	}
	f.AddToFileMap(fileNameWithPath, meta)
	f.logger.Infof(fileNameWithPath)
	return nil
}

func (f *File) handleMeta(meta *MetaData) error {
	fmt.Println("=====================handleMeta", meta.Path)
	// check if meta is present.
	totalPath := utils.CombinePathAndFile(meta.Path, meta.Name)
	_, err := f.GetMetaFromFileName(totalPath, meta.UserAddress)
	if err != nil {
		if err != ErrDeletedFeed {
			return f.uploadMeta(meta)
		}
	}
	return f.updateMeta(meta)
}

func (f *File) uploadMeta(meta *MetaData) error {
	// marshall the meta structure
	fileMetaBytes, err := json.Marshal(meta)
	if err != nil { // skipcq: TCV-001
		return err
	}

	// put the file meta as a feed
	totalPath := utils.CombinePathAndFile(meta.Path, meta.Name)
	topic := utils.HashString(totalPath)
	_, err = f.fd.CreateFeed(topic, meta.UserAddress, fileMetaBytes)
	if err != nil { // skipcq: TCV-001
		return err
	}

	return nil
}

func (f *File) deleteMeta(meta *MetaData) error {
	totalPath := utils.CombinePathAndFile(meta.Path, meta.Name)
	topic := utils.HashString(totalPath)
	// update with utils.DeletedFeedMagicWord
	_, err := f.fd.UpdateFeed(topic, meta.UserAddress, []byte(utils.DeletedFeedMagicWord))
	if err != nil { // skipcq: TCV-001
		return err
	}
	err = f.fd.DeleteFeed(topic, meta.UserAddress)
	if err != nil {
		f.logger.Warningf("failed to remove file feed %s", totalPath)
	}
	return nil
}

func (f *File) updateMeta(meta *MetaData) error {
	// marshall the meta structure
	fileMetaBytes, err := json.Marshal(meta)
	if err != nil { // skipcq: TCV-001
		return err
	}

	// put the file meta as a feed
	totalPath := utils.CombinePathAndFile(meta.Path, meta.Name)
	topic := utils.HashString(totalPath)
	_, err = f.fd.UpdateFeed(topic, meta.UserAddress, fileMetaBytes)
	if err != nil { // skipcq: TCV-001
		return err
	}

	return nil
}

func (f *File) BackupFromFileName(fileNameWithPath string) (*MetaData, error) {
	p, err := f.GetMetaFromFileName(fileNameWithPath, f.userAddress)
	if err != nil {
		return nil, err
	}

	err = f.deleteMeta(p)
	if err != nil {
		return nil, err
	}

	// change previous meta.Name
	p.Name = fmt.Sprintf("%d_%s", time.Now().Unix(), p.Name)
	p.ModificationTime = time.Now().Unix()

	// upload PreviousMeta
	err = f.uploadMeta(p)
	if err != nil {
		return nil, err
	}

	// add file to map
	f.AddToFileMap(utils.CombinePathAndFile(p.Path, p.Name), p)
	return p, nil
}

func (f *File) RenameFromFileName(fileNameWithPath, newFileNameWithPath string) (*MetaData, error) {
	fileNameWithPath = filepath.ToSlash(fileNameWithPath)
	newFileNameWithPath = filepath.ToSlash(newFileNameWithPath)
	p, err := f.GetMetaFromFileName(fileNameWithPath, f.userAddress)
	if err != nil {
		return nil, err
	}

	// remove old meta and from file map
	err = f.deleteMeta(p)
	if err != nil {
		return nil, err
	}
	f.RemoveFromFileMap(fileNameWithPath)

	newFileName := filepath.Base(newFileNameWithPath)
	newPrnt := filepath.ToSlash(filepath.Dir(newFileNameWithPath))

	// change previous meta.Name
	p.Name = newFileName
	p.Path = newPrnt
	p.ModificationTime = time.Now().Unix()

	// upload meta
	err = f.uploadMeta(p)
	if err != nil {
		return nil, err
	}

	// add file to map
	f.AddToFileMap(newFileNameWithPath, p)
	return p, nil
}

func (f *File) GetMetaFromFileName(fileNameWithPath string, userAddress utils.Address) (*MetaData, error) {
	fmt.Println("GetMetaFromFileName======", fileNameWithPath)
	topic := utils.HashString(fileNameWithPath)
	_, metaBytes, err := f.fd.GetFeedData(topic, userAddress)
	if err != nil {
		return nil, err
	}

	if string(metaBytes) == utils.DeletedFeedMagicWord {
		f.logger.Errorf("found deleted feed for %s\n", fileNameWithPath)
		return nil, ErrDeletedFeed
	}

	var meta *MetaData
	err = json.Unmarshal(metaBytes, &meta)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	return meta, nil
}

func (f *File) PutMetaForFile(meta *MetaData) error {
	return f.updateMeta(meta)
}
