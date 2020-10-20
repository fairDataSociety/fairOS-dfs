package collection

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	LeafEntry         = "L"
	IntermediateEntry = "I"
)

type Index struct {
	name        string
	user        utils.Address
	accountInfo *account.Info
	feed        *feed.API
	client      blockstore.Client
}

func OpenIndex(collectionName, IndexName string, fd *feed.API, ai *account.Info, user utils.Address, client blockstore.Client) (*Index, error) {
	idx :=  &Index{
		name:        utils.PathSeperator + collectionName + utils.PathSeperator +IndexName,
		user:        user,
		accountInfo: ai,
		feed:        fd,
		client:      client,
	}
	return idx, nil
}

func (idx *Index) DeleteIndex() error {
	// erase the top manifest
	topic := utils.HashString(idx.name)
	_, err := idx.feed.UpdateFeed(topic, idx.user, []byte(""))
	if err != nil {
		return ErrDeleteingIndex
	}
	return nil
}

func (idx *Index) Put(key string, refValue []byte) error {
	// get the first feed of the Index
	topic := utils.HashString(idx.name)
	_, data, err := idx.feed.GetFeedData(topic, idx.user)
	if err != nil {
		return idx.initIndex(key, refValue)
	}

	// unmarshall the manifest
	var manifest Manifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return idx.addOrUpdateEntry(ctx, manifest, key, refValue)
}

func (idx *Index) Get(key string) ([]byte, error) {
	manifest, i, err := idx.seekManifestAndEntry(key)
	if err != nil {
		return nil, err
	}
	return manifest.Entries[i].Ref, nil
}

func (idx *Index) Delete(key string) ([]byte, error) {
	manifest, i, err := idx.seekManifestAndEntry(key)
	if err != nil {
		return nil, err
	}

	deletedRef := manifest.Entries[i].Ref

	if len(manifest.Entries) == 1 && manifest.Entries[0].Name == "" {
		// then we have to remove the intermediate node in the parent manifest
		// so that the entire branch goes kaboom
		parentEntryKey := filepath.Base(manifest.Name)
		parentManifest, err := idx.loadManifest(filepath.Dir(manifest.Name))
		if err != nil {
			return nil, err
		}
		for i, entry := range parentManifest.Entries {
			if entry.EType == IntermediateEntry && entry.Name == parentEntryKey {
				deletedRef = entry.Ref
				parentManifest.Entries = append(parentManifest.Entries[:i], parentManifest.Entries[i+1:]...)
				break
			}
		}
		err = idx.updateManifest(parentManifest)
		if err != nil {
			return nil, err
		}
		return deletedRef, nil
	}

	manifest.Entries = append(manifest.Entries[:i], manifest.Entries[i+1:]...)
	err = idx.updateManifest(manifest)
	if err != nil {
		return nil, err
	}
	return deletedRef, nil
}

