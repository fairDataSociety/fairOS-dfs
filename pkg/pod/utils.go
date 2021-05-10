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
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (p *Pod) IsPodOpened(podName string) bool {
	p.podMu.Lock()
	defer p.podMu.Unlock()
	if _, ok := p.podMap[podName]; ok {
		return true
	}
	return false
}

func (p *Pod) GetPath(inode *d.Inode) string {
	if inode != nil {
		return inode.Meta.Path
	}
	return ""
}

func (p *Pod) GetName(inode *d.Inode) string {
	if inode != nil {
		return inode.Meta.Name
	}
	return ""
}

func (p *Pod) GetAccountInfo(podName string) (*account.Info, error) {
	podInfo, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, err
	}
	return podInfo.GetAccountInfo(), nil
}

func CleanPodName(podName string) (string, error) {
	if podName == "" {
		return "", ErrInvalidPodName
	}
	if len(podName) > utils.MaxPodNameLength {
		return "", ErrTooLongPodName
	}
	podName = strings.TrimSpace(podName)
	return podName, nil
}
