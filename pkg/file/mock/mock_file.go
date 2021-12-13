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
	"io"

	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
)

type MockFile struct {
}

func NewMockFile() *MockFile {
	return &MockFile{}
}

func (*MockFile) Upload(_ io.Reader, _ string, _ int64, _ uint32, _, _ string) error {
	return nil
}

func (*MockFile) Download(_ string) (io.ReadCloser, uint64, error) {
	return nil, 0, nil
}

func (*MockFile) ListFiles(_ []string) ([]file.Entry, error) {
	return nil, nil
}

func (*MockFile) GetStats(_, _ string) (*file.Stats, error) {
	return nil, nil
}

func (*MockFile) RmFile(_ string) error {
	return nil
}

func (*MockFile) Read(_ []byte) (n int, err error) {
	return 0, nil
}

func (*MockFile) AddFileToPath(_, _ string) error {
	return nil
}

func (*MockFile) LoadFileMeta(_ string) error {
	return nil
}
