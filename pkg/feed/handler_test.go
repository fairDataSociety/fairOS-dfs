package feed

import (
	"fmt"
	"io"
	"testing"

	"github.com/ethersphere/bee/v2/pkg/file/redundancy"

	"github.com/asabya/swarm-blockstore/bee"
	"github.com/asabya/swarm-blockstore/bee/mock"
	mockpost "github.com/ethersphere/bee/v2/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/v2/pkg/storer/mock"
	"github.com/ethersphere/bee/v2/pkg/swarm"
	bmtlegacy "github.com/ethersphere/bmt/legacy"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

func TestHandler(t *testing.T) {
	logger := logging.New(io.Discard, 0)

	storer := mockstorer.New()
	beeUrl := mock.NewTestBeeServer(t, mock.TestServerOptions{
		Storer: storer,
		Post:   mockpost.New(mockpost.WithAcceptAll()),
	})
	client := bee.NewBeeClient(beeUrl, bee.WithStamp(mock.BatchOkStr), bee.WithRedundancy(fmt.Sprintf("%d", redundancy.NONE)), bee.WithPinning(true))

	t.Run("new-handler", func(t *testing.T) {
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}

		accountInfo := acc.GetUserAccountInfo()
		bmtPool := bmtlegacy.NewTreePool(hashFunc, swarm.Branches, bmtlegacy.PoolSize)
		handler := NewHandler(accountInfo, client, bmtPool, -1, 0, logger)
		//defer handler.Close()

		if handler == nil {
			t.Fatal("handler is nil")
		}
	})

	t.Run("new-handler", func(t *testing.T) {
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}

		accountInfo := acc.GetUserAccountInfo()
		bmtPool := bmtlegacy.NewTreePool(hashFunc, swarm.Branches, bmtlegacy.PoolSize)
		handler := NewHandler(accountInfo, client, bmtPool, -1, 0, logger)
		//defer handler.Close()

		if handler == nil {
			t.Fatal("handler is nil")
		}
	})

	t.Run("new-handler-nil pool", func(t *testing.T) {
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}

		accountInfo := acc.GetUserAccountInfo()
		bmtPool := bmtlegacy.NewTreePool(hashFunc, swarm.Branches, bmtlegacy.PoolSize)
		handler := NewHandler(accountInfo, client, bmtPool, -1, 0, logger)
		//defer handler.Close()

		if handler.pool != nil {
			t.Fatal("poll is nol nil")
		}
	})
}
