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
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	LeafEntry         = "L"
	IntermediateEntry = "I"
)

func (idx *Index) Put(key string, refValue []byte, idxType IndexType, apnd bool) error {
	if idx.isReadOnlyFeed() {
		return ErrReadOnlyIndex
	}

	// get the first feed of the Index
	manifest, err := idx.loadManifest(idx.name)
	if err != nil {
		return err
	}

	stringKey := key
	if idx.indexType == NumberIndex {
		i, err := strconv.ParseInt(stringKey, 10, 64)
		if err != nil {
			return ErrKVKeyNotANumber
		}
		stringKey = fmt.Sprintf("%020d", i)
	}

	ctx := context.Background()
	return idx.addOrUpdateStringEntry(ctx, manifest, stringKey, idxType, refValue, false, apnd)
}

func (idx *Index) Get(key string) ([][]byte, error) {
	stringKey := key
	if idx.indexType == NumberIndex {
		i, err := strconv.ParseInt(stringKey, 10, 64)
		if err != nil {
			return nil, ErrKVKeyNotANumber
		}
		stringKey = fmt.Sprintf("%020d", i)
	}

	_, manifest, i, err := idx.seekManifestAndEntry(stringKey)
	if err != nil {
		return nil, err
	}

	return manifest.Entries[i].Ref, nil
}

func (idx *Index) Delete(key string) ([][]byte, error) {
	if idx.isReadOnlyFeed() {
		return nil, ErrReadOnlyIndex
	}
	stringKey := key
	if idx.indexType == NumberIndex {
		i, err := strconv.ParseInt(stringKey, 10, 64)
		if err != nil {
			return nil, ErrKVKeyNotANumber
		}
		stringKey = fmt.Sprintf("%020d", i)
	}

	_, manifest, i, err := idx.seekManifestAndEntry(stringKey)
	if err != nil {
		return nil, err
	}

	deletedRef := manifest.Entries[i].Ref

	if len(manifest.Entries) == 1 && manifest.Entries[0].Name == "" {
		// then we have to remove the intermediate node in the parent manifest
		// so that the entire branch goes kaboom
		parentEntryKey := filepath.Base(manifest.Name)
		parentManifest, err := idx.loadManifest(filepath.Dir(manifest.Name))
		if err != nil {
			return nil, err
		}
		for i, entry := range parentManifest.Entries {
			if entry.EType == IntermediateEntry && entry.Name == parentEntryKey {
				deletedRef = entry.Ref
				parentManifest.Entries = append(parentManifest.Entries[:i], parentManifest.Entries[i+1:]...)
				break
			}
		}
		err = idx.updateManifest(parentManifest)
		if err != nil {
			return nil, err
		}
		return deletedRef, nil
	}

	manifest.Entries = append(manifest.Entries[:i], manifest.Entries[i+1:]...)
	err = idx.updateManifest(manifest)
	if err != nil {
		return nil, err
	}
	return deletedRef, nil
}

