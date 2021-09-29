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

package collection

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

type IndexType int

const (
	InvalidIndex IndexType = iota
	BytesIndex
	StringIndex
	NumberIndex
	MapIndex
	ListIndex
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
	default:
		return InvalidIndex
	}
}

type Index struct {
	name        string
	mutable     bool
	indexType   IndexType
	podFile     string
	user        utils.Address
	accountInfo *account.Info
	feed        *feed.API
	client      blockstore.Client
	count       uint64
	memDB       *Manifest
	logger      logging.Logger
}

var (
	NoOfParallelWorkers = runtime.NumCPU() * 4
)

// CreateIndex creates a common index file to be used in kv or document tables.
func CreateIndex(podName, collectionName, indexName string, indexType IndexType, fd *feed.API, user utils.Address, client blockstore.Client, mutable bool) error {
	if fd.IsReadOnlyFeed() {
		return ErrReadOnlyIndex
	}
	actualIndexName := podName + collectionName + indexName
	topic := utils.HashString(actualIndexName)
	_, oldData, err := fd.GetFeedData(topic, user)
	if err == nil && len(oldData) != 0 {
		// if the feed is present and it has some data means there index is still valid
		return ErrIndexAlreadyPresent
	}

	manifest := NewManifest(actualIndexName, time.Now().Unix(), indexType, mutable)

	// marshall and store the Manifest as new feed
	data, err := json.Marshal(manifest)
	if err != nil {
		return ErrManifestUnmarshall
	}

	ref, err := client.UploadBlob(data, true, true)
	if err != nil {
		return ErrManifestUnmarshall
	}

	_, err = fd.CreateFeed(topic, user, ref)
	if err != nil {
		return ErrManifestCreate
	}
	return nil
}

// OpenIndex open the index and load any index in to the memory.
func OpenIndex(podName, collectionName, indexName string, fd *feed.API, ai *account.Info, user utils.Address, client blockstore.Client, logger logging.Logger) (*Index, error) {
	actualIndexName := podName + collectionName + indexName
	manifest := getRootManifestOfIndex(actualIndexName, fd, user, client) // this will load the entire Manifest for immutable indexes
	if manifest == nil {
		return nil, ErrIndexNotPresent
	}

	idx := &Index{
		name:        manifest.Name,
		mutable:     manifest.Mutable,
		indexType:   manifest.IdxType,
		podFile:     manifest.PodFile,
		user:        user,
		accountInfo: ai,
		feed:        fd,
		client:      client,
		count:       0,
		memDB:       manifest,
		logger:      logger,
	}
	return idx, nil
}

// DeleteIndex delete the index from file and all its entries.
func (idx *Index) DeleteIndex() error {
	if idx.isReadOnlyFeed() {
		return ErrReadOnlyIndex
	}
	manifest := getRootManifestOfIndex(idx.name, idx.feed, idx.user, idx.client)
	if manifest == nil {
		return ErrIndexNotPresent
	}

	// erase the top Manifest
	topic := utils.HashString(idx.name)
	_, err := idx.feed.UpdateFeed(topic, idx.user, []byte(utils.DeletedFeedMagicWord))
	if err != nil {
		return ErrDeleteingIndex
	}
	return nil
}

// CountIndex counts the entries in an index.
func (idx *Index) CountIndex() (uint64, error) {
	if idx.memDB == nil || idx.memDB.Entries == nil {
		manifest, err := idx.loadManifest(idx.name)
		if err != nil {
			return 0, err
		}
		idx.memDB = manifest
	}

	if len(idx.memDB.Entries) == 0 {
		return 0, nil
	}

	idx.count = 0
	errC := make(chan error, 1) // get only one error
	workers := make(chan bool, NoOfParallelWorkers)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	idx.loadIndexAndCount(ctx, cancel, workers, idx.memDB, errC)
	select {
	case err := <-errC:
		if err != nil {
			idx.count = 0
			return 0, err
		}
	default: // Default is must to avoid blocking
	}
	return idx.count, nil
}

