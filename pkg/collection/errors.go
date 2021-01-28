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
	ErrEmptyIndex                     = errors.New("empty Index")
	ErrEntryNotFound                  = errors.New("entry not found")
	ErrNoNextElement                  = errors.New("no next element")
	ErrNoManifestFound                = errors.New("no manifest found")
	ErrManifestUnmarshall             = errors.New("could not unmarshall manifest")
	ErrManifestCreate                 = errors.New("could not create new manifest")
	ErrDeleteingIndex                 = errors.New("could not delete index")
	ErrIndexAlreadyPresent            = errors.New("index already present")
	ErrIndexNotPresent                = errors.New("index not present")
	ErrIndexNotSupported              = errors.New("index not supported")
	ErrInvalidIndexType               = errors.New("invalid index type")
	ErrKvTableAlreadyPresent          = errors.New("kv table already present")
	ErrKVTableNotPresent              = errors.New("kv table not present")
	ErrKVTableNotOpened               = errors.New("kv table not opened")
	ErrKVInvalidIndexType             = errors.New("kv invalid index type")
	ErrKVIndexTypeNotSupported        = errors.New("kv index type not supported yet")
	ErrKVKeyNotANumber                = errors.New("kv key not a number")
	ErrUnmarshallingDBSchema          = errors.New("could not unmarshall document db schema")
	ErrMarshallingDBSchema            = errors.New("could not marshall document db schema")
	ErrDocumentDBAlreadyPresent       = errors.New("document db already present")
	ErrDocumentDBNotPresent           = errors.New("document db  not present")
	ErrDocumentDBNotOpened            = errors.New("document db not opened")
	ErrDocumentDBAlreadyOpened        = errors.New("document db already opened")
	ErrDocumentDBIndexFieldNotPresent = errors.New("document db index field not present")
	ErrInvalidOperator                = errors.New("invalid operator")
	ErrDocumentNotPresent             = errors.New("document not present")
	ErrInvalidDocumentId              = errors.New("invalid document id")
	ErrReadOnlyIndex                  = errors.New("read only index")
)
