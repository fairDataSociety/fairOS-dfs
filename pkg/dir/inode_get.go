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
	"encoding/json"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (d *Directory) GetDirNode(name string, fd *feed.API, accountInfo *account.Info) ([]byte, *DirInode, error) {
	topic := utils.HashString(name)
	addr, data, err := fd.GetFeedData(topic, accountInfo.GetAddress())
	if err != nil {
		return nil, nil, err
	}

	var dirInode DirInode
	err = json.Unmarshal(data, &dirInode)
	if err != nil {
		return nil, nil, err
	}
	return addr, &dirInode, nil
}
