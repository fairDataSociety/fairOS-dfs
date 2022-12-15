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

package dir

var (
	//MetaVersion
	MetaVersion uint8 = 2
)

// MetaData
type MetaData struct {
	Version          uint8  `json:"version"`
	Path             string `json:"path"`
	Name             string `json:"name"`
	CreationTime     int64  `json:"creationTime"`
	AccessTime       int64  `json:"accessTime"`
	ModificationTime int64  `json:"modificationTime"`
}
