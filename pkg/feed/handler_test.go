package feed

import (
	"io"
	"testing"

	mockpost "github.com/ethersphere/bee/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"
	"github.com/ethersphere/bee/pkg/swarm"
	bmtlegacy "github.com/ethersphere/bmt/legacy"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

func TestHandler(t *testing.T) {
	logger := logging.New(io.Discard, 0)

	storer := mockstorer.New()
	beeUrl := mock.NewTestBeeServer(t, mock.TestServerOptions{
		Storer: storer,
		Post:   mockpost.New(mockpost.WithAcceptAll()),
	})
	client := bee.NewBeeClient(beeUrl, mock.BatchOkStr, true, logger)

	t.Run("new-handler", func(t *testing.T) {
		acc := account.New(logger)
		_, _, err := acc.CreateUserAccount("")
		if err != nil {
			t.Fatal(err)
		}

		accountInfo := acc.GetUserAccountInfo()
		bmtPool := bmtlegacy.NewTreePool(hashFunc, swarm.Branches, bmtlegacy.PoolSize)
		handler := NewHandler(accountInfo, client, bmtPool)
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
		handler := NewHandler(accountInfo, client, bmtPool)
		//defer handler.Close()

		if handler == nil {
			t.Fatal("handler is nil")
		}
	})
}
