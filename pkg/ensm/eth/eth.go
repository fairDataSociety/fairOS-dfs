package eth

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts"
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts/ens"
	publicresolver "github.com/fairdatasociety/fairOS-dfs/pkg/contracts/public-resolver"
	subdomainregistrar "github.com/fairdatasociety/fairOS-dfs/pkg/contracts/subdomain-registrar"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	goens "github.com/wealdtech/go-ens/v3"
	"golang.org/x/crypto/sha3"
)

const (
	additionalConfirmations           = 1
	transactionReceiptTimeout         = time.Minute * 2
	transactionReceiptPollingInterval = time.Second * 10
)

var (
	minRequiredBalance = big.NewInt(10000000000000000) // 0.01 eth
)

type Client struct {
	eth                *ethclient.Client
	ensConfig          *contracts.Config
	ensRegistry        *ens.Ens
	subdomainRegistrar *subdomainregistrar.Subdomainregistrar
	publicResolver     *publicresolver.Publicresolver

	providerPrivateKey *ecdsa.PrivateKey
	providerAddress    common.Address
	logger             logging.Logger
}

func New(ensConfig *contracts.Config, logger logging.Logger) (*Client, error) {
	eth, err := ethclient.Dial(ensConfig.ProviderBackend)
	if err != nil {
		return nil, fmt.Errorf("dial eth ensm: %w", err)
	}

	// check connection
	_, err = eth.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("dial eth ensm: %w", err)
	}
	ensRegistry, err := ens.NewEns(common.HexToAddress(ensConfig.ENSRegistryAddress), eth)
	if err != nil {
		return nil, err
	}
	subdomainRegistrar, err := subdomainregistrar.NewSubdomainregistrar(common.HexToAddress(ensConfig.SubdomainRegistrarAddress), eth)
	if err != nil {
		return nil, err
	}
	publicResolver, err := publicresolver.NewPublicresolver(common.HexToAddress(ensConfig.PublicResolverAddress), eth)
	if err != nil {
		return nil, err
	}
	privateKey, err := crypto.HexToECDSA(ensConfig.ProviderPrivateKey)
	if err != nil {
		return nil, err
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	logger.Info("ensProviderBackend   : ", ensConfig.ProviderBackend)
	logger.Info("ensProviderAddress   : ", fromAddress.Hex())
	logger.Info("ensProviderDomain    : ", ensConfig.ProviderDomain)
	c := &Client{
		eth:                eth,
		ensConfig:          ensConfig,
		ensRegistry:        ensRegistry,
		subdomainRegistrar: subdomainRegistrar,
		publicResolver:     publicResolver,
		providerAddress:    fromAddress,
		providerPrivateKey: privateKey,
		logger:             logger,
	}
	return c, nil
}

func (c *Client) GetOwner(username string) (common.Address, error) {
	node, err := goens.NameHash(username + "." + c.ensConfig.ProviderDomain)
	if err != nil {
		return common.Address{}, err
	}
	opts := &bind.CallOpts{}
	return c.ensRegistry.Owner(opts, node)
}

func (c *Client) RegisterSubdomain(username string, owner common.Address) error {
	balance, err := c.eth.BalanceAt(context.Background(), owner, nil)
	if err != nil {
		return err
	}
	if balance.Cmp(minRequiredBalance) < 0 {
		c.logger.Error("account does not have enough balance")
		return ErrInsufficientBalance
	}

	opts, err := c.newTransactor(c.providerPrivateKey, c.providerAddress)
	if err != nil {
		return err
	}

	h := sha3.NewLegacyKeccak256()
	_, err = h.Write([]byte(username))
	if err != nil {
		return err
	}
	hash := h.Sum(nil)
	label := [32]byte{}
	copy(label[:], hash)

	tx, err := c.subdomainRegistrar.Register(opts, label, owner)
	if err != nil {
		c.logger.Error("subdomain register failed :", err)
		return err
	}
	err = c.checkReceipt(tx)
	if err != nil {
		c.logger.Error("subdomain register failed :", err)
		return err
	}
	c.logger.Info("subdomain registered with hash :", tx.Hash().Hex())
	return nil
}

