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

// IsPodOpened checks if a pod is open
func (p *Pod) IsPodOpened(podName string) bool {
	p.podMu.Lock()
	defer p.podMu.Unlock()
	if _, ok := p.podMap[podName]; ok {
		return true
	}
	return false
}

// IsPodPresent checks if a pod is already present for user
func (p *Pod) IsPodPresent(podName string) bool {
	podName, err := cleanPodName(podName)
	if err != nil {
		return false
	}
	// check if pods is present and get free index
	pods, sharedPods, err := p.loadUserPods()
	if err != nil { // skipcq: TCV-001
		return false
	}
	if p.checkIfPodPresent(pods, podName) {
		return true
	}
	if p.checkIfSharedPodPresent(sharedPods, podName) {
		return true
	}
	return false
}

// GetPath returns the path of the node in a pod
func (*Pod) GetPath(inode *d.Inode) string {
	if inode != nil {
		return inode.Meta.Path
	}
	return "" // skipcq: TCV-001
}

// GetName returns the name of the node in a pod
func (*Pod) GetName(inode *d.Inode) string {
	if inode != nil {
		return inode.Meta.Name
	}
	return "" // skipcq: TCV-001
}

// GetAccountInfo returns the pod account info
func (p *Pod) GetAccountInfo(podName string) (*account.Info, error) {
	podInfo, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, err
	}
	return podInfo.GetAccountInfo(), nil
}

// cleanPodName trims spaces from a pod name
func cleanPodName(podName string) (string, error) {
	if podName == "" {
		return "", ErrInvalidPodName
	}
	if len(podName) > utils.MaxPodNameLength {
		return "", ErrTooLongPodName
	}
	podName = strings.TrimSpace(podName)
	return podName, nil
}
