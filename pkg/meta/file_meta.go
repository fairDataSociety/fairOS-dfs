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

package datapod

var (
	FileMetaVersion uint8 = 1
)

type FileMetaData struct {
	Version          uint8
	Path             string
	Name             string
	FileSize         uint64
	BlockSize        uint32
	ContentType      string
	Compression      string
	CreationTime     int64
	AccessTime       int64
	ModificationTime int64
	MetaReference    []byte
	InodeAddress     []byte
}
