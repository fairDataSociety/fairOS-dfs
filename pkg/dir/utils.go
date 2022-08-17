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

// IsDirInodeRoot check if the node is root dir
func (in *Inode) IsDirInodeRoot() bool {
	return in.Meta.Name == utils.PathSeparator
}

// GetDirInodePathAndNameForRoot returns full path of the root node
func (in *Inode) GetDirInodePathAndNameForRoot() string {
	return in.Meta.Path + in.Meta.Name
}

// GetDirInodePathAndName returns full path of the node from root
func (in *Inode) GetDirInodePathAndName() string {
	if in.Meta.Path == "" {
		return in.Meta.Name
	} else if in.Meta.Path == utils.PathSeparator {
		return utils.PathSeparator + in.Meta.Name
	}
	return in.Meta.Path + utils.PathSeparator + in.Meta.Name
}

// GetDirInodePathOnly returns path of the node
func (in *Inode) GetDirInodePathOnly() string {
	return in.Meta.Path
}
