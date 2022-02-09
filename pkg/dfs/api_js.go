//go:build js
// +build js

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

package dfs

import (
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/spf13/afero"
)

type DfsAPI struct {
	dataDir string
	client  blockstore.Client
	users   *user.Users
	logger  logging.Logger

	os afero.Fs
}

// NewDfsAPI is the main entry point for the df controller.
func NewDfsAPI(dataDir, apiUrl, debugApiUrl, postageBlockId string, logger logging.Logger) (*DfsAPI, error) {
	c := bee.NewBeeClient(apiUrl, debugApiUrl, postageBlockId, logger)
	if !c.CheckConnection() {
		return nil, ErrBeeClient
	}
	fs := afero.NewMemMapFs()
	users := user.NewUsers(dataDir, c, logger, fs)
	return &DfsAPI{
		dataDir: dataDir,
		client:  c,
		users:   users,
		logger:  logger,

		os: fs,
	}, nil
}
