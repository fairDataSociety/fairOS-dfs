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

func TestImport(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(ioutil.Discard, 0)

	t.Run("import-user", func(t *testing.T) {
		dataDir1, err := ioutil.TempDir("", "import")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dataDir1)

		//create user to export
		userObject1 := user.NewUsers(dataDir1, mockClient, logger)
		_, _, ui, err := userObject1.CreateNewUser("user1", "password1", "", "")
		if err != nil {
			t.Fatal(err)
		}

		// export user
		userName, address, err := userObject1.ExportUser(ui)
		if err != nil {
			t.Fatal(err)
		}

		// import user
		dataDir2, err := ioutil.TempDir("", "import")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dataDir2)
		userObject2 := user.NewUsers(dataDir2, mockClient, logger)
		_, err = userObject2.ImportUsingAddress(userName, "password1", address, dataDir2, mockClient, "")
		if err != nil {
			t.Fatal(err)
		}

		// validate import
		if !userObject2.IsUsernameAvailable("user1", dataDir2) {
			t.Fatalf("user not created")
		}
		if !userObject2.IsUserNameLoggedIn("user1") {
			t.Fatalf("user not loggin in")
		}

	})
}
