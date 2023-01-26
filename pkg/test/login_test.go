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

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/plexsysio/taskmanager"
)

func TestLogin(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()

	t.Run("login-user", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()
		// create user
		userObject := user.NewUsers(mockClient, ens, logger)
		_, _, _, _, ui, err := userObject.CreateNewUserV2("7e4567e7cb003804992eef11fd5c757275a4c", "password1twelve", "", "", tm)
		if err != nil {
			t.Fatal(err)
		}

		// Logout user
		err = userObject.LogoutUser(ui.GetUserName(), ui.GetSessionId())
		if err != nil {
			t.Fatal(err)
		}

		_, _, _, err = userObject.LoginUserV2("not_an_username", "password1", mockClient, tm, "")
		if !errors.Is(err, user.ErrUserNameNotFound) {
			t.Fatal(err)
		}

		_, _, _, err = userObject.LoginUserV2("7e4567e7cb003804992eef11fd5c757275a4c", "wrong_password", mockClient, tm, "")
		if !errors.Is(err, user.ErrInvalidPassword) {
			t.Fatal(err)
		}

		// addUserAndSessionToMap user again
		ui1, _, _, err := userObject.LoginUserV2("7e4567e7cb003804992eef11fd5c757275a4c", "password1twelve", mockClient, tm, "")
		if err != nil {
			t.Fatal(err)
		}

		ui2 := userObject.GetLoggedInUserInfo(ui1.GetSessionId())

		// Validate login
		if !userObject.IsUserNameLoggedIn("7e4567e7cb003804992eef11fd5c757275a4c") {
			t.Fatalf("user not loggin in")
		}

		if ui.GetAccount().GetUserAccountInfo().GetAddress().Hex() != ui1.GetAccount().GetUserAccountInfo().GetAddress().Hex() {
			t.Fatal("loaded with different account")
		}

		if ui.GetAccount().GetUserAccountInfo().GetAddress().Hex() != ui2.GetAccount().GetUserAccountInfo().GetAddress().Hex() {
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
		//create user
		userObject := user.NewUsers(mockClient, ens, logger)
		_, mnemonic, _, _, ui, err := userObject.CreateNewUserV2(user1, pass, "", "", tm)
		if err != nil {
			t.Fatal(err)
		}

		_, _, _, _, _, err = userObject.CreateNewUserV2(user1, pass, "", "", tm)
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

		_, _, _, _, _, err = userObject.CreateNewUserV2(user1, pass+pass, mnemonic, "", tm)
		if err != nil {
			t.Fatal(err)
		}

		login1, _, _, err := userObject.LoginUserV2(user1, pass, mockClient, tm, "")
		if err != nil {
			t.Fatal(err)
		}

		login2, _, _, err := userObject.LoginUserV2(user1, pass+pass, mockClient, tm, "")
		if err != nil {
			t.Fatal(err)
		}

		if login1.GetAccount().GetUserAccountInfo().GetAddress().Hex() !=
			login2.GetAccount().GetUserAccountInfo().GetAddress().Hex() {
			t.Fatal("got different accounts with same login")
		}
	})

	t.Run("new-user-multi-cred-with-pods", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()
		user1 := "multicredtester"
		//create user
		userObject := user.NewUsers(mockClient, ens, logger)
		pass := "password1password1"
		_, mnemonic, _, _, ui, err := userObject.CreateNewUserV2(user1, pass, "", "", tm)
		if err != nil {
			t.Fatal(err)
		}

		_, _, _, _, _, err = userObject.CreateNewUserV2(user1, pass, "", "", tm)
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

		_, _, _, _, ui2, err := userObject.CreateNewUserV2(user1, pass+pass, mnemonic, "", tm)
		if err != nil {
			t.Fatal(err)
		}

		login1, _, _, err := userObject.LoginUserV2(user1, pass, mockClient, tm, "")
		if err != nil {
			t.Fatal(err)
		}

		login2, _, _, err := userObject.LoginUserV2(user1, pass+pass, mockClient, tm, "")
		if err != nil {
			t.Fatal(err)
		}

		if login1.GetAccount().GetUserAccountInfo().GetAddress().Hex() !=
			login2.GetAccount().GetUserAccountInfo().GetAddress().Hex() {
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
