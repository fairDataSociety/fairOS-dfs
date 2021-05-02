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

package mock

import (
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"io"
	"sync"
)

type MockFile struct {
	fileMap     map[string]string
	fileMu      *sync.RWMutex
}

func NewMockFile() *MockFile {
	return &MockFile{
	}
}

func (mf *MockFile) Upload(fd io.Reader, podFileName string, fileSize int64, blockSize uint32, podPath, compression string) error {
	return nil
}

func (mf *MockFile) Download(podFileWithPath string) (io.ReadCloser, uint64, error) {
	return nil, 0, nil
}

func (mf *MockFile) ListFiles(files []string) ([]file.Entry, error) {
	return nil, nil
}

func (mf *MockFile) GetStats(podName, podFileWithPath string) (*file.Stats, error) {
	return nil, nil
}

func (mf *MockFile) RmFile(podFileWithPath string) error {
	return nil
}

func (mf *MockFile) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (mf *MockFile) GetFileReference(podFile string) ([]byte, string, error) {
	return nil, "", nil
}

func (mf *MockFile) AddFileToPath(filePath, metaHexRef string) error {
	return nil
}

func (mf *MockFile) LoadFileMeta(fileNameWithPath string) error {
	return nil
}