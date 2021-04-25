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
	"strconv"
	"time"
)

func (f *File) Download(podFile string) (io.ReadCloser, string, error) {
	meta := f.GetFromFileMap(podFile)
	if meta == nil {
		return nil, "", fmt.Errorf("file not found in dfs")
	}

	fileInodeBytes, _, err := f.getClient().DownloadBlob(meta.InodeAddress)
	if err != nil {
		return nil, "", err
	}
	var fileInode INode
	err = json.Unmarshal(fileInodeBytes, &fileInode)
	if err != nil {
		return nil, "", err
	}

	//need to change the access time for podFile
	meta.AccessTime = time.Now().Unix()
	err = f.uploadMeta(meta)
	if err != nil {
		return nil, "", err
	}

	reader := NewReader(fileInode, f.getClient(), meta.Size, meta.BlockSize, meta.Compression, false)
	size := strconv.FormatUint(meta.Size, 10)
	return reader, size, nil
}
