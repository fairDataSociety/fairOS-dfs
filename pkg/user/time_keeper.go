package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

const (
	listTopic = "leveldb/storage/files-list"
)

var (
	errFileOpen = errors.New("leveldb/storage: file still open")
)

func (*Users) initFeedsTracker(address utils.Address, username, password string, fd *feed.API, client blockstore.Client, logger logging.Logger) (*leveldb.DB, error) {
	db, err := leveldb.Open(NewMemStorage(fd, client, address, username, password, logger), nil)
	if err != nil {
		return nil, err
	}
	fd.SetUpdateTracker(db)
	return db, nil
}

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
	logging  logging.Logger
}

// NewMemStorage returns a new memory-backed storage implementation.
func NewMemStorage(fd *feed.API, client blockstore.Client, address utils.Address, username string, password string, logger logging.Logger) storage.Storage {
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
		logging:  logger,
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

func (ms *memStorage) Log(str string) {
	ms.logging.Debug(str)
}

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
			ms.logging.Error("error updating list", "error", err)
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

func (mw *memWriter) Write(p []byte) (n int, err error) {
	n, err = mw.memFile.Write(p)
	if err != nil {
		return
	}

	ref, err := mw.client.UploadBlob(mw.Bytes(), 0, false)
	if err != nil {
		return
	}

	topic := utils.HashString(mw.name + mw.username + mw.password)
	_, err = mw.fd.UpdateFeed(mw.address, topic, ref, []byte(mw.password), true)
	return
}

func (mw *memWriter) Sync() error {
	return nil
}

func (mw *memWriter) Close() error {

	mw.ms.mu.Lock()
	defer mw.ms.mu.Unlock()
	if mw.closed {
		return storage.ErrClosed
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
