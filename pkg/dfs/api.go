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
	"context"
	"errors"
	"io"
	"time"

	"github.com/plexsysio/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts"
	ethClient "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
)

const (
	defaultMaxWorkers = 100
)

// API is the go api for fairOS
type API struct {
	client blockstore.Client
	users  *user.Users
	logger logging.Logger
	tm     *taskmanager.TaskManager
	io.Closer
}

// NewDfsAPI is the main entry point for the df controller.
func NewDfsAPI(apiUrl, postageBlockId string, ensConfig *contracts.Config, logger logging.Logger) (*API, error) {
	ens, err := ethClient.New(ensConfig, logger)
	if err != nil {
		if errors.Is(err, ethClient.ErrWrongChainID) {
			return nil, err
		}
		return nil, errEthClient
	}
	c := bee.NewBeeClient(apiUrl, postageBlockId, logger)
	if !c.CheckConnection() {
		return nil, ErrBeeClient
	}
	users := user.NewUsers(c, ens, logger)

	// discard tm logs as it creates too much noise
	tmLogger := logging.New(io.Discard, 0)

	return &API{
		client: c,
		users:  users,
		logger: logger,
		tm:     taskmanager.New(10, defaultMaxWorkers, time.Second*15, tmLogger),
	}, nil
}

// NewMockDfsAPI is used for tests only
func NewMockDfsAPI(client blockstore.Client, users *user.Users, logger logging.Logger) *API {
	return &API{
		client: client,
		users:  users,
		logger: logger,
		tm:     taskmanager.New(1, 100, time.Second*15, logger),
	}
}

// Close stops the taskmanager
func (a *API) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	return a.tm.Stop(ctx)
}
