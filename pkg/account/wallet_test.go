package account

import (
	"os"
	"testing"

	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"

	"github.com/tyler-smith/go-bip39"
)

func TestWallet(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pod")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		t.Fatal(err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		t.Fatal(err)
	}
	seed, err := hdwallet.NewSeedFromMnemonic(mnemonic)
	if err != nil {
		t.Fatal(err)
	}
	wallet := newWallet(seed)
	_, _, err = wallet.LoadMnemonicAndCreateRootAccount("invalid mnemonic that we are passing to check create account error message")
	if err == nil {
		t.Fatal("invalid mnemonic")
	}
	err = wallet.IsValidMnemonic("invalid mnemonic that we are passing to check create account error message")
	if err == nil {
		t.Fatal("invalid mnemonic")
	}
}
