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
	"testing"
	"time"

	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestAccount_CreateRootAccount(t *testing.T) {

	logger := logging.New(io.Discard, 0)
	acc := New(logger)

	_, _, err := acc.CreateUserAccount("invalid mnemonic that we are passing to check create account error message")
	if err == nil {
		t.Fatal("invalid mnemonic passed")
	}

	_, _, err = acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}

	if acc.wallet == nil || acc.wallet.seed == nil {
		t.Fatal("wallet creation error")
	}

	if acc.userAccount.GetPrivateKey() == nil || acc.userAccount.GetPublicKey() == nil || len(acc.userAccount.address[:]) != utils.AddressLength {
		t.Fatalf("keys not intialised")
	}
}

func TestAuthorise(t *testing.T) {
	logger := logging.New(io.Discard, 0)
	acc := New(logger)
	_, _, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateRandomKeyPair(t *testing.T) {
	pk1, err := CreateRandomKeyPair(time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	pk2, err := CreateRandomKeyPair(time.Now().Unix() + 100)
	if err != nil {
		t.Fatal(err)
	}
	if pk1.Equal(pk2) {
		t.Fatal("keys should be different")
	}
}

func TestLoadUserAccountFromSeed(t *testing.T) {
	logger := logging.New(io.Discard, 0)
	acc := New(logger)
	_, seed, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}

	acc.wallet.seed = seed

	acc2 := New(logger)
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
}

func TestPadUnpadSeed(t *testing.T) {
	password := "letmein"
	logger := logging.New(io.Discard, 0)
	acc := New(logger)
	_, seed, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}

	acc.wallet.seed = seed
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

func TestPadUnpadSeedName(t *testing.T) {
	name := "TestPadUnpadSeedName"
	password := "letmein"
	logger := logging.New(io.Discard, 0)
	acc := New(logger)
	_, seed, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}

	acc.wallet.seed = seed
	r, err := acc.userAccount.PadSeedName(seed, name, password)
	if err != nil {
		t.Fatal(err)
	}

	if len(r) != utils.MaxChunkLength {
		t.Fatal("padded string does not match chunk size")
	}

	seed2, name2, err := acc.userAccount.RemovePadFromSeedName(r, password)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(seed, seed2) {
		t.Fatal("seed and padding removed seed do not match")
	}

	if name != name2 {
		t.Fatal("name do not match")
	}
}

func TestCreatePodAccount(t *testing.T) {
	logger := logging.New(io.Discard, 0)
	acc := New(logger)
	_, seed, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}

	acc.wallet.seed = seed
	pod1AccountInfo, err := acc.CreatePodAccount(1, false)
	if err != nil {
		t.Fatal(err)
	}
	pod2AccountInfo, err := acc.CreatePodAccount(2, false)
	if err != nil {
		t.Fatal(err)
	}

	// check if different pod accounts are generated for different index
	if pod1AccountInfo.address == pod2AccountInfo.address {
		t.Fatal("address should not be same")
	}
}

func TestCreatePodAccountWithSeed(t *testing.T) {
	logger := logging.New(io.Discard, 0)
	acc := New(logger)
	_, seed, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}

	acc.wallet.seed = seed

	acc2 := New(logger)
	err = acc2.LoadUserAccountFromSeed(seed)
	if err != nil {
		t.Fatal(err)
	}
	pod1AccountInfo, err := acc2.CreatePodAccount(1, false)
	if err != nil {
		t.Fatal(err)
	}
	pod2AccountInfo, err := acc2.CreatePodAccount(2, false)
	if err != nil {
		t.Fatal(err)
	}

	// check if different pod accounts are generated for different index
	if pod1AccountInfo.address == pod2AccountInfo.address {
		t.Fatal("address should not be same")
	}
}

func TestGetAddress(t *testing.T) {
	logger := logging.New(io.Discard, 0)
	acc := New(logger)
	m, seed, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}

	acc.wallet.seed = seed
	seed2, err := hdwallet.NewSeedFromMnemonic(m)
	if err != nil {
		t.Fatal(err)
	}
	acc2 := New(logger)
	err = acc2.LoadUserAccountFromSeed(seed2)
	if err != nil {
		t.Fatal(err)
	}
	pod1AccountInfo, err := acc2.CreatePodAccount(1, false)
	if err != nil {
		t.Fatal(err)
	}
	pod2AccountInfo, err := acc2.CreatePodAccount(2, false)
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
}
