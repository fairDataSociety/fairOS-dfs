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
	"context"
	"fmt"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	c "github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// OpenPod opens a pod if it is not already opened. as part of opening the pod
// it loads all the data structures related to the pod. Also, it syncs all the
// files and directories under this pod from the Swarm network.
func (p *Pod) OpenPod(podName string) (*Info, error) {
	// check if pods is present and get the index of the pod
	podList, err := p.loadUserPods()
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	sharedPodType := false
	if !p.checkIfPodPresent(podList, podName) {
		if !p.checkIfSharedPodPresent(podList, podName) {
			return nil, ErrInvalidPodName
		} else {
			sharedPodType = true
		}
	}
	var (
		podPassword string
		accountInfo *account.Info
		file        *f.File
		fd          *feed.API
		dir         *d.Directory
		user        utils.Address
	)
	if sharedPodType {
		var addressString string
		addressString, podPassword = p.getAddressPassword(podList, podName)
		if addressString == "" { // skipcq: TCV-001
			return nil, fmt.Errorf("shared pod does not exist")
		}

		accountInfo = p.acc.GetEmptyAccountInfo()
		address := utils.HexToAddress(addressString)
		accountInfo.SetAddress(address)

		fd = feed.New(accountInfo, p.client, p.logger)
		file = f.NewFile(podName, p.client, fd, accountInfo.GetAddress(), p.tm, p.logger)
		dir = d.NewDirectory(podName, p.client, fd, accountInfo.GetAddress(), file, p.tm, p.logger)

		// set the userAddress as the pod address we got from shared pod
		user = address
	} else {
		var index int
		index, podPassword = p.getIndexPassword(podList, podName)
		if index == -1 {
			return nil, fmt.Errorf("pod does not exist")
		}
		// Create pod account and other data structures
		// create a child account for the userAddress and other data structures for the pod
		accountInfo, err = p.acc.CreatePodAccount(index, false)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}

		fd = feed.New(accountInfo, p.client, p.logger)
		file = f.NewFile(podName, p.client, fd, accountInfo.GetAddress(), p.tm, p.logger)
		dir = d.NewDirectory(podName, p.client, fd, accountInfo.GetAddress(), file, p.tm, p.logger)

		user = p.acc.GetAddress(index)
	}

	kvStore := c.NewKeyValueStore(podName, fd, accountInfo, user, p.client, p.logger)
	docStore := c.NewDocumentStore(podName, fd, accountInfo, user, file, p.tm, p.client, p.logger)

	// create the pod info and store it in the podMap
	podInfo := &Info{
		podName:     podName,
		podPassword: podPassword,
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
	if err != nil && err != d.ErrResourceDeleted { // skipcq: TCV-001
		return nil, err
	}
	return podInfo, nil
}

func (p *Pod) OpenFromReference(ref utils.Reference) (*Info, error) {
	si, err := p.ReceivePodInfo(ref)
	if err != nil {
		return nil, err
	}

	accountInfo := p.acc.GetEmptyAccountInfo()
	address := utils.HexToAddress(si.Address)
	accountInfo.SetAddress(address)

	fd := feed.New(accountInfo, p.client, p.logger)
	file := f.NewFile(si.PodName, p.client, fd, accountInfo.GetAddress(), p.tm, p.logger)
	dir := d.NewDirectory(si.PodName, p.client, fd, accountInfo.GetAddress(), file, p.tm, p.logger)

	kvStore := c.NewKeyValueStore(si.PodName, fd, accountInfo, address, p.client, p.logger)
	docStore := c.NewDocumentStore(si.PodName, fd, accountInfo, address, file, p.tm, p.client, p.logger)

	podInfo := &Info{
		podName:     si.PodName,
		podPassword: si.Password,
		userAddress: address,
		accountInfo: accountInfo,
		feed:        fd,
		dir:         dir,
		file:        file,
		kvStore:     kvStore,
		docStore:    docStore,
	}
	p.addPodToPodMap(si.PodName, podInfo)

	// sync the pod's files and directories
	err = p.SyncPod(si.PodName)
	if err != nil && err != d.ErrResourceDeleted { // skipcq: TCV-001
		return nil, err
	}

	return podInfo, nil
}

// OpenPodAsync opens a pod if it is not already opened. as part of opening the pod
// it loads all the data structures related to the pod. Also, it syncs all the
// files and directories under this pod from the Swarm network.
func (p *Pod) OpenPodAsync(ctx context.Context, podName string) (*Info, error) {
	// check if pods is present and get the index of the pod
	podList, err := p.loadUserPods()
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	sharedPodType := false
	if !p.checkIfPodPresent(podList, podName) {
		if !p.checkIfSharedPodPresent(podList, podName) {
			return nil, ErrInvalidPodName
		} else {
			sharedPodType = true
		}
	}

	var (
		podPassword string
		accountInfo *account.Info
		file        *f.File
		fd          *feed.API
		dir         *d.Directory
		user        utils.Address
	)
	if sharedPodType {
		var addressString string
		addressString, podPassword = p.getAddressPassword(podList, podName)
		if addressString == "" { // skipcq: TCV-001
			return nil, fmt.Errorf("shared pod does not exist")
		}

		accountInfo = p.acc.GetEmptyAccountInfo()
		address := utils.HexToAddress(addressString)
		accountInfo.SetAddress(address)

		fd = feed.New(accountInfo, p.client, p.logger)
		file = f.NewFile(podName, p.client, fd, accountInfo.GetAddress(), p.tm, p.logger)
		dir = d.NewDirectory(podName, p.client, fd, accountInfo.GetAddress(), file, p.tm, p.logger)

		// set the userAddress as the pod address we got from shared pod
		user = address
	} else {
		var index int
		index, podPassword = p.getIndexPassword(podList, podName)
		if index == -1 {
			return nil, fmt.Errorf("pod does not exist")
		}
		// Create pod account and other data structures
		// create a child account for the userAddress and other data structures for the pod
		accountInfo, err = p.acc.CreatePodAccount(index, false)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}

		fd = feed.New(accountInfo, p.client, p.logger)
		file = f.NewFile(podName, p.client, fd, accountInfo.GetAddress(), p.tm, p.logger)
		dir = d.NewDirectory(podName, p.client, fd, accountInfo.GetAddress(), file, p.tm, p.logger)

		user = p.acc.GetAddress(index)
	}

	kvStore := c.NewKeyValueStore(podName, fd, accountInfo, user, p.client, p.logger)
	docStore := c.NewDocumentStore(podName, fd, accountInfo, user, file, p.tm, p.client, p.logger)

	// create the pod info and store it in the podMap
	podInfo := &Info{
		podName:     podName,
		podPassword: podPassword,
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
	err = p.SyncPodAsync(ctx, podName)
	if err != nil && err != d.ErrResourceDeleted { // skipcq: TCV-001
		return nil, err
	}
	return podInfo, nil
}

func (*Pod) getIndexPassword(podList *List, podName string) (int, string) {
	for _, pod := range podList.Pods {
		if pod.Name == podName {
			return pod.Index, pod.Password
		}
	}
	return -1, "" // skipcq: TCV-001
}

func (*Pod) getAddressPassword(podList *List, podName string) (string, string) {
	for _, pod := range podList.SharedPods {
		if pod.Name == podName {
			return pod.Address, pod.Password
		}
	}
	return "", ""
}
