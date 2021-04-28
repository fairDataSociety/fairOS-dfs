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

package dir

import (
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (d *Directory) RmDir(podName, parentPath, dirName string) error {
	// validation checks of the arguments
	if podName == "" {
		return ErrInvalidPodName
	}

	if dirName == "" {
		return ErrInvalidDirectoryName
	}

	// check if directory present
	totalPath := podName + parentPath + utils.PathSeperator + dirName
	if d.GetDirFromDirectoryMap(totalPath) == nil {
		return ErrDirectoryNotPresent
	}

	// remove the feed and clear the data structure
	topic := utils.HashString(totalPath)
	err := d.fd.DeleteFeed(topic, d.userAddress)
	if err != nil {
		return err
	}
	d.RemoveFromDirectoryMap(totalPath)
	return nil
}
