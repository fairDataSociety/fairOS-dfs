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

package file

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/golang/snappy"
	lru "github.com/hashicorp/golang-lru/v2/simplelru"
	"github.com/klauspost/pgzip"
)

const (
	blockCacheSize = 500
)

var (
	// ErrInvalidOffset is returned when the offset is invalid
	ErrInvalidOffset = errors.New("invalid offset")
)

// Reader is a struct to read a file from the pod
type Reader struct {
	readOffset  int64
	client      blockstore.Client
	fileInode   INode
	fileC       chan []byte
	lastBlock   []byte
	fileSize    uint64
	blockSize   uint32
	blockCursor uint32
	totalSize   uint64
	compression string
	blockCache  *lru.LRU[string, []byte]

	rlBuffer      []byte
	rlOffset      int
	rlReadNewLine bool
}

// OpenFileForIndex opens file for indexing for document db from pod filepath
// TODO test
// skipcq: TCV-001
func (f *File) OpenFileForIndex(podFile, podPassword string) (*Reader, error) {
	meta := f.GetInode(podPassword, podFile)
	if meta == nil {
		return nil, ErrFileNotFound
	}

	encryptedFileInodeBytes, _, err := f.getClient().DownloadBlob(meta.InodeAddress)
	if err != nil {
		return nil, err
	}

	temp := make([]byte, len(encryptedFileInodeBytes))
	copy(temp, encryptedFileInodeBytes)
	fileInodeBytes, err := utils.DecryptBytes([]byte(podPassword), temp)
	if err != nil {
		return nil, err
	}

	var fileInode INode
	err = json.Unmarshal(fileInodeBytes, &fileInode)
	if err != nil {
		return nil, err
	}

	reader := NewReader(fileInode, f.getClient(), meta.Size, meta.BlockSize, meta.Compression, true)
	return reader, nil
}

// NewReader create a new reader object to read a file from the pod based on its configuration.
func NewReader(fileInode INode, client blockstore.Client, fileSize uint64, blockSize uint32, compression string, cache bool) *Reader {
	r := &Reader{
		fileInode:     fileInode,
		client:        client,
		fileC:         make(chan []byte),
		fileSize:      fileSize,
		blockSize:     blockSize,
		compression:   compression,
		rlReadNewLine: false,
	}
	if cache {
		r.blockCache, _ = lru.NewLRU[string, []byte](blockCacheSize, func(key string, value []byte) {})
	}
	return r
}

// Read reads a given segment of the file from the pod and returns it. it does all the
// related function like block extraction, block un-compression etc.
func (r *Reader) Read(b []byte) (n int, err error) {
	if r.totalSize >= r.fileSize {
		return 0, io.EOF
	}
	bytesToRead := uint32(len(b))
	bytesRead := 0
	if r.lastBlock != nil {
		remDataSize := r.blockSize - r.blockCursor
		if bytesToRead <= remDataSize {
			copy(b, r.lastBlock[r.blockCursor:r.blockCursor+bytesToRead])
			r.blockCursor += bytesToRead
			r.readOffset += int64(bytesToRead)
			bytesRead = int(bytesToRead)
			if r.blockCursor == r.blockSize {
				r.lastBlock = nil
				r.blockCursor = 0
			}
			r.totalSize += uint64(bytesToRead)
			return bytesRead, nil
		} else {
			copy(b, r.lastBlock[r.blockCursor:r.blockSize])
			r.lastBlock = nil
			r.blockCursor = 0
			r.readOffset += int64(remDataSize)
			bytesRead += int(remDataSize)
			bytesToRead -= remDataSize
			r.totalSize += uint64(remDataSize)

			// this situation comes when the block ends
			if r.totalSize >= r.fileSize {
				return bytesRead, io.EOF
			}

			// read spans across block. so flow down and read the next block
		}
	}
	if r.lastBlock == nil {
		noOfBlocks := int((bytesToRead / r.blockSize) + 1)
		for i := 0; i < noOfBlocks; i++ {
			if r.lastBlock == nil {
				blockIndex := r.readOffset / int64(r.blockSize)
				if blockIndex > int64(len(r.fileInode.Blocks)) {
					return bytesRead, io.EOF
				}
				if blockIndex >= int64(len(r.fileInode.Blocks)) { // skipcq: TCV-001
					return bytesRead, io.EOF
				}
				r.lastBlock, err = r.getBlock(r.fileInode.Blocks[blockIndex].Reference.Bytes(), r.compression, r.blockSize)
				if err != nil { // skipcq: TCV-001
					return bytesRead, err
				}
				r.blockSize = uint32(len(r.lastBlock))
			}

			// if length of bytes to read is greater than block size
			if bytesToRead > r.blockSize {
				bytesToRead = r.blockSize
			}

			if uint32(len(r.lastBlock)) < bytesToRead { // skipcq: TCV-001
				bytesToRead = uint32(len(r.lastBlock))
			}

			cursor := uint32(bytesRead)
			copy(b[cursor:cursor+bytesToRead], r.lastBlock[:bytesToRead])
			r.totalSize += uint64(bytesToRead)
			if bytesToRead == r.blockSize {
				r.lastBlock = nil
				r.blockCursor = 0
			} else {
				r.blockCursor += bytesToRead
			}
			r.readOffset += int64(bytesToRead)
			bytesRead += int(bytesToRead)
			bytesToRead -= bytesToRead

			if bytesToRead <= 0 {
				return bytesRead, nil
			}
		}
	}

	return 0, nil // skipcq: TCV-001
}

