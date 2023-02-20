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

	"github.com/plexsysio/taskmanager"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
)

func TestDelete(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
	tm := taskmanager.New(1, 10, time.Second*15, logger)
	defer func() {
		_ = tm.Stop(context.Background())
	}()

	t.Run("delete-user", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()
		// create user
		userObject := user.NewUsers(mockClient, ens, logger)
		_, _, _, _, ui, err := userObject.CreateNewUserV2("user1", "password1twelve", "", "", tm)
		if err != nil {
			t.Fatal(err)
		}
		// delete user with wrong password
		err = userObject.DeleteUserV2("user1", "password11", ui.GetSessionId(), ui)
		if err == nil {
			t.Fatal("delete should fail")
		}
		// delete user invalid sessionId
		err = userObject.DeleteUserV2("user1", "password1", "invalid_session", ui)
		if !errors.Is(err, user.ErrUserNotLoggedIn) {
			t.Fatal(err)
		}
		// delete user
		err = userObject.DeleteUserV2("user1", "password1twelve", ui.GetSessionId(), ui)
		if err != nil {
			t.Fatal(err)
		}

		// validate deletion
		if userObject.IsUserNameLoggedIn("user1") {
			t.Fatalf("user not deleted")
		}
	})
}
