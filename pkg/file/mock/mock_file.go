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

// File is the mock file
type File struct {
}

// NewMockFile is used for tests only
func NewMockFile() *File {
	return &File{}
}

// Upload is used for tests only
func (*File) Upload(_ io.Reader, _ string, _ int64, _, _ uint32, _, _, _ string) error {
	return nil
}

// Download is used for tests only
func (*File) Download(_, _ string) (io.ReadCloser, uint64, error) {
	return nil, 0, nil
}

// ListFiles is used for tests only
func (*File) ListFiles(_ []string, _ string) ([]file.Entry, error) {
	return nil, nil
}

// GetStats is used for tests only
func (*File) GetStats(_, _, _ string) (*file.Stats, error) {
	return nil, nil
}

// RmFile is used for tests only
func (*File) RmFile(_, _ string) error {
	return nil
}

// Read is used for tests only
func (*File) Read(_ []byte) (n int, err error) {
	return 0, nil
}

// AddFileToPath is used for tests only
func (*File) AddFileToPath(_, _ string) error {
	return nil
}

// LoadFileMeta is used for tests only
func (*File) LoadFileMeta(_, _ string) error {
	return nil
}
