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

package api

import (
	"context"

	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

type Handler struct {
	ctx    context.Context
	cancel context.CancelFunc
	dfsAPI *dfs.API
	logger logging.Logger

	whitelistedOrigins []string
	cookieDomain       string
}

func NewHandler(ctx context.Context, dataDir, beeApi, cookieDomain, postageBlockId string, whitelistedOrigins []string, ensConfig *contracts.Config, logger logging.Logger) (*Handler, error) {
	api, err := dfs.NewDfsAPI(dataDir, beeApi, postageBlockId, ensConfig, logger)
	if err != nil {
		return nil, err
	}
	newContext, cancel := context.WithCancel(ctx)
	return &Handler{
		dfsAPI:             api,
		logger:             logger,
		whitelistedOrigins: whitelistedOrigins,
		cookieDomain:       cookieDomain,
		ctx:                newContext,
		cancel:             cancel,
	}, nil
}

// NewMockHandler is used for tests only
func NewMockHandler(dfsAPI *dfs.API, logger logging.Logger, whitelistedOrigins []string) *Handler {
	newContext, cancel := context.WithCancel(context.Background())
	return &Handler{
		dfsAPI:             dfsAPI,
		logger:             logger,
		whitelistedOrigins: whitelistedOrigins,
		ctx:                newContext,
		cancel:             cancel,
	}
}

func (h *Handler) Close() error {
	h.cancel()
	return h.dfsAPI.Close()
}
