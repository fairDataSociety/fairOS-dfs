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
	"encoding/hex"
	"encoding/json"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	c "github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	f "github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	podFile = "Pods"
)

// CreatePod creates a new pod for a given user.
func (p *Pod) CreatePod(podName, addressString, podPassword string) (*Info, error) {
	podName, err := CleanPodName(podName)
	if err != nil {
		return nil, err
	}

	// check if pods is present and get free index
	podList, err := p.loadUserPods()
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	pods := map[int]string{}
	sharedPods := map[string]string{}
	for _, pod := range podList.Pods {
		pods[pod.Index] = pod.Name
	}

	for _, pod := range podList.SharedPods {
		sharedPods[pod.Address] = pod.Name
	}
	var accountInfo *account.Info
	var fd *feed.API
	var file *f.File
	var dir *d.Directory
	var user utils.Address
	if addressString != "" {
		if p.checkIfPodPresent(podList, podName) {
			return nil, ErrPodAlreadyExists
		}
		if p.checkIfSharedPodPresent(podList, podName) {
			return nil, ErrPodAlreadyExists
		}

		// shared pod, so add only address to the account info
		accountInfo = p.acc.GetEmptyAccountInfo()
		address := utils.HexToAddress(addressString)
		accountInfo.SetAddress(address)

		fd = feed.New(accountInfo, p.client, p.logger)
		file = f.NewFile(podName, p.client, fd, accountInfo.GetAddress(), p.tm, p.logger)
		dir = d.NewDirectory(podName, p.client, fd, accountInfo.GetAddress(), file, p.tm, p.logger)

		// store the pod file with shared pod
		sharedPod := &SharedPodListItem{
			Name:     podName,
			Address:  addressString,
			Password: podPassword,
		}
		podList.SharedPods = append(podList.SharedPods, *sharedPod)
		err = p.storeUserPods(podList)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}

		// set the userAddress as the pod address we got from shared pod
		user = address

	} else {
		// your own pod, so create a new account with private key
		if p.checkIfPodPresent(podList, podName) {
			return nil, ErrPodAlreadyExists
		}
		if p.checkIfSharedPodPresent(podList, podName) {
			return nil, ErrPodAlreadyExists
		}

		freeId, err := p.getFreeId(pods)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}

		// create a child account for the userAddress and other data structures for the pod
		accountInfo, err = p.acc.CreatePodAccount(freeId, true)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}

		fd = feed.New(accountInfo, p.client, p.logger)
		file = f.NewFile(podName, p.client, fd, accountInfo.GetAddress(), p.tm, p.logger)
		dir = d.NewDirectory(podName, p.client, fd, accountInfo.GetAddress(), file, p.tm, p.logger)

		// store the pod file
		pods[freeId] = podName
		pod := &PodListItem{
			Name:     podName,
			Index:    freeId,
			Password: podPassword,
		}
		podList.Pods = append(podList.Pods, *pod)
		err = p.storeUserPods(podList)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		user = p.acc.GetAddress(freeId)
	}

	kvStore := c.NewKeyValueStore(podName, fd, accountInfo, user, p.client, p.logger)
	docStore := c.NewDocumentStore(podName, fd, accountInfo, user, file, p.client, p.logger)

	// create the pod info and store it in the podMap
	podInfo := &Info{
		podName:     podName,
		podPassword: podPassword,
		userAddress: user,
		dir:         dir,
		file:        file,
		accountInfo: accountInfo,
		feed:        fd,
		kvStore:     kvStore,
		docStore:    docStore,
	}
	p.addPodToPodMap(podName, podInfo)
	return podInfo, nil
}

func (p *Pod) loadUserPods() (*PodList, error) {
	// The userAddress pod file topic should be in the name of the userAddress account
	topic := utils.HashString(podFile)
	privKeyBytes := crypto.FromECDSA(p.acc.GetUserAccountInfo().GetPrivateKey())
	_, data, err := p.fd.GetFeedData(topic, p.acc.GetAddress(account.UserAccountIndex), []byte(hex.EncodeToString(privKeyBytes)))
	if err != nil { // skipcq: TCV-001
		if err.Error() != "feed does not exist or was not updated yet" {
			return nil, err
		}
	}
	podList := &PodList{
		Pods:       []PodListItem{},
		SharedPods: []SharedPodListItem{},
	}
	if len(data) == 0 {
		return podList, nil
	}

	err = json.Unmarshal(data, podList)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	return podList, nil
}

func (p *Pod) storeUserPods(podList *PodList) error {
	data, err := json.Marshal(podList)
	if err != nil {
		return err
	}

	if len(data) > utils.MaxChunkLength {
		return ErrMaximumPodLimit
	}
	topic := utils.HashString(podFile)

	privKeyBytes := crypto.FromECDSA(p.acc.GetUserAccountInfo().GetPrivateKey())
	_, err = p.fd.UpdateFeed(topic, p.acc.GetAddress(account.UserAccountIndex), data, []byte(hex.EncodeToString(privKeyBytes)))
	if err != nil { // skipcq: TCV-001
		return err
	}
	return nil
}

func (*Pod) getFreeId(pods map[int]string) (int, error) {
	for i := 0; i < maxPodId; i++ {
		if _, ok := pods[i]; !ok {
			if i == 0 {
				// this is the root account patch id
				continue
			}
			return i, nil
		}
	}
	return 0, ErrMaxPodsReached // skipcq: TCV-001
}

func (*Pod) checkIfPodPresent(pods *PodList, podName string) bool {
	for _, pod := range pods.Pods {
		if pod.Name == podName {
			return true
		}
	}
	return false
}

func (*Pod) checkIfSharedPodPresent(pods *PodList, podName string) bool {
	for _, pod := range pods.SharedPods {
		if pod.Name == podName {
			return true
		}
	}
	return false
}

func (p *Pod) getPodIndex(podName string) (podIndex int, err error) {
	podList, err := p.loadUserPods()
	if err != nil {
		return -1, err
	} // skipcq: TCV-001
	podIndex = -1
	for _, pod := range podList.Pods {
		if pod.Name == podName {
			podIndex = pod.Index
			return
		}
	}
	return
}
