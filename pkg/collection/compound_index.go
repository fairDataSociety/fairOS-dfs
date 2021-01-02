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
	"encoding/json"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"strings"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

type CompoundIndex struct {
	name        string
	indexFields []string
	IndexTypes  []IndexType
	index       *Index
}

func CreateCompoundIndex(collectionName string, indexFields []string, indexTypes []IndexType, fd *feed.API, user utils.Address, client blockstore.Client) error {
	compoundIndexName := getCompoundIndexName(collectionName, indexFields)

	topic := utils.HashString(compoundIndexName)
	_, oldData, err := fd.GetFeedData(topic, user)
	if err == nil && len(oldData) != 0 {
		return ErrIndexAlreadyPresent
	}

	compoundManifest := NewCompoundManifest(compoundIndexName, indexFields, indexTypes, time.Now().Unix())

	// marshall and store the compound manifest as new feed
	data, err := json.Marshal(compoundManifest)
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

	// create the simple index inside
	return CreateIndex(collectionName, getIndexName(indexFields), StringIndex, fd, user, client)
}

func OpenCompoundIndex(collectionName string, indexFields []string, fd *feed.API, ai *account.Info, user utils.Address,
	client blockstore.Client, logger logging.Logger) (*CompoundIndex, error) {

	// check if the compound index manifest exists
	compoundIndexName := getCompoundIndexName(collectionName, indexFields)
	rootCompoundManifest := getRootCompoundManifestOfIndex(compoundIndexName, fd, user, client)
	if rootCompoundManifest == nil {
		return nil, ErrIndexNotPresent
	}

	// open the simple index inside
	idx, err := OpenIndex(collectionName, getIndexName(indexFields), fd, ai, user, client, logger)
	if err != nil {
		return nil, err
	}

	// construct the compound Index and return
	return &CompoundIndex{
		name: getCompoundIndexName(collectionName, indexFields),
		indexFields: indexFields,
		IndexTypes: rootCompoundManifest.IdxTypes,
		index: idx,
	}, nil
}

func (ci *CompoundIndex) DeleteCompoundIndex() error {
	rootCompoundManifest := getRootCompoundManifestOfIndex(ci.name, ci.index.feed, ci.index.user, ci.index.client)
	if rootCompoundManifest == nil {
		return ErrIndexNotPresent
	}

	// erase the top manifest
	topic := utils.HashString(ci.index.name)
	_, err := ci.index.feed.UpdateFeed(topic, ci.index.user, []byte(""))
	if err != nil {
		return ErrDeleteingIndex
	}

	// erase the simple index inside
	return ci.index.DeleteIndex()
}

func getIndexName(indexFields []string) string {
	return strings.Join(indexFields, "_")
}

func getCompoundIndexName(collectionName string, indexFields []string) string {
	return "COMPOUND_" + collectionName + "_" + getIndexName(indexFields)
}

func getRootCompoundManifestOfIndex(actualIndexName string, fd *feed.API, user utils.Address, client blockstore.Client) *CompoundManifest {
	var manifest CompoundManifest
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
