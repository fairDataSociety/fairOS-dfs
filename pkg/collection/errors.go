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

import "errors"

var (
	ErrEmptyIndex          = errors.New("empty Index")
	ErrEntryNotFound       = errors.New("entry not found")
	ErrNoNextElement       = errors.New("no next element")
	ErrNoManifestFound     = errors.New("no manifest found")
	ErrManifestUnmarshall  = errors.New("could not unmarshall manifest")
	ErrManifestCreate      = errors.New("could not create new manifest")
	ErrDeleteingIndex      = errors.New("could not delete index")
	ErrIndexAlreadyPresent = errors.New("index already present")
	ErrIndexNotPresent     = errors.New("index not present")
)
