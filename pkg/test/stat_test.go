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

package test_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	mockpost "github.com/ethersphere/bee/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"

	mock3 "github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager/rpc/mock"
	"github.com/sirupsen/logrus"

	"github.com/fairdatasociety/fairOS-dfs/pkg/user"

	"github.com/plexsysio/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

func TestUserStat(t *testing.T) {
	storer := mockstorer.New()
	beeUrl := mock.NewTestBeeServer(t, mock.TestServerOptions{
		Storer:          storer,
		PreventRedirect: true,
		Post:            mockpost.New(mockpost.WithAcceptAll()),
	})

	logger := logging.New(io.Discard, logrus.DebugLevel)
	mockClient := bee.NewBeeClient(beeUrl, mock.BatchOkStr, true, logger)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()
	sm := mock3.NewMockSubscriptionManager()

	t.Run("stat-nonexistent-user", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()
		// create user
		userObject := user.NewUsers(mockClient, ens, 500, 0, logger)
		ui := &user.Info{}
		//  stat the user
		_, err := userObject.GetUserStat(ui)
		if !errors.Is(err, user.ErrInvalidUserName) {
			t.Fatal("should be invalid user")
		}
	})

	t.Run("stat-user", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()
		// create user
		userObject := user.NewUsers(mockClient, ens, 500, 0, logger)
		sr, err := userObject.CreateNewUserV2("user1", "password1twelve", "", "", tm, sm)
		if err != nil {
			t.Fatal(err)
		}
		ui := sr.UserInfo
		//  stat the user
		stat, err := userObject.GetUserStat(ui)
		if err != nil {
			t.Fatal(err)
		}

		// verification
		if stat == nil {
			t.Fatalf("invalid stat")
		}
		if stat.Name != "user1" {
			t.Fatalf("invalid user name")
		}
	})
}
