package user

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"

	"github.com/akrylysov/pogreb"
	"github.com/akrylysov/pogreb/fs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	dbFileName = "timeKeeper.db"
)

type FeedTracker struct {
	db     *pogreb.DB
	client blockstore.Client
}

func (*Users) initFeedsTracker(address utils.Address, username, password string, fd *feed.API, client blockstore.Client) (*pogreb.DB, error) {
	files := map[string]*memFile{}
	opts := &pogreb.Options{
		FileSystem:                   &memFS{files: files, fd: fd, username: username, password: password, address: address, client: client},
		BackgroundSyncInterval:       time.Minute * 5,
		BackgroundCompactionInterval: time.Minute,
	}

	db, err := pogreb.Open(dbFileName, opts)
	if err != nil {
		return nil, err
	}

	return db, nil
}

type memFS struct {
	files    map[string]*memFile
	fd       *feed.API
	client   blockstore.Client
	address  utils.Address
	username string
	password string
}

var (
	errAppendModeNotSupported = errors.New("append mode is not supported")
)

// Mem is a file system backed by memory.
var Mem fs.FileSystem = &memFS{files: map[string]*memFile{}}

func (fs *memFS) OpenFile(name string, flag int, perm os.FileMode) (fs.File, error) {
	if flag&os.O_APPEND != 0 {
		// memFS doesn't support opening files in append-only mode.
		// The database doesn't currently use O_APPEND.
		return nil, errAppendModeNotSupported
	}
	f := fs.files[name]
	if f == nil || (flag&os.O_TRUNC) != 0 {
		f = &memFile{
			name: name,
			perm: perm, // Perm is saved to return it in Mode, but don't do anything else with it yet.
		}
		fs.files[name] = f
	} else if !f.closed {
		return nil, os.ErrExist
	} else {
		f.offset = 0
		f.closed = false
	}
	f.fd = fs.fd
	f.username = fs.username
	f.password = fs.password
	f.address = fs.address
	f.client = fs.client
	topic := utils.HashString(name + fs.username + fs.password)
	_, ref, err := fs.fd.GetFeedData(topic, fs.address, []byte(fs.password), true)
	if err != nil && err.Error() != "feed does not exist or was not updated yet" {
		return nil, err
	}
	data, _, err := fs.client.DownloadBlob(ref)
	if err != nil {
		data = []byte{}
	}
	f.buf = data
	f.size = int64(len(data))

	return f, nil
}

func (fs *memFS) CreateLockFile(name string, perm os.FileMode) (fs.LockFile, bool, error) {
	_, exists := fs.files[name]
	_, err := fs.OpenFile(name, 0, perm)
	if err != nil {
		return nil, false, err
	}
	return fs.files[name], exists, nil
}

func (fs *memFS) Stat(name string) (os.FileInfo, error) {
	if f, ok := fs.files[name]; ok {
		return f, nil
	}
	return nil, os.ErrNotExist
}

func (fs *memFS) Remove(name string) error {
	if _, ok := fs.files[name]; ok {
		delete(fs.files, name)

	}
	return nil
}

func (fs *memFS) Rename(oldpath, newpath string) error {
	if f, ok := fs.files[oldpath]; ok {
		delete(fs.files, oldpath)
		fs.files[newpath] = f
		f.name = newpath
		return nil
	}
	return os.ErrNotExist
}

func (fs *memFS) ReadDir(dir string) ([]os.FileInfo, error) {
	dir = filepath.Clean(dir)
	var fis []os.FileInfo
	for name, f := range fs.files {
		if filepath.Dir(name) == dir {
			fis = append(fis, f)
		}
	}
	return fis, nil
}

type memFile struct {
	name     string
	perm     os.FileMode
	buf      []byte
	size     int64
	offset   int64
	closed   bool
	fd       *feed.API
	username string
	password string
	address  utils.Address
	client   blockstore.Client
}

func (f *memFile) Close() error {
	if f.closed {
		return os.ErrClosed
	}
	f.closed = true
	ref, err := f.client.UploadBlob(f.buf, 0, false)
	if err != nil {
		return err
	}

	topic := utils.HashString(f.name + f.username + f.password)
	_, err = f.fd.UpdateFeed(f.address, topic, ref, []byte(f.password), true)
	if err != nil { // skipcq: TCV-001
		return err
	}

	return nil
}

func (f *memFile) Unlock() error {
	if err := f.Close(); err != nil {
		return err
	}
	return Mem.Remove(f.name)
}

func (f *memFile) ReadAt(p []byte, off int64) (int, error) {
	if f.closed {
		return 0, os.ErrClosed
	}
	if off >= f.size {
		return 0, io.EOF
	}
	n := int64(len(p))
	if n > f.size-off {
		copy(p, f.buf[off:])
		return int(f.size - off), nil
	}
	copy(p, f.buf[off:off+n])
	return int(n), nil
}

func (f *memFile) Read(p []byte) (int, error) {
	n, err := f.ReadAt(p, f.offset)
	if err != nil {
		return n, err
	}
	f.offset += int64(n)
	return n, err
}

func (f *memFile) WriteAt(p []byte, off int64) (int, error) {
	if f.closed {
		return 0, os.ErrClosed
	}
	n := int64(len(p))
	if off+n > f.size {
		f.truncate(off + n)
	}
	copy(f.buf[off:off+n], p)
	return int(n), nil
}

func (f *memFile) Write(p []byte) (int, error) {
	n, err := f.WriteAt(p, f.offset)
	if err != nil {
		return n, err
	}
	f.offset += int64(n)
	return n, err
}

func (f *memFile) Seek(offset int64, whence int) (int64, error) {
	if f.closed {
		return 0, os.ErrClosed
	}
	switch whence {
	case io.SeekEnd:
		f.offset = f.size + offset
	case io.SeekStart:
		f.offset = offset
	case io.SeekCurrent:
		f.offset += offset
	}
	return f.offset, nil
}

func (f *memFile) Stat() (os.FileInfo, error) {
	if f.closed {
		return f, os.ErrClosed
	}
	return f, nil
}

func (f *memFile) Sync() error {
	if f.closed {
		return os.ErrClosed
	}
	return nil
}

func (f *memFile) truncate(size int64) {
	if size > f.size {
		diff := int(size - f.size)
		f.buf = append(f.buf, make([]byte, diff)...)
	} else {
		f.buf = f.buf[:size]
	}
	f.size = size
}

func (f *memFile) Truncate(size int64) error {
	if f.closed {
		return os.ErrClosed
	}
	f.truncate(size)
	return nil
}

func (f *memFile) Name() string {
	_, name := filepath.Split(f.name)
	return name
}

func (f *memFile) Size() int64 {
	return f.size
}

func (f *memFile) Mode() os.FileMode {
	return f.perm
}

func (f *memFile) ModTime() time.Time {
	return time.Now()
}

func (f *memFile) IsDir() bool {
	return false
}

func (f *memFile) Sys() interface{} {
	return nil
}

func (f *memFile) Slice(start int64, end int64) ([]byte, error) {
	if f.closed {
		return nil, os.ErrClosed
	}
	if end > f.size {
		return nil, io.EOF
	}
	return f.buf[start:end], nil
}
