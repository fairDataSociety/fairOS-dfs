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

// Stat represents a pod name and address
type Stat struct {
	PodName    string
	PodAddress string
}

// PodStat shows all the pod related information like podname and its current address.
func (p *Pod) PodStat(podName string) (*Stat, error) {
	podInfo, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, ErrInvalidPodName
	}
	return &Stat{
		PodName:    podInfo.GetPodName(),
		PodAddress: podInfo.userAddress.String(),
	}, nil
}
