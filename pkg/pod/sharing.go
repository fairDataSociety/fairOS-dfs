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
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

type ShareInfo struct {
	PodName     string `json:"pod_name"`
	Address     string `json:"pod_address"`
	UserName    string `json:"user_name"`
	UserAddress string `json:"user_address"`
	SharedTime  string `json:"shared_time"`
}

func (p *Pod) GetMetaReferenceOfFile(podName, filePath string) ([]byte, string, error) {
	if !p.isPodOpened(podName) {
		return nil, "", fmt.Errorf("login to pod to do this operation")
	}

	podInfo, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, "", err
	}

	podDir := filepath.Dir(filePath)
	fileName := filepath.Base(filePath)
	path := p.getFilePath(podDir, podInfo)
	fpath := path + utils.PathSeperator + fileName

	return podInfo.getFile().GetFileReference(fpath)
}

func (p *Pod) ReceiveFileAndStore(podName, podDir, fileName, metaHexRef string) error {
	if !p.isPodOpened(podName) {
		return fmt.Errorf("login to pod to do this operation")
	}

	podInfo, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}

	path := p.getFilePath(podDir, podInfo)
	dir := podInfo.GetDirectory()

	_, dirInode, err := dir.GetDirNode(path, podInfo.GetFeed(), podInfo.GetAccountInfo())
	if err != nil {
		return err
	}

	// check if the file exists already
	fpath := path + utils.PathSeperator + fileName
	if podInfo.file.IsFileAlreadyPResent(fpath) {
		return fmt.Errorf("file already present in the destination dir")
	}

	// append the file meta to the parent directory and update the directory feed
	metaReference, err := utils.ParseHexReference(metaHexRef)
	if err != nil {
		return err
	}
	dirInode.Hashes = append(dirInode.Hashes, metaReference.Bytes())
	dirInode.Meta.ModificationTime = time.Now().Unix()
	topic, err := dir.UpdateDirectory(dirInode)
	if err != nil {
		return err
	}

	// if the directory path is not root.. then update all the parents too
	if path != podInfo.GetCurrentPodPathAndName() {
		err = p.UpdateTillThePod(podName, podInfo.GetDirectory(), topic, path, true)
		if err != nil {
			return err
		}
	}

	// Add to file path map
	return podInfo.getFile().AddFileToPath(fpath, metaHexRef)
}

func (p *Pod) PodShare(podName, passPhrase, userName string) (string, error) {
	// check if pods is present and get the index of the pod
	pods, _, err := p.loadUserPods()
	if err != nil {
		return "", err
	}
	if !p.checkIfPodPresent(pods, podName) {
		return "", ErrInvalidPodName
	}

	index := p.getIndex(pods, podName)
	if index == -1 {
		return "", fmt.Errorf("pod does not exist")
	}

	// Create pod account  and get the address
	accountInfo, err := p.acc.CreatePodAccount(index, passPhrase, false)
	if err != nil {
		return "", err
	}

	address := accountInfo.GetAddress()
	userAddress := p.acc.GetUserAccountInfo().GetAddress()
	shareInfo := &ShareInfo{
		PodName:     podName,
		Address:     address.String(),
		UserName:    userName,
		UserAddress: userAddress.String(),
		SharedTime:  time.Now().String(),
	}

	data, err := json.Marshal(shareInfo)
	if err != nil {
		return "", err
	}

	ref, err := p.client.UploadBlob(data, true, true)
	if err != nil {
		return "", err
	}

	shareInfoRef := utils.NewReference(ref)
	return shareInfoRef.String(), nil
}

func (p *Pod) ReceivePodInfo(ref utils.Reference) (*ShareInfo, error) {
	data, resp, err := p.client.DownloadBlob(ref.Bytes())
	if err != nil {
		return nil, err
	}

	if resp != http.StatusOK {
		return nil, fmt.Errorf("ReceivePodInfo: could not download blob")
	}

	var shareInfo ShareInfo
	err = json.Unmarshal(data, &shareInfo)
	if err != nil {
		return nil, err
	}

	return &shareInfo, nil

}

func (p *Pod) ReceivePod(ref utils.Reference) (*Info, error) {
	data, resp, err := p.client.DownloadBlob(ref.Bytes())
	if err != nil {
		return nil, err
	}
	if resp != http.StatusOK {
		return nil, fmt.Errorf("ReceivePod: could not download blob")
	}

	var shareInfo ShareInfo
	err = json.Unmarshal(data, &shareInfo)
	if err != nil {
		return nil, err
	}

	return p.CreatePod(shareInfo.PodName, "", shareInfo.Address)
}
