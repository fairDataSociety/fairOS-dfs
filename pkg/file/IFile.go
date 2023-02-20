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

import "io"

// IFile
type IFile interface {
	Upload(fd io.Reader, podFileName string, fileSize int64, blockSize uint32, podPath, compression, podPassword string) error
	Download(podFileWithPath, podPassword string) (io.ReadCloser, uint64, error)
	ListFiles(files []string, podPassword string) ([]Entry, error)
	GetStats(podName, podFileWithPath, podPassword string) (*Stats, error)
	RmFile(podFileWithPath, podPassword string) error
	LoadFileMeta(fileNameWithPath, podPassword string) error
}
