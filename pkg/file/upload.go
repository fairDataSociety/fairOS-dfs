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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/ethersphere/bee/v2/pkg/swarm"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/golang/snappy"
	"github.com/klauspost/pgzip"
)

const (
	MinBlockSize uint32 = 1000000
	MaxBlockSize uint32 = 8000000

	// S_IFREG is the regular file type
	S_IFREG     = 0100000
	defaultMode = 0600
)

var (
	noOfParallelWorkers = runtime.NumCPU()

	ErrInvalidBlockSize = fmt.Errorf("upload: block size must be between %d - %d (1Mb - 8Mb)", MinBlockSize, MaxBlockSize)
)

// Upload uploads a given blob of bytes as a file in the pod. It also splits the file into number of blocks. the
// size of the block is provided during upload. This function also does compression of the blocks gzip/snappy if it is
// requested during the upload.
func (f *File) Upload(fd io.Reader, podFileName string, fileSize int64, blockSize, mode uint32, podPath, compression, podPassword string) error {
	podPath = filepath.ToSlash(podPath)
	if blockSize < MinBlockSize || blockSize > MaxBlockSize {
		return ErrInvalidBlockSize
	}
	reader := bufio.NewReader(fd)
	now := time.Now().Unix()
	tag, err := f.client.CreateTag(swarm.ZeroAddress)
	if err != nil { // skipcq: TCV-001
		return err
	}

	if mode == 0 {
		mode = S_IFREG | defaultMode
	}
	meta := MetaData{
		Version:          MetaVersion,
		Path:             podPath,
		Name:             podFileName,
		Size:             uint64(fileSize),
		BlockSize:        blockSize,
		Compression:      compression,
		CreationTime:     now,
		AccessTime:       now,
		ModificationTime: now,
		Mode:             mode,
	}
	var totalLength uint64
	i := 0
	errC := make(chan error)
	doneC := make(chan bool)
	worker := make(chan bool, noOfParallelWorkers)
	var wg sync.WaitGroup
	refMap := make(map[int]*BlockInfo)
	refMapMu := sync.RWMutex{}
	var contentBytes []byte
	wg.Add(1)
	go func() {
		var mainErr error
		for {
			if mainErr != nil { // skipcq: TCV-001
				errC <- mainErr
				wg.Done()
				return
			}
			data := make([]byte, blockSize, blockSize+1024)
			r, err := reader.Read(data)
			totalLength += uint64(r)
			if err != nil {
				if err == io.EOF {
					if totalLength < uint64(fileSize) { // skipcq: TCV-001
						errC <- fmt.Errorf("invalid file length of file data received")
						return
					}
					wg.Done()
					break
				}
				errC <- err // skipcq: TCV-001
				return
			}

			// determine the content type from the first 512 bytes of the file
			if len(contentBytes) < 512 {
				contentBytes = append(contentBytes, data[:r]...)
				if len(contentBytes) >= 512 { // skipcq: TCV-001
					cBytes := bytes.NewReader(contentBytes[:512])
					cReader := bufio.NewReader(cBytes)
					meta.ContentType = f.getContentType(cReader)
				}
			}

			wg.Add(1)
			worker <- true
			go func(counter, size int) {
				defer func() {
					<-worker
					wg.Done()
					if mainErr != nil { // skipcq: TCV-001
						f.logger.Error("failed uploading block ", counter)
						return
					}
					f.logger.Info("done uploading block ", counter)
				}()

				f.logger.Infof("Uploading %d block", counter)
				// Compress the data
				uploadData := data[:size]
				if compression != "" {
					uploadData, err = Compress(data[:size], compression, blockSize)
					if err != nil { // skipcq: TCV-001
						mainErr = err
						return
					}
				}
				addr, uploadErr := f.client.UploadBlob(tag, "", "0", false, true, bytes.NewReader(uploadData))
				if uploadErr != nil {
					mainErr = uploadErr
					return
				}

				fileBlock := &BlockInfo{
					Size:           uint32(size),
					CompressedSize: uint32(len(uploadData)),
					Reference:      utils.NewReference(addr.Bytes()),
				}

				refMapMu.Lock()
				defer refMapMu.Unlock()
				refMap[counter] = fileBlock
			}(i, r)

			i++
		}
	}()

	go func() {
		wg.Wait()
		close(doneC)
	}()

	select {
	case <-doneC:
		break
	case err := <-errC: // skipcq: TCV-001
		close(errC)
		return err
	}

	// copy the block references to the fileInode
	fileINode := INode{}
	for i := 0; i < len(refMap); i++ {
		fileINode.Blocks = append(fileINode.Blocks, refMap[i])
	}
	fileInodeData, err := json.Marshal(fileINode)
	if err != nil { // skipcq: TCV-001
		return err
	}
	addr, err := f.client.UploadBlob(tag, "", "0", false, true, bytes.NewReader(fileInodeData))
	if err != nil { // skipcq: TCV-001
		return err
	}
	meta.InodeAddress = addr.Bytes()
	err = f.handleMeta(&meta, podPassword)
	if err != nil { // skipcq: TCV-001
		return err
	}
	totalPath := utils.CombinePathAndFile(meta.Path, meta.Name)
	f.AddToFileMap(totalPath, &meta)
	if tag > 0 {
		f.AddToTagMap(totalPath, tag)
	}
	return nil
}

// skipcq: TCV-001
func (*File) getContentType(bufferReader *bufio.Reader) string {
	buffer, err := bufferReader.Peek(512)
	if err != nil && err != io.EOF {
		return ""
	}
	return http.DetectContentType(buffer)
}

// Compress data
func Compress(dataToCompress []byte, compression string, blockSize uint32) ([]byte, error) {
	switch compression {
	case "gzip":
		var b bytes.Buffer
		w := pgzip.NewWriter(&b)
		block := int(blockSize / 10)
		err := w.SetConcurrency(block, 10)
		if err != nil {
			return nil, err
		}
		_, err = w.Write(dataToCompress)
		if err != nil { // skipcq: TCV-001
			return nil, err
		}
		err = w.Close()
		if err != nil {
			return nil, err
		}
		return b.Bytes(), nil
	case "snappy":
		return snappy.Encode(nil, dataToCompress), nil
	}
	return dataToCompress, nil
}
