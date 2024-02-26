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
	"fmt"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

const (
	defaultFeedCacheSize = -1
)

var errInvalidDuration = fmt.Errorf("invalid duration")

// Handler is the api handler
type Handler struct {
	ctx                context.Context
	cancel             context.CancelFunc
	dfsAPI             *dfs.API
	logger             logging.Logger
	whitelistedOrigins []string
	cookieDomain       string
}

type Options struct {
	BeeApiEndpoint     string
	CookieDomain       string
	Stamp              string
	WhitelistedOrigins []string
	EnsConfig          *contracts.ENSConfig
	SubscriptionConfig *contracts.SubscriptionConfig
	Logger             logging.Logger
	FeedCacheSize      int
	FeedCacheTTL       string
}

// New returns a new handler
func New(ctx context.Context, opts *Options) (*Handler, error) {
	dfsOpts := &dfs.Options{
		BeeApiEndpoint:     opts.BeeApiEndpoint,
		Stamp:              opts.Stamp,
		EnsConfig:          opts.EnsConfig,
		SubscriptionConfig: opts.SubscriptionConfig,
		Logger:             opts.Logger,
	}
	if opts.FeedCacheSize == 0 {
		opts.FeedCacheSize = defaultFeedCacheSize
	} else {
		dfsOpts.FeedCacheSize = opts.FeedCacheSize

		ttl, err := time.ParseDuration(opts.FeedCacheTTL)
		if err != nil {
			return nil, errInvalidDuration
		}
		dfsOpts.FeedCacheTTL = ttl
	}
	api, err := dfs.NewDfsAPI(ctx, dfsOpts)
	if err != nil {
		return nil, err
	}
	newContext, cancel := context.WithCancel(ctx)
	return &Handler{
		dfsAPI:             api,
		logger:             opts.Logger,
		whitelistedOrigins: opts.WhitelistedOrigins,
		cookieDomain:       opts.CookieDomain,
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

// Close closes the handler
func (h *Handler) Close() error {
	h.cancel()
	return h.dfsAPI.Close()
}
