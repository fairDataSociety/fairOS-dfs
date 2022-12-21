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
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const (
	LeafEntry         = "L"
	IntermediateEntry = "I"
)

// PutNumber inserts an entry in to index with a number as key.
func (idx *Index) PutNumber(key float64, refValue []byte, idxType IndexType, apnd bool) error {
	stringKey := fmt.Sprintf("%020.20g", key)
	return idx.Put(stringKey, refValue, idxType, apnd)
}

// Put inserts an entry in to index with a string as key.
func (idx *Index) Put(key string, refValue []byte, idxType IndexType, apnd bool) error {
	if idx.isReadOnlyFeed() { // skipcq: TCV-001
		return ErrReadOnlyIndex
	}

	if !idx.mutable { // skipcq: TCV-001
		return ErrCannotModifyImmutableIndex
	}

	// get the first feed of the Index
	manifest, err := idx.loadManifest(idx.name, idx.encryptionPassword)
	if err != nil { // skipcq: TCV-001
		return err
	}

	ctx := context.Background()
	return idx.addOrUpdateStringEntry(ctx, manifest, key, idxType, refValue, false, apnd)
}

// GetNumber retrieves an element from the index where the key is of type number.
// skipcq: TCV-001
func (idx *Index) GetNumber(key float64) ([][]byte, error) {
	stringKey := fmt.Sprintf("%020.20g", key)
	return idx.Get(stringKey)
}

// Get retrieves an element from the index where the key is of type string.
func (idx *Index) Get(key string) ([][]byte, error) {
	_, manifest, i, err := idx.seekManifestAndEntry(key)
	if err != nil {
		return nil, err
	}

	return manifest.Entries[i].Ref, nil
}

// DeleteNumber removes an entry from index where the key is of type number.
func (idx *Index) DeleteNumber(key float64) ([][]byte, error) {
	stringKey := fmt.Sprintf("%020.20g", key)
	return idx.Delete(stringKey)
}

// Delete removes an entry from index where the key is of type string.
func (idx *Index) Delete(key string) ([][]byte, error) {
	if idx.isReadOnlyFeed() { // skipcq: TCV-001
		return nil, ErrReadOnlyIndex
	}

	if !idx.mutable { // skipcq: TCV-001v
		return nil, ErrCannotModifyImmutableIndex
	}

	_, manifest, i, err := idx.seekManifestAndEntry(key)
	if err != nil {
		return nil, err
	}

	deletedRef := manifest.Entries[i].Ref

	if len(manifest.Entries) == 1 && manifest.Entries[0].Name == "" { // skipcq: TCV-001
		// then we have to remove the intermediate node in the parent Manifest
		// so that the entire branch goes kaboom
		parentEntryKey := filepath.Base(manifest.Name)
		parentManifest, err := idx.loadManifest(filepath.Dir(manifest.Name), idx.encryptionPassword)
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
		err = idx.updateManifest(parentManifest, idx.encryptionPassword)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		return deletedRef, nil
	}

	manifest.Entries = append(manifest.Entries[:i], manifest.Entries[i+1:]...)
	err = idx.updateManifest(manifest, idx.encryptionPassword)
	if err != nil {
		return nil, err
	}
	return deletedRef, nil
}

