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
	"io/ioutil"
	"os"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
)

func TestLogout(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(ioutil.Discard, 0)

	t.Run("logout-user", func(t *testing.T) {
		dataDir, err := ioutil.TempDir("", "logout")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dataDir)

		//create user
		userObject := user.NewUsers(dataDir, mockClient, logger)
		_, _, ui, err := userObject.CreateNewUser("user1", "password1", "", "")
		if err != nil {
			t.Fatal(err)
		}

		// Logout user
		err = userObject.LogoutUser(ui.GetUserName(), dataDir, ui.GetSessionId())
		if err != nil {
			t.Fatal(err)
		}

		// Validate logout
		if userObject.IsUserNameLoggedIn(ui.GetUserName()) {
			t.Fatalf("user still loggin in")
		}
		if !userObject.IsUsernameAvailable(ui.GetUserName(), dataDir) {
			t.Fatalf("user not created")
		}
	})

}
