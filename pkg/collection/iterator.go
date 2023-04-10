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

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Iterator
type Iterator struct {
	index         *Index
	indexType     IndexType
	startPrefix   string
	endPrefix     string
	limit         int64
	givenUntilNow int64
	currentKey    string
	currentValue  [][]byte
	currentDigits int
	manifestStack []*ManifestState
	error         error
}

// ManifestState
type ManifestState struct {
	currentManifest *Manifest
	currentIndex    int
}

// NewStringIterator creates a new iterator object which is used to create new index iterators.
func (idx *Index) NewStringIterator(start, end string, limit int64) (*Iterator, error) {
	var manifest *Manifest
	if idx.mutable {
		// get the first feed of the Index
		mf, err := idx.loadManifest(idx.name, idx.encryptionPassword)
		if err != nil { // skipcq: TCV-001
			return nil, ErrEmptyIndex
		}
		manifest = mf
	} else {
		manifest = idx.memDB
	}
	itr := &Iterator{
		index:         idx,
		startPrefix:   start,
		endPrefix:     end,
		limit:         limit,
		givenUntilNow: 0,
		currentKey:    "",
		currentValue:  nil,
		currentDigits: 1,
		error:         nil,
	}

	if itr.startPrefix != "" {
		err := itr.Seek(itr.startPrefix)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
	} else {
		firstManifest := &ManifestState{
			currentManifest: manifest,
			currentIndex:    0,
		}
		var stack []*ManifestState
		stack = append(stack, firstManifest)
		itr.manifestStack = stack
	}
	return itr, nil
}

// NewIntIterator creates a new index iterator with start prefix, endPrefix and the limit to iterate.
func (idx *Index) NewIntIterator(start, end, limit int64) (*Iterator, error) {
	var manifest *Manifest
	if idx.mutable {
		// get the first feed of the Index
		mf, err := idx.loadManifest(idx.name, idx.encryptionPassword)
		if err != nil { // skipcq: TCV-001
			return nil, ErrEmptyIndex
		}
		manifest = mf
	} else { // skipcq: TCV-001
		manifest = idx.memDB
	}

	startPrefix := fmt.Sprintf("%020g", float64(start))
	endPrefix := fmt.Sprintf("%020g", float64(end))
	if start == -1 {
		startPrefix = ""
	}
	if end == -1 {
		endPrefix = ""
	}

	itr := &Iterator{
		index:         idx,
		startPrefix:   startPrefix,
		endPrefix:     endPrefix,
		limit:         limit,
		givenUntilNow: 0,
		currentKey:    "",
		currentValue:  nil,
		currentDigits: 1,
		error:         nil,
	}

	if itr.startPrefix != "" {
		err := itr.Seek(itr.startPrefix)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
	} else {
		firstManifest := &ManifestState{
			currentManifest: manifest,
			currentIndex:    0,
		}
		var stack []*ManifestState
		stack = append(stack, firstManifest)
		itr.manifestStack = stack
	}
	return itr, nil
}

// Seek seeks to the given key prefix.
func (itr *Iterator) Seek(key string) error {
	var manifest *Manifest
	if itr.index.mutable {
		mf, err := itr.index.loadManifest(itr.index.name, itr.index.encryptionPassword)
		if err != nil { // skipcq: TCV-001
			return err
		}
		manifest = mf
	} else { // skipcq: TCV-001
		manifest = itr.index.memDB
	}

	// Set the index type here from the Manifest
	itr.indexType = manifest.IdxType

	err := itr.seekStringKey(manifest, key)
	if err != nil { // skipcq: TCV-001
		return err
	}

	return nil
}

// Next moves the seek pointer one step ahead.
func (itr *Iterator) Next() bool {
	return itr.nextStringKey()
}

// StringKey
func (itr *Iterator) StringKey() string {
	return itr.currentKey
}

// IntegerKey
func (itr *Iterator) IntegerKey() int64 {
	gotKey, err := strconv.ParseInt(itr.currentKey, 10, 64)
	if err != nil {
		return -1
	}
	return gotKey
}

// Value
func (itr *Iterator) Value() []byte {
	return itr.currentValue[0]
}

// ValueAll
func (itr *Iterator) ValueAll() [][]byte {
	return itr.currentValue
}

