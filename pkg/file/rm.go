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
	"io"
	"net/http"

	"github.com/ethersphere/bee/v2/pkg/swarm"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// RmFile deletes all the blocks of a file, and it related metadata from the Swarm network.
func (f *File) RmFile(podFileWithPath, podPassword string) error {
	totalFilePath := utils.CombinePathAndFile(podFileWithPath, "")
	meta := f.GetInode(podPassword, totalFilePath)
	if meta == nil {
		return ErrFileNotFound
	}
	r, respCode, err := f.client.DownloadBlob(swarm.NewAddress(meta.InodeAddress))
	if err != nil { // skipcq: TCV-001
		return err
	}
	if respCode != http.StatusOK { // skipcq: TCV-001
		f.logger.Warningf("could not remove blocks in %s", swarm.NewAddress(meta.InodeAddress).String())
		return fmt.Errorf("could not remove blocks in %v", swarm.NewAddress(meta.InodeAddress).String())
	}
	defer r.Close()

	fileInodeBytes, err := io.ReadAll(r)
	if err != nil { // skipcq: TCV-001
		f.logger.Warningf("could not read data in address %s", swarm.NewAddress(meta.InodeAddress).String())
		return fmt.Errorf("could not read data in address %v", swarm.NewAddress(meta.InodeAddress).String())
	}
	// find the inode and remove the blocks present in the inode one by one
	var fInode *INode
	err = json.Unmarshal(fileInodeBytes, &fInode)
	if err != nil { // skipcq: TCV-001
		f.logger.Warningf("could not unmarshall data in address %s", swarm.NewAddress(meta.InodeAddress).String())
		return fmt.Errorf("could not unmarshall data in address %v", swarm.NewAddress(meta.InodeAddress).String())
	}

	err = f.client.DeleteReference(swarm.NewAddress(meta.InodeAddress))
	if err != nil {
		f.logger.Errorf("could not delete file inode %s", swarm.NewAddress(meta.InodeAddress).String())
		return fmt.Errorf("could not delete file inode %s: %s", swarm.NewAddress(meta.InodeAddress).String(), err.Error())
	}
	for _, fblocks := range fInode.Blocks {
		err = f.client.DeleteReference(swarm.NewAddress(fblocks.Reference.Bytes()))
		if err != nil { // skipcq: TCV-001
			f.logger.Errorf("could not delete file block %s", swarm.NewAddress(fblocks.Reference.Bytes()).String())
			return fmt.Errorf("could not delete file inode %v", swarm.NewAddress(fblocks.Reference.Bytes()).String())
		}
	}
	// remove the meta
	topic := utils.HashString(totalFilePath)
	err = f.fd.UpdateFeed(f.userAddress, topic, []byte(utils.DeletedFeedMagicWord), []byte(podPassword), false) // empty byte array will fail, so some 1 byte
	if err != nil {                                                                                             // skipcq: TCV-001
		return err
	}
	// remove the file from file map
	f.RemoveFromFileMap(totalFilePath)

	return nil
}
