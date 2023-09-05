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

package feed_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestFeed(t *testing.T) {
	logger := logging.New(io.Discard, 0)

	acc1 := account.New(logger)
	_, _, err := acc1.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	user1 := acc1.GetAddress(account.UserAccountIndex)
	accountInfo1 := acc1.GetUserAccountInfo()
	client := mock.NewMockBeeClient()

	t.Run("create-feed", func(t *testing.T) {
		fd := feed.New(accountInfo1, client, logger)
		topic := utils.HashString("topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		addr, err := fd.CreateFeed(user1, topic, data, nil)
		if err != nil {
			t.Fatal(err)
		}
		longTopic := append(topic, topic...) // skipcq: CRT-D0001
		_, _, err = fd.GetFeedData(longTopic, user1, nil, false)
		if !errors.Is(err, feed.ErrInvalidTopicSize) {
			t.Fatal("invalid topic size")
		}

		// check if the data and address is present and is same as stored
		rcvdAddr, rcvdData, err := fd.GetFeedData(topic, user1, nil, false)
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
		_, _, err = acc2.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		accountInfo2 := acc2.GetUserAccountInfo()

		// create feed from user1
		fd1 := feed.New(accountInfo1, client, logger)
		topic := utils.HashString("topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		addr, err := fd1.CreateFeed(user1, topic, data, nil)
		if err != nil {
			t.Fatal(err)
		}

		// check if you can read the data from user2
		fd2 := feed.New(accountInfo2, client, logger)
		rcvdAddr, rcvdData, err := fd2.GetFeedData(topic, user1, nil, false)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(addr, rcvdAddr) {
			t.Fatal("addresses do not match")
		}
		if !bytes.Equal(data, rcvdData) {
			t.Fatal("data does not match")
		}
	})

	t.Run("read-feed-first-time", func(t *testing.T) {
		fd := feed.New(accountInfo1, client, logger)
		topic := utils.HashString("topic2")

		// check if the data and address is present and is same as stored
		_, _, err := fd.GetFeedData(topic, user1, nil, false)
		if err != nil && err.Error() != "feed does not exist or was not updated yet" {
			t.Fatal(err)
		}
	})

	t.Run("create-from-user1-read-from-user2-with-user2-address", func(t *testing.T) {
		// create account2
		acc2 := account.New(logger)
		_, _, err = acc2.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		accountInfo2 := acc2.GetUserAccountInfo()
		user2 := acc2.GetAddress(account.UserAccountIndex)

		// create feed from user1
		fd1 := feed.New(accountInfo1, client, logger)
		topic := utils.HashString("topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		_, err := fd1.CreateFeed(user1, topic, data, nil)
		if err != nil {
			t.Fatal(err)
		}

		// check if you can read the data from user2
		fd2 := feed.New(accountInfo2, client, logger)
		rcvdAddr, rcvdData, err := fd2.GetFeedData(topic, user2, nil, false)
		if err != nil && err.Error() != "feed does not exist or was not updated yet" {
			t.Fatal(err)
		}
		if rcvdAddr != nil || rcvdData != nil {
			t.Fatal("was able to read feed of user1 using user2's address")
		}
	})

	t.Run("update-feed", func(t *testing.T) {
		fd := feed.New(accountInfo1, client, logger)
		topic := utils.HashString("topic3")
		data := []byte{0}
		_, err = fd.CreateFeed(user1, topic, data, nil)
		if err != nil {
			t.Fatal(err)
		}

		for i := 1; i < 256; i++ {
			buf := make([]byte, 4)
			binary.LittleEndian.PutUint16(buf, uint16(i))

			_, err = fd.UpdateFeed(user1, topic, buf, nil, false)
			if err != nil {
				t.Fatal(err)
			}
			getAddr, rcvdData, err := fd.GetFeedData(topic, user1, nil, false)
			if err != nil {
				t.Fatal(err)
			}
			if getAddr == nil {
				t.Fatal("invalid update address")
			}
			if !bytes.Equal(buf, rcvdData) {
				t.Fatal("data not matching", buf, rcvdData)
			}
		}
	})

	t.Run("create-feed-from-topic", func(t *testing.T) {
		fd := feed.New(accountInfo1, client, logger)
		topic := utils.HashString("feed-topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		addr, err := fd.CreateFeedFromTopic(topic, user1, data)
		if err != nil {
			t.Fatal(err)
		}
		longTopic := append(topic, topic...) // skipcq: CRT-D0001
		_, _, err = fd.GetFeedDataFromTopic(longTopic, user1)
		if !errors.Is(err, feed.ErrInvalidTopicSize) {
			t.Fatal("invalid topic size")
		}

		// check if the data and address is present and is same as stored
		rcvdAddr, rcvdData, err := fd.GetFeedDataFromTopic(topic, user1)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(addr, rcvdAddr) {
			t.Fatal("addresses do not match")
		}
		if !bytes.Equal(data, rcvdData) {
			t.Fatal("data does not match")
		}
	})

	t.Run("delete-feed-from-topic", func(t *testing.T) {
		fd := feed.New(accountInfo1, client, logger)
		topic := utils.HashString("feed-topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		_, err := fd.CreateFeedFromTopic(topic, user1, data)
		if err != nil {
			t.Fatal(err)
		}
		err = fd.DeleteFeedFromTopic(topic, user1)
		if err != nil {
			t.Fatal(err)
		}

		_, _, err = fd.GetFeedDataFromTopic(topic, user1)
		if err != nil && err.Error() != "error downloading data" {
			t.Fatal("error should be \"error downloading data\"")
		}
	})

	t.Run("create-feed-errors", func(t *testing.T) {
		nilFd := feed.New(&account.Info{}, client, logger)

		fd := feed.New(accountInfo1, client, logger)
		topic := utils.HashString("feed-topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

		_, err = nilFd.CreateFeed(user1, topic, data, nil)
		if !errors.Is(err, feed.ErrReadOnlyFeed) {
			t.Fatal("read only feed")
		}

		longTopic := append(topic, topic...) // skipcq: CRT-D0001
		_, err = fd.CreateFeed(user1, longTopic, data, nil)
		if !errors.Is(err, feed.ErrInvalidTopicSize) {
			t.Fatal("invalid topic size")
		}

		longData, err := utils.GetRandBytes(5000)
		if err != nil {
			t.Fatal(err)
		}
		_, err = fd.CreateFeed(user1, topic, longData, nil)
		if !errors.Is(err, feed.ErrInvalidPayloadSize) {
			t.Fatal("invalid payload size")
		}
	})

	t.Run("create-feed-from-topic-errors", func(t *testing.T) {
		nilFd := feed.New(&account.Info{}, client, logger)

		fd := feed.New(accountInfo1, client, logger)
		topic := utils.HashString("feed-topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

		_, err = nilFd.CreateFeedFromTopic(topic, user1, data)
		if !errors.Is(err, feed.ErrReadOnlyFeed) {
			t.Fatal("read only feed")
		}

		longTopic := append(topic, topic...) // skipcq: CRT-D0001
		_, err = fd.CreateFeedFromTopic(longTopic, user1, data)
		if !errors.Is(err, feed.ErrInvalidTopicSize) {
			t.Fatal("invalid topic size")
		}

		longData, err := utils.GetRandBytes(5000)
		if err != nil {
			t.Fatal(err)
		}
		_, err = fd.CreateFeedFromTopic(topic, user1, longData)
		if !errors.Is(err, feed.ErrInvalidPayloadSize) {
			t.Fatal("invalid payload size")
		}
	})

	t.Run("feed-update-errors", func(t *testing.T) {
		nilFd := feed.New(&account.Info{}, client, logger)

		fd := feed.New(accountInfo1, client, logger)
		topic := utils.HashString("feed-topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

		_, err = nilFd.UpdateFeed(user1, topic, data, nil, false)
		if !errors.Is(err, feed.ErrReadOnlyFeed) {
			t.Fatal("read only feed")
		}

		longTopic := append(topic, topic...) // skipcq: CRT-D0001
		_, err = fd.UpdateFeed(user1, longTopic, data, nil, false)
		if !errors.Is(err, feed.ErrInvalidTopicSize) {
			t.Fatal("invalid topic size")
		}

		longData, err := utils.GetRandBytes(5000)
		if err != nil {
			t.Fatal(err)
		}
		_, err = fd.UpdateFeed(user1, topic, longData, nil, false)
		if !errors.Is(err, feed.ErrInvalidPayloadSize) {
			t.Fatal("invalid payload size")
		}
	})

	t.Run("feed-delete-errors", func(t *testing.T) {
		nilFd := feed.New(&account.Info{}, client, logger)

		fd := feed.New(accountInfo1, client, logger)
		topic := utils.HashString("feed-topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

		err = nilFd.DeleteFeed(topic, user1)
		if !errors.Is(err, feed.ErrReadOnlyFeed) {
			t.Fatal("read only feed")
		}

		_, err = fd.CreateFeed(user1, topic, data, nil)
		if err != nil {
			t.Fatal(err)
		}

		err = fd.DeleteFeed(topic, user1)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("feed-from-topic-delete-errors", func(t *testing.T) {
		nilFd := feed.New(&account.Info{}, client, logger)
		topic := utils.HashString("feed-topic1")
		err = nilFd.DeleteFeedFromTopic(topic, user1)
		if !errors.Is(err, feed.ErrReadOnlyFeed) {
			t.Fatal("read only feed")
		}
	})
}
