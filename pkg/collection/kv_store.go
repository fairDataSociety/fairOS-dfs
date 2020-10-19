package collection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"net/http"
	"strings"
	"time"
)

const (
	LeafEntry         = "L"
	IntermediateEntry = "I"
)

var (
	ErrEmptyDB       = errors.New("empty DB")
	ErrNoNextElement = errors.New("no next element")
)

type DB struct {
	name        string
	count       int64
	user        utils.Address
	accountInfo *account.Info
	feed        *feed.API
	client      blockstore.Client
	topManifest *Manifest
}

func NewDB(name string, fd *feed.API, ai *account.Info, user utils.Address, client blockstore.Client) (*DB, error) {
	m := NewManifest(name, time.Now().Unix())
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	// create the top feed and store the empty manifest
	topic := utils.HashString(name)
	_, err = fd.CreateFeed(topic, user, data)
	if err != nil {
		return nil, err
	}

	return &DB{
		name:        name,
		count:       0,
		user:        user,
		accountInfo: ai,
		feed:        fd,
		client:      client,
		topManifest: m,
	}, nil
}

func (db *DB) Put(key string, value []byte) error {
	// store the value first
	val, err := db.client.UploadBlob(value, true, true)
	if err != nil {
		return err
	}

	// get the first feed fo the DB
	topic := utils.HashString(utils.PathSeperator + db.name)
	_, data, err := db.feed.GetFeedData(topic, db.user)
	if err != nil {
		return db.initDB(key, val)
	}

	// unmarshall the manifest
	var manifest Manifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return err
	}
	db.topManifest = &manifest

	ctx := context.Background()
	return db.addOrUpdateEntry(ctx, manifest, key, val)
}

//func (db *DB) Get(key string) ([]byte, error) {
//
//}

//func (db *DB) Delete(key string) error {
//
//}

