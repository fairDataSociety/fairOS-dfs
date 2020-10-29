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
	"net/http"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (d *Directory) LoadDirMeta(podName string, curDirInode *DirInode, fd *feed.API, accountInfo *account.Info) error {
	for _, ref := range curDirInode.Hashes {
		_, data, err := fd.GetFeedData(ref, accountInfo.GetAddress())
		if err != nil {
			respCode, err := d.file.LoadFileMeta(podName, ref)
			if err != nil {
				return err
			}
			if respCode == http.StatusOK {
				continue
			}
			return err
		}

		var dirInode *DirInode
		err = json.Unmarshal(data, &dirInode)
		if err != nil {
			return err
		}

		path := dirInode.Meta.Path + utils.PathSeperator + dirInode.Meta.Name
		d.AddToDirectoryMap(path, dirInode)
		d.logger.Infof(path)

		_, newDirInode, err := d.GetDirNode(path, fd, accountInfo)
		if err != nil {
			return err
		}
		err = d.LoadDirMeta(podName, newDirInode, fd, accountInfo)
		if err != nil {
			return err
		}

	}
	return nil
}
