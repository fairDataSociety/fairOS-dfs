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
	"os"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (p *Pod) CopyToLocal(podName string, podFile string, localDir string) error {
	if !p.isPodOpened(podName) {
		return fmt.Errorf("login to pod to do this operation")
	}

	podInfo, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}

	dirStat, err := os.Stat(localDir)
	if err != nil {
		return err
	}

	if !dirStat.IsDir() {
		return fmt.Errorf("local path is not a directory")
	}

	var path string
	if podInfo.IsCurrentDirRoot() {
		path = podInfo.GetCurrentPodPathAndName() + podFile
	} else {
		path = podInfo.GetCurrentDirPathAndName() + utils.PathSeperator + podFile
	}

	if !podInfo.getFile().IsFileAlreadyPResent(path) {
		return fmt.Errorf("file not present in pod")
	}

	err = podInfo.getFile().CopyToFile(path, localDir)
	if err != nil {
		return err
	}
	return nil
}
