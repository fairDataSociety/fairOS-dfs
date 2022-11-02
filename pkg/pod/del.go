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
)

// DeleteOwnPod removed a pod and the list of pods belonging to a user.
func (p *Pod) DeleteOwnPod(podName string) error {
	podList, err := p.loadUserPods()
	if err != nil { // skipcq: TCV-001
		return err
	}
	found := false
	var podIndex int
	for index, pod := range podList.Pods {
		if pod.Name == podName {
			podList.Pods = append(podList.Pods[:index], podList.Pods[index+1:]...)
			podIndex = index
			found = true
		}
	}
	if !found {
		return fmt.Errorf("pod not found")
	}

	// delete tables
	podInfo, _, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}

	err = podInfo.GetDocStore().DeleteAllDocumentDBs()
	if err != nil {
		return err
	}

	err = podInfo.GetKVStore().DeleteAllKVTables()
	if err != nil {
		return err
	}

	// remove it from other data structures
	p.removePodFromPodMap(podName)
	p.acc.DeletePodAccount(podIndex)

	// remove the pod finally
	return p.storeUserPods(podList)
}

// DeleteSharedPod removed a pod and the list of pods shared by other users.
func (p *Pod) DeleteSharedPod(podName string) error {
	podList, err := p.loadUserPods()
	if err != nil { // skipcq: TCV-001
		return err
	}

	found := false
	for index, pod := range podList.SharedPods {
		if pod.Name == podName {
			podList.SharedPods = append(podList.SharedPods[:index], podList.SharedPods[index+1:]...)
			found = true
		}
	}
	if !found {
		return fmt.Errorf("pod not found")
	}

	// remove it from other data structures
	p.removePodFromPodMap(podName)

	// remove the pod finally
	return p.storeUserPods(podList)
}
