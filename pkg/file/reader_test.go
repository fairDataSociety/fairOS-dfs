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

package file_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
)

func TestFileReader(t *testing.T) {
	mockClient := mock.NewMockBeeClient()

	t.Run("read-entire-file", func(t *testing.T) {
		fileSize := uint64(100)
		blockSize := uint32(10)
		fileInode := createFile(t, fileSize, blockSize, "", mockClient)
		reader := file.NewReader(fileInode, mockClient, fileSize, blockSize, "", false)
		outputBytes := readFileContents(t, fileSize, reader)
		if !checkFileContents(t, fileInode, outputBytes, mockClient, "") {
			t.Fatalf("file contents are not same")
		}
	})

	t.Run("read-file-with-last-block-shorter", func(t *testing.T) {
		fileSize := uint64(93)
		blockSize := uint32(10)
		fileInode := createFile(t, fileSize, blockSize, "", mockClient)
		reader := file.NewReader(fileInode, mockClient, fileSize, blockSize, "", false)
		outputBytes := readFileContents(t, fileSize, reader)
		if !checkFileContents(t, fileInode, outputBytes, mockClient, "") {
			t.Fatalf("file contents are not same")
		}
	})

	t.Run("read-gzip-file", func(t *testing.T) {
		fileSize := uint64(1638500)
		blockSize := uint32(163850)
		compression := "gzip"
		fileInode := createFile(t, fileSize, blockSize, compression, mockClient)
		reader := file.NewReader(fileInode, mockClient, fileSize, blockSize, compression, false)
		outputBytes := readFileContents(t, fileSize, reader)
		if !checkFileContents(t, fileInode, outputBytes, mockClient, compression) {
			t.Fatalf("file contents are not same")
		}
	})

	t.Run("read-gzip-file-with-last-block-shorter", func(t *testing.T) {
		fileSize := uint64(1999000)
		blockSize := uint32(200000)
		compression := "gzip"
		fileInode := createFile(t, fileSize, blockSize, compression, mockClient)
		reader := file.NewReader(fileInode, mockClient, fileSize, blockSize, compression, false)
		outputBytes := readFileContents(t, fileSize, reader)
		if !checkFileContents(t, fileInode, outputBytes, mockClient, compression) {
			t.Fatalf("file contents are not same")
		}
	})

	t.Run("read-snappy-file", func(t *testing.T) {
		fileSize := uint64(100)
		blockSize := uint32(10)
		compression := "snappy"
		fileInode := createFile(t, fileSize, blockSize, compression, mockClient)
		reader := file.NewReader(fileInode, mockClient, fileSize, blockSize, compression, false)
		outputBytes := readFileContents(t, fileSize, reader)
		if !checkFileContents(t, fileInode, outputBytes, mockClient, compression) {
			t.Fatalf("file contents are not same")
		}
	})

	t.Run("read-snappy-file-with-last-block-shorter", func(t *testing.T) {
		fileSize := uint64(93)
		blockSize := uint32(10)
		compression := "snappy"
		fileInode := createFile(t, fileSize, blockSize, compression, mockClient)
		reader := file.NewReader(fileInode, mockClient, fileSize, blockSize, compression, false)
		outputBytes := readFileContents(t, fileSize, reader)
		if !checkFileContents(t, fileInode, outputBytes, mockClient, compression) {
			t.Fatalf("file contents are not same")
		}
	})

	t.Run("read-lines", func(t *testing.T) {
		fileSize := uint64(100)
		blockSize := uint32(10)
		linesPerBlock := uint32(2)
		fileInode, _, _, _, _ := createFileWithNewlines(t, fileSize, blockSize, "", mockClient, linesPerBlock)
		reader := file.NewReader(fileInode, mockClient, fileSize, blockSize, "", false)
		outputBytes := readFileContentsUsingReadline(t, fileSize, reader)
		if !checkFileContents(t, fileInode, outputBytes, mockClient, "") {
			t.Fatalf("file contents are not same")
		}
	})

	t.Run("read-lines-with-last-block-shorter", func(t *testing.T) {
		fileSize := uint64(97)
		blockSize := uint32(10)
		linesPerBlock := uint32(2)
		fileInode, _, _, _, _ := createFileWithNewlines(t, fileSize, blockSize, "", mockClient, linesPerBlock)
		reader := file.NewReader(fileInode, mockClient, fileSize, blockSize, "", false)
		outputBytes := readFileContentsUsingReadline(t, fileSize, reader)
		if !checkFileContents(t, fileInode, outputBytes, mockClient, "") {
			t.Fatalf("file contents are not same")
		}
	})

	t.Run("read-lines-with-last-block-shorter-and-compressed", func(t *testing.T) {
		fileSize := uint64(97)
		blockSize := uint32(10)
		linesPerBlock := uint32(2)
		compression := "snappy"
		fileInode, _, _, _, _ := createFileWithNewlines(t, fileSize, blockSize, compression, mockClient, linesPerBlock)
		reader := file.NewReader(fileInode, mockClient, fileSize, blockSize, compression, false)
		outputBytes := readFileContentsUsingReadline(t, fileSize, reader)
		if !checkFileContents(t, fileInode, outputBytes, mockClient, compression) {
			t.Fatalf("file contents are not same")
		}
	})

	t.Run("seek-and-read-line", func(t *testing.T) {
		fileSize := uint64(100)
		blockSize := uint32(10)
		linesPerBlock := uint32(2)
		fileInode, lineStart, line, _, _ := createFileWithNewlines(t, fileSize, blockSize, "", mockClient, linesPerBlock)
		reader := file.NewReader(fileInode, mockClient, fileSize, blockSize, "", false)
		seekN, err := reader.Seek(int64(lineStart), 0)
		if err != nil {
			t.Fatal(err)
		}
		if seekN != int64(lineStart) {
			t.Fatalf("did not seek to proper line start")
		}
		buf, err := reader.ReadLine()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf, line) {
			t.Fatalf("line contents are not same")
		}
	})

	t.Run("seek-and-read-line-spanning-block-boundary", func(t *testing.T) {
		fileSize := uint64(100)
		blockSize := uint32(10)
		linesPerBlock := uint32(2)
		fileInode, _, _, lineStart, line := createFileWithNewlines(t, fileSize, blockSize, "", mockClient, linesPerBlock)
		reader := file.NewReader(fileInode, mockClient, fileSize, blockSize, "", false)
		seekN, err := reader.Seek(int64(lineStart), 0)
		if err != nil {
			t.Fatal(err)
		}
		if seekN != int64(lineStart) {
			t.Fatalf("did not seek to proper line start")
		}
		buf, err := reader.ReadLine()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf, line) {
			t.Fatalf("line contents are not same")
		}
	})

	t.Run("seek-and-read-line-spanning-block-boundary-with-compression", func(t *testing.T) {
		fileSize := uint64(100)
		blockSize := uint32(10)
		linesPerBlock := uint32(2)
		compression := "snappy"
		fileInode, _, _, lineStart, line := createFileWithNewlines(t, fileSize, blockSize, compression, mockClient, linesPerBlock)
		reader := file.NewReader(fileInode, mockClient, fileSize, blockSize, compression, false)
		seekN, err := reader.Seek(int64(lineStart), 0)
		if err != nil {
			t.Fatal(err)
		}
		if seekN != int64(lineStart) {
			t.Fatalf("did not seek to proper line start")
		}
		buf, err := reader.ReadLine()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf, line) {
			t.Fatalf("line contents are not same")
		}
	})

	t.Run("seek-and-read-line-spanning-block-boundary-with-compression-with-cache", func(t *testing.T) {
		fileSize := uint64(100)
		blockSize := uint32(10)
		linesPerBlock := uint32(2)
		compression := "snappy"
		fileInode, _, _, lineStart, line := createFileWithNewlines(t, fileSize, blockSize, compression, mockClient, linesPerBlock)
		reader := file.NewReader(fileInode, mockClient, fileSize, blockSize, compression, true)
		seekN, err := reader.Seek(int64(lineStart), 0)
		if err != nil {
			t.Fatal(err)
		}
		if seekN != int64(lineStart) {
			t.Fatalf("did not seek to proper line start")
		}
		buf, err := reader.ReadLine()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf, line) {
			t.Fatalf("line contents are not same")
		}

		// this should come from cache
		seekN, err = reader.Seek(int64(lineStart), 0)
		if err != nil {
			t.Fatal(err)
		}
		if seekN != int64(lineStart) {
			t.Fatalf("did not seek to proper line start")
		}
		buf, err = reader.ReadLine()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf, line) {
			t.Fatalf("line contents are not same")
		}
	})
}

