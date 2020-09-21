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
	"fmt"
	"io"
	"io/ioutil"

	"github.com/golang/snappy"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/klauspost/pgzip"
)

type Reader struct {
	offset      int64
	client      blockstore.Client
	fileInode   FileINode
	fileC       chan []byte
	lastBlock   []byte
	fileSize    uint64
	blockSize   uint32
	blockCursor uint32
	totalSize   uint64
	compression string
}

func NewReader(fileInode FileINode, client blockstore.Client, fileSize uint64, blockSize uint32, compression string) *Reader {
	r := &Reader{
		fileInode:   fileInode,
		client:      client,
		fileC:       make(chan []byte),
		fileSize:    fileSize,
		blockSize:   blockSize,
		compression: compression,
	}
	return r
}

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
			r.offset += int64(bytesToRead)
			bytesRead = int(bytesToRead)
			//bytesToRead = 0
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
			r.offset += int64(remDataSize)
			bytesRead += int(remDataSize)
			bytesToRead -= remDataSize
			r.totalSize += uint64(remDataSize)

			// this situation comes when the block ends
			if r.totalSize >= r.fileSize {
				fmt.Println("returning 2", bytesRead, r.blockCursor, r.totalSize, r.fileSize)
				return bytesRead, io.EOF
			}

			// read spans across block.. so flow down and read the next block
		}
	}

	if r.lastBlock == nil {
		noOfBlocks := int((bytesToRead / r.blockSize) + 1)
		for i := 0; i < noOfBlocks; i++ {
			if r.lastBlock == nil {
				blockIndex := (r.offset / int64(r.blockSize))
				if blockIndex > int64(len(r.fileInode.FileBlocks)) {
					return bytesRead, fmt.Errorf("asking past EOF")
				}
				if blockIndex >= int64(len(r.fileInode.FileBlocks)) {
					return 0, io.EOF
				}
				r.lastBlock, err = r.getBlock(r.fileInode.FileBlocks[blockIndex].Address, r.compression, r.blockSize)
				if err != nil {
					return bytesRead, err
				}
				r.blockSize = uint32(len(r.lastBlock))
			}

			// if length of bytes to read is greater than block size
			if bytesToRead > r.blockSize {
				bytesToRead = r.blockSize
			}

			if uint32(len(r.lastBlock)) < bytesToRead {
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
			r.offset += int64(bytesToRead)
			bytesRead += int(bytesToRead)
			bytesToRead -= bytesToRead

			if bytesToRead <= 0 {
				return bytesRead, nil
			}
		}
	}
	return 0, nil
}

func (r *Reader) getBlock(addr []byte, compression string, blockSize uint32) ([]byte, error) {
	stdoutBytes, _, err := r.client.DownloadBlob(addr)
	if err != nil {
		return nil, err
	}
	decompressedData, err := decompress(stdoutBytes, compression, blockSize)
	if err != nil {
		return nil, err
	}
	return decompressedData, nil
}

func (r *Reader) Close() error {
	return nil
}

func decompress(dataToDecompress []byte, compression string, blockSize uint32) ([]byte, error) {
	switch compression {
	case "gzip":
		br := bytes.NewReader(dataToDecompress)
		block := int(blockSize / 10)
		r, err := pgzip.NewReaderN(br, block, 10)
		if err != nil {
			return nil, err
		}
		s, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		err = r.Close()
		if err != nil {
			return nil, err
		}
		return s, nil
	case "snappy":
		decoded, err := snappy.Decode(nil, dataToDecompress)
		if err != nil {
			return nil, err
		}
		return decoded, nil
	}
	return dataToDecompress, nil
}
