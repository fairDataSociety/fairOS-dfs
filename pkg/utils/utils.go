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
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"os"
	"strconv"
	"strings"

	"github.com/ethersphere/bee/pkg/bmtpool"

	"github.com/ethersphere/bee/pkg/swarm"
	bmtlegacy "github.com/ethersphere/bmt/legacy"
	"golang.org/x/crypto/sha3"
)

const (
	MaxChunkLength   = 4096
	PathSeperator    = string(os.PathSeparator)
	MaxPodNameLength = 25
	SpanLength       = 8
)

type decError struct{ msg string }

func (err decError) Error() string { return err.msg }

var (
	ErrEmptyString   = &decError{"empty hex string"}
	ErrMissingPrefix = &decError{"hex string without 0x prefix"}
	ErrSyntax        = &decError{"invalid hex string"}
	ErrOddLength     = &decError{"hex string of odd length"}
	ErrUint64Range   = &decError{"hex number > 64 bits"}
)

// Encode encodes b as a hex string with 0x prefix.
func Encode(b []byte) string {
	enc := make([]byte, len(b)*2)
	hex.Encode(enc, b)
	return string(enc)
}

// Decode decodes a hex string with 0x prefix.
func Decode(input string) ([]byte, error) {
	if input == "" {
		return nil, ErrEmptyString
	}
	if !has0xPrefix(input) {
		return nil, ErrMissingPrefix
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
			return ErrUint64Range
		case strconv.ErrSyntax:
			return ErrSyntax
		}
	}
	if _, ok := err.(hex.InvalidByteError); ok {
		return ErrSyntax
	}
	if err == hex.ErrLength {
		return ErrOddLength
	}
	return err
}

func hashFunc() hash.Hash {
	return sha3.NewLegacyKeccak256()
}

func HashString(path string) []byte {
	p := bmtlegacy.NewTreePool(hashFunc, swarm.Branches, bmtlegacy.PoolSize)
	hasher := bmtlegacy.New(p)
	hasher.Reset()
	_, err := hasher.Write([]byte(path))
	if err != nil {
		return []byte{0}
	}
	return hasher.Sum(nil)
}

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
	err := hasher.SetSpanBytes(spanBytes)
	if err != nil {
		return nil, err
	}
	_, err = hasher.Write(data)
	if err != nil {
		return nil, err
	}
	s := hasher.Sum(nil)

	payload := append(spanBytes, data...)
	address := swarm.NewAddress(s)
	return swarm.NewChunk(address, payload), nil
}

func NewChunkWithoutSpan(data []byte) (swarm.Chunk, error) {
	if len(data) > swarm.ChunkSize+swarm.SpanSize {
		return nil, errors.New("max chunk size exceeded")
	}
	hasher := bmtpool.Get()
	defer bmtpool.Put(hasher)

	// execute hash, compare and return result
	err := hasher.SetSpanBytes(data[:swarm.SpanSize])
	if err != nil {
		return nil, err
	}
	_, err = hasher.Write(data[swarm.SpanSize:])
	if err != nil {
		return nil, err
	}
	s := hasher.Sum(nil)

	address := swarm.NewAddress(s)
	return swarm.NewChunk(address, data), nil
}

func CombinePathAndFile(podName, path, fileName string) string {
	var totalPath string

	if path == PathSeperator || path == "" {
		totalPath = path + fileName
	} else {
		if fileName == "" {
			totalPath = path
		} else {
			fileName = strings.TrimPrefix(fileName, PathSeperator)
			path = strings.TrimPrefix(path, PathSeperator)
			totalPath = PathSeperator + path + PathSeperator + fileName
		}
	}
	return totalPath
}

//func CombinePathAndFile(podName, path, fileName string) string {
//	var totalPath string
//	if path == PathSeperator || path == "" {
//		totalPath = PathSeperator + podName + path + fileName
//	} else {
//		if fileName != "" {
//			totalPath = PathSeperator + podName + PathSeperator + path + PathSeperator + fileName
//		} else {
//			totalPath = PathSeperator + podName + PathSeperator + path
//		}
//	}
//	return totalPath
//}
