package file

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func (f *File) WriteAt(podFileWithPath string, update io.Reader, offset uint64) (int, error) {
	totalFilePath := utils.CombinePathAndFile(podFileWithPath, "")
	if !f.IsFileAlreadyPresent(totalFilePath) {
		return 0, ErrFileNotPresent
	}

	meta := f.GetFromFileMap(totalFilePath)
	if meta == nil { // skipcq: TCV-001
		return 0, ErrFileNotFound
	}

	fileInodeBytes, _, err := f.getClient().DownloadBlob(meta.InodeAddress)
	if err != nil { // skipcq: TCV-001
		return 0, err
	}
	var fileInode INode
	err = json.Unmarshal(fileInodeBytes, &fileInode)
	if err != nil { // skipcq: TCV-001
		return 0, err
	}
	fd := NewReader(fileInode, f.getClient(), meta.Size, meta.BlockSize, meta.Compression, false)
	reader := &bytes.Buffer{}
	_, err = reader.ReadFrom(fd)
	if err != nil {
		return 0, err
	}
	updater := &bytes.Buffer{}
	_, err = updater.ReadFrom(update)
	if err != nil {
		return 0, err
	}
	dataSize := uint64(reader.Len())
	updaterSize := uint64(updater.Len())
	if offset > dataSize {
		return 0, fmt.Errorf("wrong offset")
	}

	newDataSize := dataSize
	endofst := offset + updaterSize
	if endofst > dataSize {
		newDataSize = endofst
	}
	startingBlock := offset / uint64(meta.BlockSize)
	readStartPoint := startingBlock * uint64(meta.BlockSize)
	reader.Next(int(readStartPoint))
	blockOffset := offset - readStartPoint
	var totalLength uint64 = readStartPoint
	i := startingBlock
	errC := make(chan error)
	doneC := make(chan bool)
	worker := make(chan bool, noOfParallelWorkers)
	var wg sync.WaitGroup

	refMap := map[int]*BlockInfo{}
	for k, v := range fileInode.Blocks {
		refMap[k] = v
	}

	refMapMu := sync.RWMutex{}
	var contentBytes []byte
	wg.Add(1)
	go func() {
		var mainErr error
		for {
			if !(totalLength < newDataSize && updater.Len() != 0) {
				wg.Done()
				break
			}
			if mainErr != nil { // skipcq: TCV-001
				errC <- mainErr
				wg.Done()
				return
			}
			data := []byte{}
			n := 0
			var err error
			if totalLength < offset {
				temp := make([]byte, blockOffset)
				n, err = reader.Read(temp)
				if err != nil {
					if err == io.EOF {
						if totalLength < meta.Size { // skipcq: TCV-001
							errC <- fmt.Errorf("invalid file length of file data received")
							return
						}
						wg.Done()
						break
					}
					errC <- err // skipcq: TCV-001
					return
				}
				data = append(data, temp[:n]...)
				totalLength += uint64(n)
			}
			if totalLength >= offset && totalLength < endofst && uint32(len(data)) != meta.BlockSize {
				temp := make([]byte, meta.BlockSize-uint32(n))
				n, err = updater.Read(temp)
				if err != nil {
					if err == io.EOF {
						if totalLength < meta.Size { // skipcq: TCV-001
							errC <- fmt.Errorf("invalid file length of file data received")
							return
						}
						wg.Done()
						break
					}
					errC <- err // skipcq: TCV-001
					return
				}
				data = append(data, temp[:n]...)
				totalLength += uint64(n)
				if reader.Len() > 0 {
					reader.Next(n)
				}
			}

			if uint32(len(data)) != meta.BlockSize {
				if totalLength < dataSize && uint32(len(data)) != meta.BlockSize {
					temp := make([]byte, meta.BlockSize-uint32(len(data)))
					n, err = reader.Read(temp)
					if err != nil {
						if err == io.EOF {
							if totalLength < meta.Size { // skipcq: TCV-001
								errC <- fmt.Errorf("invalid file length of file data received")
								return
							}
							wg.Done()
							break
						}
						errC <- err // skipcq: TCV-001
						return
					}
					data = append(data, temp...)
					totalLength += uint64(n)
				}
			}
			// determine the content type from the first 512 bytes of the file
			if len(contentBytes) < 512 {
				contentBytes = append(contentBytes, data[:n]...)
				if len(contentBytes) >= 512 { // skipcq: TCV-001
					cBytes := bytes.NewReader(contentBytes[:512])
					cReader := bufio.NewReader(cBytes)
					meta.ContentType = f.getContentType(cReader)
				}
			}

			wg.Add(1)
			worker <- true
			go func(counter, size int) {
				blockName := fmt.Sprintf("block-%05d", counter)
				defer func() {
					<-worker
					wg.Done()
					if mainErr != nil { // skipcq: TCV-001
						return
					}
				}()
				f.logger.Info("Uploading ", blockName)
				// compress the data
				uploadData := data
				if meta.Compression != "" {
					uploadData, err = compress(data, meta.Compression, meta.BlockSize)
					if err != nil { // skipcq: TCV-001
						mainErr = err
						return
					}
				}

				addr, uploadErr := f.client.UploadBlob(uploadData, true, true)
				if uploadErr != nil {
					mainErr = uploadErr
					return
				}

				fileBlock := &BlockInfo{
					Name:           blockName,
					Size:           uint32(size),
					CompressedSize: uint32(len(uploadData)),
					Reference:      utils.NewReference(addr),
				}

				refMapMu.Lock()
				defer refMapMu.Unlock()
				refMap[counter] = fileBlock
			}(int(i), n)

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
		return 0, err
	}
	// copy the block references to the fileInode
	fileInode.Blocks = []*BlockInfo{}
	for i := 0; i < len(refMap); i++ {
		fileInode.Blocks = append(fileInode.Blocks, refMap[i])
	}
	fileInodeData, err := json.Marshal(fileInode)
	if err != nil { // skipcq: TCV-001
		return 0, err
	}

	addr, err := f.client.UploadBlob(fileInodeData, true, true)
	if err != nil { // skipcq: TCV-001
		return 0, err
	}

	meta.InodeAddress = addr
	meta.Size = newDataSize
	err = f.handleMeta(meta)
	if err != nil { // skipcq: TCV-001
		return 0, err
	}
	f.AddToFileMap(utils.CombinePathAndFile(meta.Path, meta.Name), meta)
	return int(updaterSize), nil
}
