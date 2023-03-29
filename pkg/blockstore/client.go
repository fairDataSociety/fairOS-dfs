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

package blockstore

import (
	"context"

	"github.com/ethersphere/bee/pkg/swarm"
)

// Client is the interface for block store
type Client interface {
	CheckConnection() bool
	UploadSOC(owner string, id string, signature string, data []byte) (address []byte, err error)
	UploadChunk(ch swarm.Chunk) (address []byte, err error)
	UploadBlob(data []byte, tag uint32, encrypt bool) (address []byte, err error)
	UploadBzz(data []byte, fileName string) (address []byte, err error)
	DownloadChunk(ctx context.Context, address []byte) (data []byte, err error)
	DownloadBlob(address []byte) (data []byte, respCode int, err error)
	DownloadBzz(address []byte) (data []byte, respCode int, err error)
	DeleteReference(address []byte) error
	CreateTag(address []byte) (uint32, error)
	GetTag(tag uint32) (int64, int64, int64, error)
}
