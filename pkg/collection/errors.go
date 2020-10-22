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
)