func (c *Client) SetResolver(username string, owner common.Address, key *ecdsa.PrivateKey) (string, error) {
	node, err := goens.NameHash(username + "." + c.ensConfig.ProviderDomain)
	if err != nil {
		return "", err
	}
	opts, err := c.newTransactor(key, owner)
	if err != nil {
		return "", err
	}
	tx, err := c.ensRegistry.SetResolver(opts, node, common.HexToAddress(c.ensConfig.PublicResolverAddress))
	if err != nil {
		c.logger.Error("ensRegistry SetResolver failed :", err)
		return "", err
	}
	err = c.checkReceipt(tx)
	if err != nil {
		c.logger.Error("ensRegistry SetResolver failed :", err)
		return "", err
	}
	c.logger.Info("set resolver called with hash :", tx.Hash().Hex())
	nameHash := node[:]
	return utils.Encode(nameHash), nil
}

func (c *Client) SetAll(username string, owner common.Address, key *ecdsa.PrivateKey) error {
	node, err := goens.NameHash(username + "." + c.ensConfig.ProviderDomain)
	if err != nil {
		return err
	}

	name := "subdomain-hidden"
	contentStr := "0x0000000000000000000000000000000000000000000000000000000000000000"
	content := [32]byte{}
	copy(content[:], contentStr)
	publicKey := key.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}
	x := [32]byte{}
	copy(x[:], publicKeyECDSA.X.Bytes())
	y := [32]byte{}
	copy(y[:], publicKeyECDSA.Y.Bytes())
	opts, err := c.newTransactor(key, owner)
	if err != nil {
		return err
	}
	tx, err := c.publicResolver.SetAll(opts, node, owner, content, []byte{}, x, y, name)
	if err != nil {
		c.logger.Error("public resolver set all failed :", err)
		return err
	}
	err = c.checkReceipt(tx)
	if err != nil {
		c.logger.Error("public resolver set all failed :", err)
		return err
	}
	c.logger.Info("public resolver setall called with hash :", tx.Hash().Hex())
	return nil
}

func (c *Client) GetInfo(username string) (*ecdsa.PublicKey, string, error) {
	node, err := goens.NameHash(username + "." + c.ensConfig.ProviderDomain)
	if err != nil {
		return nil, "", err
	}

	opts := &bind.CallOpts{}
	info, err := c.publicResolver.GetAll(opts, node)
	if err != nil {
		c.logger.Error("public resolver get all failed :", err)
		return nil, "", err
	}
	x := new(big.Int)
	x.SetBytes(info.X[:])

	y := new(big.Int)
	y.SetBytes(info.Y[:])
	pub := new(ecdsa.PublicKey)
	pub.X = x
	pub.Y = y

	pub.Curve = btcec.S256()
	nameHash := node[:]
	return pub, utils.Encode(nameHash), nil
}

func (c *Client) newTransactor(key *ecdsa.PrivateKey, account common.Address) (*bind.TransactOpts, error) {
	nonce, err := c.eth.PendingNonceAt(context.Background(), account)
	if err != nil {
		return nil, err
	}
	gasPrice, err := c.eth.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	chainID, err := c.eth.ChainID(context.Background())
	if err != nil {
		return nil, err
	}
	opts, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		return nil, err
	}
	opts.Nonce = big.NewInt(int64(nonce))
	opts.Value = big.NewInt(0)
	opts.GasLimit = uint64(300000)
	opts.GasPrice = gasPrice
	opts.From = account
	return opts, nil
}

func (c *Client) checkReceipt(tx *types.Transaction) error {
	ctx, cancel := context.WithTimeout(context.Background(), transactionReceiptTimeout)
	defer cancel()

	pollingInterval := transactionReceiptPollingInterval
	for {
		receipt, err := c.eth.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			if !errors.Is(err, ethereum.NotFound) {
				return err
			}
			select {
			case <-time.After(pollingInterval):
			case <-ctx.Done():
				return ctx.Err()
			}
			continue
		}
		bn, err := c.eth.BlockNumber(ctx)
		if err != nil {
			return err
		}

		nextBlock := receipt.BlockNumber.Uint64() + 1

		if bn >= nextBlock+additionalConfirmations {
			_, err = c.eth.HeaderByNumber(ctx, new(big.Int).SetUint64(nextBlock))
			if err != nil {
				if !errors.Is(err, ethereum.NotFound) {
					return err
				}
			} else {
				return nil
			}
		}

		select {
		case <-time.After(pollingInterval):
		case <-ctx.Done():
			return errors.New("context timeout")
		}
	}
}
