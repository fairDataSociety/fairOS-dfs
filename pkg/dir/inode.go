/*
Copyright © 2021 FairOS Authors

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
	"encoding/json"
	"errors"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

type Inode struct {
	Meta           *MetaData
	FileOrDirNames []string
}

var (
	ErrResourceDeleted = errors.New("resource was deleted")
)

func (in *Inode) GetMeta() *MetaData {
	return in.Meta
}

func (in *Inode) GetFileOrDirNames() []string {
	return in.FileOrDirNames
}

func (in *Inode) SetFileOrDirNames(fileOrDirNames []string) {
	in.FileOrDirNames = fileOrDirNames
}

func (in *Inode) Unmarshal(data []byte) error {
	if string(data) == utils.DeletedFeedMagicWord {
		return ErrResourceDeleted
	}
	err := json.Unmarshal(data, in)
	if err != nil {
		return err
	}
	return nil
}
