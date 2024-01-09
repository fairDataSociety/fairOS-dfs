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
	"io"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/file"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"

	mockpost "github.com/ethersphere/bee/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/sirupsen/logrus"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
)

func TestGroupNew(t *testing.T) {
	storer := mockstorer.New()
	beeUrl := mock.NewTestBeeServer(t, mock.TestServerOptions{
		Storer:          storer,
		PreventRedirect: true,
		Post:            mockpost.New(mockpost.WithAcceptAll()),
	})

	logger := logging.New(io.Discard, logrus.DebugLevel)
	mockClient := bee.NewBeeClient(beeUrl, mock.BatchOkStr, true, logger)
	acc := account.New(logger)
	_, _, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("create-first-group", func(t *testing.T) {
		fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)

		group := pod.NewGroup(mockClient, fd, acc, logger)
		groupName1 := "test1"
		_, err = group.CreateGroup(groupName1)
		if err != nil {
			t.Fatalf("error creating group %s: %s", groupName1, err.Error())
		}

		groups, err := group.ListGroup()
		if err != nil {
			t.Fatalf("error getting groups")
		}

		if len(groups.Groups) != 1 {
			t.Fatalf("length of groups is not 1")
		}

		_, err = group.OpenGroup(groupName1)
		if err != nil {
			t.Fatalf("error opening group %s: %s", groupName1, err.Error())
		}
	})

	t.Run("create-second-group", func(t *testing.T) {
		fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)

		group := pod.NewGroup(mockClient, fd, acc, logger)
		groupName1 := "test11"
		groupName2 := "test21"
		_, err = group.CreateGroup(groupName1)
		if err != nil {
			t.Fatalf("error creating group %s: %s", groupName1, err.Error())
		}
		_, err = group.CreateGroup(groupName2)
		if err != nil {
			t.Fatalf("error creating group %s: %s", groupName2, err.Error())
		}
		_, err = group.ListGroup()
		if err != nil {
			t.Fatalf("error getting groups")
		}

		_, err = group.OpenGroup(groupName2)
		if err != nil {
			t.Fatalf("error opening group %s: %s", groupName2, err.Error())
		}
	})

	t.Run("group-file-upload", func(t *testing.T) {
		fd := feed.New(acc.GetUserAccountInfo(), mockClient, -1, 0, logger)

		group := pod.NewGroup(mockClient, fd, acc, logger)
		groupName1 := "test12"
		_, err = group.CreateGroup(groupName1)
		if err != nil {
			t.Fatalf("error creating group %s: %s", groupName1, err.Error())
		}

		_, err = group.ListGroup()
		if err != nil {
			t.Fatalf("error getting groups")
		}

		g, err := group.OpenGroup(groupName1)
		if err != nil {
			t.Fatalf("error opening group %s: %s", groupName1, err.Error())
		}

		maxfiles := 100
		filePath := "/"
		for i := 1; i <= maxfiles; i++ {
			fileName, _ := utils.GetRandString(100)
			compression := ""
			fileSize := int64(1000)
			blockSize := file.MinBlockSize
			_, err = uploadFile(t, g.GetFile(), filePath, fileName, compression, g.GetPodPassword(), fileSize, blockSize)
			if err != nil {
				t.Fatal(err)
			}
			err = g.GetDirectory().AddEntryToDir("/", g.GetPodPassword(), fileName, true)
			if err != nil {
				t.Fatal(i, err)
			}
		}

		dirInode, err := g.GetDirectory().GetInode(g.GetPodPassword(), filePath)
		if err != nil {
			t.Fatal(err)
		}
		if len(dirInode.FileOrDirNames) != maxfiles {
			t.Fatal("files not present")
		}
	})
}
