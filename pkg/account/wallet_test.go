package account

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

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

func TestSignatureToWallet(t *testing.T) {
	signature := "b7f4346174a6ff79983bdb10348523de3a4bd2b4772b9f7217b997c6ca1f6abd3de015eab01818e459fad3c067e00969d9f02b808df027574da2f7fd50170a911c"
	addrs := []string{"0x61E18Ac267f4d5af06D421DeA020818255678649", "0x13543e7BA5ff28AD8B203BB8e93b47D76ee2aE05"}
	w := newWallet(nil)
	acc, _, err := w.GenerateWalletFromSignature(signature, "")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, addrs[0], acc.Address.String())
	acc, _, err = w.GenerateWalletFromSignature(signature, "111111111111")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, addrs[1], acc.Address.String())
}
