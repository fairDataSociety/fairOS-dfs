package user

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed/lookup"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestTimeKeeper(t *testing.T) {
	logger := logging.New(io.Discard, 0)

	acc1 := account.New(logger)
	_, _, err := acc1.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	user1 := acc1.GetAddress(account.UserAccountIndex)
	accountInfo1 := acc1.GetUserAccountInfo()
	client := mock.NewMockBeeClient()

	t.Run("level-get-from-same-feed-pointer", func(t *testing.T) {
		fd1 := feed.New(accountInfo1, client, logger)
		db, err := leveldb.Open(NewMemStorage(fd1, client, user1, "username", "password"), nil)
		if err != nil {
			t.Fatal(err)
		}
		fd1.SetUpdateTracker(db)

		topicOne := utils.HashString("topicOne")
		_, err = fd1.GetFeedUpdateEpoch(topicOne)
		if err == nil {
			t.Fatal("feed should not exist")
		}

		now := time.Now().Unix()
		err = fd1.PutFeedUpdateEpoch(topicOne, lookup.Epoch{
			Time:  uint64(now),
			Level: 31,
		})
		if err != nil {
			t.Fatal(err)
		}

		epoch, err := fd1.GetFeedUpdateEpoch(topicOne)
		if err != nil {
			t.Fatal(err)
		}

		require.Equal(t, uint64(now), epoch.Time)

		err = db.Close()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("level-get-from-different-feed-pointer", func(t *testing.T) {
		fd1 := feed.New(accountInfo1, client, logger)
		db, err := leveldb.Open(NewMemStorage(fd1, client, user1, "username", "password"), nil)
		if err != nil {
			t.Fatal(err)
		}
		fd1.SetUpdateTracker(db)

		topic := utils.HashString("topicTwo")
		_, err = fd1.GetFeedUpdateEpoch(topic)
		if err == nil {
			t.Fatal("feed should not exist")
		}

		now := time.Now().Unix()
		err = fd1.PutFeedUpdateEpoch(topic, lookup.Epoch{
			Time:  uint64(now),
			Level: 31,
		})
		if err != nil {
			t.Fatal(err)
		}
		err = db.Close()
		if err != nil {
			t.Fatal(err)
		}

		fd2 := feed.New(accountInfo1, client, logger)
		db2, err := leveldb.Open(NewMemStorage(fd2, client, user1, "username", "password"), nil)
		if err != nil {
			t.Fatal(err)
		}
		fd2.SetUpdateTracker(db2)
		epoch, err := fd2.GetFeedUpdateEpoch(topic)
		if err != nil {
			t.Fatal(err)
		}

		require.Equal(t, uint64(now), epoch.Time)
	})
}
