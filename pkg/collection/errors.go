/*
Copyright © 2020 FairOS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http:// www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package collection

import "errors"

var (
	// ErrEmptyIndex is returned when the index is empty
	ErrEmptyIndex = errors.New("empty Index")
	// ErrEntryNotFound is returned when the entry is not found
	ErrEntryNotFound = errors.New("entry not found")
	// ErrNoNextElement is returned when there is no next element
	ErrNoNextElement = errors.New("no next element")
	// ErrNoManifestFound is returned when there is no manifest found
	ErrNoManifestFound = errors.New("no Manifest found")
	// ErrManifestUnmarshall is returned when the manifest cannot be unmarshalled
	ErrManifestUnmarshall = errors.New("could not unmarshall Manifest")
	// ErrManifestCreate is returned when the manifest cannot be created
	ErrManifestCreate = errors.New("could not create new Manifest")
	// ErrDeleteingIndex is returned when the index cannot be deleted
	ErrDeleteingIndex = errors.New("could not delete index")
	// ErrIndexAlreadyPresent is returned when the index is already present
	ErrIndexAlreadyPresent = errors.New("index already present")
	// ErrIndexNotPresent is returned when the index is not present
	ErrIndexNotPresent = errors.New("index not present")
	// ErrIndexNotSupported is returned when the index is not supported
	ErrIndexNotSupported = errors.New("index not supported")
	// ErrInvalidIndexType is returned when the index type is invalid
	ErrInvalidIndexType = errors.New("invalid index type")
	// ErrKvTableAlreadyPresent is returned when the kv table is already present
	ErrKvTableAlreadyPresent = errors.New("kv table already present")
	// ErrKVTableNotPresent is returned when the kv table is not present
	ErrKVTableNotPresent = errors.New("kv table not present")
	// ErrKVTableNotOpened is returned when the kv table is not opened
	ErrKVTableNotOpened = errors.New("kv table not opened")
	// ErrKVInvalidIndexType is returned when the kv index type is invalid
	ErrKVInvalidIndexType = errors.New("kv invalid index type")
	// ErrKVNilIterator is returned when the kv iterator is nil
	ErrKVNilIterator = errors.New("iterator not set, seek first")
	// ErrKVIndexTypeNotSupported is returned when the kv index type is not supported
	ErrKVIndexTypeNotSupported = errors.New("kv index type not supported yet")
	// ErrKVKeyNotANumber is returned when the kv key is not a number
	ErrKVKeyNotANumber = errors.New("kv key not a number")
	// ErrUnmarshallingDBSchema is returned when the db schema cannot be unmarshalled
	ErrUnmarshallingDBSchema = errors.New("could not unmarshall document db schema")
	// ErrMarshallingDBSchema is returned when the db schema cannot be marshalled
	ErrMarshallingDBSchema = errors.New("could not marshall document db schema")
	// ErrDocumentDBAlreadyPresent is returned when the document db is already present
	ErrDocumentDBAlreadyPresent = errors.New("document db already present")
	// ErrDocumentDBNotPresent is returned when the document db is not present
	ErrDocumentDBNotPresent = errors.New("document db  not present")
	// ErrDocumentDBNotOpened is returned when the document db is not opened
	ErrDocumentDBNotOpened = errors.New("document db not opened")
	// ErrDocumentDBAlreadyOpened is returned when the document db is already opened
	ErrDocumentDBAlreadyOpened = errors.New("document db already opened")
	// ErrDocumentDBIndexFieldNotPresent is returned when the document db index field is not present
	ErrDocumentDBIndexFieldNotPresent = errors.New("document db index field not present")
	// ErrModifyingImmutableDocDB is returned when the document db is immutable
	ErrModifyingImmutableDocDB = errors.New("trying to modify immutable document db")
	// ErrInvalidOperator is returned when the operator is invalid
	ErrInvalidOperator = errors.New("invalid operator")
	// ErrDocumentNotPresent is returned when the document is not present
	ErrDocumentNotPresent = errors.New("document not present")
	// ErrInvalidDocumentId is returned when the document id is invalid
	ErrInvalidDocumentId = errors.New("invalid document id")
	// ErrReadOnlyIndex is returned when the index is read only
	ErrReadOnlyIndex = errors.New("read only index")
	// ErrCannotModifyImmutableIndex is returned when the index is immutable
	ErrCannotModifyImmutableIndex = errors.New("trying to modify immutable index")
	// ErrUnknownJsonFormat is returned when the json format is unknown
	ErrUnknownJsonFormat = errors.New("unknown json format")
)
