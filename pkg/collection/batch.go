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
	"sync/atomic"
	"time"
)

const (
	maxManifestDepth = 3
)

// Batch is to be used in KV table or a Document database
type Batch struct {
	idx           *Index
	memDb         *Manifest
	manifestStack []*Manifest
	storageCount  uint64
}

// NewBatch creates a new batch index to be used in a KV table or a Document database.
func NewBatch(idx *Index) (*Batch, error) {
	return &Batch{
		idx: idx,
	}, nil
}

// PutNumber inserts index as a number.
func (b *Batch) PutNumber(key float64, refValue []byte, apnd, memory bool) error {
	stringKey := fmt.Sprintf("%020.20g", key)
	return b.Put(stringKey, refValue, apnd, memory)
}

// Put creates an index entry given a key string and value.
func (b *Batch) Put(key string, value []byte, apnd, memory bool) error {
	if b.idx.isReadOnlyFeed() { // skipcq: TCV-001
		return ErrReadOnlyIndex
	}

	if b.memDb == nil {
		manifest := &Manifest{
			Name:         b.idx.name,
			IdxType:      b.idx.indexType,
			CreationTime: time.Now().Unix(),
			dirtyFlag:    true,
		}
		b.memDb = manifest
	}
	ctx := context.Background()

	stringKey := key
	if b.idx.indexType == NumberIndex {
		i, err := strconv.ParseInt(stringKey, 10, 64)
		if err != nil { // skipcq: TCV-001
			return ErrKVKeyNotANumber
		}
		stringKey = fmt.Sprintf("%020d", i)
	} else if b.idx.indexType == BytesIndex {
		ref, err := b.idx.client.UploadBlob(value, 0, true)
		if err != nil { // skipcq: TCV-001
			return err
		}
		value = ref
	}
	return b.idx.addOrUpdateStringEntry(ctx, b.memDb, stringKey, b.idx.indexType, value, memory, apnd)
}

// Get extracts an index value from an index given a key.
func (b *Batch) Get(key string) ([][]byte, error) {
	if b.memDb == nil {
		return nil, ErrEntryNotFound
	}
	if len(b.memDb.Entries) > 0 {
		stringKey := key
		if b.idx.indexType == NumberIndex { // skipcq: TCV-001
			i, err := strconv.ParseInt(stringKey, 10, 64)
			if err != nil {
				return nil, ErrKVKeyNotANumber
			}
			stringKey = fmt.Sprintf("%020d", i)
		}

		_, manifest, i, err := b.idx.findManifest(nil, b.memDb, stringKey)
		if err != nil {
			return nil, err
		}
		return manifest.Entries[i].Ref, nil // skipcq: TCV-001
	}
	return nil, ErrEntryNotFound // skipcq: TCV-001
}

// DelNumber deletes a number index key and value.
// skipcq: TCV-001
func (b *Batch) DelNumber(key float64) ([][]byte, error) {
	stringKey := fmt.Sprintf("%020.20g", key)
	return b.Del(stringKey)
}

// Del deletes a index entry.
func (b *Batch) Del(key string) ([][]byte, error) {
	if b.idx.isReadOnlyFeed() { // skipcq: TCV-001
		return nil, ErrReadOnlyIndex
	}

	if !b.idx.mutable { // skipcq: TCV-001
		return nil, ErrCannotModifyImmutableIndex
	}

	if b.memDb == nil { // skipcq: TCV-001
		return nil, ErrEntryNotFound
	}
	if len(b.memDb.Entries) > 0 {
		stringKey := key
		if b.idx.indexType == NumberIndex { // skipcq: TCV-001
			i, err := strconv.ParseInt(stringKey, 10, 64)
			if err != nil {
				return nil, ErrKVKeyNotANumber
			}
			stringKey = fmt.Sprintf("%020d", i)
		}
		parentManifest, manifest, i, err := b.idx.findManifest(nil, b.memDb, stringKey)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}

		deletedRef := manifest.Entries[i].Ref

		if parentManifest != nil && len(manifest.Entries) == 1 && manifest.Entries[0].Name == "" { // skipcq: TCV-001
			// then we have to remove the intermediate node in the parent Manifest
			// so that the entire branch goes kaboom
			parentEntryKey := filepath.Base(manifest.Name)
			for i, entry := range parentManifest.Entries {
				if entry.EType == intermediateEntry && entry.Name == parentEntryKey {
					deletedRef = entry.Ref
					parentManifest.Entries = append(parentManifest.Entries[:i], parentManifest.Entries[i+1:]...)
					break
				}
			}
			return deletedRef, nil
		}
		manifest.Entries = append(manifest.Entries[:i], manifest.Entries[i+1:]...)
		return deletedRef, nil
	}
	return nil, ErrEntryNotFound // skipcq: TCV-001
}

