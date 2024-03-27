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
	"sort"
	"testing"
	"time"

	mockpost "github.com/ethersphere/bee/v2/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/v2/pkg/storer/mock"

	mock3 "github.com/fairdatasociety/fairOS-dfs/pkg/subscriptionManager/rpc/mock"
	"github.com/sirupsen/logrus"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
)

func TestLogin(t *testing.T) {
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

	t.Run("login-user", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()
		// create user
		userObject := user.NewUsers(mockClient, ens, -1, 0, logger)
		sr, err := userObject.CreateNewUserV2("7e4567e7cb003804992eef11fd5c757275a4c", "password1twelve", "", "", tm, sm)
		if err != nil {
			t.Fatal(err)
		}
		ui := sr.UserInfo
		// Logout user
		err = userObject.LogoutUser(ui.GetUserName(), ui.GetSessionId())
		if err != nil {
			t.Fatal(err)
		}

		_, err = userObject.LoginUserV2("not_an_username", "password1", mockClient, tm, sm, "")
		if !errors.Is(err, user.ErrUserNameNotFound) {
			t.Fatal(err)
		}

		_, err = userObject.LoginUserV2("7e4567e7cb003804992eef11fd5c757275a4c", "wrong_password", mockClient, tm, sm, "")
		if !errors.Is(err, user.ErrInvalidPassword) {
			t.Fatal(err)
		}

		// addUserAndSessionToMap user again
		sr1, err := userObject.LoginUserV2("7e4567e7cb003804992eef11fd5c757275a4c", "password1twelve", mockClient, tm, sm, "")
		if err != nil {
			t.Fatal(err)
		}
		ui1 := sr1.UserInfo
		ui2 := userObject.GetLoggedInUserInfo(ui1.GetSessionId())

		// Validate login
		if !userObject.IsUserNameLoggedIn("7e4567e7cb003804992eef11fd5c757275a4c") {
			t.Fatalf("user not loggin in")
		}
		addr := ui.GetAccount().GetUserAccountInfo().GetAddress()
		addr1 := ui1.GetAccount().GetUserAccountInfo().GetAddress()
		addr2 := ui2.GetAccount().GetUserAccountInfo().GetAddress()
		if addr.Hex() != addr1.Hex() {
			t.Fatal("loaded with different account")
		}

		if addr.Hex() != addr2.Hex() {
			t.Fatal("got different userinfo")
		}

		if ui.GetUserDirectory() == nil {
			t.Fatal("user directory handler should not be nil")
		}
	})

	t.Run("new-user-multi-cred", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()
		user1 := "multicredtester"
		pass := "password1password1"
		// create user
		userObject := user.NewUsers(mockClient, ens, -1, 0, logger)
		sr, err := userObject.CreateNewUserV2(user1, pass, "", "", tm, sm)
		if err != nil {
			t.Fatal(err)
		}
		ui := sr.UserInfo
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
		err = ui.GetAccount().GetWallet().IsValidMnemonic(sr.Mnemonic)
		if err != nil {
			t.Fatalf("invalid mnemonic")
		}

		_, err = userObject.CreateNewUserV2(user1, pass+pass, sr.Mnemonic, "", tm, sm)
		if err != nil {
			t.Fatal(err)
		}

		lr1, err := userObject.LoginUserV2(user1, pass, mockClient, tm, sm, "")
		if err != nil {
			t.Fatal(err)
		}
		login1 := lr1.UserInfo
		lr2, err := userObject.LoginUserV2(user1, pass+pass, mockClient, tm, sm, "")
		if err != nil {
			t.Fatal(err)
		}
		login2 := lr2.UserInfo

		addr1 := login1.GetAccount().GetUserAccountInfo().GetAddress()
		addr2 := login2.GetAccount().GetUserAccountInfo().GetAddress()
		if addr1.Hex() != addr2.Hex() {
			t.Fatal("got different accounts with same login")
		}
	})

	t.Run("new-user-multi-cred-with-pods", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()
		user1 := "multicredtester"
		// create user
		userObject := user.NewUsers(mockClient, ens, -1, 0, logger)
		pass := "password1password1"
		sr, err := userObject.CreateNewUserV2(user1, pass, "", "", tm, sm)
		if err != nil {
			t.Fatal(err)
		}

		_, err = userObject.CreateNewUserV2(user1, pass, "", "", tm, sm)
		if !errors.Is(err, user.ErrUserAlreadyPresent) {
			t.Fatal(err)
		}
		ui := sr.UserInfo
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
		err = ui.GetAccount().GetWallet().IsValidMnemonic(sr.Mnemonic)
		if err != nil {
			t.Fatalf("invalid mnemonic")
		}
		pod1 := ui.GetPod()
		podName1 := "test1"
		podName2 := "test2"
		podPassword, _ := utils.GetRandString(pod.PasswordLength)
		_, err = pod1.CreatePod(podName1, "", podPassword)
		if err != nil {
			t.Fatalf("error creating pod %s : %s", podName1, err.Error())
		}
		_, err = pod1.CreatePod(podName2, "", podPassword)
		if err != nil {
			t.Fatalf("error creating pod %s : %s", podName1, err.Error())
		}

		sr2, err := userObject.CreateNewUserV2(user1, pass+pass, sr.Mnemonic, "", tm, sm)
		if err != nil {
			t.Fatal(err)
		}
		ui2 := sr2.UserInfo

		lr1, err := userObject.LoginUserV2(user1, pass, mockClient, tm, sm, "")
		if err != nil {
			t.Fatal(err)
		}
		login1 := lr1.UserInfo
		lr2, err := userObject.LoginUserV2(user1, pass+pass, mockClient, tm, sm, "")
		if err != nil {
			t.Fatal(err)
		}
		login2 := lr2.UserInfo
		addr1 := login1.GetAccount().GetUserAccountInfo().GetAddress()
		addr2 := login2.GetAccount().GetUserAccountInfo().GetAddress()
		if addr1.Hex() != addr2.Hex() {
			t.Fatal("got different accounts with same login")
		}

		login1Pods, _, err := login1.GetPod().ListPods()
		if err != nil {
			t.Fatal(err)
		}
		login2Pods, _, err := login2.GetPod().ListPods()
		if err != nil {
			t.Fatal(err)
		}
		ui2Pods, _, err := ui2.GetPod().ListPods()
		if err != nil {
			t.Fatal(err)
		}

		sort.Strings(login1Pods)
		sort.Strings(login2Pods)
		sort.Strings(ui2Pods)

		for i, v := range login1Pods {
			if login2Pods[i] != v {
				t.Fatal("login two pod are different", login2Pods[i], v)
			}
			if ui2Pods[i] != v {
				t.Fatal("login two pod are different", ui2Pods[i], v)
			}
		}
	})
}
