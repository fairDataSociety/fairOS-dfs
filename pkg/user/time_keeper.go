package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/akrylysov/pogreb"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

const (
	dbFileName = "timeKeeper.db"
	listTopic  = "leveldb/storage/list"
)

var (
	errFileOpen = errors.New("leveldb/storage: file still open")
)

type FeedTracker struct {
	db     *pogreb.DB
	client blockstore.Client
}

func (*Users) initFeedsTracker(address utils.Address, username, password string, fd *feed.API, client blockstore.Client) (*leveldb.DB, error) {
	db, err := leveldb.Open(NewMemStorage(fd, client, address, username, password), &opt.Options{ErrorIfMissing: true, ErrorIfExist: true})
	if err != nil {
		return nil, err
	}
	fd.SetUpdateTracker(db)
	return db, nil
}

//type memFS struct {
//	files    map[string]*memFile
//	fd       *feed.API
//	client   blockstore.Client
//	address  utils.Address
//	username string
//	password string
//}
//
//var (
//	errAppendModeNotSupported = errors.New("append mode is not supported")
//)
//
//// Mem is a file system backed by memory.
//var Mem fs.FileSystem = &memFS{files: map[string]*memFile{}}
//
//func (fs *memFS) OpenFile(name string, flag int, perm os.FileMode) (fs.File, error) {
//	if flag&os.O_APPEND != 0 {
//		// memFS doesn't support opening files in append-only mode.
//		// The database doesn't currently use O_APPEND.
//		return nil, errAppendModeNotSupported
//	}
//	f := fs.files[name]
//	if f == nil || (flag&os.O_TRUNC) != 0 {
//		f = &memFile{
//			name: name,
//			perm: perm, // Perm is saved to return it in Mode, but don't do anything else with it yet.
//		}
//		fs.files[name] = f
//	} else if !f.closed {
//		return nil, os.ErrExist
//	} else {
//		f.offset = 0
//		f.closed = false
//	}
//	f.fd = fs.fd
//	f.username = fs.username
//	f.password = fs.password
//	f.address = fs.address
//	f.client = fs.client
//	topic := utils.HashString(name + fs.username + fs.password)
//	_, ref, err := fs.fd.GetFeedData(topic, fs.address, []byte(fs.password), true)
//	if err != nil && err.Error() != "feed does not exist or was not updated yet" {
//		return nil, err
//	}
//	data, _, err := fs.client.DownloadBlob(ref)
//	if err != nil {
//		data = []byte{}
//	}
//	f.buf = data
//	f.size = int64(len(data))
//
//	return f, nil
//}
//
//func (fs *memFS) CreateLockFile(name string, perm os.FileMode) (fs.LockFile, bool, error) {
//	_, exists := fs.files[name]
//	_, err := fs.OpenFile(name, 0, perm)
//	if err != nil {
//		return nil, false, err
//	}
//	return fs.files[name], exists, nil
//}
//
//func (fs *memFS) Stat(name string) (os.FileInfo, error) {
//	if f, ok := fs.files[name]; ok {
//		return f, nil
//	}
//	return nil, os.ErrNotExist
//}
//
//func (fs *memFS) Remove(name string) error {
//	if _, ok := fs.files[name]; ok {
//		delete(fs.files, name)
//
//	}
//	return nil
//}
//
//func (fs *memFS) Rename(oldpath, newpath string) error {
//	if f, ok := fs.files[oldpath]; ok {
//		delete(fs.files, oldpath)
//		fs.files[newpath] = f
//		f.name = newpath
//		return nil
//	}
//	return os.ErrNotExist
//}
//
//func (fs *memFS) ReadDir(dir string) ([]os.FileInfo, error) {
//	dir = filepath.Clean(dir)
//	var fis []os.FileInfo
//	for name, f := range fs.files {
//		if filepath.Dir(name) == dir {
//			fis = append(fis, f)
//		}
//	}
//	return fis, nil
//}
//
//type memFile struct {
//	name     string
//	perm     os.FileMode
//	buf      []byte
//	size     int64
//	offset   int64
//	closed   bool
//	fd       *feed.API
//	username string
//	password string
//	address  utils.Address
//	client   blockstore.Client
//}
//
//func (f *memFile) Close() error {
//	if f.closed {
//		return os.ErrClosed
//	}
//	f.closed = true
//	ref, err := f.client.UploadBlob(f.buf, 0, false)
//	if err != nil {
//		return err
//	}
//
//	topic := utils.HashString(f.name + f.username + f.password)
//	_, err = f.fd.UpdateFeed(f.address, topic, ref, []byte(f.password), true)
//	if err != nil { // skipcq: TCV-001
//		return err
//	}
//
//	return nil
//}
//
//func (f *memFile) Unlock() error {
//	if err := f.Close(); err != nil {
//		return err
//	}
//	return Mem.Remove(f.name)
//}
//
//func (f *memFile) ReadAt(p []byte, off int64) (int, error) {
//	if f.closed {
//		return 0, os.ErrClosed
//	}
//	if off >= f.size {
//		return 0, io.EOF
//	}
//	n := int64(len(p))
//	if n > f.size-off {
//		copy(p, f.buf[off:])
//		return int(f.size - off), nil
//	}
//	copy(p, f.buf[off:off+n])
//	return int(n), nil
//}
//
//func (f *memFile) Read(p []byte) (int, error) {
//	n, err := f.ReadAt(p, f.offset)
//	if err != nil {
//		return n, err
//	}
//	f.offset += int64(n)
//	return n, err
//}
//
//func (f *memFile) WriteAt(p []byte, off int64) (int, error) {
//	if f.closed {
//		return 0, os.ErrClosed
//	}
//	n := int64(len(p))
//	if off+n > f.size {
//		f.truncate(off + n)
//	}
//	copy(f.buf[off:off+n], p)
//	return int(n), nil
//}
//
//func (f *memFile) Write(p []byte) (int, error) {
//	n, err := f.WriteAt(p, f.offset)
//	if err != nil {
//		return n, err
//	}
//	f.offset += int64(n)
//	return n, err
//}
//
//func (f *memFile) Seek(offset int64, whence int) (int64, error) {
//	if f.closed {
//		return 0, os.ErrClosed
//	}
//	switch whence {
//	case io.SeekEnd:
//		f.offset = f.size + offset
//	case io.SeekStart:
//		f.offset = offset
//	case io.SeekCurrent:
//		f.offset += offset
//	}
//	return f.offset, nil
//}
//
//func (f *memFile) Stat() (os.FileInfo, error) {
//	if f.closed {
//		return f, os.ErrClosed
//	}
//	return f, nil
//}
//
//func (f *memFile) Sync() error {
//	if f.closed {
//		return os.ErrClosed
//	}
//	return nil
//}
//
//func (f *memFile) truncate(size int64) {
//	if size > f.size {
//		diff := int(size - f.size)
//		f.buf = append(f.buf, make([]byte, diff)...)
//	} else {
//		f.buf = f.buf[:size]
//	}
//	f.size = size
//}
//
//func (f *memFile) Truncate(size int64) error {
//	if f.closed {
//		return os.ErrClosed
//	}
//	f.truncate(size)
//	return nil
//}
//
//func (f *memFile) Name() string {
//	_, name := filepath.Split(f.name)
//	return name
//}
//
//func (f *memFile) Size() int64 {
//	return f.size
//}
//
//func (f *memFile) Mode() os.FileMode {
//	return f.perm
//}
//
//func (f *memFile) ModTime() time.Time {
//	return time.Now()
//}
//
//func (f *memFile) IsDir() bool {
//	return false
//}
//
//func (f *memFile) Sys() interface{} {
//	return nil
//}
//
//func (f *memFile) Slice(start int64, end int64) ([]byte, error) {
//	if f.closed {
//		return nil, os.ErrClosed
//	}
//	if end > f.size {
//		return nil, io.EOF
//	}
//	return f.buf[start:end], nil
//}

