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
	"encoding/json"
	"fmt"
	"net/http"
	gopath "path"
	"time"

	"github.com/ethersphere/bee/pkg/swarm"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	m "github.com/fairdatasociety/fairOS-dfs/pkg/meta"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (p *Pod) RemoveFile(podName, podFile string) error {
	if !p.isPodOpened(podName) {
		return fmt.Errorf("login to pod to do this operation")
	}

	podInfo, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}
	dir := podInfo.GetDirectory()

	var path string
	if podInfo.IsCurrentDirRoot() {
		path = podInfo.GetCurrentPodPathAndName() + podFile
	} else {
		path = podInfo.GetCurrentDirPathAndName() + utils.PathSeperator + podFile
	}

	if !podInfo.getFile().IsFileAlreadyPResent(path) {
		return fmt.Errorf("file not present in pod")
	}

	_, dirInode, err := dir.GetDirNode(gopath.Dir(path), podInfo.GetFeed(), podInfo.GetAccountInfo())
	if err != nil {
		return err
	}

	// remove the file
	var newHashes [][]byte
	for _, hash := range dirInode.Hashes {
		_, _, err := podInfo.GetFeed().GetFeedData(hash, podInfo.GetAccountInfo().GetAddress())
		if err != nil {
			data, respCode, err := p.GetClient().DownloadBlob(hash)
			if err != nil || respCode != http.StatusOK {
				p.logger.Warningf("could not load address ", swarm.NewAddress(hash).String())
				continue
			}
			var meta *m.FileMetaData
			err = json.Unmarshal(data, &meta)
			if err != nil {
				p.logger.Warningf("could not unmarshall data in address ", swarm.NewAddress(hash).String())
				continue
			}
			if meta.Name != gopath.Base(path) {
				newHashes = append(newHashes, hash)
			} else {
				err = p.client.DeleteBlob(hash)
				if err != nil {
					p.logger.Errorf("could not delete file meta ", swarm.NewAddress(hash).String())
					continue
				}
				fdata, respCode, err := p.GetClient().DownloadBlob(meta.InodeAddress)
				if err != nil || respCode != http.StatusOK {
					p.logger.Warningf("could not load address ", swarm.NewAddress(meta.InodeAddress).String())
					continue
				}
				var fInode *file.FileINode
				err = json.Unmarshal(fdata, &fInode)
				if err != nil {
					p.logger.Warningf("could not unmarshall data in address ", swarm.NewAddress(meta.InodeAddress).String())
					continue
				}
				err = p.client.DeleteBlob(meta.InodeAddress)
				if err != nil {
					p.logger.Errorf("could not delete file inode ", swarm.NewAddress(meta.InodeAddress).String())
					continue
				}
				for _, fblocks := range fInode.FileBlocks {
					err = p.client.DeleteBlob(fblocks.Address)
					if err != nil {
						p.logger.Errorf("could not delete file block ", swarm.NewAddress(fblocks.Address).String())
						continue
					}
				}

				podInfo.getFile().RemoveFromFileMap(path)
			}
		}
	}
	dirInode.Hashes = newHashes

	dirInode.Meta.ModificationTime = time.Now().Unix()
	topic, err := dir.UpdateDirectory(dirInode)
	if err != nil {
		return err
	}

	if path != podInfo.GetCurrentPodPathAndName() {
		err = p.UpdateTillThePod(podName, podInfo.GetDirectory(), topic, gopath.Dir(path), true)
		if err != nil {
			return err
		}
	}
	return nil

}
