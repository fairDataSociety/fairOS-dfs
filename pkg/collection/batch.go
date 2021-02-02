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
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

type Batch struct {
	idx   *Index
	memDb *Manifest
}

func NewBatch(idx *Index) (*Batch, error) {
	return &Batch{
		idx: idx,
	}, nil
}

func (b *Batch) PutNumber(key float64, refValue []byte, apnd, memory bool) error {
	stringKey := fmt.Sprintf("%020.20g", key)
	return b.Put(stringKey, refValue, apnd, memory)
}

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

		_, manifest, i, err := b.idx.findManifest(nil, b.memDb, stringKey, true)
		if err != nil {
			return nil, err
		}
		return manifest.Entries[i].Ref, nil
	}
	return nil, ErrEntryNotFound
}

func (b *Batch) DelNumber(key float64) ([][]byte, error) {
	stringKey := fmt.Sprintf("%020.20g", key)
	return b.Del(stringKey)
}

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
		memory := !b.idx.mutable
		parentManifest, manifest, i, err := b.idx.findManifest(nil, b.memDb, stringKey, memory)
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

		if b.idx.mutable {
			err = b.storeMutableMemoryManifest(manifest)
			if err != nil {
				return nil, err
			}
		}
		return deletedRef, nil
	}
	return nil, ErrEntryNotFound
}

func (b *Batch) Write(podFile string) error {
	if b.idx.isReadOnlyFeed() {
		return ErrReadOnlyIndex
	}
	if b.memDb == nil {
		return ErrEntryNotFound
	}

	if b.memDb.dirtyFlag {
		diskManifest, err := b.idx.loadManifest(b.memDb.Name)
		if err != nil && errors.Is(err, ErrNoManifestFound) {
			return err
		}
		b.memDb.Mutable = b.idx.mutable
		if podFile != "" {
			b.memDb.PodFile = podFile
			b.idx.podFile = podFile
		}
		if b.idx.mutable {
			_, err := b.mergeAndWriteManifest(b.memDb, nil)
			if err != nil {
				return err
			}
			return nil
		}

		dm, err := b.mergeAndWriteManifest(diskManifest, b.memDb)
		if err != nil {
			return err
		}
		b.idx.memDB = dm
		b.idx.memDB.dirtyFlag = false
	}
	return nil
}

func (b *Batch) mergeAndWriteManifest(diskManifest, memManifest *Manifest) (*Manifest, error) {
	if !diskManifest.Mutable {
		if !memManifest.dirtyFlag {
			return nil, nil
		}
		for _, dirtyEntry := range memManifest.Entries {
			diskManifest.dirtyFlag = true
			b.idx.addEntryToManifestSortedLexicographically(diskManifest, dirtyEntry)
		}
		diskManifest.Mutable = memManifest.Mutable
		diskManifest.PodFile = memManifest.PodFile

		// save th entire Manifest in one shot
		data, err := json.Marshal(diskManifest)
		if err != nil {
			return nil, err
		}
		ref, err := b.idx.client.UploadBlob(data, true, true)
		if err != nil {
			return nil, err
		}

		// update the feed to point to this Manifest
		topic := utils.HashString(diskManifest.Name)
		_, err = b.idx.feed.UpdateFeed(topic, b.idx.user, ref)
		if err != nil {
			return nil, err
		}
		return diskManifest, nil
	} else {
		// merge the mem Manifest with the disk version
		//for _, dirtyEntry := range memManifest.Entries {
		//	diskManifest.dirtyFlag = true
		//	b.idx.addEntryToManifestSortedLexicographically(diskManifest, dirtyEntry)
		//	if dirtyEntry.EType == IntermediateEntry && dirtyEntry.Manifest != nil {
		//		err := b.storeMutableMemoryManifest(dirtyEntry.Manifest)
		//		if err != nil {
		//			return nil, err
		//		}
		//	}
		//}

		if diskManifest.dirtyFlag {
			// save th disk Manifest
			err := b.idx.updateManifest(diskManifest)
			if err != nil {
				return nil, err
			}
		}
		return diskManifest, nil
	}
}

func (b *Batch) storeMutableMemoryManifest(manifest *Manifest) error {
	// store this Manifest
	err := b.idx.storeManifest(manifest)
	if err != nil {
		return err
	}

	// store any branches in this Manifest
	for _, entry := range manifest.Entries {
		if entry.EType == IntermediateEntry && entry.Manifest != nil {
			err := b.storeMutableMemoryManifest(entry.Manifest)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
