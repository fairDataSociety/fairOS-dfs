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

package file

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ethersphere/bee/pkg/swarm"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// RmFile deletes all the blocks of a file and it related meta data from the Swarm network.
func (f *File) RmFile(podFileWithPath string) error {
	meta, err := f.GetMetaFromFileName(podFileWithPath)
	if err != nil {
		return err
	}

	fdata, respCode, err := f.client.DownloadBlob(meta.InodeAddress)
	if err != nil {
		return err
	}

	if respCode != http.StatusOK {
		f.logger.Warningf("could not remove blocks in %s", swarm.NewAddress(meta.InodeAddress).String())
		return fmt.Errorf("could not remove blocks in %v", swarm.NewAddress(meta.InodeAddress).String())
	}

	// find the inode and remove the blocks present in the inode one by one
	var fInode *INode
	err = json.Unmarshal(fdata, &fInode)
	if err != nil {
		f.logger.Warningf("could not unmarshall data in address %s", swarm.NewAddress(meta.InodeAddress).String())
		return fmt.Errorf("could not unmarshall data in address %v", swarm.NewAddress(meta.InodeAddress).String())
	}
	err = f.client.DeleteBlob(meta.InodeAddress)
	if err != nil {
		f.logger.Errorf("could not delete file inode %s", swarm.NewAddress(meta.InodeAddress).String())
		return fmt.Errorf("could not delete file inode %v", swarm.NewAddress(meta.InodeAddress).String())
	}
	for _, fblocks := range fInode.Blocks {
		err = f.client.DeleteBlob(fblocks.Reference.Bytes())
		if err != nil {
			f.logger.Errorf("could not delete file block %s", swarm.NewAddress(fblocks.Reference.Bytes()).String())
			return fmt.Errorf("could not delete file inode %v", swarm.NewAddress(fblocks.Reference.Bytes()).String())
		}
	}

	// remove the meta
	topic := utils.HashString(podFileWithPath)
	_, err = f.fd.UpdateFeed(topic, f.userAddress, []byte{})
	if err != nil {
		return err
	}

	// remove the file from file map
	f.RemoveFromFileMap(podFileWithPath)

	return nil
}
