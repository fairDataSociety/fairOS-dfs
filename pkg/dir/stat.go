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
	"strconv"
	"strings"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

type DirStats struct {
	Account          string `json:"account"`
	PodAddress       string `json:"pod_address"`
	PodName          string `json:"pod_name"`
	DirPath          string `json:"dir_path"`
	DirName          string `json:"dir_name"`
	CreationTime     string `json:"creation_time"`
	ModificationTime string `json:"modification_time"`
	AccessTime       string `json:"access_time"`
	NoOfDirectories  string `json:"no_of_directories"`
	NoOfFiles        string `json:"no_of_files"`
}

func (d *Directory) DirStat(podName, dirName string, dirInode *DirInode, account, podAddr string, printNames bool) (*DirStats, error) {
	meta := dirInode.Meta
	fl, dl := d.ListDirOnlyNames(podName, dirName, printNames)

	files := 0
	dirs := 0
	for _, list := range dl {
		if strings.HasPrefix(list, "<Dir>") {
			dirs++
		} else {
			files++
		}
	}
	path := meta.Path
	if meta.Path == podName {
		path = utils.PathSeperator
	}

	return &DirStats{
		Account:          account,
		PodAddress:       podAddr,
		PodName:          podName,
		DirPath:          path,
		DirName:          meta.Name,
		CreationTime:     strconv.FormatInt(meta.CreationTime, 10),
		ModificationTime: strconv.FormatInt(meta.ModificationTime, 10),
		AccessTime:       strconv.FormatInt(meta.AccessTime, 10),
		NoOfDirectories:  strconv.FormatInt(int64(len(dl)), 10),
		NoOfFiles:        strconv.FormatInt(int64(len(fl)), 10),
	}, nil
}