// leveldb

const typeShift = 4

// Verify at compile-time that typeShift is large enough to cover all FileType
// values by confirming that 0 == 0.
var _ [0]struct{} = [storage.TypeAll >> typeShift]struct{}{}

type memStorageLock struct {
	ms *memStorage
}

func (lock *memStorageLock) Unlock() {
	ms := lock.ms
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if ms.slock == lock {
		ms.slock = nil
	}
	return
}

// memStorage is a memory-backed storage.
type memStorage struct {
	mu       sync.Mutex
	slock    *memStorageLock
	files    map[string]*memFile
	list     map[string]storage.FileDesc
	meta     storage.FileDesc
	fd       *feed.API
	client   blockstore.Client
	address  utils.Address
	username string
	password string
}

// NewMemStorage returns a new memory-backed storage implementation.
func NewMemStorage(fd *feed.API, client blockstore.Client, address utils.Address, username string, password string) storage.Storage {
	list := make(map[string]storage.FileDesc)
	topic := utils.HashString(listTopic + username + password)
	_, dt, err := fd.GetFeedData(topic, address, []byte(password), true)
	if err == nil {
		_ = json.Unmarshal(dt, &list)
	}
	return &memStorage{
		files:    make(map[string]*memFile),
		list:     list,
		fd:       fd,
		client:   client,
		address:  address,
		username: username,
		password: password,
	}
}

