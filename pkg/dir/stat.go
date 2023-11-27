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
)

// Stats represents a given directory
type Stats struct {
	PodName          string `json:"podName"`
	DirPath          string `json:"dirPath"`
	DirName          string `json:"dirName"`
	Mode             uint32 `json:"mode"`
	CreationTime     string `json:"creationTime"`
	ModificationTime string `json:"modificationTime"`
	AccessTime       string `json:"accessTime"`
	NoOfDirectories  string `json:"noOfDirectories"`
	NoOfFiles        string `json:"noOfFiles"`
}

// DirStat returns all the information related to a given directory.
func (d *Directory) DirStat(podName, podPassword, dirNameWithPath string) (*Stats, error) {
	dirInode, err := d.GetInode(podPassword, dirNameWithPath)
	if err != nil { // skipcq: TCV-001
		d.logger.Errorf("dir stat : %v", err)
		return nil, err
	}

	if dirInode.Meta == nil && dirInode.FileOrDirNames == nil { // skipcq: TCV-001
		return nil, ErrDirectoryNotPresent
	}

	files := 0
	dirs := 0
	for _, k := range dirInode.FileOrDirNames {
		if strings.HasPrefix(k, "_D_") {
			dirs++
		} else if strings.HasPrefix(k, "_F_") {
			files++
		}
	}

	meta := dirInode.Meta
	return &Stats{
		PodName:          podName,
		DirPath:          meta.Path,
		DirName:          meta.Name,
		Mode:             meta.Mode,
		CreationTime:     strconv.FormatInt(meta.CreationTime, 10),
		ModificationTime: strconv.FormatInt(meta.ModificationTime, 10),
		AccessTime:       strconv.FormatInt(meta.AccessTime, 10),
		NoOfDirectories:  strconv.FormatInt(int64(dirs), 10),
		NoOfFiles:        strconv.FormatInt(int64(files), 10),
	}, nil
}
