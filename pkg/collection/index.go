/*
Copyright © 2020 FairOS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http:// www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package collection

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

// IndexType is the type of the index
type IndexType int

const (
	// InvalidIndex is returned when the index type is invalid
	InvalidIndex IndexType = iota
	// BytesIndex is returned when the index type is bytes
	BytesIndex
	// StringIndex is returned when the index type is string
	StringIndex
	// NumberIndex is returned when the index type is number
	NumberIndex
	// MapIndex is returned when the index type is map
	MapIndex
	// ListIndex is returned when the index type is list
	ListIndex

	VectorIndex
)

func (e IndexType) String() string {
	switch e {
	case BytesIndex:
		return "BytesIndex"
	case StringIndex:
		return "StringIndex"
	case NumberIndex:
		return "NumberIndex"
	case MapIndex:
		return "MapIndex"
	case ListIndex:
		return "ListIndex"
	case VectorIndex: //  skipcq: TCV-001
		return "VectorIndex"
	default:
		return "InvalidIndex"
	}
}

func toIndexTypeEnum(s string) IndexType {
	switch s {
	case "BytesIndex":
		return BytesIndex
	case "StringIndex":
		return StringIndex
	case "NumberIndex":
		return NumberIndex
	case "MapIndex":
		return MapIndex
	case "ListIndex":
		return ListIndex
	case "VectorIndex": //  skipcq: TCV-001
		return VectorIndex
	default:
		return InvalidIndex
	}
}

// Index is the structure of the index
type Index struct {
	name               string
	mutable            bool
	indexType          IndexType
	podFile            string
	encryptionPassword string
	user               utils.Address
	accountInfo        *account.Info
	feed               *feed.API
	client             blockstore.Client
	count              uint64
	memDB              *Manifest
	logger             logging.Logger
}

var (
	//  NoOfParallelWorkers is the number of parallel workers to be used for index creation
	NoOfParallelWorkers = runtime.NumCPU() * 4
)

// CreateIndex creates a common index file to be used in kv or document tables.
func CreateIndex(podName, collectionName, indexName, encryptionPassword string, indexType IndexType, fd *feed.API, user utils.Address, client blockstore.Client, mutable bool) error {
	if fd.IsReadOnlyFeed() { //  skipcq: TCV-001
		return ErrReadOnlyIndex
	}
	actualIndexName := podName + collectionName + indexName
	topic := utils.HashString(actualIndexName)
	_, oldData, err := fd.GetFeedData(topic, user, []byte(encryptionPassword))
	if err == nil && len(oldData) != 0 && string(oldData) != utils.DeletedFeedMagicWord {
		//  if the feed is present, and it has some data means there index is still valid
		return ErrIndexAlreadyPresent
	}

	manifest := NewManifest(actualIndexName, time.Now().Unix(), indexType, mutable)

	//  marshall and store the Manifest as new feed
	data, err := json.Marshal(manifest)
	if err != nil { //  skipcq: TCV-001
		return ErrManifestUnmarshall
	}

	ref, err := client.UploadBlob(data, 0, true)
	if err != nil { //  skipcq: TCV-001
		return ErrManifestUnmarshall
	}

	if string(oldData) == utils.DeletedFeedMagicWord { //  skipcq: TCV-001
		_, err = fd.UpdateFeed(user, topic, ref, []byte(encryptionPassword))
		if err != nil {
			return ErrManifestCreate
		}
		return nil
	}
	_, err = fd.CreateFeed(user, topic, ref, []byte(encryptionPassword))
	if err != nil { //  skipcq: TCV-001
		return ErrManifestCreate
	}
	return nil
}

// OpenIndex open the index and load any index in to the memory.
func OpenIndex(podName, collectionName, indexName, podPassword string, fd *feed.API, ai *account.Info, user utils.Address, client blockstore.Client, logger logging.Logger) (*Index, error) {
	actualIndexName := podName + collectionName + indexName
	manifest := getRootManifestOfIndex(actualIndexName, podPassword, fd, user, client) //  this will load the entire Manifest for immutable indexes
	if manifest == nil {
		return nil, ErrIndexNotPresent
	}
	idx := &Index{
		name:               manifest.Name,
		encryptionPassword: podPassword,
		mutable:            manifest.Mutable,
		indexType:          manifest.IdxType,
		podFile:            manifest.PodFile,
		user:               user,
		accountInfo:        ai,
		feed:               fd,
		client:             client,
		count:              manifest.Count,
		memDB:              manifest,
		logger:             logger,
	}
	return idx, nil
}

// DeleteIndex delete the index from file and all its entries.
func (idx *Index) DeleteIndex(encryptionPassword string) error {
	if idx.isReadOnlyFeed() { //  skipcq: TCV-001
		return ErrReadOnlyIndex
	}
	manifest := getRootManifestOfIndex(idx.name, encryptionPassword, idx.feed, idx.user, idx.client)
	if manifest == nil {
		return ErrIndexNotPresent
	}

	//  erase the top Manifest
	topic := utils.HashString(idx.name)
	_, err := idx.feed.UpdateFeed(idx.user, topic, []byte(utils.DeletedFeedMagicWord), []byte(encryptionPassword))
	if err != nil { //  skipcq: TCV-001
		return ErrDeleteingIndex
	}
	return nil
}

func (idx *Index) IsEmpty(encryptionPassword string) (bool, error) {
	if idx.memDB == nil || idx.memDB.Entries == nil {
		manifest, err := idx.loadManifest(idx.name, encryptionPassword)
		if err != nil {
			return true, err
		}
		idx.memDB = manifest
	}

	return len(idx.memDB.Entries) == 0, nil
}

// CountIndex counts the entries in an index.
func (idx *Index) CountIndex(encryptionPassword string) (uint64, error) {
	if idx.memDB == nil || idx.memDB.Entries == nil {
		manifest, err := idx.loadManifest(idx.name, encryptionPassword)
		if err != nil {
			return 0, err
		}
		idx.memDB = manifest
	}

	if len(idx.memDB.Entries) == 0 {
		return 0, nil
	}

	idx.count = 0
	errC := make(chan error, 1) //  get only one error
	workers := make(chan bool, NoOfParallelWorkers)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	idx.loadIndexAndCount(ctx, cancel, workers, idx.memDB, encryptionPassword, errC)
	select {
	case err := <-errC: //  skipcq: TCV-001
		if err != nil {
			idx.count = 0
			return 0, err
		}
	default: //  Default is must avoid blocking
	}
	return idx.count, nil
}

func (idx *Index) loadIndexAndCount(ctx context.Context, cancel context.CancelFunc, workers chan bool, manifest *Manifest,
	encryptionPassword string, errC chan error) {
	var count uint64
	for _, entry := range manifest.Entries {
		if entry.EType == intermediateEntry {
			var newManifest *Manifest
			if entry.Manifest == nil {

				man, err := idx.loadManifest(manifest.Name+entry.Name, encryptionPassword)
				if err != nil { //  skipcq: TCV-001
					idx.logger.Error("Manifest load error: ", manifest.Name+entry.Name)
					return
				}
				newManifest = man
				entry.Manifest = newManifest
			} else { //  skipcq: TCV-001
				newManifest = entry.Manifest
			}
			idx.loadIndexAndCount(ctx, cancel, workers, newManifest, encryptionPassword, errC)
		} else {
			count++
		}
	}
	atomic.AddUint64(&idx.count, count)
}

// Manifest related functions
func (idx *Index) loadManifest(manifestPath, encryptionPassword string) (*Manifest, error) {
	//  get feed data and unmarshall the Manifest
	idx.logger.Info("loading Manifest: ", manifestPath)
	topic := utils.HashString(manifestPath)
	_, refData, err := idx.feed.GetFeedData(topic, idx.user, []byte(encryptionPassword))
	if err != nil { //  skipcq: TCV-001
		return nil, ErrNoManifestFound
	}
	data, respCode, err := idx.client.DownloadBlob(refData)
	if err != nil { //  skipcq: TCV-001
		return nil, ErrNoManifestFound
	}
	if respCode != http.StatusOK { //  skipcq: TCV-001
		return nil, ErrNoManifestFound
	}

	var manifest Manifest
	err = json.Unmarshal(data, &manifest)
	if err != nil { //  skipcq: TCV-001
		return nil, ErrManifestUnmarshall
	}

	return &manifest, nil
}

func (idx *Index) updateManifest(manifest *Manifest, encryptionPassword string) error {
	//  marshall and update the Manifest in the feed
	idx.logger.Info("updating Manifest: ", manifest.Name)
	data, err := json.Marshal(manifest)
	if err != nil { //  skipcq: TCV-001
		return ErrManifestUnmarshall
	}

	ref, err := idx.client.UploadBlob(data, 0, true)
	if err != nil { //  skipcq: TCV-001
		return ErrManifestUnmarshall
	}

	topic := utils.HashString(manifest.Name)
	_, err = idx.feed.UpdateFeed(idx.user, topic, ref, []byte(encryptionPassword))
	if err != nil { //  skipcq: TCV-001
		return ErrManifestCreate
	}
	return nil
}

func (idx *Index) storeManifest(manifest *Manifest, encryptionPassword string) error {
	//  marshall and store the Manifest as new feed
	data, err := json.Marshal(manifest)
	if err != nil { //  skipcq: TCV-001
		return ErrManifestUnmarshall
	}
	logStr := fmt.Sprintf("storing Manifest: %s, data len = %d", manifest.Name, len(data))
	idx.logger.Debug(logStr)

	ref, err := idx.client.UploadBlob(data, 0, true)
	// TODO: once the tags issue is fixed i bytes.
	//  remove the error string check
	if err != nil { //  skipcq: TCV-001
		idx.logger.Errorf("uploadBlob failed in storeManifest : %s", err.Error())
		return ErrManifestCreate
	}
	topic := utils.HashString(manifest.Name)
	_, err = idx.feed.CreateFeed(idx.user, topic, ref, []byte(encryptionPassword))
	if err != nil { //  skipcq: TCV-001
		if strings.Contains(err.Error(), "chunk already exists") {
			_, err = idx.feed.UpdateFeed(idx.user, topic, ref, []byte(encryptionPassword))
			if err != nil { //  skipcq: TCV-001
				return ErrManifestCreate
			}
		} else {
			return ErrManifestCreate
		}
	}
	return nil
}

func (idx *Index) isReadOnlyFeed() bool {
	return idx.feed.IsReadOnlyFeed()
}

func longestCommonPrefix(str1, str2 string) (string, string, string) {
	if str1 == "" || str2 == "" {
		return "", "", ""
	}
	maxLen := len(str2)
	if len(str1) < len(str2) {
		maxLen = len(str1)
	}

	matchLen := 0
	for i := 0; i < maxLen; i++ {
		if str1[i] != str2[i] {
			break
		}
		matchLen++
	}
	if matchLen == 0 {
		return "", str1, str2
	}
	return str1[:matchLen], str1[matchLen:], str2[matchLen:]
}

func getRootManifestOfIndex(actualIndexName, encryptionPassword string, fd *feed.API, user utils.Address, client blockstore.Client) *Manifest {
	var manifest Manifest
	topic := utils.HashString(actualIndexName)
	_, addr, err := fd.GetFeedData(topic, user, []byte(encryptionPassword))
	if err != nil {
		return nil
	}
	data, _, err := client.DownloadBlob(addr)
	if err != nil {
		return nil
	}
	err = json.Unmarshal(data, &manifest)
	if err != nil { //  skipcq: TCV-001
		return nil
	}
	return &manifest
}
