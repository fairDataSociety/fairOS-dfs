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

// ListPods List all the available pods belonging to a user.
func (p *Pod) ListPods() ([]string, []string, error) {
	podList, err := p.loadUserPods()
	if err != nil { // skipcq: TCV-001
		return nil, nil, err
	}

	var listPods []string
	for _, pod := range podList.Pods {
		listPods = append(listPods, pod.Name)
	}

	var listSharedPods []string
	for _, pod := range podList.SharedPods {
		listSharedPods = append(listSharedPods, pod.Name)
	}

	return listPods, listSharedPods, nil
}

// PodList List all the available pods belonging to a user in json format.
func (p *Pod) PodList() (*PodList, error) {
	return p.loadUserPods()
}
