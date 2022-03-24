package client

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts"
	ens3 "github.com/fairdatasociety/fairOS-dfs/pkg/contracts/ens"
	goens "github.com/wealdtech/go-ens/v3"
)

func TestInitChain(t *testing.T) {
	c, err := InitChain("https://xdai.fairdatasociety.org")
	if err != nil {
		t.Fatal()
	}

	ens, err := ens3.NewEns(common.HexToAddress(contracts.ENSRegistry), c)
	if err != nil {
		t.Fatal()
	}
	node, err := goens.NameHash(contracts.ProviderDomain)
	if err != nil {
		t.Fatal()
	}
	opts := &bind.CallOpts{}
	addr, err := ens.Owner(opts, node)
	if err != nil {
		t.Fatal()
	}
	fmt.Println(addr.Hex())
}
