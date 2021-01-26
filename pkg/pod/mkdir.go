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

package pod

import (
	"bytes"
	"fmt"
	gopath "path"
	"time"

	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (p *Pod) MakeDir(podName, dirName string) error {
	dirs, err := CleanDirName(dirName)
	if err != nil {
		return err
	}

	if !p.isPodOpened(podName) {
		return ErrPodNotOpened
	}

	podInfo, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}

	if podInfo.accountInfo.IsReadOnlyPod() {
		return ErrReadOnlyPod
	}

	directory := podInfo.GetDirectory()

	var firstTopic []byte
	var topic []byte
	var dirInode *d.DirInode
	var previousDirINode *d.DirInode
	addToPod := false

	// ex: mkdir make/all/this/dir
	if len(dirs) > 1 {
		for i, dirName := range dirs {
			path := p.buildPath(podInfo, dirs, i)
			_, dirInode, err = directory.GetDirNode(path, podInfo.GetFeed(), podInfo.GetAccountInfo())
			if err != nil {
				if previousDirINode == nil {
					if podInfo.IsCurrentDirRoot() {
						addToPod = true
					}
					dirInode, topic, err = directory.CreateDirINode(podName, dirName, podInfo.GetCurrentDirInode())
				} else {
					dirInode, topic, err = directory.CreateDirINode(podName, dirName, previousDirINode)
				}
				if err != nil {
					return err
				}
				if i == 0 {
					firstTopic = topic
				}

				if previousDirINode != nil {
					found := false
					for _, hash := range previousDirINode.Hashes {
						if bytes.Equal(hash, topic) {
							found = true
						}
					}
					if !found {
						previousDirINode.Hashes = append(previousDirINode.Hashes, topic)
						dirInode.Meta.Path = previousDirINode.Meta.Path + utils.PathSeperator + previousDirINode.Meta.Name
						previousDirINode.Meta.ModificationTime = time.Now().Unix()
						_, err = directory.UpdateDirectory(previousDirINode)
						if err != nil {
							return err
						}
					}
				}
			}
			previousDirINode = dirInode
		}
		topic = firstTopic
	} else {
		// see if the dir is present in dirMap
		inode, err := p.GetInodeFromName(dirName, podInfo.GetCurrentDirInode(), directory, podInfo)
		if err != nil {
			return err
		}
		if inode != nil {
			return fmt.Errorf("directory already present")
		}

		dirInode = podInfo.GetCurrentDirInode()
		_, topic, err = directory.CreateDirINode(podName, dirs[0], dirInode)
		if err != nil {
			return err
		}
		addToPod = true
	}

	if addToPod {
		path := podInfo.GetCurrentDirPathAndName()
		if podInfo.IsCurrentDirRoot() {
			path = podInfo.GetCurrentPodPathAndName()
		}
		err = p.UpdateTillThePod(podName, directory, topic, path, true)
		if err != nil {
			return err
		}
	}
	return nil
}

// Assumption is that the d.currentDirInode is the newly updated one
func (p *Pod) UpdateTillThePod(podName string, directory *d.Directory, topic []byte, path string, isAddHash bool) error {
	podInfo, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}

	var dirInode *d.DirInode
	for path != utils.PathSeperator {
		_, dirInode, err = directory.GetDirNode(path, podInfo.GetFeed(), podInfo.GetAccountInfo())
		if err != nil {
			return err
		}
		if isAddHash {
			// Add or update a hash
			found := false
			for i, hash := range dirInode.Hashes {
				if bytes.Equal(hash, topic) {
					found = true
					dirInode.Hashes[i] = topic
				}
			}
			// ignore if it is the current dir, otherwise there will be a loop
			pathTopic := utils.HashString(path)
			if bytes.Equal(pathTopic, topic) {
				path = gopath.Dir(path)
				continue
			}
			if !found {
				dirInode.Hashes = append(dirInode.Hashes, topic)
			}
		} else {
			// remove hash
			var newHashes [][]byte
			for _, hash := range dirInode.Hashes {
				if !bytes.Equal(hash, topic) {
					newHashes = append(newHashes, hash)
				}
			}
			dirInode.Hashes = newHashes
			isAddHash = true // after the first deletion, the rest of the parent links should be updated
		}
		dirInode.Meta.ModificationTime = time.Now().Unix()
		topic, err = directory.UpdateDirectory(dirInode)
		if err != nil {
			return err
		}
		path = gopath.Dir(path)
	}
	podInfo.SetCurrentPodInode(dirInode)
	p.addPodToPodMap(podName, podInfo)
	return nil
}

func (p *Pod) buildPath(podInfo *Info, dirs []string, index int) string {
	var path string
	i := 0
	if podInfo.IsCurrentDirRoot() {
		path = podInfo.GetCurrentPodPathAndName()
	}
	for ; i <= index; i++ {
		path = path + utils.PathSeperator + dirs[i]
	}
	return path
}