func (idx *Index) addOrUpdateStringEntry(ctx context.Context, manifest *Manifest, key string, idxType IndexType, value []byte, memory, apnd bool) error {
	entryAdded := false

	for i := range manifest.Entries {
		entry := manifest.Entries[i] // we change the entry so dont simplify this

		// add new entry with key equal to the manifest name
		if key == "" {
			break
		}

		// this is the update of an existing entry
		if entry.EType == LeafEntry && entry.Name == key {
			var refs [][]byte
			if apnd {
				refs = entry.Ref
			}
			entry.Ref = append(refs, value)
			manifest.dirtyFlag = true
			entryAdded = true
			break
		}

		// if there is no common prefix, skip to next entry
		prefix, entrySuffix, keySuffix := longestCommonPrefix(entry.Name, key)
		if prefix == "" {
			continue
		}

		if entry.EType == LeafEntry {
			var newManifest Manifest
			newManifest.Name = manifest.Name + prefix
			newManifest.IdxType = idxType
			newManifest.CreationTime = time.Now().Unix()
			var refs1 [][]byte
			refs1 = append(refs1, value)
			entry1 := &Entry{
				Name:  keySuffix,
				EType: LeafEntry,
				Ref:   refs1,
			}
			idx.addEntryToManifestSortedLexicographically(&newManifest, entry1)
			entry2 := &Entry{
				Name:  entrySuffix,
				EType: LeafEntry,
				Ref:   entry.Ref,
			}
			idx.addEntryToManifestSortedLexicographically(&newManifest, entry2)

			// store the new manifest with two leaves
			if !memory {
				err := idx.storeManifest(&newManifest)
				if err != nil {
					return err
				}
			} else {
				entry.manifest = &newManifest
				manifest.dirtyFlag = true
			}

			// convert the existing leaf to intermediate node
			entry.Name = prefix
			entry.EType = IntermediateEntry
			manifest.dirtyFlag = true
			entryAdded = true
			break
		}

		if entry.EType == IntermediateEntry {
			if len(keySuffix) > 0 && len(entrySuffix) > 0 {
				// create the new manifest with two entries
				var newManifest Manifest
				newManifest.Name = manifest.Name + prefix
				newManifest.IdxType = idxType
				newManifest.CreationTime = time.Now().Unix()
				// add the new entry as a leaf
				var refs2 [][]byte
				refs2 = append(refs2, value)
				entry1 := &Entry{
					Name:  keySuffix,
					EType: LeafEntry,
					Ref:   refs2,
				}
				idx.addEntryToManifestSortedLexicographically(&newManifest, entry1)
				// add the old intermediate branch as another entry
				entry2 := &Entry{
					Name:  entrySuffix,
					EType: IntermediateEntry,
				}
				idx.addEntryToManifestSortedLexicographically(&newManifest, entry2)
				if !memory {
					err := idx.storeManifest(&newManifest)
					if err != nil {
						return err
					}
				} else {
					// update the old manifest name and add the new manifest to the existing entry
					oldManifest := entry.manifest
					entry2.manifest = oldManifest
					entry.manifest = &newManifest
				}

				// update the existing intermediate nodes name
				entry.Name = prefix
				entry.EType = IntermediateEntry
				manifest.dirtyFlag = true
				entryAdded = true
				break
			} else if len(keySuffix) > 0 {
				// load the entry's manifest and add the keySuffix as a new leaf
				if !memory {
					intermediateManifest, err := idx.loadManifest(manifest.Name + entry.Name)
					if err != nil {
						return err
					}
					return idx.addOrUpdateStringEntry(ctx, intermediateManifest, keySuffix, idxType, value, memory, apnd)
				} else {
					return idx.addOrUpdateStringEntry(ctx, entry.manifest, keySuffix, idxType, value, memory, apnd)
				}

			} else if entrySuffix == "" && keySuffix == "" {
				// load the entry's manifest and add the keySuffix as a new leaf
				if !memory {
					intermediateManifest, err := idx.loadManifest(manifest.Name + prefix)
					if err != nil {
						return err
					}
					return idx.addOrUpdateStringEntry(ctx, intermediateManifest, keySuffix, idxType, value, memory, apnd)
				} else {
					return idx.addOrUpdateStringEntry(ctx, entry.manifest, keySuffix, idxType, value, memory, apnd)
				}

			} else if len(entrySuffix) > 0 {
				// create the new manifest with two entries
				var newManifest Manifest
				newManifest.Name = manifest.Name + prefix
				newManifest.IdxType = idxType
				newManifest.CreationTime = time.Now().Unix()
				// add the new entry as a leaf
				var refs3 [][]byte
				refs3 = append(refs3, value)
				entry1 := &Entry{
					Name:  keySuffix,
					EType: LeafEntry,
					Ref:   refs3,
				}
				idx.addEntryToManifestSortedLexicographically(&newManifest, entry1)
				// add the old intermediate branch as another entry
				entry2 := &Entry{
					Name:  entrySuffix,
					EType: IntermediateEntry,
				}
				idx.addEntryToManifestSortedLexicographically(&newManifest, entry2)
				if !memory {
					err := idx.storeManifest(&newManifest)
					if err != nil {
						return err
					}
				} else {
					oldManifest := entry.manifest
					entry2.manifest = oldManifest
					entry.manifest = &newManifest
				}

				// update the existing intermediate nodes name
				entry.Name = key
				entry.EType = IntermediateEntry
				manifest.dirtyFlag = true
				entryAdded = true
				break
			}
		}
	}

	// if the manifest is not already changed, then this is a new entry
	if !entryAdded {
		var refs [][]byte
		newEntry := Entry{
			Name:  key,
			EType: LeafEntry,
			Ref:   append(refs, value),
		}
		idx.addEntryToManifestSortedLexicographically(manifest, &newEntry)
		manifest.dirtyFlag = true
		entryAdded = true
	}

	if entryAdded && !memory {
		return idx.updateManifest(manifest)
	}
	return nil
}

