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
	"io"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (p *Pod) DownloadFile(podName, podFile string) (io.ReadCloser, string, error) {
	if !p.isPodOpened(podName) {
		return nil, "", fmt.Errorf("login to pod to do this operation")
	}

	podInfo, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return nil, "", err
	}

	var path string
	if podInfo.IsCurrentDirRoot() {
		path = podInfo.GetCurrentPodPathAndName() + podFile
	} else {
		path = podInfo.GetCurrentDirPathAndName() + utils.PathSeperator + podFile
	}

	if !podInfo.getFile().IsFileAlreadyPresent(path) {
		return nil, "", fmt.Errorf("file not present in pod")
	}

	reader, size, err := podInfo.getFile().Download(path)
	if err != nil {
		return nil, "",  err
	}
	return reader, size, nil
}