func (db *DB) addOrUpdateEntry(ctx context.Context, manifest Manifest, key string, value []byte) error {
	// go through the manifest to find the key
	dirtyFlag := false
	for i, _ := range manifest.Entries {
		entry := &manifest.Entries[i]
		if entry.EType == LeafEntry {

			// an entry with the same key is already present... so update it
			if entry.Name == key {
				entry.Ref = value
				dirtyFlag = true
				break
			}

			prefix := longestCommonPrefix(key, entry.Name)
			if prefix != "" {
				// the new element is a prefix of the existing leaf..
				// add a new branch with two new leafs
				var newManifest Manifest
				newManifest.Name = manifest.Name + utils.PathSeperator + prefix
				newManifest.CreationTime = time.Now().Unix()
				newManifest.Entries = append(newManifest.Entries, Entry{
					Name:  strings.TrimPrefix(entry.Name, prefix),
					EType: LeafEntry,
					Ref:   entry.Ref,
				})
				newManifest.Entries = append(newManifest.Entries, Entry{
					Name:  strings.TrimPrefix(key, prefix),
					EType: LeafEntry,
					Ref:   value,
				})

				data, err := json.Marshal(newManifest)
				if err != nil {
					return err
				}

				prefixTopic := utils.HashString(newManifest.Name)
				_, err = db.feed.CreateFeed(prefixTopic, db.user, data)
				if err != nil {
					return err
				}

				// convert the existing leaf to intermediate node
				entry.Name = prefix
				entry.EType = IntermediateEntry
				dirtyFlag = true
			}
		} else {
			// go inside the branch and search
			if entry.EType == IntermediateEntry && strings.HasPrefix(key, entry.Name) {
				newKey := strings.TrimPrefix(key, entry.Name)
				topic := utils.HashString(manifest.Name + utils.PathSeperator + entry.Name)
				_, data, err := db.feed.GetFeedData(topic, db.user)
				if err != nil {
					return err
				}
				var intermediateManifest Manifest
				err = json.Unmarshal(data, &intermediateManifest)
				if err != nil {
					return err
				}
				return db.addOrUpdateEntry(ctx, intermediateManifest, newKey, value)
			}
		}
	}

	// if the manifest is not already changed, then this is a new entry
	if !dirtyFlag {
		db.count++
		newEntry := Entry{
			Name:  key,
			EType: LeafEntry,
			Ref:   value,
		}
		manifest.Entries = append(manifest.Entries, newEntry)
		dirtyFlag = true
	}

	if dirtyFlag {
		data, err := json.Marshal(manifest)
		if err != nil {
			return err
		}
		topic := utils.HashString(manifest.Name)
		_, err = db.feed.UpdateFeed(topic, db.user, data)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (db *DB) initDB(key string, value []byte) error {
	// This is a new branch, so crreate a new manifest
	var manifest Manifest
	manifest.Name = utils.PathSeperator + db.name
	manifest.CreationTime = time.Now().Unix()

	// addd the entry to the new manifest
	newEntry := Entry{
		Name:  key,
		EType: LeafEntry,
		Ref:   value,
	}
	manifest.Entries = append(manifest.Entries, newEntry)

	// marshall and store the manifest as new feed
	data, err := json.Marshal(manifest)
	if err != nil {
		return err
	}

	topic := utils.HashString(manifest.Name)
	_, err = db.feed.CreateFeed(topic, db.user, data)
	if err != nil {
		return err
	}
	return nil
}

type Iterator struct {
	db            *DB
	startPrefix   string
	endPrefix     string
	limit         int64
	givenUntilNow int64
	currentKey    string
	currentValue  []byte
	manifestStack []*ManifestState
	error         error
}

type ManifestState struct {
	currentManifest Manifest
	currentIndex    int
}

func (db *DB) NewIterator(start, end string, limit int64) (*Iterator, error) {
	// get the first feed of the DB
	topic := utils.HashString(utils.PathSeperator + db.name)
	_, data, err := db.feed.GetFeedData(topic, db.user)
	if err != nil {
		return nil, ErrEmptyDB
	}

	var manifest Manifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return nil, err
	}

	firstManifest := &ManifestState{
		currentManifest: manifest,
		currentIndex:    0,
	}
	var stack []*ManifestState
	stack = append(stack, firstManifest)

	itr := &Iterator{
		db:            db,
		startPrefix:   start,
		endPrefix:     end,
		limit:         limit,
		givenUntilNow: 0,
		manifestStack: stack,
		currentKey:    "",
		currentValue:  nil,
		error:         nil,
	}
	return itr, nil
}

func (itr *Iterator) Next() bool {

	// get the current manifest at the top of the stack
	depthOfStack := len(itr.manifestStack)
	if depthOfStack == 0 {
		itr.error = ErrNoNextElement
		return false
	}

	// take the top manifest to find the next entry
	manifestState := itr.manifestStack[depthOfStack-1]

	entriesExhausted := true
	for entriesExhausted {
		// see if we have exhausted the entries in the current manifest
		if manifestState.currentIndex >= len(manifestState.currentManifest.Entries) {
			// pop the exhausted manifest from the top and pick the next manifest to find the entry
			n := depthOfStack - 1
			if n == 0 {
				itr.error = ErrNoNextElement
				return false
			}
			manifestState = itr.manifestStack[n-1]
			itr.manifestStack[n] = nil
			itr.manifestStack = itr.manifestStack[:n]
			depthOfStack = n
		} else {
			entriesExhausted = false
		}
	}

	// We have a manifest whose entries are not yet exhausted,
	// so get the next entry and check for valid conditions of the Iterator()
	entry := &manifestState.currentManifest.Entries[manifestState.currentIndex]
	manifestState.currentIndex++

	// if it is a leaf entry, set the key and value
	if entry.EType == LeafEntry {
		data, respCode, err := itr.db.client.DownloadBlob(entry.Ref)
		if err != nil {
			itr.error = err
			return false
		}
		if respCode != http.StatusOK {
			itr.error = errors.New(fmt.Sprintf("could not download value: %d", respCode))
			return false
		}
		actualKey := manifestState.currentManifest.Name + utils.PathSeperator + entry.Name
		actualKey = strings.TrimPrefix(actualKey, utils.PathSeperator+itr.db.name)
		actualKey = strings.Replace(actualKey, utils.PathSeperator, "", -1)
		itr.currentKey = actualKey
		itr.currentValue = data
		itr.givenUntilNow++
		return true
	}

	// if it is an intermediate entry, get the branch manifest and push in to the stack
	if entry.EType == IntermediateEntry {
		topic := utils.HashString(manifestState.currentManifest.Name + utils.PathSeperator + entry.Name)
		_, data, err := itr.db.feed.GetFeedData(topic, itr.db.user)
		if err != nil {
			itr.error = err
			return false
		}

		var newManifest Manifest
		err = json.Unmarshal(data, &newManifest)
		if err != nil {
			itr.error = err
			return false
		}
		newManifestState := &ManifestState{
			currentManifest: newManifest,
			currentIndex:    0,
		}
		itr.manifestStack = append(itr.manifestStack, newManifestState)
		return itr.Next()
	}
	return false
}

func (itr *Iterator) Key() string {
	return itr.currentKey
}

func (itr *Iterator) Value() []byte {
	return itr.currentValue
}

func (itr *Iterator) Seek(key string) bool {
	return false
}

func (itr *Iterator) getNextManifest(name string) (*Manifest, error) {
	// get the first feegit add .d fo the DB
	topic := utils.HashString(name)
	_, data, err := itr.db.feed.GetFeedData(topic, itr.db.user)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return nil, err
	}
	return &manifest, err
}

func longestCommonPrefix(str1, str2 string) string {
	if str1 == "" || str2 == "" {
		return ""
	}
	maxLen := len(str2)
	if len(str1) < len(str2) {
		maxLen = len(str1)
	}

	matchLen := 0
	for i := 0; i < maxLen; i++ {
		if str1[i] != str2[i] {
			break
		} else {
			matchLen++
		}
	}
	if matchLen == 0 {
		return ""
	}
	return str1[:matchLen]
}
