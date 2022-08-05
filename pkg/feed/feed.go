// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package feed

import (
	"hash"
	"unsafe"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// Feed represents a particular user's stream of updates on a topic
type Feed struct {
	Topic Topic         `json:"topic"`
	User  utils.Address `json:"user"`
}

// Feed layout:
// TopicLength bytes
// userAddr common.AddressLength bytes
const feedLength = TopicLength + utils.AddressLength

// mapKey calculates a unique id for this feed. Used by the cache map in `Handler`
func (f *Feed) mapKey() (uint64, error) {
	serializedData := make([]byte, feedLength)
	err := f.binaryPut(serializedData)
	if err != nil { // skipcq: TCV-001
		return 0, err
	}
	hasher := hashPool.Get().(hash.Hash)
	defer hashPool.Put(hasher)
	hasher.Reset()
	_, err = hasher.Write(serializedData)
	if err != nil { // skipcq: TCV-001
		return 0, err
	}
	sumHash := hasher.Sum(nil)
	return *(*uint64)(unsafe.Pointer(&sumHash[0])), nil // skipcq: GSC-G103
}

// binaryPut serializes this feed instance into the provided slice
func (f *Feed) binaryPut(serializedData []byte) error {
	if len(serializedData) != feedLength { // skipcq: TCV-001
		return NewErrorf(errInvalidValue, "Incorrect slice size to serialize feed. Expected %d, got %d", feedLength, len(serializedData))
	}
	var cursor int
	copy(serializedData[cursor:cursor+TopicLength], f.Topic[:TopicLength])
	cursor += TopicLength

	copy(serializedData[cursor:cursor+utils.AddressLength], f.User[:])
	return nil
}
