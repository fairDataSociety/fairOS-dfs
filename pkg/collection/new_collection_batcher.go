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
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	"github.com/ethersphere/bee/pkg/file"
	"github.com/ethersphere/bee/pkg/file/pipeline/builder"

	"github.com/ethersphere/bee/pkg/storage"
	"github.com/ethersphere/bee/pkg/swarm"
)

// Store implements a simple Putter and Getter which can be used to temporarily cache
// chunks. Currently, this is used in the bootstrapping process of new nodes where
// we sync the postage events from the swarm network.
type Store struct {
	mtx   sync.Mutex
	store map[string]swarm.Chunk
}

func NewStore() *Store {
	return &Store{
		store: make(map[string]swarm.Chunk),
	}
}

func (s *Store) Get(_ context.Context, _ storage.ModeGet, addr swarm.Address) (ch swarm.Chunk, err error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if ch, ok := s.store[addr.ByteString()]; ok {
		return ch, nil
	}

	return nil, storage.ErrNotFound
}

func (s *Store) Put(_ context.Context, _ storage.ModePut, chs ...swarm.Chunk) (exist []bool, err error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for _, ch := range chs {
		s.store[ch.Address().ByteString()] = ch
	}

	exist = make([]bool, len(chs))

	return exist, err
}

func (s *Store) GetMulti(_ context.Context, _ storage.ModeGet, _ ...swarm.Address) (ch []swarm.Chunk, err error) {
	panic("not implemented")
}

func (s *Store) Has(_ context.Context, _ swarm.Address) (yes bool, err error) {
	panic("not implemented")
}

func (s *Store) HasMulti(_ context.Context, _ ...swarm.Address) (yes []bool, err error) {
	panic("not implemented")
}

func (s *Store) Set(_ context.Context, _ storage.ModeSet, _ ...swarm.Address) (err error) {
	panic("not implemented")
}

func (s *Store) LastPullSubscriptionBinID(_ uint8) (id uint64, err error) {
	panic("not implemented")
}

func (s *Store) SubscribePull(_ context.Context, _ uint8, _ uint64, _ uint64) (c <-chan storage.Descriptor, closed <-chan struct{}, stop func()) {
	panic("not implemented")
}

func (s *Store) SubscribePush(_ context.Context, _ func([]byte) bool) (c <-chan swarm.Chunk, repeat func(), stop func()) {
	panic("not implemented")
}

func (s *Store) ReserveSample(_ context.Context, _ []byte, _ uint8, _ uint64) (storage.Sample, error) {
	panic("not implemented")
}

func (s *Store) Close() error {
	return nil
}

type SocStore struct {
	Mtx   sync.Mutex
	Store map[string][]byte
}

func (s *SocStore) Get(topic string) ([]byte, error) {
	s.Mtx.Lock()
	defer s.Mtx.Unlock()

	if data, ok := s.Store[topic]; ok {
		return data, nil
	}

	return nil, fmt.Errorf("not found")
}

func (s *SocStore) Put(topic string, data []byte) error {
	s.Mtx.Lock()
	defer s.Mtx.Unlock()

	s.Store[topic] = data

	return nil
}

func NewSocStore() *SocStore {
	return &SocStore{
		Store: make(map[string][]byte),
	}
}

// Batcher is to be used in KV table or a Document database for generating manifest locally and store then in one update
type Batcher struct {
	idx           *Index
	memDb         *Manifest
	manifestStack []*Manifest
	storageCount  uint64
	store         *Store
	socStore      *SocStore
}

// NewBatcher creates a new batch index to be used in a KV table or a Document database.
func NewBatcher(idx *Index) (*Batcher, error) {
	b := &Batcher{
		idx:      idx,
		store:    NewStore(),
		socStore: NewSocStore(),
	}

	memDb, err := b.idx.loadManifest(idx.name, b.idx.encryptionPassword)
	if err == nil && memDb != nil {
		b.memDb = memDb
	}
	return b, nil
}

// PutNumber inserts index as a number.
func (b *Batcher) PutNumber(key float64, refValue []byte, apnd, memory bool) error {
	stringKey := fmt.Sprintf("%020.20g", key)
	return b.Put(stringKey, refValue, apnd, memory)
}

// Put creates an index entry given a key string and value.
func (b *Batcher) Put(key string, value []byte, apnd, memory bool) error {
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
		p := builder.NewPipelineBuilder(context.Background(), b.store, storage.ModePutUpload, false)
		dataReader := file.NewSimpleReadCloser(value)
		ref, err := builder.FeedPipeline(context.Background(), p, dataReader)
		if err != nil {
			return err
		}
		value = ref.Bytes()
	}
	return b.idx.addOrUpdateStringEntryIntoStore(ctx, b.memDb, stringKey, b.idx.indexType, value, b.store, b.socStore)
}

