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
	"fmt"
	"strings"

	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (p *Pod) RemoveDir(podName, dirName string) error {
	if !p.isPodOpened(podName) {
		return ErrPodNotOpened
	}
	dirName = strings.TrimPrefix(dirName, utils.PathSeperator)

	info, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}

	directory := info.GetDirectory()

	dirInode, err := p.GetInodeFromName(dirName, info.GetCurrentDirInode(), directory, info)
	if err != nil {
		return err
	}

	if dirInode == nil || dirInode.Meta == nil {
		return fmt.Errorf("name is not a directory")
	}

	topic := info.GetCurrentDirPathAndName() + utils.PathSeperator + dirName
	if info.IsCurrentDirRoot() {
		topic = info.GetCurrentPodPathAndName() + utils.PathSeperator + dirName
	}
	topicBytes := utils.HashString(topic)
	err = p.UpdateTillThePod(podName, directory, topicBytes, dirInode.GetDirInodePathOnly(), false)
	if err != nil {
		return err
	}
	directory.GetPrefixPodFromPathMap(topic)

	// delete the directory inode
	err = info.dir.DeletePodInode(topic)
	if err != nil {
		return err
	}
	return nil
}

func (p *Pod) GetInodeFromName(nameToGetMeta string, curDirInode *d.DirInode, directory *d.Directory, info *Info) (*d.DirInode, error) {
	path := info.GetCurrentDirPathAndName() + utils.PathSeperator + nameToGetMeta
	if info.IsCurrentDirRoot() {
		if strings.HasPrefix(nameToGetMeta, utils.PathSeperator) {
			path = curDirInode.Meta.Path + curDirInode.Meta.Name + nameToGetMeta
		} else {
			path = curDirInode.Meta.Path + curDirInode.Meta.Name + utils.PathSeperator + nameToGetMeta
		}
	}
	return directory.GetDirFromDirectoryMap(path), nil
}
