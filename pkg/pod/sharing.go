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
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

type ShareInfo struct {
	PodName     string `json:"pod_name"`
	Address     string `json:"pod_address"`
	UserName    string `json:"user_name"`
	UserAddress string `json:"user_address"`
	SharedTime  string `json:"shared_time"`
}

// PodShare makes a pod public by exporting all the pod related information and its
// address. it does this by creating a sharing reference which points to the information
// required to import this pod.
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
	accountInfo, err := p.acc.CreatePodAccountV2(index, crypto.FromECDSA(p.acc.GetUserAccountInfo().GetPrivateKey()))
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
