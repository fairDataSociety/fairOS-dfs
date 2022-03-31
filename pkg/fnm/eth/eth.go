package eth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts"
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts/ens"
	publicresolver "github.com/fairdatasociety/fairOS-dfs/pkg/contracts/public-resolver"
	subdomainregistrar "github.com/fairdatasociety/fairOS-dfs/pkg/contracts/subdomain-registrar"
	goens "github.com/wealdtech/go-ens/v3"
	"golang.org/x/crypto/sha3"
)

type Client struct {
	eth                *ethclient.Client
	ensRegistry        *ens.Ens
	subdomainRegistrar *subdomainregistrar.Subdomainregistrar
	publicResolver     *publicresolver.Publicresolver
}

func New(endpoint string) (*Client, error) {
	rpcClient, err := rpc.DialContext(context.Background(), endpoint)
	if err != nil {
		return nil, fmt.Errorf("dial eth fnm: %w", err)
	}
	eth := ethclient.NewClient(rpcClient)
	ensRegistry, err := ens.NewEns(common.HexToAddress(contracts.ENSRegistryAddress), eth)
	if err != nil {
		return nil, err
	}
	subdomainRegistrar, err := subdomainregistrar.NewSubdomainregistrar(common.HexToAddress(contracts.SubdomainRegistrarAddress), eth)
	if err != nil {
		return nil, err
	}
	publicResolver, err := publicresolver.NewPublicresolver(common.HexToAddress(contracts.PublicResolverAddress), eth)
	if err != nil {
		return nil, err
	}
	c := &Client{
		eth:                ethclient.NewClient(rpcClient),
		ensRegistry:        ensRegistry,
		subdomainRegistrar: subdomainRegistrar,
		publicResolver:     publicResolver,
	}
	return c, nil
}

func (c *Client) GetOwner(username string) (common.Address, error) {
	node, err := goens.NameHash(username + "." + contracts.ProviderDomain)
	if err != nil {
		return common.Address{}, err
	}
	opts := &bind.CallOpts{}
	return c.ensRegistry.Owner(opts, node)
}

func (c *Client) RegisterSubdomain(username string, owner common.Address) error {
	privateKey, err := crypto.HexToECDSA("8de4ecbaa1804ac86b1cd10b61848133446a20d0b5fd5ab6b2c9bd9d11c34bfa")
	if err != nil {
		return err
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := c.eth.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	gasPrice, err := c.eth.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	opts := bind.NewKeyedTransactor(privateKey)
	opts.Nonce = big.NewInt(int64(nonce))
	opts.Value = big.NewInt(0)     // in wei
	opts.GasLimit = uint64(300000) // in units
	opts.GasPrice = gasPrice
	opts.From = fromAddress

	h := sha3.NewLegacyKeccak256()
	h.Write([]byte(username))
	hash := h.Sum(nil)
	label := [32]byte{}
	copy(label[:], hash)
	tx, err := c.subdomainRegistrar.Register(opts, label, owner)
	if err != nil {
		fmt.Println("RegisterSubdomain err", err)
		return err
	}
	fmt.Println("RegisterSubdomain", tx.Hash().Hex())
	return nil
}

func (c *Client) SetResolver(username string) error {
	node, err := goens.NameHash(username + "." + contracts.ProviderDomain)
	if err != nil {
		return err
	}
	privateKey, err := crypto.HexToECDSA("8de4ecbaa1804ac86b1cd10b61848133446a20d0b5fd5ab6b2c9bd9d11c34bfa")
	if err != nil {
		return err
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := c.eth.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	gasPrice, err := c.eth.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	opts := bind.NewKeyedTransactor(privateKey)
	opts.Nonce = big.NewInt(int64(nonce))
	opts.Value = big.NewInt(0)     // in wei
	opts.GasLimit = uint64(300000) // in units
	opts.GasPrice = gasPrice
	opts.From = fromAddress
	tx, err := c.ensRegistry.SetResolver(opts, node, common.HexToAddress(contracts.PublicResolverAddress))
	if err != nil {
		fmt.Println("SetResolver err", err)
		return err
	}
	fmt.Println("SetResolver", tx.Hash().Hex())
	return nil
}

func (c *Client) SetAll(username string, owner common.Address, key *ecdsa.PublicKey) error {
	node, err := goens.NameHash(username + "." + contracts.ProviderDomain)
	if err != nil {
		return err
	}

	contentStr := "0x0000000000000000000000000000000000000000000000000000000000000000"
	content := [32]byte{}
	copy(content[:], contentStr)

	x := [32]byte{}
	copy(x[:], key.X.Bytes())
	y := [32]byte{}
	copy(y[:], key.Y.Bytes())

	name := "subdomain-hidden"
	privateKey, err := crypto.HexToECDSA("8de4ecbaa1804ac86b1cd10b61848133446a20d0b5fd5ab6b2c9bd9d11c34bfa")
	if err != nil {
		return err
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := c.eth.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	gasPrice, err := c.eth.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	opts := bind.NewKeyedTransactor(privateKey)
	opts.Nonce = big.NewInt(int64(nonce))
	opts.Value = big.NewInt(0)     // in wei
	opts.GasLimit = uint64(300000) // in units
	opts.GasPrice = gasPrice
	opts.From = fromAddress
	tx, err := c.publicResolver.SetAll(opts, node, owner, content, []byte{}, x, y, name)
	if err != nil {
		fmt.Println("SetAll err", err)
		return err
	}
	fmt.Println("SetAll", tx.Hash().Hex())
	return err
}
