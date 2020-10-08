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

	"github.com/ethersphere/bee/pkg/swarm"

	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
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

func (p *Pod) DirectoryStat(podName, podFileOrDir string, printNames bool) (*dir.DirStats, error) {
	if !p.isPodOpened(podName) {
		return nil, ErrPodNotOpened
	}

	info, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, err
	}

	acc := info.getAccountInfo().GetAddress()

	path := p.getDirectoryPath(podFileOrDir, info)
	dirInode := info.GetDirectory().GetDirFromDirectoryMap(path)
	if dirInode != nil {
		meta := dirInode.Meta
		addr, dirInode, err := info.GetDirectory().GetDirNode(meta.Path+utils.PathSeperator+meta.Name, info.getFeed(), info.getAccountInfo())
		if err != nil {
			return nil, err
		}
		podAddress := swarm.NewAddress(addr).String()
		return info.GetDirectory().DirStat(podName, path, dirInode, acc.String(), podAddress, printNames)
	}
	return nil, fmt.Errorf("directory not found")
}

func (p *Pod) FileStat(podName, podFileOrDir string) (*file.FileStats, error) {
	if !p.isPodOpened(podName) {
		return nil, fmt.Errorf("login to pod to do this operation")
	}

	info, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, err
	}

	acc := info.getAccountInfo().GetAddress()

	path := p.getDirectoryPath(podFileOrDir, info)
	if !info.file.IsFileAlreadyPResent(path) {
		return nil, fmt.Errorf("file not present in pod")
	}
	return info.file.FileStat(podName, path, acc.String())
}
