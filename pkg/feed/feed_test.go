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
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/asabya/swarm-blockstore/bee"
	"github.com/asabya/swarm-blockstore/bee/mock"
	"github.com/ethersphere/bee/v2/pkg/file/redundancy"
	mockpost "github.com/ethersphere/bee/v2/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/v2/pkg/storer/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
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

	storer := mockstorer.New()
	beeUrl := mock.NewTestBeeServer(t, mock.TestServerOptions{
		Storer: storer,
		Post:   mockpost.New(mockpost.WithAcceptAll()),
	})
	client := bee.NewBeeClient(beeUrl, bee.WithStamp(mock.BatchOkStr), bee.WithRedundancy(fmt.Sprintf("%d", redundancy.NONE)), bee.WithPinning(true))

	t.Run("create-feed", func(t *testing.T) {
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		user := acc.GetAddress(account.UserAccountIndex)
		accountInfo := acc.GetUserAccountInfo()
		fd := feed.New(accountInfo, client, 500, 0, logger)
		topic := utils.HashString("topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		err = fd.CreateFeed(user, topic, data, nil)
		if err != nil {
			t.Fatal(err)
		}
		longTopic := append(topic, topic...) // skipcq: CRT-D0001
		_, _, err = fd.GetFeedData(longTopic, user, nil, false)
		if !errors.Is(err, feed.ErrInvalidTopicSize) {
			t.Fatal("invalid topic size")
		}
		<-time.After(3 * time.Second)
		// check if the data and address is present and is same as stored
		_, rcvdData, err := fd.GetFeedData(topic, user, nil, false)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(data, rcvdData) {
			t.Fatal(err)
		}
	})

	t.Run("create-feed-nil-feed-cache", func(t *testing.T) {
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		user := acc.GetAddress(account.UserAccountIndex)
		accountInfo := acc.GetUserAccountInfo()
		fd := feed.New(accountInfo, client, -1, 0, logger)
		topic := utils.HashString("topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		err = fd.CreateFeed(user, topic, data, nil)
		if err != nil {
			t.Fatal(err)
		}
		longTopic := append(topic, topic...) // skipcq: CRT-D0001
		_, _, err = fd.GetFeedData(longTopic, user, nil, false)
		if !errors.Is(err, feed.ErrInvalidTopicSize) {
			t.Fatal("invalid topic size")
		}
		<-time.After(3 * time.Second)
		// check if the data and address is present and is same as stored
		_, rcvdData, err := fd.GetFeedData(topic, user, nil, false)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(data, rcvdData) {
			t.Fatal(err)
		}
	})

	t.Run("create-from-user-read-from-user2-with-user-address", func(t *testing.T) {
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		user := acc.GetAddress(account.UserAccountIndex)
		accountInfo := acc.GetUserAccountInfo()

		// create feed from user
		fd1 := feed.New(accountInfo, client, 500, 0, logger)
		topic := utils.HashString("topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		err = fd1.CreateFeed(user, topic, data, nil)
		if err != nil {
			t.Fatal(err)
		}
		fd1.CommitFeeds()
		// check if you can read the data from user2
		fd2 := feed.New(accountInfo, client, -1, 0, logger)
		_, rcvdData, err := fd2.GetFeedData(topic, user, nil, false)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(data, rcvdData) {
			t.Fatal("data does not match")
		}
	})

	t.Run("create-from-user-read-from-user2-with-user-address-nil-feed-cache", func(t *testing.T) {
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		user := acc.GetAddress(account.UserAccountIndex)
		accountInfo := acc.GetUserAccountInfo()

		// create feed from user
		fd1 := feed.New(accountInfo, client, -1, 0, logger)
		topic := utils.HashString("topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		err = fd1.CreateFeed(user, topic, data, nil)
		if err != nil {
			t.Fatal(err)
		}
		fd1.CommitFeeds()
		// check if you can read the data from user2
		fd2 := feed.New(accountInfo, client, -1, 0, logger)
		_, rcvdData, err := fd2.GetFeedData(topic, user, nil, false)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(data, rcvdData) {
			t.Fatal("data does not match")
		}
	})

	t.Run("read-feed-first-time", func(t *testing.T) {
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		user := acc.GetAddress(account.UserAccountIndex)
		accountInfo := acc.GetUserAccountInfo()

		fd := feed.New(accountInfo, client, -1, 0, logger)
		topic := utils.HashString("topic2")

		// check if the data and address is present and is same as stored
		_, _, err = fd.GetFeedData(topic, user, nil, false)
		if err != nil && err.Error() != "feed does not exist or was not updated yet" {
			t.Fatal(err)
		}
	})

	t.Run("read-feed-created-from-different-user", func(t *testing.T) {
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		accountInfo := acc.GetUserAccountInfo()
		user := acc.GetAddress(account.UserAccountIndex)

		// create feed from user
		fd1 := feed.New(accountInfo, client, 500, 0, logger)
		topic := utils.HashString("topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		err = fd1.CreateFeed(user, topic, data, nil)
		if err != nil {
			t.Fatal(err)
		}

		acc2 := account.New(logger)
		_, _, err = acc2.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		accountInfo2 := acc2.GetUserAccountInfo()
		user2 := acc2.GetAddress(account.UserAccountIndex)

		// check if you can read the data from user2
		fd2 := feed.New(accountInfo2, client, -1, 0, logger)
		rcvdAddr, rcvdData, err := fd2.GetFeedData(topic, user2, nil, false)
		if err != nil && err.Error() != "feed does not exist or was not updated yet" {
			t.Fatal(err)
		}
		if rcvdAddr != nil || rcvdData != nil {
			t.Fatal("was able to read feed of user using user2's address")
		}
	})

	t.Run("read-feed-created-from-different-user-nil-feed-cache", func(t *testing.T) {
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		accountInfo := acc.GetUserAccountInfo()
		user := acc.GetAddress(account.UserAccountIndex)

		// create feed from user
		fd1 := feed.New(accountInfo, client, -1, 0, logger)
		topic := utils.HashString("topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		err = fd1.CreateFeed(user, topic, data, nil)
		if err != nil {
			t.Fatal(err)
		}

		acc2 := account.New(logger)
		_, _, err = acc2.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		accountInfo2 := acc2.GetUserAccountInfo()
		user2 := acc2.GetAddress(account.UserAccountIndex)

		// check if you can read the data from user2
		fd2 := feed.New(accountInfo2, client, -1, 0, logger)
		rcvdAddr, rcvdData, err := fd2.GetFeedData(topic, user2, nil, false)
		if err != nil && err.Error() != "feed does not exist or was not updated yet" {
			t.Fatal(err)
		}
		if rcvdAddr != nil || rcvdData != nil {
			t.Fatal("was able to read feed of user using user2's address")
		}
	})

	t.Run("update-feed", func(t *testing.T) {

		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		user := acc.GetAddress(account.UserAccountIndex)
		accountInfo := acc.GetUserAccountInfo()
		fd := feed.New(accountInfo, client, 500, 0, logger)
		topic := utils.HashString("topic3")
		data := []byte{0}
		err = fd.CreateFeed(user, topic, data, nil)
		if err != nil {
			t.Fatal(err)
		}
		var finalData []byte

		for i := 1; i < 256; i++ {
			buf := make([]byte, 4)
			_, _ = rand.Read(buf)
			if i == 255 {
				finalData = buf
			}
			err = fd.UpdateFeed(user, topic, buf, nil, false)
			if err != nil {
				t.Fatal(err)
			}
			_, rcvdData, err := fd.GetFeedData(topic, user, nil, false)
			if err != nil {
				t.Fatal(err)
			}

			require.Equal(t, buf, rcvdData)
		}

		fd.CommitFeeds()
		<-time.After(time.Second)
		_, rcvdData, err := fd.GetFeedData(topic, user, nil, false)
		if err != nil {
			t.Fatal(err)
		}

		require.Equal(t, finalData, rcvdData)
	})

	t.Run("update-feed-read-from-different-user", func(t *testing.T) {

		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		user := acc.GetAddress(account.UserAccountIndex)
		accountInfo := acc.GetUserAccountInfo()
		fd := feed.New(accountInfo, client, 500, 0, logger)

		topic := utils.HashString("topic3")
		data := []byte{0}
		err = fd.CreateFeed(user, topic, data, nil)
		if err != nil {
			t.Fatal(err)
		}

		for i := 1; i < 256; i++ {
			buf := make([]byte, 4)
			_, _ = rand.Read(buf)
			err = fd.UpdateFeed(user, topic, buf, nil, false)
			if err != nil {
				t.Fatal(err)
			}
			fd.CommitFeeds()
			<-time.After(time.Second)
			acc2 := account.New(logger)
			_, _, err = acc2.CreateUserAccount("")
			if err != nil {
				t.Fatal(err)
			}
			accountInfo2 := acc2.GetUserAccountInfo()

			// check if you can read the data from user2
			fd2 := feed.New(accountInfo2, client, -1, 0, logger)
			_, rcvdData2, err := fd2.GetFeedData(topic, user, nil, false)
			if err != nil && err.Error() != "feed does not exist or was not updated yet" {
				t.Fatal(err)
			}

			require.Equal(t, buf, rcvdData2)
		}
	})

	t.Run("update-feed-nil-feed-cache", func(t *testing.T) {

		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		user := acc.GetAddress(account.UserAccountIndex)
		accountInfo := acc.GetUserAccountInfo()
		fd := feed.New(accountInfo, client, 500, 0, logger)
		topic := utils.HashString("topic3")
		data := []byte{0}
		err = fd.CreateFeed(user, topic, data, nil)
		if err != nil {
			t.Fatal(err)
		}
		var finalData []byte

		for i := 1; i < 256; i++ {
			buf := make([]byte, 4)
			_, _ = rand.Read(buf)
			if i == 255 {
				finalData = buf
			}
			err = fd.UpdateFeed(user, topic, buf, nil, false)
			if err != nil {
				t.Fatal(err)
			}
			_, rcvdData, err := fd.GetFeedData(topic, user, nil, false)
			if err != nil {
				t.Fatal(err)
			}

			require.Equal(t, buf, rcvdData)
			<-time.After(time.Second)
		}

		fd.CommitFeeds()
		<-time.After(time.Second)
		_, rcvdData, err := fd.GetFeedData(topic, user, nil, false)
		if err != nil {
			t.Fatal(err)
		}

		require.Equal(t, finalData, rcvdData)
	})

	t.Run("create-feed-from-topic", func(t *testing.T) {
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		user := acc.GetAddress(account.UserAccountIndex)
		accountInfo := acc.GetUserAccountInfo()
		fd := feed.New(accountInfo, client, -1, 0, logger)
		topic := utils.HashString("feed-topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		addr, err := fd.CreateFeedFromTopic(topic, user, data)
		if err != nil {
			t.Fatal(err)
		}
		longTopic := append(topic, topic...) // skipcq: CRT-D0001
		_, _, err = fd.GetFeedDataFromTopic(longTopic, user)
		if !errors.Is(err, feed.ErrInvalidTopicSize) {
			t.Fatal("invalid topic size")
		}

		// check if the data and address is present and is same as stored
		rcvdAddr, rcvdData, err := fd.GetFeedDataFromTopic(topic, user)
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
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		user := acc.GetAddress(account.UserAccountIndex)
		accountInfo := acc.GetUserAccountInfo()
		fd := feed.New(accountInfo, client, -1, 0, logger)
		topic := utils.HashString("feed-topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		_, err = fd.CreateFeedFromTopic(topic, user, data)
		if err != nil {
			t.Fatal(err)
		}
		err = fd.DeleteFeedFromTopic(topic, user)
		if err != nil {
			t.Fatal(err)
		}

		_, _, err = fd.GetFeedDataFromTopic(topic, user)
		if err != nil && err.Error() != "error downloading data" {
			t.Fatal("error should be \"error downloading data\"")
		}
	})

	t.Run("create-feed-errors", func(t *testing.T) {
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		user := acc.GetAddress(account.UserAccountIndex)
		accountInfo := acc.GetUserAccountInfo()
		nilFd := feed.New(&account.Info{}, client, -1, 0, logger)

		fd := feed.New(accountInfo, client, -1, 0, logger)
		topic := utils.HashString("feed-topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

		err = nilFd.CreateFeed(user, topic, data, nil)
		if !errors.Is(err, feed.ErrReadOnlyFeed) {
			t.Fatal("read only feed")
		}

		longTopic := append(topic, topic...) // skipcq: CRT-D0001
		err = fd.CreateFeed(user, longTopic, data, nil)
		if !errors.Is(err, feed.ErrInvalidTopicSize) {
			t.Fatal("invalid topic size")
		}

		longData, err := utils.GetRandBytes(5000)
		if err != nil {
			t.Fatal(err)
		}
		err = fd.CreateFeed(user, topic, longData, nil)
		if !errors.Is(err, feed.ErrInvalidPayloadSize) {
			t.Fatal("invalid payload size")
		}
	})

	t.Run("create-feed-from-topic-errors", func(t *testing.T) {
		nilFd := feed.New(&account.Info{}, client, -1, 0, logger)
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		user := acc.GetAddress(account.UserAccountIndex)
		accountInfo := acc.GetUserAccountInfo()
		fd := feed.New(accountInfo, client, -1, 0, logger)
		topic := utils.HashString("feed-topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

		_, err = nilFd.CreateFeedFromTopic(topic, user, data)
		if !errors.Is(err, feed.ErrReadOnlyFeed) {
			t.Fatal("read only feed")
		}

		longTopic := append(topic, topic...) // skipcq: CRT-D0001
		_, err = fd.CreateFeedFromTopic(longTopic, user, data)
		if !errors.Is(err, feed.ErrInvalidTopicSize) {
			t.Fatal("invalid topic size")
		}

		longData, err := utils.GetRandBytes(5000)
		if err != nil {
			t.Fatal(err)
		}
		_, err = fd.CreateFeedFromTopic(topic, user, longData)
		if !errors.Is(err, feed.ErrInvalidPayloadSize) {
			t.Fatal("invalid payload size")
		}
	})

	t.Run("feed-update-errors", func(t *testing.T) {
		nilFd := feed.New(&account.Info{}, client, -1, 0, logger)
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		user := acc.GetAddress(account.UserAccountIndex)
		accountInfo := acc.GetUserAccountInfo()
		fd := feed.New(accountInfo, client, -1, 0, logger)
		topic := utils.HashString("feed-topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

		err = nilFd.UpdateFeed(user, topic, data, nil, false)
		if !errors.Is(err, feed.ErrReadOnlyFeed) {
			t.Fatal("read only feed")
		}

		longTopic := append(topic, topic...) // skipcq: CRT-D0001
		err = fd.UpdateFeed(user, longTopic, data, nil, false)
		if !errors.Is(err, feed.ErrInvalidTopicSize) {
			t.Fatal("invalid topic size")
		}

		longData, err := utils.GetRandBytes(5000)
		if err != nil {
			t.Fatal(err)
		}
		err = fd.UpdateFeed(user, topic, longData, nil, false)
		if !errors.Is(err, feed.ErrInvalidPayloadSize) {
			t.Fatal("invalid payload size")
		}
	})

	t.Run("feed-delete-errors", func(t *testing.T) {
		nilFd := feed.New(&account.Info{}, client, -1, 0, logger)
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		user := acc.GetAddress(account.UserAccountIndex)
		accountInfo := acc.GetUserAccountInfo()
		fd := feed.New(accountInfo, client, -1, 0, logger)
		topic := utils.HashString("feed-topic1")
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

		err = nilFd.DeleteFeed(topic, user)
		if !errors.Is(err, feed.ErrReadOnlyFeed) {
			t.Fatal("read only feed")
		}

		err = fd.CreateFeed(user, topic, data, nil)
		if err != nil {
			t.Fatal(err)
		}

		err = fd.DeleteFeed(topic, user)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("feed-from-topic-delete-errors", func(t *testing.T) {
		nilFd := feed.New(&account.Info{}, client, -1, 0, logger)
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}
		user := acc.GetAddress(account.UserAccountIndex)
		topic := utils.HashString("feed-topic1")
		err = nilFd.DeleteFeedFromTopic(topic, user)
		if !errors.Is(err, feed.ErrReadOnlyFeed) {
			t.Fatal("read only feed")
		}
	})
}
