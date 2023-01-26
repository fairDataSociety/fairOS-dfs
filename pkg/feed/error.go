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
	"fmt"
)

const (
	errInit = iota
	errNotFound
	errIO
	errUnauthorized
	errInvalidValue
	errDataOverflow
	errNothingToReturn
	errCorruptData
	errInvalidSignature
	errNotSynced
	errPeriodDepth
	errCnt
)

// Error is the typed error object used for Swarm feeds
type Error struct {
	code int
	err  string
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.err
}

// Code returns the error code
// Error codes are enumerated in the error.go file within the feeds package
// skipcq: TCV-001
func (e *Error) Code() int {
	return e.code
}

// NewError creates a new Swarm feeds Error object with the specified code and custom error message
func NewError(code int, s string) error {
	if code < 0 || code >= errCnt { // skipcq: TCV-001
		panic("no such error code!")
	}
	r := &Error{
		err: s,
	}
	switch code {
	case errNotFound, errIO, errUnauthorized, errInvalidValue, errDataOverflow, errNothingToReturn, errInvalidSignature, errNotSynced, errPeriodDepth, errCorruptData:
		r.code = code
	}
	return r
}

// NewErrorf is a convenience version of NewError that incorporates printf-style formatting
// skipcq: TCV-001
func NewErrorf(code int, format string, args ...interface{}) error {
	return NewError(code, fmt.Sprintf(format, args...))
}
