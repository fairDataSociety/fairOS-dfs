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
	"strconv"
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

type PodStat struct {
	Version          string
	PodName          string
	PodPath          string
	CreationTime     string
	AccessTime       string
	ModificationTime string
}

func (p *Pod) PodStat(podName string) (*PodStat, error) {
	podInfo, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, ErrInvalidPodName
	}
	podInode := podInfo.GetCurrentPodInode()
	return &PodStat{
		Version:          strconv.Itoa(int(podInode.Meta.Version)),
		PodName:          podInode.Meta.Name,
		PodPath:          podInode.Meta.Path,
		CreationTime:     strconv.FormatInt(podInode.Meta.CreationTime, 10),
		AccessTime:       strconv.FormatInt(podInode.Meta.AccessTime, 10),
		ModificationTime: strconv.FormatInt(podInode.Meta.AccessTime, 10),
	}, nil
}
func (p *Pod) ExpandFilePath(podName, podFileOrDir string) (string, error) {
	if !p.IsPodOpened(podName) {
		return "", fmt.Errorf("login to pod to do this operation")
	}

	info, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return "", err
	}

	path := p.getDirectoryPath(podFileOrDir, info)
	return path, nil
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
