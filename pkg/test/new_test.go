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

	"github.com/plexsysio/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
)

func TestNew(t *testing.T) {
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

	t.Run("new-blank-username", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()

		// create user
		userObject := user.NewUsers(mockClient, ens, 500, 0, logger)
		_, err := userObject.CreateNewUserV2("", "password1", "", "", tm, sm)
		if !errors.Is(err, user.ErrBlankUsername) {
			t.Fatal(err)
		}
	})

	t.Run("new-user", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()

		// create user
		userObject := user.NewUsers(mockClient, ens, 500, 0, logger)
		_, err := userObject.CreateNewUserV2("user1", "password1", "", "", tm, sm)
		if err != nil && !errors.Is(err, user.ErrPasswordTooSmall) {
			t.Fatal(err)
		}

		sr, err := userObject.CreateNewUserV2("user1", "password1twelve", "", "", tm, sm)
		if err != nil {
			t.Fatal(err)
		}
		ui := sr.UserInfo
		mnemonic := sr.Mnemonic
		_, err = userObject.CreateNewUserV2("user1", "password1twelve", "", "", tm, sm)
		if !errors.Is(err, user.ErrUserAlreadyPresent) {
			t.Fatal(err)
		}

		// validate user
		if !userObject.IsUsernameAvailableV2(ui.GetUserName()) {
			t.Fatalf("user not created")
		}
		if !userObject.IsUserNameLoggedIn(ui.GetUserName()) {
			t.Fatalf("user not loggin in")
		}
		if ui == nil {
			t.Fatalf("invalid user info")
		}
		if ui.GetUserName() != "user1" {
			t.Fatalf("invalid user name")
		}
		if ui.GetFeed() == nil || ui.GetAccount() == nil || ui.GetPod() == nil {
			t.Fatalf("invalid feed, account or pod")
		}
		err = ui.GetAccount().GetWallet().IsValidMnemonic(mnemonic)
		if err != nil {
			t.Fatalf("invalid mnemonic")
		}
	})

	t.Run("new-user-multi-cred", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()
		user1 := "multicredtester"
		// create user
		userObject := user.NewUsers(mockClient, ens, 500, 0, logger)
		pass := "password1password1"
		sr, err := userObject.CreateNewUserV2(user1, pass, "", "", tm, sm)
		if err != nil {
			t.Fatal(err)
		}
		ui := sr.UserInfo
		mnemonic := sr.Mnemonic
		_, err = userObject.CreateNewUserV2(user1, pass, "", "", tm, sm)
		if !errors.Is(err, user.ErrUserAlreadyPresent) {
			t.Fatal(err)
		}

		// validate user
		if !userObject.IsUsernameAvailableV2(ui.GetUserName()) {
			t.Fatalf("user not created")
		}
		if !userObject.IsUserNameLoggedIn(ui.GetUserName()) {
			t.Fatalf("user not loggin in")
		}
		if ui == nil {
			t.Fatalf("invalid user info")
		}
		if ui.GetUserName() != user1 {
			t.Fatalf("invalid user name")
		}
		if ui.GetFeed() == nil || ui.GetAccount() == nil {
			t.Fatalf("invalid feed or account")
		}
		err = ui.GetAccount().GetWallet().IsValidMnemonic(mnemonic)
		if err != nil {
			t.Fatalf("invalid mnemonic")
		}

		_, err = userObject.CreateNewUserV2(user1, pass+pass, mnemonic, "", tm, sm)
		if err != nil {
			t.Fatal(err)
		}
	})
}
