package account

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/tyler-smith/go-bip39"
)

func TestWallet(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "pod")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	password := "letmein"
	logger := logging.New(io.Discard, 0)
	acc := New(logger)
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		t.Fatal(err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		t.Fatal(err)
	}
	enMnemonic, err := acc.encryptMnemonic(mnemonic, password)
	if err != nil {
		t.Fatal(err)
	}

	wallet := newWalletFromMnemonic(enMnemonic)
	_, _, err = wallet.LoadMnemonicAndCreateRootAccount("invalid mnemonic that we are passing to check create account error message")
	if err == nil {
		t.Fatal("invalid mnemonic")
	}
	err = wallet.IsValidMnemonic("invalid mnemonic that we are passing to check create account error message")
	if err == nil {
		t.Fatal("invalid mnemonic")
	}

	_, err = wallet.LoadSeedFromMnemonic("wrongpassword")
	if err == nil {
		t.Fatal("wrong password")
	}

	w := &Wallet{}
	_, err = w.LoadSeedFromMnemonic("pass")
	if err == nil {
		t.Fatal("wrong password")
	}
}
