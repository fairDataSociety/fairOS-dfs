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
	"context"
	"strings"
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// SyncDirectory syncs all the latest entries under a given directory.
func (d *Directory) SyncDirectory(dirNameWithPath, podPassword string) error {
	dirInode, err := d.GetInode(podPassword, dirNameWithPath)
	if err != nil { // skipcq: TCV-001
		return nil // pod is empty
	}

	d.AddToDirectoryMap(dirNameWithPath, dirInode)
	for _, fileOrDirName := range dirInode.FileOrDirNames {
		if strings.HasPrefix(fileOrDirName, "_F_") {
			fileName := strings.TrimPrefix(fileOrDirName, "_F_")
			filePath := utils.CombinePathAndFile(dirNameWithPath, fileName)
			err := d.file.LoadFileMeta(filePath, podPassword)
			if err != nil { // skipcq: TCV-001
				d.logger.Errorf("loading metadata failed %s: %s", filePath, err.Error())
			}
		} else if strings.HasPrefix(fileOrDirName, "_D_") {
			dirName := strings.TrimPrefix(fileOrDirName, "_D_")
			path := utils.CombinePathAndFile(dirNameWithPath, dirName)
			d.logger.Infof(dirNameWithPath)

			err = d.SyncDirectory(path, podPassword)
			if err != nil { // skipcq: TCV-001
				return err
			}
		}
	}
	return nil
}

// SyncDirectoryAsync syncs all the latest entries under a given directory concurrently.
func (d *Directory) SyncDirectoryAsync(ctx context.Context, dirNameWithPath, podPassword string, wg *sync.WaitGroup) error {
	dirInode, err := d.GetInode(podPassword, dirNameWithPath)
	if err != nil { // skipcq: TCV-001
		return nil // pod is empty
	}

	for _, fileOrDirName := range dirInode.FileOrDirNames {
		if strings.HasPrefix(fileOrDirName, "_F_") {
			wg.Add(1)
			fileName := strings.TrimPrefix(fileOrDirName, "_F_")
			filePath := utils.CombinePathAndFile(dirNameWithPath, fileName)
			syncTask := newSyncTask(d, filePath, podPassword, wg)
			_, err = d.syncManager.Go(syncTask)
			if err != nil { // skipcq: TCV-001
				wg.Done()
			}
		} else if strings.HasPrefix(fileOrDirName, "_D_") {
			dirName := strings.TrimPrefix(fileOrDirName, "_D_")
			path := utils.CombinePathAndFile(dirNameWithPath, dirName)
			d.logger.Infof(dirNameWithPath)

			err = d.SyncDirectoryAsync(ctx, path, podPassword, wg)
			if err != nil { // skipcq: TCV-001
				return err
			}
		}
	}
	return nil
}
