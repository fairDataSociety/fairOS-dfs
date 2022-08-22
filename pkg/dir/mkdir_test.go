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

package dir_test

import (
	"io"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	bm "github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dir"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	fm "github.com/fairdatasociety/fairOS-dfs/pkg/file/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

func TestMkdir(t *testing.T) {
	mockClient := bm.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("password", "")
	if err != nil {
		t.Fatal(err)
	}
	pod1AccountInfo, err := acc.CreatePodAccount(1, "password", false)
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(pod1AccountInfo, mockClient, logger)
	user := acc.GetAddress(1)
	mockFile := fm.NewMockFile()
	t.Run("simple-mkdir", func(t *testing.T) {
		dirObject := dir.NewDirectory("pod1", mockClient, fd, user, mockFile, logger)

		// make root dir so that other directories can be added
		err = dirObject.MkRootDir("pod1", user, fd)
		if err != nil {
			t.Fatal(err)
		}

		// create a new dir
		err := dirObject.MkDir("/baseDir")
		if err != nil {
			t.Fatal(err)
		}

		// validate dir
		dirs, _, err := dirObject.ListDir("/")
		if err != nil {
			t.Fatal(err)
		}
		if len(dirs) != 1 {
			t.Fatalf("invalid directory count")
		}
		if dirs[0].Name != "baseDir" {
			t.Fatalf("invalid directory name")
		}
	})
	t.Run("too-many-dirs", func(t *testing.T) {
		dirObject := dir.NewDirectory("pod1", mockClient, fd, user, mockFile, logger)

		// make root dir so that other directories can be added
		err = dirObject.MkRootDir("pod1", user, fd)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 37; i++ {
			name, err := utils.GetRandString(100)
			if err != nil {
				t.Fatal(err)
			}
			// create a new dir
			err = dirObject.MkDir("/" + name)
			if err != nil {
				t.Fatal("i", i, err)
			}
			for j := 0; j < 36; j++ {
				name2, err := utils.GetRandString(100)
				if err != nil {
					t.Fatal(err)
				}
				path := "/" + name + "/" + name2
				// create a new dir
				err = dirObject.MkDir("/" + path)
				if err != nil {
					t.Fatal("j", j, err)
				}
			}
		}
	})
	t.Run("complicated-mkdir", func(t *testing.T) {
		dirObject := dir.NewDirectory("pod1", mockClient, fd, user, mockFile, logger)

		// make root dir so that other directories can be added
		err = dirObject.MkRootDir("pod1", user, fd)
		if err != nil {
			t.Fatal(err)
		}

		// try to create a new dir without creating root
		err := dirObject.MkDir("/baseDir/baseDir2/baseDir3/baseDir4")
		if err == nil || err != dir.ErrDirectoryNotPresent {
			t.Fatal(err)
		}

		err = dirObject.MkDir("/baseDir")
		if err != nil {
			t.Fatal(err)
		}

		err = dirObject.MkDir("/baseDir/baseDir2")
		if err != nil {
			t.Fatal(err)
		}

		err = dirObject.MkDir("/baseDir/baseDir2/baseDir3")
		if err != nil {
			t.Fatal(err)
		}

		// validate dir
		dirs, _, err := dirObject.ListDir("/baseDir")
		if err != nil {
			t.Fatal(err)
		}
		if len(dirs) != 1 {
			t.Fatalf("invalid directory count")
		}
		if dirs[0].Name != "baseDir2" {
			t.Fatalf("invalid directory name")
		}

		dirs, _, err = dirObject.ListDir("/baseDir/baseDir2")
		if err != nil {
			t.Fatal(err)
		}
		if len(dirs) != 1 {
			t.Fatalf("invalid directory count")
		}
		if dirs[0].Name != "baseDir3" {
			t.Fatalf("invalid directory name")
		}
	})
}
