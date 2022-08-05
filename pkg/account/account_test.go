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

package account

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

func TestAccount_CreateRootAccount(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "pod")
	if err != nil {
		t.Fatal(err)
	}

	password := "letmein"
	logger := logging.New(io.Discard, 0)
	acc := New(logger)

	_, _, err = acc.CreateUserAccount(password, "invalid mnemonic that we are passing to check create account error message")
	if err == nil {
		t.Fatal("invalid mnemonic passed")
	}

	_, _, err = acc.CreateUserAccount(password, "")
	if err != nil {
		t.Fatal(err)
	}

	if acc.wallet == nil || acc.wallet.encryptedmnemonic == "" {
		t.Fatal("wallet creation error")
	}

	plainMnemonic, err := acc.wallet.decryptMnemonic(password)
	if err != nil {
		t.Fatal(err)
	}

	words := strings.Split(plainMnemonic, " ")
	if len(words) != 12 {
		t.Fatal("mnemonic is not 12 words")
	}

	if acc.userAccount.GetPrivateKey() == nil || acc.userAccount.GetPublicKey() == nil || len(acc.userAccount.address[:]) != utils.AddressLength {
		t.Fatalf("keys not intialised")
	}

	err = os.RemoveAll(tempDir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAuthorise(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "pod")
	if err != nil {
		t.Fatal(err)
	}

	password := "letmein"
	logger := logging.New(io.Discard, 0)
	acc := New(logger)
	_, _, err = acc.CreateUserAccount(password, "")
	if err != nil {
		t.Fatal(err)
	}

	authorised := acc.Authorise("")
	if authorised {
		t.Fatal("authorised with blank password")
	}
	authorised = acc.Authorise("wrong password")
	if authorised {
		t.Fatal("authorised with wrong password")
	}
	authorised = acc.Authorise(password)
	if !authorised {
		t.Fatal("authorisation failed")
	}

	err = os.RemoveAll(tempDir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadAndStoreMnemonic(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "pod")
	if err != nil {
		t.Fatal(err)
	}
	password := "letmein"
	logger := logging.New(io.Discard, 0)
	acc := New(logger)
	_, em, err := acc.CreateUserAccount(password, "")
	if err != nil {
		t.Fatal(err)
	}

	expectedMnemonic, err := acc.wallet.decryptMnemonic(password)
	if err != nil {
		t.Fatal(err)
	}

	acc.wallet.encryptedmnemonic = em

	gotMnemonic, err := acc.wallet.decryptMnemonic(password)
	if err != nil {
		t.Fatal(err)
	}

	if gotMnemonic != expectedMnemonic {
		t.Fatalf("mnemonics does not match. expected %s and got %s", expectedMnemonic, gotMnemonic)
	}

	err = os.RemoveAll(tempDir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateRandomKeyPair(t *testing.T) {
	pk1, err := CreateRandomKeyPair(1)
	if err != nil {
		t.Fatal(err)
	}
	pk2, err := CreateRandomKeyPair(2)
	if err != nil {
		t.Fatal(err)
	}
	if pk1.Equal(pk2) {
		t.Fatal("keys should be different")
	}
}

func TestLoadUserAccount(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "pod")
	if err != nil {
		t.Fatal(err)
	}
	password := "letmein"
	logger := logging.New(io.Discard, 0)
	acc := New(logger)
	_, em, err := acc.CreateUserAccount(password, "")
	if err != nil {
		t.Fatal(err)
	}

	acc.wallet.encryptedmnemonic = em

	acc2 := New(logger)
	err = acc2.LoadUserAccount("", em)
	if err == nil {
		t.Fatal("blank password")
	}
	err = acc2.LoadUserAccount("asdasd", em)
	if err == nil {
		t.Fatal("wrong password password")
	}

	err = acc2.LoadUserAccount(password, em)
	if err != nil {
		t.Fatal(err)
	}
	if acc.userAccount.address != acc2.userAccount.address {
		t.Fatal("address do not match")
	}
	err = os.RemoveAll(tempDir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadUserAccountFromSeed(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "pod")
	if err != nil {
		t.Fatal(err)
	}
	password := "letmein"
	logger := logging.New(io.Discard, 0)
	acc := New(logger)
	m, em, err := acc.CreateUserAccount(password, "")
	if err != nil {
		t.Fatal(err)
	}

	acc.wallet.encryptedmnemonic = em

	acc2 := New(logger)
	seed, err := hdwallet.NewSeedFromMnemonic(m)
	if err != nil {
		t.Fatal(err)
	}
	err = acc2.LoadUserAccountFromSeed([]byte{})
	if err == nil {
		t.Fatal("nil seed provided")
	}

	err = acc2.LoadUserAccountFromSeed(seed)
	if err != nil {
		t.Fatal(err)
	}
	if acc.userAccount.address != acc2.userAccount.address {
		t.Fatal("address do not match")
	}
	err = os.RemoveAll(tempDir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPadUnpadSeed(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "pod")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	password := "letmein"
	logger := logging.New(io.Discard, 0)
	acc := New(logger)
	m, em, err := acc.CreateUserAccount(password, "")
	if err != nil {
		t.Fatal(err)
	}

	acc.wallet.encryptedmnemonic = em
	seed, err := hdwallet.NewSeedFromMnemonic(m)
	if err != nil {
		t.Fatal(err)
	}
	r, err := acc.userAccount.PadSeed(seed, password)
	if err != nil {
		t.Fatal(err)
	}

	if len(r) != utils.MaxChunkLength {
		t.Fatal("padded string does not match chunk size")
	}

	seed2, err := acc.userAccount.RemovePadFromSeed(r, password)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(seed, seed2) {
		t.Fatal("seed and padding removed seed do not match")
	}
}

func TestCreatePodAccount(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "pod")
	if err != nil {
		t.Fatal(err)
	}
	password := "letmein"
	logger := logging.New(io.Discard, 0)
	acc := New(logger)
	_, em, err := acc.CreateUserAccount(password, "")
	if err != nil {
		t.Fatal(err)
	}

	acc.wallet.encryptedmnemonic = em
	pod1AccountInfo, err := acc.CreatePodAccount(1, password, false)
	if err != nil {
		t.Fatal(err)
	}
	pod2AccountInfo, err := acc.CreatePodAccount(2, password, false)
	if err != nil {
		t.Fatal(err)
	}

	// check if different pod accounts are generated for different index
	if pod1AccountInfo.address == pod2AccountInfo.address {
		t.Fatal("address should not be same")
	}
	err = os.RemoveAll(tempDir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreatePodAccountWithSeed(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "pod")
	if err != nil {
		t.Fatal(err)
	}
	password := "letmein"
	logger := logging.New(io.Discard, 0)
	acc := New(logger)
	m, em, err := acc.CreateUserAccount(password, "")
	if err != nil {
		t.Fatal(err)
	}

	acc.wallet.encryptedmnemonic = em
	seed, err := hdwallet.NewSeedFromMnemonic(m)
	if err != nil {
		t.Fatal(err)
	}
	acc2 := New(logger)
	err = acc2.LoadUserAccountFromSeed(seed)
	if err != nil {
		t.Fatal(err)
	}
	pod1AccountInfo, err := acc2.CreatePodAccount(1, password, false)
	if err != nil {
		t.Fatal(err)
	}
	pod2AccountInfo, err := acc2.CreatePodAccount(2, password, false)
	if err != nil {
		t.Fatal(err)
	}

	// check if different pod accounts are generated for different index
	if pod1AccountInfo.address == pod2AccountInfo.address {
		t.Fatal("address should not be same")
	}
	err = os.RemoveAll(tempDir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetAddress(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "pod")
	if err != nil {
		t.Fatal(err)
	}
	password := "letmein"
	logger := logging.New(io.Discard, 0)
	acc := New(logger)
	m, em, err := acc.CreateUserAccount(password, "")
	if err != nil {
		t.Fatal(err)
	}

	acc.wallet.encryptedmnemonic = em
	seed, err := hdwallet.NewSeedFromMnemonic(m)
	if err != nil {
		t.Fatal(err)
	}
	acc2 := New(logger)
	err = acc2.LoadUserAccountFromSeed(seed)
	if err != nil {
		t.Fatal(err)
	}
	pod1AccountInfo, err := acc2.CreatePodAccount(1, "password", false)
	if err != nil {
		t.Fatal(err)
	}
	pod2AccountInfo, err := acc2.CreatePodAccount(2, "password", false)
	if err != nil {
		t.Fatal(err)
	}

	userAddress := acc2.GetAddress(UserAccountIndex)
	if acc2.userAccount.address != userAddress {
		t.Fatal("user address do not match")
	}
	pod1Address := acc2.GetAddress(1)
	if pod1AccountInfo.address != pod1Address {
		t.Fatal("pod1 address do not match")
	}
	pod2Address := acc2.GetAddress(2)
	if pod2AccountInfo.address != pod2Address {
		t.Fatal("pod2 address do not match")
	}
	err = os.RemoveAll(tempDir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadSeedFromMnemonic(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "pod")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	password := "letmein"
	logger := logging.New(io.Discard, 0)
	acc := New(logger)
	m, em, err := acc.CreateUserAccount(password, "")
	if err != nil {
		t.Fatal(err)
	}

	acc.wallet.encryptedmnemonic = em
	seed, err := hdwallet.NewSeedFromMnemonic(m)
	if err != nil {
		t.Fatal(err)
	}
	seed2, err := acc.wallet.LoadSeedFromMnemonic(password)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(seed, seed2) {
		t.Fatal("seeds do not match")
	}
}
