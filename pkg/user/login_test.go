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

package user_test

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/plexsysio/taskmanager"
)

func TestLogin(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)

	t.Run("login-user", func(t *testing.T) {
		tm := taskmanager.New(1, 10, time.Second*15, logger)

		ens := mock2.NewMockNamespaceManager()
		//create user
		userObject := user.NewUsers("", mockClient, ens, logger)
		_, _, _, _, ui, err := userObject.CreateNewUserV2("7e4567e7cb003804992eef11fd5c757275a4c", "password1", "", "", tm)
		if err != nil {
			t.Fatal(err)
		}

		// Logout user
		err = userObject.LogoutUser(ui.GetUserName(), ui.GetSessionId())
		if err != nil {
			t.Fatal(err)
		}

		_, _, _, err = userObject.LoginUserV2("not_an_username", "password1", mockClient, tm, "")
		if !errors.Is(err, user.ErrInvalidUserName) {
			t.Fatal(err)
		}

		_, _, _, err = userObject.LoginUserV2("7e4567e7cb003804992eef11fd5c757275a4c", "wrong_password", mockClient, tm, "")
		if !errors.Is(err, user.ErrInvalidPassword) {
			t.Fatal(err)
		}

		// addUserAndSessionToMap user again
		ui1, _, _, err := userObject.LoginUserV2("7e4567e7cb003804992eef11fd5c757275a4c", "password1", mockClient, tm, "")
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

}