func (ms *memStorage) Lock() (storage.Locker, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if ms.slock != nil {
		return nil, storage.ErrLocked
	}
	ms.slock = &memStorageLock{ms: ms}
	return ms.slock, nil
}

func (*memStorage) Log(str string) {}

//
//func (ms *memStorage) SetMeta(fd storage.FileDesc) error {
//	if !storage.FileDescOk(fd) {
//		return storage.ErrInvalidFile
//	}
//
//	ms.mu.Lock()
//	ms.meta = fd
//	ms.mu.Unlock()
//
//	return nil
//}
//
//func (ms *memStorage) GetMeta() (storage.FileDesc, error) {
//	ms.mu.Lock()
//	defer ms.mu.Unlock()
//	if ms.meta.Zero() {
//		return storage.FileDesc{}, os.ErrNotExist
//	}
//	return ms.meta, nil
//}

func (ms *memStorage) List(ft storage.FileType) ([]storage.FileDesc, error) {
	ms.mu.Lock()
	var fds []storage.FileDesc
	for _, fd := range ms.list {
		if fd.Type&ft != 0 {
			fds = append(fds, fd)
		}
	}
	ms.mu.Unlock()
	return fds, nil
}

func (ms *memStorage) Open(fd storage.FileDesc) (storage.Reader, error) {
	if !storage.FileDescOk(fd) {
		return nil, storage.ErrInvalidFile
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()
	if m, exist := ms.files[fd.String()]; exist {
		if m.open {
			return nil, errFileOpen
		}
		m.open = true
		return &memReader{Reader: bytes.NewReader(m.Bytes()), ms: ms, m: m}, nil
	}
	m := &memFile{}
	m.fd = ms.fd
	m.name = fd.String()
	m.username = ms.username
	m.password = ms.password
	m.address = ms.address
	m.client = ms.client
	topic := utils.HashString(fd.String() + ms.username + ms.password)
	_, ref, err := ms.fd.GetFeedData(topic, ms.address, []byte(ms.password), true)
	if err != nil && err.Error() != "feed does not exist or was not updated yet" {
		return nil, os.ErrNotExist
	}

	data, _, err := ms.client.DownloadBlob(ref)
	if err != nil {
		return nil, os.ErrNotExist
	}
	m.Buffer = bytes.NewBuffer(data)
	ms.files[fd.String()] = m
	m.open = true
	return &memReader{Reader: bytes.NewReader(m.Bytes()), ms: ms, m: m}, nil
}

func (ms *memStorage) Create(fd storage.FileDesc) (storage.Writer, error) {
	if !storage.FileDescOk(fd) {
		return nil, storage.ErrInvalidFile
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()
	m, exist := ms.files[fd.String()]
	if exist {
		if m.open {
			return nil, errFileOpen
		}
		m.Reset()
	} else {
		m = &memFile{}
		m.fd = ms.fd
		m.name = fd.String()
		m.username = ms.username
		m.password = ms.password
		m.address = ms.address
		m.client = ms.client

		m.Buffer = bytes.NewBuffer([]byte{})
		ms.files[fd.String()] = m
	}
	m.open = true
	ms.list[fd.String()] = fd
	dt, err := json.Marshal(ms.list)
	if err == nil {
		topic := utils.HashString(listTopic + ms.username + ms.password)
		_, err = ms.fd.UpdateFeed(ms.address, topic, dt, []byte(ms.password), true)
		if err != nil {
			// TODO log err
		}
	}

	return &memWriter{memFile: m, ms: ms, name: fd.String()}, nil
}

func (ms *memStorage) Remove(fd storage.FileDesc) error {
	if !storage.FileDescOk(fd) {
		return storage.ErrInvalidFile
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()
	if _, exist := ms.files[fd.String()]; exist {
		delete(ms.files, fd.String())
		delete(ms.list, fd.String())
		dt, err := json.Marshal(ms.list)
		if err != nil {
			return err
		}
		topic := utils.HashString(listTopic + ms.username + ms.password)
		_, err = ms.fd.UpdateFeed(ms.address, topic, dt, []byte(ms.password), true)
		if err != nil {
			return err
		}
		return nil
	}
	return os.ErrNotExist
}

func (ms *memStorage) Rename(oldfd, newfd storage.FileDesc) error {
	if !storage.FileDescOk(oldfd) || !storage.FileDescOk(newfd) {
		return storage.ErrInvalidFile
	}
	if oldfd == newfd {
		return nil
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()
	oldm, exist := ms.files[oldfd.String()]
	if !exist {
		return os.ErrNotExist
	}
	newm, exist := ms.files[newfd.String()]
	if (exist && newm.open) || oldm.open {
		return errFileOpen
	}
	delete(ms.files, oldfd.String())
	ms.files[newfd.String()] = oldm
	return nil
}

func (ms *memStorage) Close() error {
	return nil
}

func (ms *memStorage) setMeta(fd storage.FileDesc) error {
	content := fd.String()
	// Check and backup old CURRENT file.
	currentPath := "CURRENT"

	topic := utils.HashString(currentPath + ms.username + ms.password)
	_, dt, err := ms.fd.GetFeedData(topic, ms.address, []byte(ms.password), true)
	if err != nil && err.Error() != "feed does not exist or was not updated yet" {
		return err
	}
	if string(dt) == content {
		// Content not changed, do nothing.
		return nil
	}

	_, err = ms.fd.UpdateFeed(ms.address, topic, []byte(content), []byte(ms.password), true)
	if err != nil { // skipcq: TCV-001
		return err
	}
	ms.meta = fd
	return nil
}

func (ms *memStorage) SetMeta(fd storage.FileDesc) error {
	if !storage.FileDescOk(fd) {
		return storage.ErrInvalidFile
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.setMeta(fd)
}

func (ms *memStorage) GetMeta() (storage.FileDesc, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	meta := storage.FileDesc{}
	if ms.meta.Zero() {
		// Try
		// - CURRENT
		currentPath := "CURRENT"
		topic := utils.HashString(currentPath + ms.username + ms.password)
		_, dt, err := ms.fd.GetFeedData(topic, ms.address, []byte(ms.password), true)
		if err != nil {
			return meta, os.ErrNotExist
		}
		if !fsParseNamePtr(string(dt), &meta) {
			return meta, os.ErrNotExist
		}
		ms.meta = meta
		return meta, nil
	}

	return ms.meta, nil
}

type memFile struct {
	name string
	*bytes.Buffer
	open     bool
	fd       *feed.API
	username string
	password string
	address  utils.Address
	client   blockstore.Client
}

type memReader struct {
	*bytes.Reader
	ms     *memStorage
	m      *memFile
	closed bool
}

func (mr *memReader) Close() error {

	mr.ms.mu.Lock()
	defer mr.ms.mu.Unlock()
	if mr.closed {
		return storage.ErrClosed
	}
	mr.m.open = false
	return nil
}

type memWriter struct {
	name string
	*memFile
	ms     *memStorage
	closed bool
}

func (*memWriter) Sync() error {
	return nil
}

func (mw *memWriter) Close() error {

	mw.ms.mu.Lock()
	defer mw.ms.mu.Unlock()
	if mw.closed {
		return storage.ErrClosed
	}
	ref, err := mw.client.UploadBlob(mw.Bytes(), 0, false)
	if err != nil {
		return err
	}

	topic := utils.HashString(mw.name + mw.username + mw.password)
	_, err = mw.fd.UpdateFeed(mw.address, topic, ref, []byte(mw.password), true)
	if err != nil { // skipcq: TCV-001
		return err
	}
	mw.memFile.open = false
	return nil
}

func fsParseName(name string) (fd storage.FileDesc, ok bool) {
	var tail string
	_, err := fmt.Sscanf(name, "%d.%s", &fd.Num, &tail)
	if err == nil {
		switch tail {
		case "log":
			fd.Type = storage.TypeJournal
		case "ldb", "sst":
			fd.Type = storage.TypeTable
		case "tmp":
			fd.Type = storage.TypeTemp
		default:
			return
		}
		return fd, true
	}
	n, _ := fmt.Sscanf(name, "MANIFEST-%d%s", &fd.Num, &tail)
	if n == 1 {
		fd.Type = storage.TypeManifest
		return fd, true
	}
	return
}

func fsParseNamePtr(name string, fd *storage.FileDesc) bool {
	_fd, ok := fsParseName(name)
	if fd != nil {
		*fd = _fd
	}
	return ok
}