// Write commits the raw index file in to the Swarm network.
func (b *Batch) Write(podFile string) (*Manifest, error) {
	if b.idx.isReadOnlyFeed() { // skipcq: TCV-001
		return nil, ErrReadOnlyIndex
	}
	if b.memDb == nil { // skipcq: TCV-001
		return nil, ErrEntryNotFound
	}

	if b.memDb.dirtyFlag {
		diskManifest, err := b.idx.loadManifest(b.memDb.Name, b.idx.encryptionPassword)
		if err != nil && errors.Is(err, ErrNoManifestFound) { // skipcq: TCV-001
			return nil, err
		}
		diskManifest.PodFile = podFile
		b.memDb.PodFile = podFile
		b.idx.podFile = podFile
		return b.mergeAndWriteManifest(diskManifest, b.memDb)
	}
	return b.memDb, nil // skipcq: TCV-001
}

func (b *Batch) mergeAndWriteManifest(diskManifest, memManifest *Manifest) (*Manifest, error) {
	// merge the mem manifest with the disk version
	if memManifest.dirtyFlag {
		for _, dirtyEntry := range memManifest.Entries {
			diskManifest.dirtyFlag = true
			b.idx.addEntryToManifestSortedLexicographically(diskManifest, dirtyEntry)
			if dirtyEntry.EType == intermediateEntry && dirtyEntry.Manifest != nil { // skipcq: TCV-001
				err := b.storeMemoryManifest(dirtyEntry.Manifest, 0)
				if err != nil {
					return nil, err
				}
				dirtyEntry.Manifest = nil
			}
		}
		diskManifest.Mutable = memManifest.Mutable

		for _, dirtyEntry := range diskManifest.Entries {
			dirtyEntry.Manifest = nil
		}

		if diskManifest.dirtyFlag {
			// save th disk manifest
			err := b.idx.updateManifest(diskManifest, b.idx.encryptionPassword)
			if err != nil { // skipcq: TCV-001
				return nil, err
			}
		}

		err := b.emptyManifestStack()
		if err != nil { // skipcq: TCV-001
			return nil, err
		}

		return diskManifest, nil
	}
	return diskManifest, nil
}

func (b *Batch) emptyManifestStack() error {
	var tempStack []*Manifest

	// copy the data to tempStack
	tempStack = append(tempStack, b.manifestStack...)
	b.manifestStack = nil

	for _, man := range tempStack { // skipcq: TCV-001
		err := b.storeMemoryManifest(man, 0)
		if err != nil {
			return err
		}
	}

	if len(b.manifestStack) > 0 { // skipcq: TCV-001
		return b.emptyManifestStack()
	}

	return nil
}

// skipcq: TCV-001
func (b *Batch) storeMemoryManifest(manifest *Manifest, depth int) error {
	/*
		var wg sync.WaitGroup
		errC := make(chan error)
		wgDone := make(chan bool)
	*/

	// store any branches in this manifest
	for _, entry := range manifest.Entries {
		if entry.EType == intermediateEntry && entry.Manifest != nil {
			if depth >= maxManifestDepth {
				// process later
				b.manifestStack = append(b.manifestStack, entry.Manifest)
				entry.Manifest = nil
				return nil
			}
			// wg.Add(1)
			// go func() {
			// defer func() {
			//	 wg.Done()
			// }()
			err := b.storeMemoryManifest(entry.Manifest, depth+1)
			if err != nil {
				return err
			}
			// }()

		}
	}

	// go func() {
	//	 wg.Wait()
	//	 close(wgDone)
	// }()
	//
	// select {
	// case <-wgDone:
	//	 break
	// case err := <-errC:
	//	 close(errC)
	//	 return err
	// }

	// store this manifest
	// go func() {
	err := b.idx.storeManifest(manifest, b.idx.encryptionPassword)
	if err != nil {
		return err
	}
	atomic.AddUint64(&b.storageCount, 1)
	count := atomic.LoadUint64(&b.storageCount)
	if count%100 == 0 {
		fmt.Println(count)
	}

	// }()
	return nil
}