func createFile(t *testing.T, fileSize uint64, blockSize uint32, compression string, mockClient *mock.MockBeeClient) file.FileINode {
	var fileBlocks []*file.FileBlock
	noOfBlocks := fileSize / uint64(blockSize)
	if fileSize%uint64(blockSize) != 0 {
		noOfBlocks += 1
	}
	bytesRemaining := fileSize
	for i := uint64(0); i < noOfBlocks; i++ {
		bytesToWrite := blockSize
		if bytesRemaining < uint64(blockSize) {
			bytesToWrite = uint32(bytesRemaining)
		}
		buf := make([]byte, bytesToWrite)
		rand.Read(buf)
		if compression != "" {
			compressedData, err := file.Compress(buf, compression, bytesToWrite)
			if err != nil {
				t.Fatal(err)
			}
			buf = compressedData
		}

		addr, err := mockClient.UploadBlob(buf, true, true)
		if err != nil {
			t.Fatal(err)
		}
		blockName := fmt.Sprintf("block-%05d", i)
		fileBlock := &file.FileBlock{
			Name:           blockName,
			Size:           bytesToWrite,
			CompressedSize: uint32(len(buf)),
			Address:        addr,
		}
		fileBlocks = append(fileBlocks, fileBlock)
		bytesRemaining -= uint64(bytesToWrite)
	}

	return file.FileINode{
		FileBlocks: fileBlocks,
	}
}