// Non-API functions
func (itr *Iterator) seekStringKey(manifest *Manifest, key string) error {
	// if there are any elements in the index, then search for the entry
	if len(manifest.Entries) > 0 {
		for i, entry := range manifest.Entries {

			// even if the entry is not found, add the pointer to seek so that
			// seek can continue from the next element
			if len(entry.Name) > 0 {
				if key == "" || entry.Name[0] > key[0] {
					manifestState := &ManifestState{
						currentManifest: manifest,
						currentIndex:    i,
					}
					itr.manifestStack = append(itr.manifestStack, manifestState)
					return nil
				}

				if entry.EType == leafEntry && entry.Name > key {
					manifestState := &ManifestState{
						currentManifest: manifest,
						currentIndex:    i,
					}
					itr.manifestStack = append(itr.manifestStack, manifestState)
					return nil
				}
			}

			if entry.EType == leafEntry && entry.Name == key {
				manifestState := &ManifestState{
					currentManifest: manifest,
					currentIndex:    i,
				}
				itr.manifestStack = append(itr.manifestStack, manifestState)
				return nil
			}

			if entry.EType == intermediateEntry && strings.HasPrefix(key, entry.Name) {
				// found a branch, push the current Manifest state
				manifestState := &ManifestState{
					currentManifest: manifest,
					currentIndex:    i + 1,
				}
				itr.manifestStack = append(itr.manifestStack, manifestState)
				var childManifest *Manifest
				if itr.index.mutable || entry.Manifest == nil {
					// now load the child Manifest and re-seek
					cf, err := itr.index.loadManifest(manifest.Name+entry.Name, itr.index.encryptionPassword)
					if err != nil { // skipcq: TCV-001
						return err
					}
					childManifest = cf
				} else { // skipcq: TCV-001
					childManifest = entry.Manifest
				}

				childKey := strings.TrimPrefix(key, entry.Name)
				err := itr.seekStringKey(childManifest, childKey)
				if err != nil { // skipcq: TCV-001
					if errors.Is(err, ErrEntryNotFound) {
						return nil
					}
				}
				return err
			}

			if entry.EType == intermediateEntry && (len(entry.Name) < len(key)) {
				reducedKey := key[:len(entry.Name)]
				for kk := 0; kk < len(entry.Name); kk++ {
					if reducedKey[kk] == entry.Name[kk] { // skipcq: TCV-001
						continue
					} else if reducedKey[kk] > entry.Name[kk] {
						break
					} else if reducedKey[kk] < entry.Name[kk] { // skipcq: TCV-001
						manifestState := &ManifestState{
							currentManifest: manifest,
							currentIndex:    i + 1,
						}
						itr.manifestStack = append(itr.manifestStack, manifestState)

						var childManifest *Manifest
						if itr.index.mutable {
							// now load the child Manifest and re-seek
							cf, err := itr.index.loadManifest(manifest.Name+entry.Name, itr.index.encryptionPassword)
							if err != nil {
								return err
							}
							childManifest = cf
						} else {
							childManifest = entry.Manifest
						}

						childKey := key[len(reducedKey):]
						err := itr.seekStringKey(childManifest, childKey)
						if err != nil {
							if errors.Is(err, ErrEntryNotFound) {
								return nil
							}
						}
						return err
					}
				}
			}
		}
	}
	return ErrEntryNotFound // skipcq: TCV-001
}

func (itr *Iterator) nextStringKey() bool {
	// don't go beyond the limit
	if itr.limit >= 0 {
		if itr.givenUntilNow >= itr.limit {
			return false
		}
	}

	// get the current Manifest at the top of the stack
	depthOfStack := len(itr.manifestStack)
	if depthOfStack == 0 { // skipcq: TCV-001
		itr.error = ErrNoNextElement
		return false
	}

	// take the top Manifest to find the next entry
	manifestState := itr.manifestStack[depthOfStack-1]

	entriesExhausted := true
	for entriesExhausted {
		if manifestState.currentManifest == nil {
			return false
		}
		// see if we have exhausted the entries in the current Manifest
		if manifestState.currentIndex >= len(manifestState.currentManifest.Entries) {
			// pop the exhausted Manifest from the top and pick the next Manifest to find the entry
			n := depthOfStack - 1
			if n == 0 {
				itr.error = ErrNoNextElement
				return false
			}
			manifestState = itr.manifestStack[n-1]
			itr.manifestStack[n] = nil
			itr.manifestStack = itr.manifestStack[:n]
			depthOfStack = n
		} else {
			entriesExhausted = false
		}
	}

	// We have a Manifest whose entries are not yet exhausted,
	// so get the next entry and check for valid conditions of the Iterator()
	entry := manifestState.currentManifest.Entries[manifestState.currentIndex]
	manifestState.currentIndex++

	// check if the search has reached the end key
	if itr.endPrefix != "" {
		actualKey := manifestState.currentManifest.Name + entry.Name
		actualKey = strings.TrimPrefix(actualKey, itr.index.name)
		if actualKey > itr.endPrefix || actualKey == itr.endPrefix {
			return false
		}
	}

	// if it is a leaf entry, set the key and value
	if entry.EType == leafEntry {
		actualKey := manifestState.currentManifest.Name + entry.Name
		actualKey = strings.TrimPrefix(actualKey, itr.index.name)
		itr.currentKey = actualKey
		itr.currentValue = entry.Ref
		itr.givenUntilNow++
		return true
	}

	// if it is an intermediate entry, get the branch Manifest and push in to the stack
	if entry.EType == intermediateEntry {
		var newManifest *Manifest
		if itr.index.mutable {
			mf, err := itr.index.loadManifest(manifestState.currentManifest.Name+entry.Name, itr.index.encryptionPassword)
			if err != nil { // skipcq: TCV-001
				itr.error = err
				return false
			}
			newManifest = mf
		} else {
			newManifest = entry.Manifest
		}

		newManifestState := &ManifestState{
			currentManifest: newManifest,
			currentIndex:    0,
		}
		itr.manifestStack = append(itr.manifestStack, newManifestState)
		return itr.nextStringKey()
	}
	return false // skipcq: TCV-001
}
