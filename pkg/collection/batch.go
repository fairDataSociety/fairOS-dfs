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
	"time"
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

func (b *Batch) Put(key string, refValue []byte) error {
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
	return b.idx.addOrUpdateStringEntry(ctx, b.memDb, key, b.idx.indexType, refValue, true, false)
}

func (b *Batch) Write() error {
	if b.memDb == nil {
		return ErrEntryNotFound
	}

	if b.memDb.dirtyFlag {
		diskManifest, err := b.idx.loadManifest(b.memDb.Name)
		if err != nil && errors.Is(err, ErrNoManifestFound) {
			return err
		}
		return b.mergeAndWriteManifest(diskManifest, b.memDb)
	}
	return nil
}

func (b *Batch) mergeAndWriteManifest(diskManifest, memManifest *Manifest) error {
	// merge the mem manifest with the disk version
	if memManifest.dirtyFlag {
		for _, dirtyEntry := range memManifest.Entries {
			diskManifest.dirtyFlag = true
			b.idx.addEntryToManifestSortedLexicographically(diskManifest, dirtyEntry)
			if dirtyEntry.EType == IntermediateEntry && dirtyEntry.manifest != nil {
				err := b.storeMemoryManifest(dirtyEntry.manifest)
				if err != nil {
					return err
				}
			}
		}

		if diskManifest.dirtyFlag {
			// save th disk manifest
			err := b.idx.updateManifest(diskManifest)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Batch) storeMemoryManifest(manifest *Manifest) error {
	// store this manifest
	err := b.idx.storeManifest(manifest)
	if err != nil {
		return err
	}

	// store any branches in this manifest
	for _, entry := range manifest.Entries {
		if entry.EType == IntermediateEntry && entry.manifest != nil {
			err := b.storeMemoryManifest(entry.manifest)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
