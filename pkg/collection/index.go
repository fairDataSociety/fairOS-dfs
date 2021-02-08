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

func CreateIndex(collectionName, indexName string, indexType IndexType, fd *feed.API, user utils.Address, client blockstore.Client, mutable bool) error {
	if fd.IsReadOnlyFeed() {
		return ErrReadOnlyIndex
	}
	actualIndexName := collectionName + indexName
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

func OpenIndex(collectionName, indexName string, fd *feed.API, ai *account.Info, user utils.Address, client blockstore.Client, logger logging.Logger) (*Index, error) {
	actualIndexName := collectionName + indexName
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
	_, err := idx.feed.UpdateFeed(topic, idx.user, []byte(""))
	if err != nil {
		return ErrDeleteingIndex
	}
	return nil
}

func (idx *Index) CountIndex() (uint64, error) {
	var parentManifest *Manifest
	if idx.mutable {
		manifest, err := idx.loadManifest(idx.name)
		if err != nil {
			return 0, err
		}
		parentManifest = manifest
	} else {
		parentManifest = idx.memDB
	}

	if len(parentManifest.Entries) == 0 {
		return 0, nil
	}

	idx.count = 0
	errC := make(chan error, 1) // get only one error
	workers := make(chan bool, NoOfParallelWorkers)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	idx.loadIndexAndCount(ctx, cancel, workers, parentManifest, errC)
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
	//var wg sync.WaitGroup

	for _, entry := range manifest.Entries {
		if entry.EType == IntermediateEntry {
			//wg.Add(1)
			//workers <- true
			//go func(ent *Entry) {
			//	defer func() {
			//		//<- workers
			//		wg.Done()
			//	}()

			var newManifest *Manifest
			if idx.mutable {
				man, err := idx.loadManifest(manifest.Name + entry.Name)
				if err != nil {
					fmt.Println("Manifest load error: ", manifest.Name+entry.Name)
					//select {
					//case errC <- err:
					//default: // Default is must to avoid blocking
					//}
					//cancel()
					return
				}
				newManifest = man
			} else {
				if entry.Manifest != nil {
					newManifest = entry.Manifest
				} else {
					return
				}

			}

			//if some other goroutine fails, terminate this one too
			select {
			case <-ctx.Done():
				return
			default: // Default is must to avoid blocking
			}
			idx.loadIndexAndCount(ctx, cancel, workers, newManifest, errC)
			//}(entry)
		} else {
			count++
		}
	}
	//wg.Wait()
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
	idx.logger.Info("storing Manifest: ", manifest.Name)
	data, err := json.Marshal(manifest)
	if err != nil {
		return ErrManifestUnmarshall
	}

	ref, err := idx.client.UploadBlob(data, true, true)
	if err != nil {
		return ErrManifestUnmarshall
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