func createFileWithNewlines(t *testing.T, fileSize uint64, blockSize uint32, compression string, mockClient *mock.MockBeeClient, linesPerBlock uint32) (file.FileINode, int, []byte, int, []byte) {
	var fileBlocks []*file.FileBlock
	noOfBlocks := fileSize / uint64(blockSize)
	if fileSize%uint64(blockSize) != 0 {
		noOfBlocks += 1
	}
	bytesRemaining := fileSize

	randomLineStartPoint := 0
	var randomLine []byte

	borderCrossingLineStartingPoint := 0
	var borderCrossingLine []byte

	bytesWritten := 0
	for i := uint64(0); i < noOfBlocks; i++ {
		bytesToWrite := blockSize
		if bytesRemaining < uint64(blockSize) {
			bytesToWrite = uint32(bytesRemaining)
		}
		buf := make([]byte, bytesToWrite)
		rand.Read(buf)

		for j := uint32(0); j < linesPerBlock; j++ {
			idx := rand.Intn(int(bytesToWrite))
			if buf[idx] == '\n' {
				idx = rand.Intn(int(bytesToWrite))
			}
			buf[idx] = '\n'
		}

		if i == 2 {
			start := false
			startIndex := 0
			endIndex := 0
			for k, ch := range buf {
				if ch == '\n' && start {
					endIndex = k + 1
				}
				if ch == '\n' && !start {
					startIndex = k + 1
					randomLineStartPoint = (int(blockSize) * int(i)) + startIndex
					start = true
				}
			}
			if startIndex > endIndex {
				startIndex, endIndex = endIndex, startIndex
			}
			randomLine = append(randomLine, buf[startIndex:endIndex]...)
		}

		gotFromFirstBlock := false
		if i >= 4 && borderCrossingLineStartingPoint == 0 && buf[int(blockSize)-1] != '\n' {
			gotFirstNewLine := false
			startIndex := 0
			for k, ch := range buf {
				if ch == '\n' {
					if !gotFirstNewLine {
						gotFirstNewLine = true
						continue
					} else {
						startIndex = k + 1
						borderCrossingLineStartingPoint = (int(blockSize) * int(i)) + startIndex
					}
				}
			}
			borderCrossingLine = append(borderCrossingLine, buf[startIndex:]...)
			gotFromFirstBlock = true
		}

		if i >= 4 && !gotFromFirstBlock && borderCrossingLine != nil && borderCrossingLine[len(borderCrossingLine)-1] != '\n' {
			endIndex := 0
			for k, ch := range buf {
				if ch == '\n' {
					endIndex = k + 1
					borderCrossingLine = append(borderCrossingLine, buf[:endIndex]...)
					break
				}
			}
		}

		if compression != "" {
			compressedData, err := file.Compress(buf, compression, bytesToWrite)
			if err != nil {
				t.Fatal(err)
			}
			buf = compressedData
		}

		addr, err := mockClient.UploadBlob(buf, true, true)
		if err != nil {
			t.Fatal(err)
		}
		blockName := fmt.Sprintf("block-%05d", i)
		fileBlock := &file.FileBlock{
			Name:           blockName,
			Size:           bytesToWrite,
			CompressedSize: uint32(len(buf)),
			Address:        addr,
		}
		fileBlocks = append(fileBlocks, fileBlock)
		bytesRemaining -= uint64(bytesToWrite)
		bytesWritten += int(bytesToWrite)
	}

	return file.FileINode{
		FileBlocks: fileBlocks,
	}, randomLineStartPoint, randomLine, borderCrossingLineStartingPoint, borderCrossingLine
}

func checkFileContents(t *testing.T, fileInode file.FileINode, outputBytes []byte, mockClient *mock.MockBeeClient, compression string) bool {
	var inpBuf []byte
	fileSize := uint32(0)
	for _, block := range fileInode.FileBlocks {
		buf, _, err := mockClient.DownloadBlob(block.Address)
		if err != nil {
			t.Fatal(err)
		}

		deflatedBuf, err := file.Decompress(buf, compression, block.Size)
		if err != nil {
			t.Fatal(err)
		}
		fileSize += block.Size
		inpBuf = append(inpBuf, deflatedBuf...)
	}

	inputBytes := make([]byte, fileSize)
	copy(inputBytes, inpBuf[:fileSize])

	for i := range inputBytes {
		if inputBytes[i] != outputBytes[i] {
			fmt.Println(i)
		}
	}
	return bytes.Equal(inputBytes, outputBytes)
}

func readFileContents(t *testing.T, fileSize uint64, reader *file.Reader) []byte {
	outputBytes := make([]byte, fileSize)
	count := uint64(0)
	for count < fileSize {
		n, err := reader.Read(outputBytes[count:])
		if err != nil {
			if !errors.Is(err, io.EOF) {
				t.Fatal(err)
			}
		}
		count += uint64(n)
	}
	return outputBytes
}

func readFileContentsUsingReadline(t *testing.T, fileSize uint64, reader *file.Reader) []byte {
	var outputBytes []byte
	count := uint64(0)
	for count < fileSize {
		buf, err := reader.ReadLine()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				t.Fatal(err)
			}
		}
		count += uint64(len(buf))
		outputBytes = append(outputBytes, buf...)
	}
	return outputBytes
}
