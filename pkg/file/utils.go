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
	"strings"

	m "github.com/fairdatasociety/fairOS-dfs/pkg/meta"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (f *File) LoadFileMeta(podName string, addr []byte) (int, error) {
	data, respCode, err := f.getClient().DownloadBlob(addr)
	if err != nil {
		return respCode, fmt.Errorf("not a file")
	}
	var meta *m.FileMetaData
	err = json.Unmarshal(data, &meta)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	meta.MetaReference = addr
	f.AddToFileMap(meta.Path+utils.PathSeperator+meta.Name, meta)
	fileName := strings.TrimPrefix(meta.Path+utils.PathSeperator+meta.Name, podName)
	f.logger.Infof(fileName)
	return http.StatusOK, nil
}
