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
func (b *Batch) Put(key string, refValue []byte, apnd, memory bool) error {
	if b.idx.isReadOnlyFeed() {
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
		if err != nil {
			return ErrKVKeyNotANumber
		}
		stringKey = fmt.Sprintf("%020d", i)
	}
	return b.idx.addOrUpdateStringEntry(ctx, b.memDb, stringKey, b.idx.indexType, refValue, memory, apnd)
}

// Get extracts an index value from an index given key.
func (b *Batch) Get(key string) ([][]byte, error) {
	if b.memDb == nil {
		return nil, ErrEntryNotFound
	}
	if len(b.memDb.Entries) > 0 {
		stringKey := key
		if b.idx.indexType == NumberIndex {
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
		return manifest.Entries[i].Ref, nil
	}
	return nil, ErrEntryNotFound
}

// DelNumber deletes a number index key and value.
func (b *Batch) DelNumber(key float64) ([][]byte, error) {
	stringKey := fmt.Sprintf("%020.20g", key)
	return b.Del(stringKey)
}

// Del deletes a index entry.
func (b *Batch) Del(key string) ([][]byte, error) {
	if b.idx.isReadOnlyFeed() {
		return nil, ErrReadOnlyIndex
	}

	if !b.idx.mutable {
		return nil, ErrCannotModifyImmutableIndex
	}

	if b.memDb == nil {
		return nil, ErrEntryNotFound
	}
	if len(b.memDb.Entries) > 0 {
		stringKey := key
		if b.idx.indexType == NumberIndex {
			i, err := strconv.ParseInt(stringKey, 10, 64)
			if err != nil {
				return nil, ErrKVKeyNotANumber
			}
			stringKey = fmt.Sprintf("%020d", i)
		}
		parentManifest, manifest, i, err := b.idx.findManifest(nil, b.memDb, stringKey)
		if err != nil {
			return nil, err
		}

		deletedRef := manifest.Entries[i].Ref

		if parentManifest != nil && len(manifest.Entries) == 1 && manifest.Entries[0].Name == "" {
			// then we have to remove the intermediate node in the parent Manifest
			// so that the entire branch goes kaboom
			parentEntryKey := filepath.Base(manifest.Name)
			for i, entry := range parentManifest.Entries {
				if entry.EType == IntermediateEntry && entry.Name == parentEntryKey {
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
	return nil, ErrEntryNotFound
}

// Write commits the raw index file in to the Swarm network.
func (b *Batch) Write(podFile string) (*Manifest, error) {
	if b.idx.isReadOnlyFeed() {
		return nil, ErrReadOnlyIndex
	}
	if b.memDb == nil {
		return nil, ErrEntryNotFound
	}

	if b.memDb.dirtyFlag {
		diskManifest, err := b.idx.loadManifest(b.memDb.Name)
		if err != nil && errors.Is(err, ErrNoManifestFound) {
			return nil, err
		}
		diskManifest.PodFile = podFile
		b.memDb.PodFile = podFile
		b.idx.podFile = podFile
		return b.mergeAndWriteManifest(diskManifest, b.memDb)
	}
	return b.memDb, nil
}

func (b *Batch) mergeAndWriteManifest(diskManifest, memManifest *Manifest) (*Manifest, error) {
	// merge the mem manifest with the disk version
	if memManifest.dirtyFlag {
		for _, dirtyEntry := range memManifest.Entries {
			diskManifest.dirtyFlag = true
			b.idx.addEntryToManifestSortedLexicographically(diskManifest, dirtyEntry)
			if dirtyEntry.EType == IntermediateEntry && dirtyEntry.Manifest != nil {
				err := b.storeMemoryManifest(dirtyEntry.Manifest, 0)
				if err != nil {
					return nil, err
				}
				dirtyEntry.Manifest = nil
				fmt.Println(atomic.LoadUint64(&b.storageCount))
			}
		}
		diskManifest.Mutable = memManifest.Mutable

		for _, dirtyEntry := range diskManifest.Entries {
			dirtyEntry.Manifest = nil
		}

		if diskManifest.dirtyFlag {
			// save th disk manifest
			err := b.idx.updateManifest(diskManifest)
			if err != nil {
				return nil, err
			}
		}

		err := b.emptyManifestStack()
		if err != nil {
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

	for _, man := range tempStack {
		err := b.storeMemoryManifest(man, 0)
		if err != nil {
			return err
		}
	}

	if len(b.manifestStack) > 0 {
		return b.emptyManifestStack()
	}

	return nil
}

func (b *Batch) storeMemoryManifest(manifest *Manifest, depth int) error {
	//var wg sync.WaitGroup
	//errC := make(chan error)
	//wgDone := make(chan bool)

	// store any branches in this manifest
	for _, entry := range manifest.Entries {
		if entry.EType == IntermediateEntry && entry.Manifest != nil {
			if depth >= maxManifestDepth {
				// process later
				b.manifestStack = append(b.manifestStack, entry.Manifest)
				entry.Manifest = nil
				return nil
			}
			//wg.Add(1)
			//go func() {
			//	defer func() {
			//		wg.Done()
			//	}()
			err := b.storeMemoryManifest(entry.Manifest, depth+1)
			if err != nil {
				return err
			}
			//}()

		}
	}

	//go func() {
	//	wg.Wait()
	//	close(wgDone)
	//}()
	//
	//select {
	//case <-wgDone:
	//	break
	//case err := <-errC:
	//	close(errC)
	//	return err
	//}

	// store this manifest
	//go func() {
	err := b.idx.storeManifest(manifest)
	if err != nil {
		return err
	}
	atomic.AddUint64(&b.storageCount, 1)
	count := atomic.LoadUint64(&b.storageCount)
	if count%100 == 0 {
		fmt.Println(count)
	}

	//}()
	return nil
}
