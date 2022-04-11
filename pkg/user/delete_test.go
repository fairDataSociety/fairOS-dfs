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
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/fnm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
)

func TestDelete(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(ioutil.Discard, 0)

	t.Run("delete-user", func(t *testing.T) {
		fnm := mock2.NewMockNamespaceManager()
		//create user
		userObject := user.NewUsers(mockClient, fnm, logger)
		_, _, _, _, ui, err := userObject.CreateNewUser("user1", "password1", "", "")
		if err != nil {
			t.Fatal(err)
		}

		// delete user
		err = userObject.DeleteUser("user1", "password1", ui.GetSessionId(), ui)
		if err != nil {
			t.Fatal(err)
		}

		// validate deletion
		if userObject.IsUserNameLoggedIn("user1") {
			t.Fatalf("user not deleted")
		}
		// TODO test if user is deleted
	})
}
