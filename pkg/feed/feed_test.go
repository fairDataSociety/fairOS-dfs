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

package feed

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func TestFeed(t *testing.T) {
	logger := logging.New(ioutil.Discard, 0)

	acc1 := account.New(logger)
	_, _, err := acc1.CreateUserAccount("password", "")
	if err != nil {
		t.Fatal(err)
	}
	user1 := acc1.GetAddress(account.UserAccountIndex)
	accountInfo1 := acc1.GetUserAccountInfo()
	client := mock.NewMockBeeClient()

	t.Run("create-feed", func(t *testing.T) {
		fd := New(accountInfo1, client, logger)
		topic := utils.HashString("topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		addr, err := fd.CreateFeed(topic, user1, data)
		if err != nil {
			t.Fatal(err)
		}

		// check if the data and address is present and is same as stored
		rcvdAddr, rcvdData, err := fd.GetFeedData(topic, user1)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(addr, rcvdAddr) {
			t.Fatal(err)
		}
		if !bytes.Equal(data, rcvdData) {
			t.Fatal(err)
		}
	})

	t.Run("create-from-user1-read-from-user2-with-user1-address", func(t *testing.T) {
		// create account2
		acc2 := account.New(logger)
		_, _, err = acc2.CreateUserAccount("password", "")
		if err != nil {
			t.Fatal(err)
		}
		accountInfo2 := acc2.GetUserAccountInfo()

		// create feed from user1
		fd1 := New(accountInfo1, client, logger)
		topic := utils.HashString("topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		addr, err := fd1.CreateFeed(topic, user1, data)
		if err != nil {
			t.Fatal(err)
		}

		// check if you can read the data from user2
		fd2 := New(accountInfo2, client, logger)
		rcvdAddr, rcvdData, err := fd2.GetFeedData(topic, user1)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(addr, rcvdAddr) {
			t.Fatal(err)
		}
		if !bytes.Equal(data, rcvdData) {
			t.Fatal(err)
		}
	})

	t.Run("read-feed-first-time", func(t *testing.T) {
		fd := New(accountInfo1, client, logger)
		topic := utils.HashString("topic2")

		// check if the data and address is present and is same as stored
		_, _, err := fd.GetFeedData(topic, user1)
		if err != nil && err.Error() != "no feed updates found" {
			t.Fatal(err)
		}
	})

	t.Run("create-from-user1-read-from-user2-with-user2-address", func(t *testing.T) {
		// create account2
		acc2 := account.New(logger)
		_, _, err = acc2.CreateUserAccount("password", "")
		if err != nil {
			t.Fatal(err)
		}
		accountInfo2 := acc2.GetUserAccountInfo()
		user2 := acc2.GetAddress(account.UserAccountIndex)

		// create feed from user1
		fd1 := New(accountInfo1, client, logger)
		topic := utils.HashString("topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		_, err := fd1.CreateFeed(topic, user1, data)
		if err != nil {
			t.Fatal(err)
		}

		// check if you can read the data from user2
		fd2 := New(accountInfo2, client, logger)
		rcvdAddr, rcvdData, err := fd2.GetFeedData(topic, user2)
		if err != nil  && err.Error() != "no feed updates found" {
			t.Fatal(err)
		}
		if rcvdAddr != nil || rcvdData != nil {
			t.Fatal("was able to read feed of user1 using user2's address")
		}
	})


	t.Run("update-feed", func(t *testing.T) {
		fd := New(accountInfo1, client, logger)
		topic := utils.HashString("topic3")
		data := []byte{0}
		_, err = fd.CreateFeed(topic, user1, data)
		if err != nil {
			t.Fatal(err)
		}

		for i := 1; i < 256; i++ {
			buf := make([]byte, 4)
			binary.LittleEndian.PutUint16(buf, uint16(i))
			_, err = fd.UpdateFeed(topic, user1, buf)
			if err != nil {
				t.Fatal(err)
			}
			getAddr, rcvdData, err := fd.GetFeedData(topic, user1)
			if err != nil {
				t.Fatal(err)
			}
			if getAddr == nil {
				t.Fatal("invalid update address")
			}
			if !bytes.Equal(buf, rcvdData) {
				t.Fatal(err)
			}
			fmt.Println("update ", i, " Done")
		}
	})
}
