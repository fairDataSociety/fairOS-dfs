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
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts"
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts/ens"
	fdsregistrar "github.com/fairdatasociety/fairOS-dfs/pkg/contracts/fds-registrar"
	publicresolver "github.com/fairdatasociety/fairOS-dfs/pkg/contracts/public-resolver"
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

	// ErrWrongChainID denotes the rpc endpoint returned different chainId than the configured one
	ErrWrongChainID = fmt.Errorf("chainID does not match or not supported")
)

// Client is used to manage ENS
type Client struct {
	eth            *ethclient.Client
	ensConfig      *contracts.Config
	ensRegistry    *ens.ENSRegistry
	fdsRegistrar   *fdsregistrar.FDSRegistrar
	publicResolver *publicresolver.PublicResolver

	logger logging.Logger
}

// New returns a new ENS manager Client
func New(ensConfig *contracts.Config, logger logging.Logger) (*Client, error) {
	eth, err := ethclient.Dial(ensConfig.ProviderBackend)
	if err != nil {
		return nil, fmt.Errorf("dial eth ensm: %w", err)
	}

	// check connection
	chainID, err := eth.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("dial eth ensm: %w", err)
	}
	if chainID.String() != ensConfig.ChainID {
		return nil, ErrWrongChainID
	}
	ensRegistry, err := ens.NewENSRegistry(common.HexToAddress(ensConfig.ENSRegistryAddress), eth)
	if err != nil {
		return nil, err
	}
	fdsRegistrar, err := fdsregistrar.NewFDSRegistrar(common.HexToAddress(ensConfig.FDSRegistrarAddress), eth)
	if err != nil {
		return nil, err
	}
	publicResolver, err := publicresolver.NewPublicResolver(common.HexToAddress(ensConfig.PublicResolverAddress), eth)
	if err != nil {
		return nil, err
	}

	logger.Info("ensProviderBackend   : ", ensConfig.ProviderBackend)
	logger.Info("ensProviderDomain    : ", ensConfig.ProviderDomain)
	logger.Info("ENSRegistryAddress    : ", common.HexToAddress(ensConfig.ENSRegistryAddress).String())
	logger.Info("FDSRegistrarAddress    : ", common.HexToAddress(ensConfig.FDSRegistrarAddress).String())
	logger.Info("PublicResolverAddress    : ", common.HexToAddress(ensConfig.PublicResolverAddress).String())
	c := &Client{
		eth:            eth,
		ensConfig:      ensConfig,
		ensRegistry:    ensRegistry,
		fdsRegistrar:   fdsRegistrar,
		publicResolver: publicResolver,
		logger:         logger,
	}
	return c, nil
}

// GetOwner returns the owner of the username
func (c *Client) GetOwner(username string) (common.Address, error) {
	node, err := goens.NameHash(username + "." + c.ensConfig.ProviderDomain)
	if err != nil {
		return common.Address{}, err
	}
	opts := &bind.CallOpts{}
	return c.ensRegistry.Owner(opts, node)
}

// RegisterSubdomain registers the username
func (c *Client) RegisterSubdomain(username string, owner common.Address, key *ecdsa.PrivateKey) error {
	balance, err := c.eth.BalanceAt(context.Background(), owner, nil)
	if err != nil {
		return err
	}
	if balance.Cmp(minRequiredBalance) < 0 {
		c.logger.Error("account does not have enough balance")
		return ErrInsufficientBalance
	}
	opts, err := c.newTransactor(key, owner)
	if err != nil {
		return err
	}

	h := sha3.NewLegacyKeccak256()
	_, err = h.Write([]byte(username))
	if err != nil {
		return err
	}
	hash := h.Sum(nil)
	label := big.NewInt(0).SetBytes(hash)
	exp := big.NewInt(0).SetBytes([]byte("86400"))
	tx, err := c.fdsRegistrar.Register(opts, label, owner, exp)
	if err != nil {
		c.logger.Error("fds register failed : ", err)
		return err
	}
	err = c.checkReceipt(tx)
	if err != nil {
		c.logger.Error("fds register failed : ", err)
		return err
	}
	c.logger.Info("fds registered with hash : ", tx.Hash().Hex())
	return nil
}

// SetResolver sets the resolver for the username
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
		c.logger.Error("ensRegistry SetResolver failed : ", err)
		return "", err
	}
	err = c.checkReceipt(tx)
	if err != nil {
		c.logger.Error("ensRegistry SetResolver failed : ", err)
		return "", err
	}
	c.logger.Info("set resolver called with hash : ", tx.Hash().Hex())
	nameHash := node[:]
	return utils.Encode(nameHash), nil
}

// SetAll sets all the necessary information of the user
func (c *Client) SetAll(username string, owner common.Address, key *ecdsa.PrivateKey) error {
	node, err := goens.NameHash(username + "." + c.ensConfig.ProviderDomain)
	if err != nil {
		return err
	}

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
	tx, err := c.publicResolver.SetPubkey(opts, node, x, y)
	if err != nil {
		c.logger.Error("public resolver set all failed : ", err)
		return err
	}
	err = c.checkReceipt(tx)
	if err != nil {
		c.logger.Error("public resolver set all failed : ", err)
		return err
	}
	c.logger.Info("public resolver setall called with hash : ", tx.Hash().Hex())
	return nil
}

// GetInfo returns the public key of the user
func (c *Client) GetInfo(username string) (*ecdsa.PublicKey, string, error) {
	node, err := goens.NameHash(username + "." + c.ensConfig.ProviderDomain)
	if err != nil {
		return nil, "", err
	}

	opts := &bind.CallOpts{}
	info, err := c.publicResolver.GetAll(opts, node)
	if err != nil {
		c.logger.Error("public resolver get all failed : ", err)
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
	opts.GasLimit = uint64(1000000)
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
		if receipt.Status == types.ReceiptStatusFailed {
			return fmt.Errorf("transaction %s failed", tx.Hash().Hex())
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
