package test_test

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/plexsysio/taskmanager"
)

func TestLite(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(io.Discard, 0)
	tm := taskmanager.New(1, 10, time.Second*15, logger)

	t.Run("new-blank-username", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()

		// create user
		userObject := user.NewUsers("", mockClient, ens, logger)
		_, _, err := userObject.LoadLiteUser("", "password1", "", "", tm)
		if !errors.Is(err, user.ErrInvalidUserName) {
			t.Fatal(err)
		}
	})

	t.Run("new-user", func(t *testing.T) {
		ens := mock2.NewMockNamespaceManager()

		// create user
		userObject := user.NewUsers("", mockClient, ens, logger)
		mnemonic, ui, err := userObject.LoadLiteUser("user1", "password1", "", "", tm)
		if err != nil {
			t.Fatal(err)
		}

		// validate user
		if !userObject.IsUserNameLoggedIn(ui.GetUserName()) {
			t.Fatalf("user not loggin in")
		}
		if ui == nil {
			t.Fatalf("invalid user info")
		}
		if ui.GetUserName() != "user1" {
			t.Fatalf("invalid user name")
		}
		if ui.GetFeed() == nil || ui.GetAccount() == nil || ui.GetPod() == nil {
			t.Fatalf("invalid feed, account or pod")
		}
		err = ui.GetAccount().GetWallet().IsValidMnemonic(mnemonic)
		if err != nil {
			t.Fatalf("invalid mnemonic")
		}
	})

	t.Run("stat-nonexistent-user", func(t *testing.T) {

	})

	t.Run("stat-user", func(t *testing.T) {

	})
}
