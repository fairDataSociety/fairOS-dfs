//go:build !js
// +build !js

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
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts"
	ethClient "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
)

type DfsAPI struct {
	client  blockstore.Client
	users   *user.Users
	logger  logging.Logger
	dataDir string
}

// NewDfsAPI is the main entry point for the df controller.
func NewDfsAPI(dataDir, apiUrl, postageBlockId string, isGatewayProxy bool, ensConfig *contracts.Config, logger logging.Logger) (*DfsAPI, error) {
	ens, err := ethClient.New(ensConfig, logger)
	if err != nil {
		return nil, ErrEthClient
	}
	c := bee.NewBeeClient(apiUrl, postageBlockId, logger)
	if !c.CheckConnection(isGatewayProxy) {
		return nil, ErrBeeClient
	}
	users := user.NewUsers(dataDir, c, ens, logger)
	return &DfsAPI{
		client:  c,
		users:   users,
		logger:  logger,
		dataDir: dataDir,
	}, nil
}
