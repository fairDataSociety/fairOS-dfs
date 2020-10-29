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

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (p *Pod) ChangeDir(podName, dirName string) (*Info, error) {
	directoryName, err := CleanDirName(dirName)
	if err != nil {
		return nil, err
	}

	if len(directoryName) > utils.MaxDirectoryNameLength {
		return nil, fmt.Errorf("directory Name length is > %v", utils.MaxDirectoryNameLength)
	}

	if !p.isPodOpened(podName) {
		return nil, fmt.Errorf("login to pod to do this operation")
	}

	podInfo, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, err
	}

	if directoryName[0] == "" || directoryName[0] == utils.PathSeperator {
		podInfo.SetCurrentDirInode(podInfo.GetCurrentPodInode())
		return podInfo, nil
	}

	directory := podInfo.GetDirectory()
	fd := podInfo.GetFeed()
	accountInfo := podInfo.GetAccountInfo()

	if directoryName[0] == ".." {
		if podInfo.IsCurrentDirRoot() {
			return podInfo, nil
		}
		_, dirInode, err := directory.GetDirNode(podInfo.GetCurrentDirPathOnly(), fd, accountInfo)
		if err != nil {
			return nil, err
		}
		podInfo.SetCurrentDirInode(dirInode)
		return podInfo, nil
	}

	path := p.getDirectoryPath(directoryName[0], podInfo)
	dirInode := directory.GetDirFromDirectoryMap(path)
	if dirInode != nil {
		podInfo.SetCurrentDirInode(dirInode)
	}
	return podInfo, nil
}

func (p *Pod) getDirectoryPath(directoryName string, podInfo *Info) string {
	path := podInfo.GetCurrentDirPathAndName() + utils.PathSeperator + directoryName

	if podInfo.IsCurrentDirRoot() {
		if strings.HasPrefix(directoryName, utils.PathSeperator) {
			path = podInfo.GetCurrentPodPathAndName() + directoryName
		} else {
			path = podInfo.GetCurrentPodPathAndName() + utils.PathSeperator + directoryName
		}
	}
	return path
}