// Seek seeks to a given offset in the file and returns the offset.
func (r *Reader) Seek(seekOffset int64, _ int) (int64, error) {
	// TODO: use whence
	if seekOffset < 0 || seekOffset > int64(r.fileSize) {
		return 0, ErrInvalidOffset
	}

	// seek to start if offset is zero
	if seekOffset == 0 {
		blockData, err := r.getBlock(r.fileInode.Blocks[0].Reference.Bytes(), r.compression, r.blockSize)
		if err != nil { // skipcq: TCV-001
			return 0, err
		}
		r.lastBlock = blockData
		r.blockCursor = 0
		r.readOffset = 0
		r.blockSize = uint32(len(r.lastBlock))
		r.totalSize = 0
		r.rlBuffer = nil
		r.rlOffset = 0
		return 0, nil
	}

	blockIndex := seekOffset / int64(r.blockSize)
	blockOffset := seekOffset % int64(r.blockSize)

	blockData, err := r.getBlock(r.fileInode.Blocks[blockIndex].Reference.Bytes(), r.compression, r.blockSize)
	if err != nil {
		return 0, err
	}
	r.lastBlock = blockData
	r.blockCursor = uint32(blockOffset)
	r.readOffset = seekOffset
	r.blockSize = uint32(len(r.lastBlock))
	r.totalSize = uint64(seekOffset)
	r.rlBuffer = nil
	r.rlOffset = 0
	return seekOffset, nil
}

// ReadLine reads a line from the file
func (r *Reader) ReadLine() ([]byte, error) {
	if r.rlBuffer == nil {
		buf := make([]byte, r.blockSize)
		n, err := r.Read(buf)
		if err != nil { // skipcq: TCV-001
			if errors.Is(err, io.EOF) {
				if n == 0 || buf[n-1] != '\n' {
					return nil, err
				} else {
					goto SUCC
				}
			}
			return nil, err
		}
	SUCC:
		r.rlBuffer = buf[:n]
		r.rlOffset = 0
	}

	var destBuf []byte
	foundNewLine := false
READ:
	readOffset := r.rlOffset
	if readOffset != 0 || r.rlReadNewLine {
		readOffset += 1
		r.rlReadNewLine = false
	}
	for idx, char := range r.rlBuffer[readOffset:] {
		if char == '\n' {
			destBuf = append(destBuf, r.rlBuffer[readOffset:readOffset+idx+1]...)
			r.rlOffset = readOffset + idx
			foundNewLine = true
			// if the first byte is the new line
			if r.rlOffset == 0 && r.rlBuffer[0] == '\n' {
				r.rlReadNewLine = true
			}
			if len(r.rlBuffer) == readOffset+idx+1 {
				r.rlBuffer = nil
				r.rlOffset = 0
			}
			break
		}
	}

	// check if the newline is crossing the read buffer boundary
	if !foundNewLine {
		destBuf = append(destBuf, r.rlBuffer[readOffset:r.blockSize]...)
		if r.totalSize == r.fileSize {
			return destBuf, io.EOF
		}
		buf := make([]byte, r.blockSize)
		_, err := r.Read(buf)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		r.rlBuffer = buf
		r.rlOffset = 0
		goto READ
	}
	return destBuf, nil
}

// Close closes the reader
func (r *Reader) Close() error {
	if r.blockCache != nil {
		r.blockCache.Purge()
	}
	if r.rlBuffer != nil {
		r.rlBuffer = nil
	}
	return nil
}

func (r *Reader) getBlock(ref []byte, compression string, blockSize uint32) ([]byte, error) {
	refStr := utils.NewReference(ref).String()
	if r.blockCache != nil {
		if data, found := r.blockCache.Get(refStr); found {
			return data, nil
		}
	}
	stdoutBytes, _, err := r.client.DownloadBlob(ref)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	decompressedData, err := Decompress(stdoutBytes, compression, blockSize)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	if r.blockCache != nil {
		r.blockCache.Add(refStr, decompressedData)
	}
	return decompressedData, nil
}

// Decompress decompresses the data
func Decompress(dataToDecompress []byte, compression string, blockSize uint32) ([]byte, error) {
	switch compression {
	case "gzip":
		br := bytes.NewReader(dataToDecompress)
		block := int(blockSize / 10)
		r, err := pgzip.NewReaderN(br, block, 10)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		s, err := io.ReadAll(r)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		err = r.Close()
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		return s, nil
	case "snappy":
		decoded, err := snappy.Decode(nil, dataToDecompress)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		return decoded, nil
	}
	return dataToDecompress, nil
}