// Write commits the raw index file in to the Swarm network.
func (b *Batcher) Write(podFile string) (*Manifest, error) {
	fmt.Println("Write", len(b.store.store), len(b.socStore.Store))
	if b.idx.isReadOnlyFeed() { // skipcq: TCV-001
		return nil, ErrReadOnlyIndex
	}
	if b.memDb == nil { // skipcq: TCV-001
		return nil, ErrEntryNotFound
	}

	if b.memDb.dirtyFlag {
		b.memDb.PodFile = podFile
		b.idx.podFile = podFile
		err := b.mergeAndWriteManifest(b.memDb)
		if err != nil {
			return nil, err
		}
	}

	// commit to swarm
	b.store.mtx.Lock()
	defer b.store.mtx.Unlock()
	chunks := make([]swarm.Chunk, 0)
	for _, ch := range b.store.store {
		//fmt.Println("uploading chunk", ch.Address().String())
		chunks = append(chunks, ch)
		//_, err := b.idx.client.UploadChunk(ch)
		//if err != nil {
		//	return nil, err
		//}
	}
	_, err := b.idx.client.StreamChunks(context.Background(), chunks...)
	if err != nil {
		return nil, err
	}
	b.socStore.Mtx.Lock()
	defer b.socStore.Mtx.Unlock()
	for topic, data := range b.socStore.Store {
		fmt.Println("topic", topic)
		_, err := b.idx.feed.CreateFeed(b.idx.user, utils.HashString(topic), data, []byte(b.idx.encryptionPassword))
		if err != nil {
			if strings.Contains(err.Error(), "chunk already exists") {
				_, err = b.idx.feed.UpdateFeed(b.idx.user, utils.HashString(topic), data, []byte(b.idx.encryptionPassword), false)
				if err != nil { //  skipcq: TCV-001
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	}
	return b.memDb, nil // skipcq: TCV-001
}

func (b *Batcher) mergeAndWriteManifest(memManifest *Manifest) error {
	// merge the mem manifest with the disk version
	if memManifest.dirtyFlag {
		for _, dirtyEntry := range memManifest.Entries {
			if dirtyEntry.EType == intermediateEntry && dirtyEntry.Manifest != nil { // skipcq: TCV-001
				err := b.storeMemoryManifest(dirtyEntry.Manifest, 0)
				if err != nil {
					return err
				}
				dirtyEntry.Manifest = nil
			}
		}

		data, err := json.Marshal(b.memDb)
		if err != nil {
			return err
		}
		p := builder.NewPipelineBuilder(context.Background(), b.store, storage.ModePutUpload, false)
		dataReader := file.NewSimpleReadCloser(data)
		ref, err := builder.FeedPipeline(context.Background(), p, dataReader)
		if err != nil {
			return err
		}
		err = b.socStore.Put(b.memDb.Name, ref.Bytes())
		if err != nil {
			return err
		}

		err = b.emptyManifestStack()
		if err != nil { // skipcq: TCV-001
			return err
		}

		return nil
	}
	return nil
}

func (b *Batcher) emptyManifestStack() error {
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
func (b *Batcher) storeMemoryManifest(manifest *Manifest, depth int) error {
	//var wg sync.WaitGroup
	//errCh := make(chan error)

	// store any branches in this manifest
	for _, entry := range manifest.Entries {
		if entry.EType == intermediateEntry && entry.Manifest != nil {
			if depth >= maxManifestDepth {
				// process later
				b.manifestStack = append(b.manifestStack, entry.Manifest)
				entry.Manifest = nil
				return nil
			}

			err := b.storeMemoryManifest(entry.Manifest, depth+1)
			if err != nil {
				return err
			}

			//wg.Add(1)
			//go func(entry *Entry) {
			//	defer wg.Done()
			//	err := b.storeMemoryManifest(entry.Manifest, depth+1)
			//	if err != nil {
			//		errCh <- err
			//	}
			//}(entry)
		}
	}
	//
	//go func() {
	//	wg.Wait()
	//	close(errCh)
	//}()
	//
	//for err := range errCh {
	//	if err != nil {
	//		return err
	//	}
	//}

	data, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	p := builder.NewPipelineBuilder(context.Background(), b.store, storage.ModePutUpload, false)
	dataReader := file.NewSimpleReadCloser(data)
	ref, err := builder.FeedPipeline(context.Background(), p, dataReader)
	if err != nil {
		return err
	}
	err = b.socStore.Put(manifest.Name, ref.Bytes())
	if err != nil {
		return err
	}

	atomic.AddUint64(&b.storageCount, 1)
	count := atomic.LoadUint64(&b.storageCount)
	if count%100 == 0 {
		fmt.Println(count)
	}

	return nil
}
