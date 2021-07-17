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

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// GetFileReference given a file name this function extracts the current
// metadata from the Swarm network.
func (f *File) GetFileReference(podFile string) (*MetaData, error) {
	// Get the meta of the file to share
	meta := f.GetFromFileMap(podFile)
	if meta == nil {
		return nil, fmt.Errorf("file not found in dfs")
	}

	totalPodFile := utils.CombinePathAndFile(f.podName, podFile, "")
	meta, err := f.GetMetaFromFileName(totalPodFile, f.userAddress)
	if err != nil {
		return nil, fmt.Errorf("file not found in dfs")
	}

	return meta, nil
}

// AddFileToPath adds a given files meta data to the main file data structure.
func (f *File) AddFileToPath(filePath, metaHexRef string) error {
	metaReferenace, err := utils.ParseHexReference(metaHexRef)
	if err != nil {
		return err
	}

	data, respCode, err := f.getClient().DownloadBlob(metaReferenace.Bytes())
	if err != nil || respCode != http.StatusOK {
		return err
	}
	meta := &MetaData{}
	err = json.Unmarshal(data, meta)
	if err != nil {
		return err
	}
	f.AddToFileMap(filePath, meta)
	return nil
}