func (idx *Index) addOrUpdateEntry(ctx context.Context, manifest Manifest, key string, value []byte) error {
	// go through the manifest to find the key
	manifest.dirtyFlag = false
	for i, _ := range manifest.Entries {
		entry := &manifest.Entries[i]
		if entry.EType == LeafEntry {

			// an entry with the same key is already present... so update it
			if entry.Name == key {
				entry.Ref = value
				manifest.dirtyFlag = true
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
				_, err = idx.feed.CreateFeed(prefixTopic, idx.user, data)
				if err != nil {
					return err
				}

				// convert the existing leaf to intermediate node
				entry.Name = prefix
				entry.EType = IntermediateEntry
				manifest.dirtyFlag = true
				break
			}
		} else {
			// go inside the branch and search
			if entry.EType == IntermediateEntry && strings.HasPrefix(key, entry.Name) {
				newKey := strings.TrimPrefix(key, entry.Name)
				topic := utils.HashString(manifest.Name + utils.PathSeperator + entry.Name)
				_, data, err := idx.feed.GetFeedData(topic, idx.user)
				if err != nil {
					return err
				}
				var intermediateManifest Manifest
				err = json.Unmarshal(data, &intermediateManifest)
				if err != nil {
					return err
				}
				return idx.addOrUpdateEntry(ctx, intermediateManifest, newKey, value)
			}
		}
	}

	// if the manifest is not already changed, then this is a new entry
	if !manifest.dirtyFlag {
		newEntry := Entry{
			Name:  key,
			EType: LeafEntry,
			Ref:   value,
		}
		manifest.Entries = append(manifest.Entries, newEntry)
		manifest.dirtyFlag = true
	}

	if manifest.dirtyFlag {
		data, err := json.Marshal(manifest)
		if err != nil {
			return err
		}
		topic := utils.HashString(manifest.Name)
		_, err = idx.feed.UpdateFeed(topic, idx.user, data)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (idx *Index) initIndex(key string, value []byte) error {
	// This is the first insert, so create a new manifest index
	var manifest Manifest
	manifest.Name = idx.name
	manifest.CreationTime = time.Now().Unix()

	// add the entry to the new manifest
	newEntry := Entry{
		Name:  key,
		EType: LeafEntry,
		Ref:   value,
	}
	manifest.Entries = append(manifest.Entries, newEntry)

	// marshall and store the manifest as new feed
	data, err := json.Marshal(manifest)
	if err != nil {
		return ErrManifestUnmarshall
	}

	topic := utils.HashString(manifest.Name)
	_, err = idx.feed.CreateFeed(topic, idx.user, data)
	if err != nil {
		return ErrManifestCreate
	}
	manifest.dirtyFlag = false  // since the manifest is synced to Swarm
	return nil
}

func (idx *Index) seekManifestAndEntry(key string) (*Manifest, int, error) {
	// load the first manifest of the index
	firstManifest, err := idx.loadManifest(idx.name)
	if err != nil && !errors.Is(err, ErrNoManifestFound){
		return nil, 0, err
	}

	// if there are any elements in the index, then search for the entry
	if len(firstManifest.Entries) > 0 {
		return idx.findManifest(firstManifest, key)
	}
	return nil, 0, ErrEntryNotFound
}

func (idx *Index) findManifest(parentManifest *Manifest, key string) (*Manifest, int, error) {
	for i, entry := range parentManifest.Entries {
		if entry.EType == LeafEntry && entry.Name == key {
			return parentManifest, i, nil
		}

		if entry.EType == IntermediateEntry && strings.HasPrefix(key, entry.Name) {
			childManifestPath := parentManifest.Name + utils.PathSeperator + entry.Name
			childManifest, err := idx.loadManifest(childManifestPath)
			if err != nil {
				return nil, 0, err
			}
			childKey := strings.TrimPrefix(key, entry.Name)
			return idx.findManifest(childManifest, childKey)
		}
	}
	return nil, 0, ErrEntryNotFound
}

func (idx *Index) loadManifest(manifestPath string) (*Manifest, error) {
	topic := utils.HashString(manifestPath)
	_, data, err := idx.feed.GetFeedData(topic, idx.user)
	if err != nil {
		return nil, ErrNoManifestFound
	}

	var manifest Manifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return nil, ErrManifestUnmarshall
	}
	return &manifest, nil
}

func (idx *Index) updateManifest(manifest *Manifest) error {
	// marshall and store the manifest as new feed
	data, err := json.Marshal(manifest)
	if err != nil {
		return ErrManifestUnmarshall
	}

	topic := utils.HashString(manifest.Name)
	_, err = idx.feed.UpdateFeed(topic, idx.user, data)
	if err != nil {
		return ErrManifestCreate
	}
	return nil
}

type Iterator struct {
	index         *Index
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
	currentManifest *Manifest
	currentIndex    int
}

func (idx *Index) NewIterator(start, end string, limit int64) (*Iterator, error) {
	// get the first feed of the Index
	topic := utils.HashString(idx.name)
	_, data, err := idx.feed.GetFeedData(topic, idx.user)
	if err != nil {
		return nil, ErrEmptyIndex
	}

	var manifest Manifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return nil, err
	}

	firstManifest := &ManifestState{
		currentManifest: &manifest,
		currentIndex:    0,
	}
	var stack []*ManifestState
	stack = append(stack, firstManifest)

	itr := &Iterator{
		index:         idx,
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
		actualKey := manifestState.currentManifest.Name + utils.PathSeperator + entry.Name
		actualKey = strings.TrimPrefix(actualKey, itr.index.name)
		actualKey = strings.Replace(actualKey, utils.PathSeperator, "", -1)
		itr.currentKey = actualKey
		itr.currentValue = entry.Ref
		itr.givenUntilNow++
		return true
	}

	// if it is an intermediate entry, get the branch manifest and push in to the stack
	if entry.EType == IntermediateEntry {
		topic := utils.HashString(manifestState.currentManifest.Name + utils.PathSeperator + entry.Name)
		_, data, err := itr.index.feed.GetFeedData(topic, itr.index.user)
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
			currentManifest: &newManifest,
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
	err := itr.seekKey(key)
	if err != nil {
		return false
	}
	return true
}


func (itr *Iterator) seekKey(key string)  error {
	manifest, err := itr.index.loadManifest(itr.index.name)
	if err != nil {
		return err
	}

	// if there are any elements in the index, then search for the entry
	if len(manifest.Entries) > 0 {
		for i, entry := range manifest.Entries {
			if entry.EType == LeafEntry && entry.Name == key {
				manifestState := &ManifestState{
					currentManifest: manifest,
					currentIndex: i,
				}
				itr.manifestStack = append(itr.manifestStack, manifestState)
				actualKey := manifest.Name + utils.PathSeperator + entry.Name
				actualKey = strings.TrimPrefix(actualKey, utils.PathSeperator+itr.index.name)
				actualKey = strings.Replace(actualKey, utils.PathSeperator, "", -1)
				itr.currentKey = actualKey
				itr.currentValue = entry.Ref
				itr.givenUntilNow++
				return nil
			}

			if entry.EType == IntermediateEntry && strings.HasPrefix(key, entry.Name) {
				newKey := strings.TrimPrefix(key, entry.Name)
				topic := utils.HashString(manifest.Name + utils.PathSeperator + entry.Name)
				_, data, err := itr.index.feed.GetFeedData(topic, itr.index.user)
				if err != nil {
					return err
				}

				var newManifest Manifest
				err = json.Unmarshal(data, &newManifest)
				if err != nil {
					return err
				}
				newManifestState := &ManifestState{
					currentManifest: &newManifest,
					currentIndex:    0,
				}
				itr.manifestStack = append(itr.manifestStack, newManifestState)
				return itr.seekKey(newKey)
			}
		}
	}
	return ErrEntryNotFound
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
