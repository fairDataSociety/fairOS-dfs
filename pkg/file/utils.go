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

import "github.com/fairdatasociety/fairOS-dfs/pkg/utils"

func (f *File) LoadFileMeta(fileNameWithPath string) error {
	_, meta, err := f.getMetaFromFileName(fileNameWithPath)
	if err != nil {
		return err
	}
	f.AddToFileMap(fileNameWithPath, meta)
	f.logger.Infof(fileNameWithPath)
	return nil
}

func combinePathAndFile(path, fileName string) string {
	var totalPath string
	if path == utils.PathSeperator {
		totalPath = path + fileName
	} else {
		totalPath = path + utils.PathSeperator + fileName
	}
	return totalPath
}