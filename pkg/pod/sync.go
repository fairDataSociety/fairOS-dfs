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
	"strings"
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/file"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	d "github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (p *Pod) SyncPod(podName string) error {
	podName, err := CleanPodName(podName)
	if err != nil {
		return err
	}

	if !p.IsPodOpened(podName) {
		return ErrPodNotOpened
	}

	podInfo, err := p.GetPodInfoFromPodMap(podName)
	if err != nil {
		return err
	}

	err = podInfo.SyncPod(podName, "", p.client, p.logger)
	if err != nil {
		return err
	}
	return nil
}

func (pi *Info) SyncPod(podName string, path string, client blockstore.Client, logger logging.Logger) error {
	fd := pi.GetFeed()
	accountInfo := pi.GetAccountInfo()

	logger.Infof("Syncing pod: %v", podName)
	var wg sync.WaitGroup
	for _, name := range pi.GetCurrentPodInode().GetFileOrDirNames() {
		wg.Add(1)
		go func(fileOrDirName string) {
			defer wg.Done()
			if strings.HasPrefix(fileOrDirName, "_D_") {
				dirName := strings.TrimLeft(fileOrDirName, "_D_")
				dirPath := path + utils.PathSeperator + dirName
				dirTopic := utils.HashString(dirPath)
				_, data, err := fd.GetFeedData(dirTopic, accountInfo.GetAddress())
				if err != nil {
					logger.Warningf("sync: : %w", err)
					return
				}

				var dirInode *d.Inode
				err = json.Unmarshal(data, &dirInode)
				if err != nil {
					logger.Errorf("sync: unmarshall error: %w", err)
					return
				}

				pi.GetDirectory().AddToDirectoryMap(path, dirInode)
				path = strings.TrimPrefix(path, podName)
				logger.Infof(path)

			} else if strings.HasPrefix(fileOrDirName, "_F_") {
				fileName := strings.TrimLeft(fileOrDirName, "_F_")
				filePath := path + utils.PathSeperator + fileName
				fileTopic := utils.HashString(filePath)
				_, data, err := fd.GetFeedData(fileTopic, accountInfo.GetAddress())
				var meta *file.MetaData
				err = json.Unmarshal(data, &meta)
				if err != nil {
					logger.Errorf("sync: unmarshall error: %w", err)
					return
				}
				pi.file.AddToFileMap(path, meta)
				path = strings.TrimPrefix(path, podName)
				logger.Infof(path)
			}
		}(name)
	}
	wg.Wait()
	return nil
}
