package file_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
)

func TestWriteAt(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	pod1AccountInfo, err := acc.CreatePodAccount(1, false)
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(pod1AccountInfo, mockClient, logger)
	user := acc.GetAddress(1)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	podPassword, _ := utils.GetRandString(pod.PasswordLength)

	t.Run("writeAt-non-existent-file", func(t *testing.T) {
		filePath := string(os.PathSeparator)
		fileName := "file1"

		var offset uint64 = 3

		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)

		fp := utils.CombinePathAndFile(filepath.ToSlash(filePath+fileName), "")

		update := []byte("123")
		rewrite := &bytes.Buffer{}
		rewrite.Write(update)
		_, err = fileObject.WriteAt(fp, podPassword, rewrite, offset, false)
		if !errors.Is(file.ErrFileNotPresent, err) {
			t.Fatal("file should not be present")
		}
	})

	t.Run("upload-update-known-very-small-file", func(t *testing.T) {
		filePath := string(os.PathSeparator)
		fileName := "file1"
		compression := ""
		blockSize := uint32(10)
		var offset uint64 = 3

		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)
		dt, err := uploadFileKnownContent(t, fileObject, filePath, fileName, compression, podPassword, blockSize)
		if err != nil {
			t.Fatal(err)
		}

		fp := utils.CombinePathAndFile(filepath.ToSlash(filePath+fileName), "")
		// check for meta
		meta := fileObject.GetFromFileMap(fp)
		if meta == nil {
			t.Fatalf("file not added in file map")
		}

		// validate meta items
		if meta.Path != filepath.ToSlash(filePath) {
			t.Fatalf("invalid path in meta")
		}
		if meta.Name != fileName {
			t.Fatalf("invalid file name in meta")
		}
		if meta.Size != uint64(len(dt)) {
			t.Fatalf("invalid file size in meta")
		}
		if meta.BlockSize != blockSize {
			t.Fatalf("invalid block size in meta")
		}
		reader, _, err := fileObject.Download(fp, podPassword)
		if err != nil {
			t.Fatal(err)
		}
		rcvdBuffer := new(bytes.Buffer)
		_, err = rcvdBuffer.ReadFrom(reader)
		if err != nil {
			t.Fatal(err)
		}
		reader.Close()
		reader2, _, err := fileObject.Download(fp, podPassword)
		if err != nil {
			t.Fatal(err)
		}
		rcvdBuffer2 := new(bytes.Buffer)
		_, err = rcvdBuffer2.ReadFrom(reader2)
		if err != nil {
			t.Fatal(err)
		}
		reader, _, err = fileObject.Download(fp, podPassword)
		if err != nil {
			t.Fatal(err)
		}

		rcvdBuffer3 := new(bytes.Buffer)
		_, err = rcvdBuffer3.ReadFrom(reader)
		if err != nil {
			t.Fatal(err)
		}

		update := []byte("123")
		rewrite := &bytes.Buffer{}
		rewrite.Write(update)
		_, err = fileObject.WriteAt(fp, podPassword, rewrite, offset, false)
		if err != nil {
			t.Fatal(err)
		}
		reader, _, err = fileObject.Download(fp, podPassword)
		if err != nil {
			t.Fatal(err)
		}
		rcvdBuffer = new(bytes.Buffer)
		_, err = rcvdBuffer.ReadFrom(reader)
		if err != nil {
			t.Fatal(err)
		}

		updatedContent := append(dt[:offset], update...)

		if uint64(len(update))+offset < uint64(len(dt)) {
			updatedContent = append(updatedContent, dt[uint64(len(update))+offset:]...)
		}

		if !bytes.Equal(updatedContent, rcvdBuffer.Bytes()) {
			t.Fatal("content is different")
		}

		fileObject.RemoveAllFromFileMap()

		meta2 := fileObject.GetFromFileMap(fp)
		if meta2 != nil {
			t.Fatal("meta2 should be nil")
		}
	})

	t.Run("upload-update-truncate-known-very-small-file", func(t *testing.T) {
		filePath := string(os.PathSeparator)
		fileName := "file1"
		compression := ""
		blockSize := uint32(20)
		var offset uint64 = 0

		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)
		dt, err := uploadFileKnownContent(t, fileObject, filePath, fileName, compression, podPassword, blockSize)
		if err != nil {
			t.Fatal(err)
		}

		// check for meta
		fp := utils.CombinePathAndFile(filepath.ToSlash(filePath+fileName), "")
		meta := fileObject.GetFromFileMap(fp)
		if meta == nil {
			t.Fatalf("file not added in file map")
		}

		// validate meta items
		if meta.Path != filepath.ToSlash(filePath) {
			t.Fatalf("invalid path in meta")
		}
		if meta.Name != fileName {
			t.Fatalf("invalid file name in meta")
		}
		if meta.Size != uint64(len(dt)) {
			t.Fatalf("invalid file size in meta")
		}
		if meta.BlockSize != blockSize {
			t.Fatalf("invalid block size in meta")
		}

		reader, _, err := fileObject.Download(fp, podPassword)
		if err != nil {
			t.Fatal(err)
		}
		rcvdBuffer := new(bytes.Buffer)
		_, err = rcvdBuffer.ReadFrom(reader)
		if err != nil {
			t.Fatal(err)
		}

		update := []byte("abcdefg 12345")
		rewrite := &bytes.Buffer{}
		rewrite.Write(update)
		_, err = fileObject.WriteAt(fp, podPassword, rewrite, offset, true)
		if err != nil {
			t.Fatal(err)
		}

		reader, _, err = fileObject.Download(fp, podPassword)
		if err != nil {
			t.Fatal(err)
		}
		rcvdBuffer = new(bytes.Buffer)
		_, err = rcvdBuffer.ReadFrom(reader)
		if err != nil {
			t.Fatal(err)
		}
		updatedContent := append(dt[:offset], update...)
		if !bytes.Equal(updatedContent, rcvdBuffer.Bytes()) {
			t.Fatal("content is different")
		}

		fileObject.RemoveAllFromFileMap()

		meta2 := fileObject.GetFromFileMap(fp)
		if meta2 != nil {
			t.Fatal("meta2 should be nil")
		}
	})

	t.Run("upload-update-small-file", func(t *testing.T) {
		filePath := "/dir1"
		fileName := "file1"
		compression := ""
		fileSize := int64(100)
		blockSize := uint32(10)
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)
		dt, err := uploadFile(t, fileObject, filePath, fileName, compression, podPassword, fileSize, blockSize)
		if err != nil {
			t.Fatal(err)
		}

		// check for meta
		fp := utils.CombinePathAndFile(filepath.ToSlash(filePath), fileName)
		meta := fileObject.GetFromFileMap(fp)
		if meta == nil {
			t.Fatalf("file not added in file map")
		}

		// validate meta items
		if meta.Path != filepath.ToSlash(filePath) {
			t.Fatalf("invalid path in meta")
		}
		if meta.Name != fileName {
			t.Fatalf("invalid file name in meta")
		}
		if meta.Size != uint64(fileSize) {
			t.Fatalf("invalid file size in meta")
		}
		if meta.BlockSize != blockSize {
			t.Fatalf("invalid block size in meta")
		}

		err = fileObject.LoadFileMeta(filePath+"/"+fileName, podPassword)
		if err != nil {
			t.Fatal(err)
		}

		rand.Seed(time.Now().UnixNano())
		min := 0
		max := int(fileSize)
		offset := rand.Intn((max - min + 1) + min)
		content, err := utils.GetRandBytes(offset)
		if err != nil {
			t.Fatal(err)
		}
		r := bytes.NewReader(content)
		n, err := fileObject.WriteAt(fp, podPassword, r, uint64(offset), false)
		if n != offset {
			t.Fatalf("Failed to update %d bytes", offset-n)
		}
		if err != nil {
			t.Fatal(err)
		}
		reader, _, err := fileObject.Download(fp, podPassword)
		if err != nil {
			t.Fatal(err)
		}
		rcvdBuffer := new(bytes.Buffer)
		_, err = rcvdBuffer.ReadFrom(reader)
		if err != nil {
			t.Fatal(err)
		}
		updatedContent := append(dt[:offset], content...)

		if uint64(len(content)+offset) < uint64(len(dt)) {
			updatedContent = append(updatedContent, dt[uint64(len(content)+offset):]...)
		}

		if !bytes.Equal(updatedContent, rcvdBuffer.Bytes()) {
			t.Fatal("content is different")
		}
		fileObject.RemoveAllFromFileMap()

		meta2 := fileObject.GetFromFileMap(fp)
		if meta2 != nil {
			t.Fatal("meta2 should be nil")
		}
	})

	t.Run("upload-update-small-file-at-root-with-prefix-snappy", func(t *testing.T) {
		filePath := string(os.PathSeparator)
		fileName := "file2"
		compression := "snappy"
		fileSize := int64(100)
		blockSize := uint32(10)
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)
		dt, err := uploadFile(t, fileObject, filePath, fileName, compression, podPassword, fileSize, blockSize)
		if err != nil {
			t.Fatal(err)
		}

		// check for meta
		fp := utils.CombinePathAndFile(filepath.ToSlash(filePath), fileName)
		meta := fileObject.GetFromFileMap(fp)
		if meta == nil {
			t.Fatalf("file not added in file map")
		}

		// validate meta items
		if meta.Path != filepath.ToSlash(filePath) {
			t.Fatalf("invalid path in meta")
		}
		if meta.Name != fileName {
			t.Fatalf("invalid file name in meta")
		}
		if meta.Size != uint64(fileSize) {
			t.Fatalf("invalid file size in meta")
		}
		if meta.BlockSize != blockSize {
			t.Fatalf("invalid block size in meta")
		}

		rand.Seed(time.Now().UnixNano())
		min := 0
		max := int(fileSize)
		offset := rand.Intn((max - min + 1) + min)
		content, err := utils.GetRandBytes(offset)
		if err != nil {
			t.Fatal(err)
		}
		r := bytes.NewReader(content)
		n, err := fileObject.WriteAt(fp, podPassword, r, uint64(offset), false)
		if n != offset {
			t.Fatalf("Failed to update %d bytes", offset-n)
		}
		if err != nil {
			t.Fatal(err)
		}
		reader, _, err := fileObject.Download(fp, podPassword)
		if err != nil {
			t.Fatal(err)
		}
		rcvdBuffer := new(bytes.Buffer)
		_, err = rcvdBuffer.ReadFrom(reader)
		if err != nil {
			t.Fatal(err)
		}
		updatedContent := append(dt[:offset], content...)

		if uint64(len(content)+offset) < uint64(len(dt)) {
			updatedContent = append(updatedContent, dt[uint64(len(content)+offset):]...)
		}

		if !bytes.Equal(updatedContent, rcvdBuffer.Bytes()) {
			t.Fatal("content is different")
		}

		fileObject.RemoveAllFromFileMap()

		meta2 := fileObject.GetFromFileMap(fp)
		if meta2 != nil {
			t.Fatal("meta2 should be nil")
		}
	})

	t.Run("upload-update-small-file-at-root-with-prefix-gzip", func(t *testing.T) {
		filePath := "/dir1"
		fileName := "file10"
		compression := "gzip"
		fileSize := int64(100)
		blockSize := uint32(164000)
		fileObject := file.NewFile("pod1", mockClient, fd, user, tm, logger)
		dt, err := uploadFile(t, fileObject, filePath, fileName, compression, podPassword, fileSize, blockSize)
		if err != nil {
			t.Fatal(err)
		}
		err = fileObject.LoadFileMeta(filePath+"/"+fileName, podPassword)
		if err != nil {
			t.Fatal(err)
		}

		// check for meta
		fp := utils.CombinePathAndFile(filepath.ToSlash(filePath), fileName)
		meta := fileObject.GetFromFileMap(fp)
		if meta == nil {
			t.Fatalf("file not added in file map")
		}

		// validate meta items
		if meta.Path != filepath.ToSlash(filePath) {
			t.Fatalf("invalid path in meta")
		}
		if meta.Name != fileName {
			t.Fatalf("invalid file name in meta")
		}
		if meta.Size != uint64(fileSize) {
			t.Fatalf("invalid file size in meta")
		}
		if meta.BlockSize != blockSize {
			t.Fatalf("invalid block size in meta")
		}

		rand.Seed(time.Now().UnixNano())
		min := 0
		max := int(fileSize)
		offset := rand.Intn((max - min + 1) + min)
		content, err := utils.GetRandBytes(offset)
		if err != nil {
			t.Fatal(err)
		}
		r := bytes.NewReader(content)
		_, err = fileObject.WriteAt(fp, podPassword, r, uint64(offset), false)
		if err != nil {
			t.Fatal(err)
		}
		reader, n1, err := fileObject.Download(fp, podPassword)
		if err != nil {
			t.Fatal(err)
		}
		rcvdBuffer := new(bytes.Buffer)
		_, err = rcvdBuffer.ReadFrom(reader)
		if err != nil {
			t.Fatal(err)
		}
		updatedContent := append(dt[:offset], content...)

		if uint64(len(content)+offset) < uint64(len(dt)) {
			updatedContent = append(updatedContent, dt[uint64(len(content)+offset):]...)
		}

		if !bytes.Equal(updatedContent, rcvdBuffer.Bytes()[:n1]) {
			t.Log("updatedContent", updatedContent)
			t.Log("downloadedContent", rcvdBuffer.Bytes())
			t.Fatal("content is different ")
		}

		fileObject.RemoveAllFromFileMap()

		meta2 := fileObject.GetFromFileMap(fp)
		if meta2 != nil {
			t.Fatal("meta2 should be nil")
		}
	})
}

func uploadFileKnownContent(t *testing.T, fileObject *file.File, filePath, fileName, compression, podPassword string, blockSize uint32) ([]byte, error) {
	f1 := &bytes.Buffer{}
	content := []byte("abcdefghijk abcdefghijk abcdefghijk")
	_, err := f1.Write(content)
	if err != nil {
		t.Fatal(err)
	}
	// upload  the temp file
	return content, fileObject.Upload(f1, fileName, int64(len(content)), blockSize, filePath, compression, podPassword)
}