func (idx *Index) addEntryToManifestSortedLexicographically(manifest *Manifest, entryToAdd *Entry) {
	var entries []*Entry

	// this is the first element
	if len(manifest.Entries) == 0 {
		manifest.Entries = append(manifest.Entries, entryToAdd)
		return
	}

	// new element has an empty name, so add it in the beginning
	if entryToAdd.Name == "" {
		entries = append(entries, entryToAdd)
		manifest.Entries = append(entries, manifest.Entries...)
		return
	}

	entryAdded := false
	for _, entry := range manifest.Entries {
		if entry.Name == "" {
			entries = append(entries, entry)
			continue
		} else {
			if !entryAdded {
				a := entry.Name[0]
				b := entryToAdd.Name[0]
				if a > b {
					entries = append(entries, entryToAdd)
					entryAdded = true
				}
			}
			entries = append(entries, entry)
		}
	}

	if !entryAdded {
		entries = append(entries, entryToAdd)
	}

	manifest.Entries = entries
}

func (idx *Index) seekManifestAndEntry(key string) (*Manifest, *Manifest, int, error) {
	// load the first manifest of the index
	firstManifest, err := idx.loadManifest(idx.name)
	if err != nil && !errors.Is(err, ErrNoManifestFound) {
		return nil, nil, 0, err
	}

	// if there are any elements in the index, then search for the entry
	if len(firstManifest.Entries) > 0 {
		return idx.findManifest(nil, firstManifest, key, false)
	}
	return nil, nil, 0, ErrEntryNotFound
}

func (idx *Index) findManifest(grandParentManifest, parentManifest *Manifest, key string, memory bool) (*Manifest, *Manifest, int, error) {
	for i, entry := range parentManifest.Entries {

		// if the first char is > keys first char, then the key wont be found
		if len(entry.Name) > 0 {
			if key == "" { // to check for empty entry
				return nil, nil, 0, ErrEntryNotFound
			}
			if entry.Name[0] > key[0] { // to check for greater entries
				return nil, parentManifest, 0, ErrEntryNotFound
			}
		}

		if entry.EType == LeafEntry && entry.Name == key {
			return grandParentManifest, parentManifest, i, nil
		}

		if entry.EType == IntermediateEntry && strings.HasPrefix(key, entry.Name) {
			childKey := strings.TrimPrefix(key, entry.Name)
			if !memory {
				childManifestPath := parentManifest.Name + entry.Name
				childManifest, err := idx.loadManifest(childManifestPath)
				if err != nil {
					return nil, nil, 0, err
				}
				return idx.findManifest(parentManifest, childManifest, childKey, memory)
			} else {
				childManifest := entry.manifest
				return idx.findManifest(parentManifest, childManifest, childKey, memory)
			}

		}
	}
	return nil, nil, 0, ErrEntryNotFound
}
