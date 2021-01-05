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
	"fmt"
	"strconv"
	"strings"
)

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

type ManifestState struct {
	currentManifest *Manifest
	currentIndex    int
}

func (idx *Index) NewStringIterator(start, end string, limit int64) (*Iterator, error) {
	// get the first feed of the Index
	manifest, err := idx.loadManifest(idx.name)
	if err != nil {
		return nil, ErrEmptyIndex
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
		if err != nil {
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

func (idx *Index) NewIntIterator(start, end, limit int64) (*Iterator, error) {
	// get the first feed of the Index
	manifest, err := idx.loadManifest(idx.name)
	if err != nil {
		return nil, ErrEmptyIndex
	}

	startPrefix := fmt.Sprintf("%020d", start)
	endPrefix := fmt.Sprintf("%020d", end)
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
		if err != nil {
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

func (itr *Iterator) Seek(key string) error {
	manifest, err := itr.index.loadManifest(itr.index.name)
	if err != nil {
		return err
	}

	// Set the index type here from the manifest
	itr.indexType = manifest.IdxType

	err = itr.seekStringKey(manifest, key)
	if err != nil {
		return err
	}

	return nil
}

func (itr *Iterator) Next() bool {
	return itr.nextStringKey()
}

func (itr *Iterator) StringKey() string {
	return itr.currentKey
}

func (itr *Iterator) IntegerKey() int64 {
	gotKey, err := strconv.ParseInt(itr.currentKey, 10, 64)
	if err != nil {
		return -1
	}
	return gotKey
}

func (itr *Iterator) Value() []byte {
	return itr.currentValue[0]
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
			}

			if entry.EType == LeafEntry && entry.Name == key {
				manifestState := &ManifestState{
					currentManifest: manifest,
					currentIndex:    i,
				}
				itr.manifestStack = append(itr.manifestStack, manifestState)
				return nil
			}

			if entry.EType == IntermediateEntry && strings.HasPrefix(key, entry.Name) {
				// found a branch, push the current manifest state
				manifestState := &ManifestState{
					currentManifest: manifest,
					currentIndex:    i + 1,
				}
				itr.manifestStack = append(itr.manifestStack, manifestState)

				// now load the child manifest and re-seek
				childManifest, err := itr.index.loadManifest(manifest.Name + entry.Name)
				if err != nil {
					return err
				}

				childKey := strings.TrimPrefix(key, entry.Name)
				return itr.seekStringKey(childManifest, childKey)
			}
		}
	}
	return ErrEntryNotFound
}

func (itr *Iterator) nextStringKey() bool {
	// dont go beyond the limit
	if itr.limit >= 0 {
		if itr.givenUntilNow >= itr.limit {
			return false
		}
	}

	// get the current manifest at the top of the stack
	depthOfStack := len(itr.manifestStack)
	if depthOfStack == 0 {
		itr.error = ErrNoNextElement
		return false
	}

	// take the top manifest to find the next entry
	manifestState := itr.manifestStack[depthOfStack-1]

	entriesExhausted := true
	for entriesExhausted {
		// see if we have exhausted the entries in the current manifest
		if manifestState.currentIndex >= len(manifestState.currentManifest.Entries) {
			// pop the exhausted manifest from the top and pick the next manifest to find the entry
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

	// We have a manifest whose entries are not yet exhausted,
	// so get the next entry and check for valid conditions of the Iterator()
	entry := manifestState.currentManifest.Entries[manifestState.currentIndex]
	manifestState.currentIndex++

	// check if the search has reached the end key
	if itr.endPrefix != "" {
		actualKey := manifestState.currentManifest.Name + entry.Name
		actualKey = strings.TrimPrefix(actualKey, itr.index.name)
		if actualKey > itr.endPrefix {
			return false
		}
	}

	// if it is a leaf entry, set the key and value
	if entry.EType == LeafEntry {
		actualKey := manifestState.currentManifest.Name + entry.Name
		actualKey = strings.TrimPrefix(actualKey, itr.index.name)
		itr.currentKey = actualKey
		itr.currentValue = entry.Ref
		itr.givenUntilNow++
		return true
	}

	// if it is an intermediate entry, get the branch manifest and push in to the stack
	if entry.EType == IntermediateEntry {
		newManifest, err := itr.index.loadManifest(manifestState.currentManifest.Name + entry.Name)
		if err != nil {
			itr.error = err
			return false
		}

		newManifestState := &ManifestState{
			currentManifest: newManifest,
			currentIndex:    0,
		}
		itr.manifestStack = append(itr.manifestStack, newManifestState)
		return itr.nextStringKey()
	}
	return false
}
