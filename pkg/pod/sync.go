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
	"net/http"
	"strings"
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	m "github.com/fairdatasociety/fairOS-dfs/pkg/meta"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (p *Pod) SyncPod(podName string) error {
	podName, err := CleanPodName(podName)
	if err != nil {
		return err
	}

	if !p.isPodOpened(podName) {
		return ErrPodNotOpened
	}

	podInfo, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}

	err = podInfo.SyncPod(podName, p.client, p.logger)
	if err != nil {
		return err
	}
	return nil
}

func (pi *Info) SyncPod(podName string, client blockstore.Client, logger logging.Logger) error {
	fd := pi.GetFeed()
	accountInfo := pi.GetAccountInfo()

	logger.Infof("Syncing pod: %v", podName)
	var wg sync.WaitGroup
	for _, ref := range pi.currentPodInode.Hashes {
		wg.Add(1)
		go func(reference []byte) {
			defer wg.Done()
			_, data, err := fd.GetFeedData(reference, accountInfo.GetAddress())
			if err != nil {
				data, respCode, err := client.DownloadBlob(reference)
				if err != nil {
					logger.Warningf("sync: download error: ", err)
					return
				}
				if respCode != http.StatusOK {
					logger.Warningf("sync: download status not okay: ", respCode)
					return
				}
				var meta *m.FileMetaData
				err = json.Unmarshal(data, &meta)
				if err != nil {
					logger.Errorf("sync: unmarshall error: ", err)
					return
				}

				path := meta.Path + utils.PathSeperator + meta.Name
				meta.MetaReference = reference
				pi.file.AddToFileMap(path, meta)
				path = strings.TrimPrefix(path, podName)
				logger.Infof(path)
				return
			}

			var dirInode *d.DirInode
			err = json.Unmarshal(data, &dirInode)
			if err != nil {
				logger.Warningf("sync: unmarshall error: %w", err)
				return
			}

			path := dirInode.Meta.Path + utils.PathSeperator + dirInode.Meta.Name
			err = pi.GetDirectory().LoadDirMeta(podName, dirInode, fd, accountInfo)
			if err != nil {
				logger.Warningf("sync: load meta error: %w", err)
				return
			}
			pi.GetDirectory().AddToDirectoryMap(path, dirInode)
			path = strings.TrimPrefix(path, podName)
			logger.Infof(path)
		}(ref)
	}
	wg.Wait()
	return nil
}
