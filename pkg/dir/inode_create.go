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
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (d *Directory) CreateDirINode(podName, dirName string, parent *Inode) (*Inode, []byte, error) {
	// create the meta data
	parentPath := getPath(podName, parent)
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
	data, err := json.Marshal(dirInode)
	if err != nil {
		return nil, nil, err
	}

	// create a feed for the directory and add data to it
	totalPath := parentPath + utils.PathSeperator + dirName
	topic := utils.HashString(totalPath)
	_, err = d.fd.CreateFeed(topic, d.getAddress(), data)
	if err != nil {
		return nil, nil, err
	}

	d.AddToDirectoryMap(totalPath, dirInode)
	return dirInode, topic, nil
}

func (d *Directory) IsDirINodePresent(podName, dirName string, parent *Inode) bool {
	parentPath := getPath(podName, parent)
	totalPath := parentPath + utils.PathSeperator + dirName
	topic := utils.HashString(totalPath)
	_, _, err := d.fd.GetFeedData(topic, d.getAddress())
	return err == nil
}

func getPath(podName string, parent *Inode) string {
	var path string
	if parent.Meta.Path == utils.PathSeperator {
		path = parent.Meta.Path + parent.Meta.Name
	} else {
		path = parent.Meta.Path + utils.PathSeperator + parent.Meta.Name
	}
	return path
}

func (d *Directory) CreatePodINode(podName string) (*Inode, []byte, error) {
	// create the metadata
	now := time.Now().Unix()
	meta := MetaData{
		Version:          MetaVersion,
		Path:             "/",
		Name:             podName,
		CreationTime:     now,
		ModificationTime: now,
		AccessTime:       now,
	}
	dirInode := &Inode{
		Meta: &meta,
	}
	data, err := json.Marshal(dirInode)
	if err != nil {
		return nil, nil, err
	}

	// create a feed and store the metadata of the pod
	totalPath := utils.PathSeperator + podName
	topic := utils.HashString(totalPath)
	_, err = d.fd.CreateFeed(topic, d.getAddress(), data)
	if err != nil {
		return nil, nil, err
	}

	d.AddToDirectoryMap(totalPath, dirInode)
	return dirInode, topic, nil
}

func (d *Directory) DeletePodInode(podName string) error {
	totalPath := utils.PathSeperator + podName
	topic := utils.HashString(totalPath)
	return d.fd.DeleteFeed(topic, d.getAddress())
}

func (d *Directory) DeleteDirectoryInode(dirPath string) error {
	topic := utils.HashString(dirPath)
	return d.fd.DeleteFeed(topic, d.getAddress())
}
