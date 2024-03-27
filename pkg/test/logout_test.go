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

	mockpost "github.com/ethersphere/bee/v2/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/v2/pkg/storer/mock"

	mock3 "github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager/rpc/mock"
	"github.com/sirupsen/logrus"

	"github.com/plexsysio/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
)

func TestLogout(t *testing.T) {
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

	t.Run("logout-user", func(t *testing.T) {

		ens := mock2.NewMockNamespaceManager()
		// create user
		userObject := user.NewUsers(mockClient, ens, -1, 0, logger)
		sr, err := userObject.CreateNewUserV2("user1", "password1twelve", "", "", tm, sm)
		if err != nil {
			t.Fatal(err)
		}
		ui := sr.UserInfo
		// Logout user
		err = userObject.LogoutUser(ui.GetUserName(), "invalid sessionID")
		if !errors.Is(err, user.ErrUserNotLoggedIn) {
			t.Fatal(err)
		}

		// Logout user
		err = userObject.LogoutUser(ui.GetUserName(), ui.GetSessionId())
		if err != nil {
			t.Fatal(err)
		}

		// Validate logout
		if userObject.IsUserNameLoggedIn(ui.GetUserName()) {
			t.Fatalf("user still loggin in")
		}
		if !userObject.IsUsernameAvailableV2(ui.GetUserName()) {
			t.Fatalf("user not created")
		}
	})
}
