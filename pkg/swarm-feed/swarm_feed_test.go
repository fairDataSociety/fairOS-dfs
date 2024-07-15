package swarm_feed

import (
	"crypto/ecdsa"
	"fmt"
	"io"
	"testing"

	"github.com/ethersphere/bee/v2/pkg/swarm"

	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/sirupsen/logrus"

	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"

	eCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethersphere/bee/v2/pkg/crypto"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

func TestFeed(t *testing.T) {
	t.Skip()
	logger := logging.New(io.Discard, logrus.DebugLevel)
	mockClient := bee.NewBeeClient("http://localhost:1633", "1ba87c174b66150dacde56df0b914661cff548afcd96957d46f8201694f4a983", true, 0, logger)
	bzzAddr := "ea615d35603a3606426f97822b03020761b2e496568424c0b309103a7f66fb9f"
	bzzRef, err := swarm.ParseHexAddress(bzzAddr)
	pk, err := eCrypto.HexToECDSA("31b0713ac15ac3082180963d288975131a1f63658400fa88f00cf00adedf2609")
	if err != nil {
		t.Fatal(err)

	}

	// get Address from private key
	publicKey := pk.Public().(*ecdsa.PublicKey)
	addr, err := crypto.NewEthereumAddress(*publicKey)
	if err != nil {
		t.Fatal(err)
	}
	topic := "bzzUpdate js 105"
	signer := crypto.NewDefaultSigner(pk)

	f := NewFeed(mockClient)

	manifest, err := f.Upload(utils.Encode(addr), topic, signer, bzzRef)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("manifest: ", manifest)
}
