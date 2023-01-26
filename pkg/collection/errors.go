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
	//ErrEmptyIndex
	ErrEmptyIndex = errors.New("empty Index")
	//ErrEntryNotFound
	ErrEntryNotFound = errors.New("entry not found")
	//ErrNoNextElement
	ErrNoNextElement = errors.New("no next element")
	//ErrNoManifestFound
	ErrNoManifestFound = errors.New("no Manifest found")
	//ErrManifestUnmarshall
	ErrManifestUnmarshall = errors.New("could not unmarshall Manifest")
	//ErrManifestCreate
	ErrManifestCreate = errors.New("could not create new Manifest")
	//ErrDeleteingIndex
	ErrDeleteingIndex = errors.New("could not delete index")
	//ErrIndexAlreadyPresent
	ErrIndexAlreadyPresent = errors.New("index already present")
	//ErrIndexNotPresent
	ErrIndexNotPresent = errors.New("index not present")
	//ErrIndexNotSupported
	ErrIndexNotSupported = errors.New("index not supported")
	//ErrInvalidIndexType
	ErrInvalidIndexType = errors.New("invalid index type")
	//ErrKvTableAlreadyPresent
	ErrKvTableAlreadyPresent = errors.New("kv table already present")
	//ErrKVTableNotPresent
	ErrKVTableNotPresent = errors.New("kv table not present")
	//ErrKVTableNotOpened
	ErrKVTableNotOpened = errors.New("kv table not opened")
	//ErrKVInvalidIndexType
	ErrKVInvalidIndexType = errors.New("kv invalid index type")
	//ErrKVNilIterator
	ErrKVNilIterator = errors.New("iterator not set, seek first")
	//ErrKVIndexTypeNotSupported
	ErrKVIndexTypeNotSupported = errors.New("kv index type not supported yet")
	//ErrKVKeyNotANumber
	ErrKVKeyNotANumber = errors.New("kv key not a number")
	//ErrUnmarshallingDBSchema
	ErrUnmarshallingDBSchema = errors.New("could not unmarshall document db schema")
	//ErrMarshallingDBSchema
	ErrMarshallingDBSchema = errors.New("could not marshall document db schema")
	//ErrDocumentDBAlreadyPresent
	ErrDocumentDBAlreadyPresent = errors.New("document db already present")
	//ErrDocumentDBNotPresent
	ErrDocumentDBNotPresent = errors.New("document db  not present")
	//ErrDocumentDBNotOpened
	ErrDocumentDBNotOpened = errors.New("document db not opened")
	//ErrDocumentDBAlreadyOpened
	ErrDocumentDBAlreadyOpened = errors.New("document db already opened")
	//ErrDocumentDBIndexFieldNotPresent
	ErrDocumentDBIndexFieldNotPresent = errors.New("document db index field not present")
	//ErrModifyingImmutableDocDB
	ErrModifyingImmutableDocDB = errors.New("trying to modify immutable document db")
	//ErrInvalidOperator
	ErrInvalidOperator = errors.New("invalid operator")
	//ErrDocumentNotPresent
	ErrDocumentNotPresent = errors.New("document not present")
	//ErrInvalidDocumentId
	ErrInvalidDocumentId = errors.New("invalid document id")
	//ErrReadOnlyIndex
	ErrReadOnlyIndex = errors.New("read only index")
	//ErrCannotModifyImmutableIndex
	ErrCannotModifyImmutableIndex = errors.New("trying to modify immutable index")
	//ErrCouldNotUpdatePostageBatch
	ErrCouldNotUpdatePostageBatch = errors.New("could not procure new postage batch")
	//ErrUnknownJsonFormat
	ErrUnknownJsonFormat = errors.New("unknown json format")
)
