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
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed/lookup"
)

// Query is used to specify constraints when performing an update lookup
// TimeLimit indicates an upper bound for the search. Set to 0 for "now"
type Query struct {
	Feed
	Hint      lookup.Epoch
	TimeLimit uint64
}

// NewQuery constructs an Query structure to find updates on or before `time`
// if time == 0, the latest update will be looked up
func NewQuery(feed *Feed, time uint64, hint lookup.Epoch) *Query {
	return &Query{
		TimeLimit: time,
		Feed:      *feed,
		Hint:      hint,
	}
}

// NewQueryLatest generates lookup parameters that look for the latest update to a feed
func NewQueryLatest(feed *Feed, hint lookup.Epoch) *Query {
	return NewQuery(feed, 0, hint)
}
