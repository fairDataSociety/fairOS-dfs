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
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	c "github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (p *Pod) OpenPod(podName, passPhrase string) (*Info, error) {
	// check if pods is present and get the index of the pod
	pods, err := p.loadUserPods()
	if err != nil {
		return nil, err
	}
	if !p.checkIfPodPresent(pods, podName) {
		return nil, ErrInvalidPodName
	}

	index := p.getIndex(pods, podName)
	if index == -1 {
		return nil, fmt.Errorf("pod does not exist")
	}

	// Create pod account and other data structures
	// create a child account for the user and other data structures for the pod
	err = p.acc.CreatePodAccount(index, passPhrase, false)
	if err != nil {
		return nil, err
	}
	accountInfo, err := p.acc.GetPodAccountInfo(index)
	if err != nil {
		return nil, err
	}
	file := f.NewFile(podName, p.client, p.fd, accountInfo, p.logger)
	dir := d.NewDirectory(podName, p.client, p.fd, accountInfo, file, p.logger)

	// get the pod's inode
	_, dirInode, err := dir.GetDirNode(utils.PathSeperator+podName, p.fd, accountInfo)
	if err != nil {
		return nil, err
	}

	user := p.acc.GetAddress(account.UserAccountIndex)
	collection := c.NewKeyValueStore(p.fd, accountInfo, user, p.client)

	// create the pod info and store it in the podMap
	podInfo := &Info{
		podName:         podName,
		user:            user,
		accountInfo:     accountInfo,
		feed:            p.fd,
		dir:             dir,
		file:            file,
		currentPodInode: dirInode,
		curPodMu:        sync.RWMutex{},
		currentDirInode: dirInode,
		curDirMu:        sync.RWMutex{},
		collection:      collection,
	}

	p.addPodToPodMap(podName, podInfo)
	dir.AddToDirectoryMap(podName, dirInode)

	// sync the pod's files and directories
	err = p.SyncPod(podName)
	if err != nil {
		return nil, err
	}

	return podInfo, nil
}

func (p *Pod) getIndex(pods map[int]string, podName string) int {
	for index, pod := range pods {
		if strings.Trim(pod, "\n") == podName {
			return index
		}
	}
	return -1
}
