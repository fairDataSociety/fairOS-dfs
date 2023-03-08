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

import "fmt"

// Stat represents a pod name and address
type Stat struct {
	PodName    string `json:"podName"`
	PodAddress string `json:"address"`
}

// PodStat shows all the pod related information like podname and its current address.
func (p *Pod) PodStat(podName string) (*Stat, error) {
	podInfo, _, err := p.GetPodInfoFromPodMap(podName)
	if err == nil {
		return &Stat{
			PodName:    podInfo.GetPodName(),
			PodAddress: podInfo.userAddress.String(),
		}, nil
	}
	podList, err := p.loadUserPods()
	if err != nil {
		return nil, err
	}

	index, _ := p.getIndexPassword(podList, podName)
	if index == -1 {
		return nil, fmt.Errorf("pod does not exist")
	}
	// Create pod account and other data structures
	// create a child account for the userAddress and other data structures for the pod
	accountInfo, err := p.acc.CreatePodAccount(index, false)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	addr := accountInfo.GetAddress()
	return &Stat{
		PodName:    podName,
		PodAddress: addr.String(),
	}, nil
}