func (idx *Index) loadIndexAndCount(ctx context.Context, cancel context.CancelFunc, workers chan bool, manifest *Manifest, errC chan error) {
	var count uint64
	for _, entry := range manifest.Entries {
		if entry.EType == IntermediateEntry {
			var newManifest *Manifest
			if entry.Manifest == nil {

				man, err := idx.loadManifest(manifest.Name + entry.Name)
				if err != nil {
					fmt.Println("Manifest load error: ", manifest.Name+entry.Name)
					return
				}
				newManifest = man
				entry.Manifest = newManifest
			} else {
				newManifest = entry.Manifest
			}
			idx.loadIndexAndCount(ctx, cancel, workers, newManifest, errC)
		} else {
			count++
		}
	}
	atomic.AddUint64(&idx.count, count)
}

// Manifest related functions
func (idx *Index) loadManifest(manifestPath string) (*Manifest, error) {
	// get feed data and unmarshall the Manifest
	idx.logger.Info("loading Manifest: ", manifestPath)
	topic := utils.HashString(manifestPath)
	_, refData, err := idx.feed.GetFeedData(topic, idx.user)
	if err != nil {
		return nil, ErrNoManifestFound
	}

	data, respCode, err := idx.client.DownloadBlob(refData)
	if err != nil {
		return nil, ErrNoManifestFound
	}
	if respCode != http.StatusOK {
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
	// marshall and update the Manifest in the feed
	idx.logger.Info("updating Manifest: ", manifest.Name)
	data, err := json.Marshal(manifest)
	if err != nil {
		return ErrManifestUnmarshall
	}

	ref, err := idx.client.UploadBlob(data, true, true)
	if err != nil {
		return ErrManifestUnmarshall
	}

	topic := utils.HashString(manifest.Name)
	_, err = idx.feed.UpdateFeed(topic, idx.user, ref)
	if err != nil {
		return ErrManifestCreate
	}
	return nil
}

func (idx *Index) storeManifest(manifest *Manifest) error {
	// marshall and store the Manifest as new feed
	data, err := json.Marshal(manifest)
	if err != nil {
		return ErrManifestUnmarshall
	}
	logStr := fmt.Sprintf("storing Manifest: %s, data len = %d", manifest.Name, len(data))
	idx.logger.Debug(logStr)

retryUpload:
	ref, err := idx.client.UploadBlob(data, true, true)
	//TODO: once the tags issue is fixed i bytes..
	// remove the error string check
	if err != nil && err.Error() != "error uploading blob" {
		return ErrManifestUnmarshall
	}

	// if ref is nil, the stamp might be exhausted.
	// get a new stamp here and proceed.
	// in the newer bee version HTTP Payment Required is sent if batch is over.
	// we might have to use it if we upgrade to newer bee version.
	if ref == nil {
		// get new stamp here and set it as the new postage id
		idx.logger.Warning("postage stamp exhausted")
		err := idx.client.GetNewPostageBatch()
		if err != nil {
			return ErrCouldNotUpdatePostageBatch
		}
		idx.logger.Info("proceeding with new postage stamp")
		goto retryUpload
	}

	topic := utils.HashString(manifest.Name)
	_, err = idx.feed.CreateFeed(topic, idx.user, ref)
	if err != nil {
		return ErrManifestCreate
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
		} else {
			matchLen++
		}
	}
	if matchLen == 0 {
		return "", str1, str2
	}
	return str1[:matchLen], str1[matchLen:], str2[matchLen:]
}

func getRootManifestOfIndex(actualIndexName string, fd *feed.API, user utils.Address, client blockstore.Client) *Manifest {
	var manifest Manifest
	topic := utils.HashString(actualIndexName)
	_, addr, err := fd.GetFeedData(topic, user)
	if err != nil {
		return nil
	}
	data, _, err := client.DownloadBlob(addr)
	if err != nil {
		return nil
	}
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return nil
	}
	return &manifest
}
