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
	"bytes"
	"context"
	"time"
)

const (
	hasherCount = 8

	defaultRetrieveTimeout = 2 * time.Second
)

// CacheEntry caches the last known update of a specific Swarm feed.
type CacheEntry struct {
	Update
	*bytes.Reader
	lastKey []byte
}

// Size implements storage.LazySectionReader
func (r *CacheEntry) Size(ctx context.Context, _ chan bool) (int64, error) {
	return int64(len(r.Update.data)), nil
}

// Topic returns the feed's topic
func (r *CacheEntry) Topic() Topic {
	return r.Feed.Topic
}
