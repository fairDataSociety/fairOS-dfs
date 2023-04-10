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

package collection

// Manifest
type Manifest struct {
	Name         string    `json:"name"`
	Mutable      bool      `json:"mutable"`
	PodFile      string    `json:"pod_file,omitempty"`
	IdxType      IndexType `json:"index_type"`
	CreationTime int64     `json:"creation_time"`
	Entries      []*Entry  `json:"entries,omitempty"`
	dirtyFlag    bool
}

// Entry
type Entry struct {
	Name     string    `json:"name"`
	EType    string    `json:"type"`
	Ref      [][]byte  `json:"ref,omitempty"`
	Manifest *Manifest `json:"Manifest,omitempty"`
}

// NewManifest creates a new manifest
func NewManifest(name string, time int64, idxType IndexType, mutable bool) *Manifest {
	var entries []*Entry
	return &Manifest{
		Name:         name,
		Mutable:      mutable,
		IdxType:      idxType,
		CreationTime: time,
		Entries:      entries,
		dirtyFlag:    true,
	}
}
