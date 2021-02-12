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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

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

func (p *Pod) CreatePod(podName, passPhrase, addressString string) (*Info, error) {
	podName, err := CleanPodName(podName)
	if err != nil {
		return nil, err
	}

	// check if pods is present and get free index
	pods, sharedPods, err := p.loadUserPods()
	if err != nil {
		return nil, err
	}

	var accountInfo *account.Info
	var fd *feed.API
	var file *f.File
	var dir *d.Directory
	var dirInode *d.DirInode
	var user utils.Address
	if addressString != "" {
		if p.checkIfPodPresent(pods, podName) {
			return nil, ErrPodAlreadyExists
		}
		if p.checkIfSharedPodPresent(sharedPods, podName) {
			return nil, ErrPodAlreadyExists
		}

		// shared pod, so add only address to the account info
		accountInfo = p.acc.GetEmptyAccountInfo()
		address := utils.HexToAddress(addressString)
		accountInfo.SetAddress(address)

		fd = feed.New(accountInfo, p.client, p.logger)
		file = f.NewFile(podName, p.client, fd, accountInfo, p.logger)
		dir = d.NewDirectory(podName, p.client, fd, accountInfo, file, p.logger)

		// get the inode instead of creating
		_, dirInode, err = dir.GetDirNode(utils.PathSeperator+podName, fd, accountInfo)
		if err != nil {
			return nil, err
		}

		// store the pod file with shared pod
		sharedPods[addressString] = podName
		err = p.storeUserPods(pods, sharedPods)
		if err != nil {
			return nil, err
		}

		// set the user as the pod address we got from shared pod
		user = address
	} else {
		// your own pod, so create a new account with private key
		if p.checkIfPodPresent(pods, podName) {
			return nil, ErrPodAlreadyExists
		}

		if p.checkIfSharedPodPresent(sharedPods, podName) {
			return nil, ErrPodAlreadyExists
		}

		freeId, err := p.getFreeId(pods)
		if err != nil {
			return nil, err
		}

		// create a child account for the user and other data structures for the pod
		accountInfo, err = p.acc.CreatePodAccount(freeId, passPhrase, true)
		if err != nil {
			return nil, err
		}

		fd = feed.New(accountInfo, p.client, p.logger)
		file = f.NewFile(podName, p.client, fd, accountInfo, p.logger)
		dir = d.NewDirectory(podName, p.client, fd, accountInfo, file, p.logger)

		// create the pod inode
		dirInode, _, err = dir.CreatePodINode(podName)
		if err != nil {
			return nil, err
		}

		// store the pod file
		pods[freeId] = podName
		err = p.storeUserPods(pods, sharedPods)
		if err != nil {
			return nil, err
		}

		user = p.acc.GetAddress(freeId)
	}

	kvStore := c.NewKeyValueStore(fd, accountInfo, user, p.client, p.logger)
	docStore := c.NewDocumentStore(fd, accountInfo, user, file, p.client, p.logger)

	// create the pod info and store it in the podMap
	podInfo := &Info{
		podName:         podName,
		user:            user,
		dir:             dir,
		file:            file,
		accountInfo:     accountInfo,
		feed:            fd,
		currentPodInode: dirInode,
		curPodMu:        sync.RWMutex{},
		currentDirInode: dirInode,
		curDirMu:        sync.RWMutex{},
		kvStore:         kvStore,
		docStore:        docStore,
	}
	p.addPodToPodMap(podName, podInfo)
	dir.AddToDirectoryMap(podName, dirInode)

	return podInfo, nil
}

func (p *Pod) loadUserPods() (map[int]string, map[string]string, error) {
	// The user pod file topic should be in the name of the user account
	topic := utils.HashString(podFile)
	_, data, err := p.fd.GetFeedData(topic, p.acc.GetAddress(account.UserAccountIndex))
	if err != nil {
		if err.Error() != "no feed updates found" {
			return nil, nil, err
		}
	}

	buf := bytes.NewBuffer(data)
	rd := bufio.NewReader(buf)
	pods := make(map[int]string)
	sharedPods := make(map[string]string)
	for {
		line, err := rd.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("loading pods: %w", err)
		}
		line = strings.Trim(line, "\n")
		lines := strings.Split(line, ",")
		index, err := strconv.ParseInt(lines[1], 10, 64)
		if err != nil {
			sharedPods[lines[1]] = lines[0]
			continue
		}
		pods[int(index)] = lines[0]
	}
	return pods, sharedPods, nil
}

func (p *Pod) storeUserPods(pods map[int]string, sharedPods map[string]string) error {
	buf := bytes.NewBuffer(nil)
	podLen := len(pods)
	for index, pod := range pods {
		pod := strings.Trim(pod, "\n")
		if podLen > 1 && pod == "" {
			continue
		}
		line := fmt.Sprintf("%s,%d", pod, index)
		buf.WriteString(line + "\n")
	}

	for addr, pod := range sharedPods {
		pod := strings.Trim(pod, "\n")
		if podLen > 1 && pod == "" {
			continue
		}
		line := fmt.Sprintf("%s,%s", pod, addr)
		buf.WriteString(line + "\n")
	}

	topic := utils.HashString(podFile)
	_, err := p.fd.UpdateFeed(topic, p.acc.GetAddress(account.UserAccountIndex), buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (p *Pod) getFreeId(pods map[int]string) (int, error) {
	for i := 0; i < maxPodId; i++ {
		if _, ok := pods[i]; !ok {
			return i, nil
		}
	}
	return 0, ErrMaxPodsReached
}

func (p *Pod) checkIfPodPresent(pods map[int]string, podName string) bool {
	for _, pod := range pods {
		if strings.Trim(pod, "\n") == podName {
			return true
		}
	}
	return false
}

func (p *Pod) checkIfSharedPodPresent(sharedPods map[string]string, podName string) bool {
	for _, pod := range sharedPods {
		if strings.Trim(pod, "\n") == podName {
			return true
		}
	}
	return false
}

func (p *Pod) getPodIndex(podName string) (int, error) {
	pods, _, err := p.loadUserPods()
	if err != nil {
		return -1, err
	}
	podIndex := -1
	for index, pod := range pods {
		if strings.Trim(pod, "\n") == podName {
			delete(pods, index)
			podIndex = index
		}
	}
	return podIndex, nil
}
