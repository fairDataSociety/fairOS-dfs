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

	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	c "github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// OpenPod opens a pod if it is not already opened. as part of opening the pod
// it loads all the data structures related to the pod. Also it syncs all the
// files and directories under this pod from the Swarm network.
func (p *Pod) OpenPod(podName, passPhrase string) (*Info, error) {
	// check if pods is present and get the index of the pod
	pods, sharedPods, err := p.loadUserPods()
	if err != nil {
		return nil, err
	}

	sharedPodType := false
	if !p.checkIfPodPresent(pods, podName) {
		if !p.checkIfSharedPodPresent(sharedPods, podName) {
			return nil, ErrInvalidPodName
		} else {
			sharedPodType = true
		}
	}

	var accountInfo *account.Info
	var file *f.File
	var fd *feed.API
	var dir *d.Directory
	var user utils.Address
	if sharedPodType {
		addressString := p.getAddress(sharedPods, podName)
		if addressString == "" {
			return nil, fmt.Errorf("shared pod does not exist")
		}

		accountInfo = p.acc.GetEmptyAccountInfo()
		address := utils.HexToAddress(addressString)
		accountInfo.SetAddress(address)

		fd = feed.New(accountInfo, p.client, p.logger)
		file = f.NewFile(podName, p.client, fd, accountInfo.GetAddress(), p.logger)
		dir = d.NewDirectory(podName, p.client, fd, accountInfo.GetAddress(), file, p.logger)

		// set the userAddress as the pod address we got from shared pod
		user = address
	} else {
		index := p.getIndex(pods, podName)
		if index == -1 {
			return nil, fmt.Errorf("pod does not exist")
		}
		// Create pod account and other data structures
		// create a child account for the userAddress and other data structures for the pod
		accountInfo, err = p.acc.CreatePodAccount(index, passPhrase, false)
		if err != nil {
			return nil, err
		}

		fd = feed.New(accountInfo, p.client, p.logger)
		file = f.NewFile(podName, p.client, fd, accountInfo.GetAddress(), p.logger)
		dir = d.NewDirectory(podName, p.client, fd, accountInfo.GetAddress(), file, p.logger)

		user = p.acc.GetAddress(index)
	}

	kvStore := c.NewKeyValueStore(podName, fd, accountInfo, user, p.client, p.logger)
	docStore := c.NewDocumentStore(podName, fd, accountInfo, user, file, p.client, p.logger)

	// create the pod info and store it in the podMap
	podInfo := &Info{
		podName:     podName,
		userAddress: user,
		accountInfo: accountInfo,
		feed:        fd,
		dir:         dir,
		file:        file,
		kvStore:     kvStore,
		docStore:    docStore,
	}

	p.addPodToPodMap(podName, podInfo)

	// sync the pod's files and directories
	err = p.SyncPod(podName)
	if err != nil && err != d.ErrResourceDeleted {
		return nil, err
	}
	return podInfo, nil
}

func (*Pod) getIndex(pods map[int]string, podName string) int {
	for index, pod := range pods {
		if strings.Trim(pod, "\n") == podName {
			return index
		}
	}
	return -1 // skipcq: TCV-001
}

func (*Pod) getAddress(sharedPods map[string]string, podName string) string {
	for address, pod := range sharedPods {
		if strings.Trim(pod, "\n") == podName {
			return address
		}
	}
	return ""
}
