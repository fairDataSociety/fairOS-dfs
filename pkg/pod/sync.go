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
	"sync"
)

// SyncPod syncs the pod to the latest version by extracting the current meta information
// of files and directories of the pod.
func (p *Pod) SyncPod(podName string) error {
	podName, err := CleanPodName(podName)
	if err != nil { // skipcq: TCV-001
		return err
	}

	if !p.IsPodOpened(podName) {
		return ErrPodNotOpened
	}

	podInfo, _, err := p.GetPodInfoFromPodMap(podName)
	if err != nil { // skipcq: TCV-001
		return err
	}

	// sync from the root directory
	err = podInfo.GetDirectory().SyncDirectory("/", podInfo.GetPodPassword())
	if err != nil {
		return err
	}
	return nil
}

// SyncPodAsync syncs the pod to the latest version by extracting the current meta information
// of files and directories of the pod, concurrently.
func (p *Pod) SyncPodAsync(ctx context.Context, podName string) error {
	podName, err := CleanPodName(podName)
	if err != nil { // skipcq: TCV-001
		return err
	}

	if !p.IsPodOpened(podName) {
		return ErrPodNotOpened
	}

	podInfo, _, err := p.GetPodInfoFromPodMap(podName)
	if err != nil { // skipcq: TCV-001
		return err
	}

	// sync from the root directory
	wg := new(sync.WaitGroup)
	err = podInfo.GetDirectory().SyncDirectoryAsync(ctx, "/", podInfo.GetPodPassword(), wg)
	if err != nil {
		return err
	}
	wg.Wait()
	return nil
}
