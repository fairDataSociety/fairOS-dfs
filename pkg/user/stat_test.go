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
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
)

func TestStat(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(ioutil.Discard, 0)

	t.Run("stat-user", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()
		//create user
		userObject := user.NewUsers("", mockClient, ens, logger)
		_, _, _, _, ui, err := userObject.CreateNewUserV2("user1", "password1", "", "")
		if err != nil {
			t.Fatal(err)
		}

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
