package lookup_test

/*
This file contains components to mock a storage for testing
lookup algorithms and measure the number of reads.
*/

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/feed/lookup"
)

// Data is a struct to keep a value to store/retrieve during testing
type Data struct {
	Payload uint64
	Time    uint64
}

// String implements fmt.Stringer
func (d *Data) String() string {
	return fmt.Sprintf("%d-%d", d.Payload, d.Time)
}

// Datamap is an internal map to hold the mocked storage
type DataMap map[lookup.EpochID]*Data

// StoreConfig allows to specify the simulated delays for each type of
// read operation
type StoreConfig struct {
	CacheReadTime      time.Duration // time it takes to read from the cache
	FailedReadTime     time.Duration // time it takes to acknowledge a read as failed
	SuccessfulReadTime time.Duration // time it takes to fetch data
}

// Store simulates a store and keeps track of performance counters
type Store struct {
	StoreConfig
	data        DataMap
	cache       DataMap
	lock        sync.RWMutex
	activeReads int
}

// NewStore returns a new mock store ready for use
func NewStore(config *StoreConfig) *Store {
	store := &Store{
		StoreConfig: *config,
		data:        make(DataMap),
	}

	store.Reset()
	return store
}

// Reset performance counters and clears the cache
func (s *Store) Reset() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cache = make(DataMap)
}

// KVPut stores a value in the mock store at the given epoch
func (s *Store) Put(epoch lookup.Epoch, value *Data) {
	s.data[epoch.ID()] = value
}

// Update runs the seed algorithm to place the update in the appropriate epoch
func (s *Store) Update(last lookup.Epoch, now uint64, value *Data) lookup.Epoch {
	epoch := lookup.GetNextEpoch(last, now)
	s.Put(epoch, value)
	return epoch
}

// KVGet retrieves data at the specified epoch, simulating a delay
func (s *Store) Get(ctx context.Context, epoch lookup.Epoch, now uint64) (value interface{}, err error) {
	epochID := epoch.ID()
	var operationTime time.Duration

	defer func() { // simulate a delay according to what has actually happened
		select {
		case <-lookup.TimeAfter(operationTime):
		case <-ctx.Done():
			value = nil
			err = ctx.Err()
		}
		s.lock.Lock()
		s.activeReads--
		s.lock.Unlock()
	}()

	s.lock.Lock()
	defer s.lock.Unlock()
	s.activeReads++

	// 1.- Simulate a cache read
	item := s.cache[epochID]
	operationTime += s.CacheReadTime

	if item != nil {
		if item.Time <= now {
			return item, nil
		}
		return nil, nil
	}

	// 2.- simulate a full read

	item = s.data[epochID]
	if item != nil {
		operationTime += s.SuccessfulReadTime
		s.cache[epochID] = item
		if item.Time <= now {
			return item, nil
		}
	} else {
		operationTime += s.FailedReadTime
	}
	return nil, nil
}

// MakeReadFunc returns a read function suitable for the lookup algorithm, mapped
// to this mock storage
func (s *Store) MakeReadFunc() lookup.ReadFunc {
	return s.Get
}