func (idx *Index) addOrUpdateStringEntry(ctx context.Context, manifest *Manifest, key string, idxType IndexType, value []byte, memory, apnd bool) error {
	entryAdded := false

	for i := range manifest.Entries {
		entry := manifest.Entries[i] // we change the entry so dont simplify this

		// add new entry with key equal to the Manifest name
		if key == "" {
			break
		}

		// this is the update of an existing entry
		if entry.EType == LeafEntry && entry.Name == key {
			var refs [][]byte
			if apnd {
				refs = entry.Ref
			}
			entry.Ref = append(refs, value) // skipcq: CRT-D0001
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

			// store the new Manifest with two leaves
			if !memory {
				err := idx.storeManifest(&newManifest, idx.encryptionPassword)
				if err != nil { // skipcq: TCV-001
					return err
				}
			} else { // skipcq: TCV-001
				entry.Manifest = &newManifest
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
				// create the new Manifest with two entries
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
					err := idx.storeManifest(&newManifest, idx.encryptionPassword)
					if err != nil { // skipcq: TCV-001
						return err
					}
				} else { // skipcq: TCV-001
					// update the old Manifest name and add the new Manifest to the existing entry
					oldManifest := entry.Manifest
					entry2.Manifest = oldManifest
					entry.Manifest = &newManifest
				}

				// update the existing intermediate nodes name
				entry.Name = prefix
				entry.EType = IntermediateEntry
				manifest.dirtyFlag = true
				entryAdded = true
				break
			} else if len(keySuffix) > 0 {
				// load the entry's Manifest and add the keySuffix as a new leaf
				if !memory {
					intermediateManifest, err := idx.loadManifest(manifest.Name+entry.Name, idx.encryptionPassword)
					if err != nil { // skipcq: TCV-001
						return err
					}
					return idx.addOrUpdateStringEntry(ctx, intermediateManifest, keySuffix, idxType, value, memory, apnd)
				} else { // skipcq: TCV-001
					return idx.addOrUpdateStringEntry(ctx, entry.Manifest, keySuffix, idxType, value, memory, apnd)
				}

			} else if entrySuffix == "" && keySuffix == "" {
				// load the entry's Manifest and add the keySuffix as a new leaf
				if !memory {
					intermediateManifest, err := idx.loadManifest(manifest.Name+prefix, idx.encryptionPassword)
					if err != nil { // skipcq: TCV-001
						return err
					}
					return idx.addOrUpdateStringEntry(ctx, intermediateManifest, keySuffix, idxType, value, memory, apnd)
				} else { // skipcq: TCV-001
					return idx.addOrUpdateStringEntry(ctx, entry.Manifest, keySuffix, idxType, value, memory, apnd)
				}

			} else if len(entrySuffix) > 0 {
				// create the new Manifest with two entries
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
					err := idx.storeManifest(&newManifest, idx.encryptionPassword)
					if err != nil { // skipcq: TCV-001
						return err
					}
				} else { // skipcq: TCV-001
					oldManifest := entry.Manifest
					entry2.Manifest = oldManifest
					entry.Manifest = &newManifest
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

	// if the Manifest is not already changed, then this is a new entry
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
		return idx.updateManifest(manifest, idx.encryptionPassword)
	}
	return nil // skipcq: TCV-001
}

func (*Index) addEntryToManifestSortedLexicographically(manifest *Manifest, entryToAdd *Entry) {
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
		// if the enty is already there, just return
		if entry.Name == entryToAdd.Name &&
			entry.EType == entryToAdd.EType {
			if len(entry.Ref) == len(entryToAdd.Ref) {
				equal := true
				for kk, r := range entry.Ref {
					if !bytes.Equal(r, entryToAdd.Ref[kk]) { // skipcq: TCV-001
						equal = false
					}
				}
				if equal {
					return
				}
			}
		}

		if entry.Name == "" {
			entries = append(entries, entry)
			continue
		}
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

	if !entryAdded {
		entries = append(entries, entryToAdd)
	}

	manifest.Entries = entries
}

func (idx *Index) seekManifestAndEntry(key string) (*Manifest, *Manifest, int, error) {

	// load the first Manifest of the index
	fm, err := idx.loadManifest(idx.name, idx.encryptionPassword)
	if err != nil && !errors.Is(err, ErrNoManifestFound) { // skipcq: TCV-001
		return nil, nil, 0, err
	}

	// if there are any elements in the index, then search for the entry
	if fm.Entries != nil && len(fm.Entries) > 0 {
		return idx.findManifest(nil, fm, key)
	}
	return nil, nil, 0, ErrEntryNotFound
}

func (idx *Index) findManifest(grandParentManifest, parentManifest *Manifest, key string) (*Manifest, *Manifest, int, error) {
	if parentManifest != nil {
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
				var childManifest *Manifest
				childKey := strings.TrimPrefix(key, entry.Name)
				if entry.Manifest == nil {
					childManifestPath := parentManifest.Name + entry.Name
					var err error
					childManifest, err = idx.loadManifest(childManifestPath, idx.encryptionPassword)
					if err != nil { // skipcq: TCV-001
						return nil, nil, 0, err
					}
				} else { // skipcq: TCV-001
					childManifest = entry.Manifest
				}
				return idx.findManifest(parentManifest, childManifest, childKey)

			}
		}
	}
	return nil, nil, 0, ErrEntryNotFound
}
