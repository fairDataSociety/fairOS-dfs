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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func (f *File) Cat(fileName string) error {
	//TODO: need to change the access time
	meta := f.GetFromFileMap(fileName)
	if meta == nil {
		return fmt.Errorf("file not found")
	}

	fileInodeBytes, _, err := f.getClient().DownloadBlob(meta.InodeAddress)
	if err != nil {
		return err
	}
	var fileInode FileINode
	err = json.Unmarshal(fileInodeBytes, &fileInode)
	if err != nil {
		return err
	}

	totalBytes := uint32(0)
	for _, fb := range fileInode.FileBlocks {
		stdoutBytes, _, err := f.getClient().DownloadBlob(fb.Address)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("could not find file block")
		}

		if uint32(len(stdoutBytes)) != fb.Size {
			return fmt.Errorf("received less bytes than expected in a block")
		}

		buf := bytes.NewBuffer(stdoutBytes)
		n, err := io.Copy(os.Stdout, buf)
		if err != nil || uint32(n) != fb.Size {
			return fmt.Errorf("could not write to stdout")
		}
		totalBytes += fb.Size
	}
	return nil
}
