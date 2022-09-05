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

package utils

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/ethersphere/bee/pkg/bmtpool"
	"github.com/ethersphere/bee/pkg/swarm"
	bmtlegacy "github.com/ethersphere/bmt/legacy"
	"golang.org/x/crypto/sha3"
)

const (
	// MaxChunkLength is the maximum size of a chunk
	MaxChunkLength = 4096

	// PathSeparator is string of os.PathSeparator
	PathSeparator = string(os.PathSeparator)

	// MaxPodNameLength defines how long a pod name can be
	MaxPodNameLength = 64

	// SpanLength of a chunk
	SpanLength = 8

	// DeletedFeedMagicWord is written in a feed after it gets deleted from fairOS
	DeletedFeedMagicWord = "__Fair__"

	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type decError struct{ msg string }

func (err decError) Error() string { return err.msg }

var (
	errEmptyString   = &decError{"empty hex string"}
	errMissingPrefix = &decError{"hex string without 0x prefix"}
	errSyntax        = &decError{"invalid hex string"}
	errOddLength     = &decError{"hex string of odd length"}
	errUint64Range   = &decError{"hex number > 64 bits"}
)

// Encode encodes b as a hex string with 0x prefix.
// skipcq: TCV-001
func Encode(b []byte) string {
	enc := make([]byte, len(b)*2)
	hex.Encode(enc, b)
	return string(enc)
}

// Decode decodes a hex string with 0x prefix.
func Decode(input string) ([]byte, error) {
	if input == "" {
		return nil, errEmptyString
	}
	if !has0xPrefix(input) {
		return nil, errMissingPrefix
	}
	b, err := hex.DecodeString(input[2:])
	if err != nil {
		err = mapError(err)
	}
	return b, err
}

func has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}

func mapError(err error) error {
	if err, ok := err.(*strconv.NumError); ok {
		switch err.Err {
		case strconv.ErrRange:
			return errUint64Range
		case strconv.ErrSyntax:
			return errSyntax
		}
	}
	if _, ok := err.(hex.InvalidByteError); ok {
		return errSyntax
	}
	if err == hex.ErrLength {
		return errOddLength
	}
	return err // skipcq: TCV-001
}

// skipcq: TCV-001
func hashFunc() hash.Hash {
	return sha3.NewLegacyKeccak256()
}

// HashString returns the bmt hash of a string
// skipcq: TCV-001
func HashString(path string) []byte {
	p := bmtlegacy.NewTreePool(hashFunc, swarm.Branches, bmtlegacy.PoolSize)
	hasher := bmtlegacy.New(p)
	hasher.Reset()
	_, err := hasher.Write([]byte(path))
	if err != nil { // skipcq: TCV-001
		return []byte{0}
	}
	return hasher.Sum(nil) // skipcq: TCV-001
}

// NewChunkWithSpan returns a chunk with span
func NewChunkWithSpan(data []byte) (swarm.Chunk, error) {
	span := int64(len(data))

	if len(data) > swarm.ChunkSize {
		return nil, errors.New("max chunk size exceeded")
	}
	if span < swarm.ChunkSize && span != int64(len(data)) {
		return nil, fmt.Errorf("single-span chunk size mismatch; span is %d, chunk data length %d", span, len(data))
	}

	hasher := bmtpool.Get()
	defer bmtpool.Put(hasher)

	// execute hash, compare and return result
	spanBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(spanBytes, uint64(span))
	hasher.SetHeader(spanBytes)
	_, err := hasher.Write(data)
	if err != nil {
		return nil, err
	}
	s := hasher.Sum(nil)

	payload := append(spanBytes, data...) // skipcq: CRT-D0001
	address := swarm.NewAddress(s)
	return swarm.NewChunk(address, payload), nil
}

// NewChunkWithoutSpan returns a chunk without span
func NewChunkWithoutSpan(data []byte) (swarm.Chunk, error) {
	if len(data) > swarm.ChunkSize+swarm.SpanSize {
		return nil, errors.New("max chunk size exceeded")
	}
	hasher := bmtpool.Get()
	defer bmtpool.Put(hasher)

	// execute hash, compare and return result
	hasher.SetHeader(data[:swarm.SpanSize])
	_, err := hasher.Write(data[swarm.SpanSize:])
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	s := hasher.Sum(nil)

	address := swarm.NewAddress(s)
	return swarm.NewChunk(address, data), nil
}

// CombinePathAndFile joins filename with provided path
func CombinePathAndFile(path, fileName string) string {
	var totalPath string

	if path == PathSeparator || path == "" {
		fileName = strings.TrimPrefix(fileName, PathSeparator)
		totalPath = PathSeparator + fileName
	} else {
		if fileName == "" {
			totalPath = path
		} else {
			fileName = strings.TrimPrefix(fileName, PathSeparator)
			path = strings.TrimPrefix(path, PathSeparator)
			totalPath = PathSeparator + path + PathSeparator + fileName
		}
	}
	return totalPath
}

// GetRandString return random string of length n
func GetRandString(n int) (string, error) {
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		if err != nil { // skipcq: TCV-001
			return "", err
		}
		b[i] = letterBytes[num.Int64()]
	}
	return string(b), nil
}

// GetRandBytes return random bytes array of length n
func GetRandBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		b[i] = letterBytes[num.Int64()]
	}
	return b, nil
}
