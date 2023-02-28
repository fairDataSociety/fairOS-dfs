package rpc

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	swarmMail "github.com/fairdatasociety/fairOS-dfs/pkg/contracts/smail"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	additionalConfirmations           = 1
	transactionReceiptTimeout         = time.Minute * 2
	transactionReceiptPollingInterval = time.Second * 10
)

type SubscriptionInfoPutter interface {
	UploadBlob(data []byte, tag uint32, pin bool, encrypt bool) (address []byte, err error)
}

type SubscriptionInfoGetter interface {
	DownloadBlob(address []byte) (data []byte, respCode int, err error)
}

type SubscriptionItemInfo struct {
	Category          string `json:"category"`
	Description       string `json:"description"`
	FdpSellerNameHash string `json:"fdpSellerNameHash"`
	ImageURL          string `json:"imageUrl"`
	PodAddress        string `json:"podAddress"`
	PodName           string `json:"podName"`
	Price             string `json:"price"`
	Title             string `json:"title"`
}

type Client struct {
	c         *ethclient.Client
	putter    SubscriptionInfoPutter
	getter    SubscriptionInfoGetter
	swarmMail *swarmMail.SwarmMail

	logger logging.Logger
}

// ShareInfo
type ShareInfo struct {
	PodName     string `json:"podName"`
	Address     string `json:"podAddress"`
	Password    string `json:"password"`
	UserAddress string `json:"userAddress"`
}

func (c *Client) AddPodToMarketplace(podAddress, owner common.Address, pod, title, desc, thumbnail string, price uint64, category, nameHash [32]byte, key *ecdsa.PrivateKey) error {
	info := &SubscriptionItemInfo{
		Category:          utils.Encode(category[:]),
		Description:       desc,
		FdpSellerNameHash: utils.Encode(nameHash[:]),
		ImageURL:          thumbnail,
		PodAddress:        podAddress.Hex(),
		PodName:           pod,
		Price:             fmt.Sprintf("%d", price),
		Title:             title,
	}
	opts, err := c.newTransactor(key, owner)
	if err != nil {
		return err
	}

	data, err := json.Marshal(info)
	if err != nil { // skipcq: TCV-001
		return err
	}
	ref, err := c.putter.UploadBlob(data, 0, true, false)
	if err != nil { // skipcq: TCV-001
		return err
	}
	var a [32]byte
	copy(a[:], ref)

	i := new(big.Int).SetUint64(price)

	tx, err := c.swarmMail.ListSub(opts, nameHash, a, i, category, podAddress)
	if err != nil {
		return err
	}
	err = c.checkReceipt(tx)
	if err != nil {
		c.logger.Error("ListSub failed : ", err)
		return err
	}
	c.logger.Info("ListSub with hash : ", tx.Hash().Hex())

	return nil
}

func (c *Client) HidePodFromMarketplace(owner common.Address, subHash [32]byte, hide bool, key *ecdsa.PrivateKey) error {
	opts, err := c.newTransactor(key, owner)
	if err != nil {
		return err
	}

	tx, err := c.swarmMail.EnableSub(opts, subHash, !hide)
	if err != nil {
		return err
	}
	err = c.checkReceipt(tx)
	if err != nil {
		c.logger.Error("EnableSub failed : ", err)
		return err
	}
	c.logger.Info("EnableSub with hash : ", tx.Hash().Hex())
	return nil
}

func (c *Client) RequestAccess(subscriber common.Address, subHash, nameHash [32]byte, key *ecdsa.PrivateKey) error {
	opts, err := c.newTransactor(key, subscriber)
	if err != nil {
		return err
	}

	tx, err := c.swarmMail.BidSub(opts, subHash, nameHash)
	if err != nil {
		return err
	}
	err = c.checkReceipt(tx)
	if err != nil {
		c.logger.Error("BidSub failed : ", err)
		return err
	}
	c.logger.Info("BidSub with hash : ", tx.Hash().Hex())
	return nil
}

func (c *Client) AllowAccess(owner common.Address, shareInfo *ShareInfo, requestHash, secret [32]byte, key *ecdsa.PrivateKey) error {
	opts, err := c.newTransactor(key, owner)
	if err != nil {
		return err
	}

	data, err := json.Marshal(shareInfo)
	if err != nil { // skipcq: TCV-001
		return err
	}
	encData, err := utils.EncryptBytes(secret[:], data)
	if err != nil {
		return err
	}

	ref, err := c.putter.UploadBlob(encData, 0, true, false)
	var fixedRef [32]byte
	copy(fixedRef[:], ref)

	tx, err := c.swarmMail.SellSub(opts, requestHash, fixedRef)
	if err != nil {
		return err
	}
	err = c.checkReceipt(tx)
	if err != nil {
		c.logger.Error("SellSub failed : ", err)
		return err
	}
	c.logger.Info("SellSub with hash : ", tx.Hash().Hex())
	return nil
}

func (c *Client) GetSubscription(subscriber common.Address, subHash, secret [32]byte) (*ShareInfo, error) {
	opts := &bind.CallOpts{}
	item, err := c.swarmMail.GetSubItemBy(opts, subscriber, subHash)
	if err != nil {
		return nil, err
	}

	encData, resp, err := c.getter.DownloadBlob(item.UnlockKeyLocation[:])
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	if resp != http.StatusOK { // skipcq: TCV-001
		return nil, fmt.Errorf("ReceivePodInfo: could not download blob")
	}

	data, err := utils.DecryptBytes(secret[:], encData)
	if err != nil {
		return nil, err
	}
	var shareInfo *ShareInfo
	err = json.Unmarshal(data, &shareInfo)
	if err != nil {
		return nil, err
	}

	return shareInfo, nil
}

func (c *Client) GetSubscribablePodInfo(subHash [32]byte) (*SubscriptionItemInfo, error) {
	opts := &bind.CallOpts{}
	item, err := c.swarmMail.GetSubBy(opts, subHash)
	if err != nil {
		return nil, err
	}

	data, respCode, err := c.getter.DownloadBlob(item.SwarmLocation[:])
	if err != nil { // skipcq: TCV-001
		return nil, err
	}
	if respCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get subscribable podInfo")
	}

	info := &SubscriptionItemInfo{}
	err = json.Unmarshal(data, info)
	if err != nil { // skipcq: TCV-001
		return nil, err
	}

	return info, nil
}

func (c *Client) GetSubscriptions(subscriber common.Address, start, limit uint64) ([]swarmMail.SwarmMailSubItem, error) {
	opts := &bind.CallOpts{}
	return c.swarmMail.GetSubItems(opts, subscriber, new(big.Int).SetUint64(start), new(big.Int).SetUint64(limit))
}

func (c *Client) GetAllSubscribablePods() ([]swarmMail.SwarmMailSub, error) {
	opts := &bind.CallOpts{}
	return c.swarmMail.GetSubs(opts)
}

func (c *Client) GetOwnSubscribablePods(owner common.Address) ([]swarmMail.SwarmMailSub, error) {
	opts := &bind.CallOpts{}
	s, err := c.swarmMail.GetSubs(opts)
	if err != nil {
		return nil, err
	}
	osp := []swarmMail.SwarmMailSub{}
	for _, p := range s {
		if p.Seller == owner {
			osp = append(osp, p)
		}
	}
	return osp, nil
}

func (c *Client) GetSubRequests(owner common.Address) ([]swarmMail.SwarmMailSubRequest, error) {
	opts := &bind.CallOpts{}
	return c.swarmMail.GetSubRequests(opts, owner)
}

func (c *Client) GetSub(subHash [32]byte) (*swarmMail.SwarmMailSub, error) {
	opts := &bind.CallOpts{}
	sub, err := c.swarmMail.GetSubBy(opts, subHash)
	if err != nil {
		return nil, err
	}
	return &sub, err
}

func New(logger logging.Logger, getter SubscriptionInfoGetter, putter SubscriptionInfoPutter) (*Client, error) {
	c, err := ethclient.Dial("http://localhost:9545")
	if err != nil {
		return nil, fmt.Errorf("dial eth ensm: %w", err)
	}
	sMail, err := swarmMail.NewSwarmMail(common.HexToAddress("0x21a59654176f2689d12E828B77a783072CD26680"), c)
	if err != nil {
		return nil, err
	}

	return &Client{
		c:         c,
		getter:    getter,
		putter:    putter,
		logger:    logger,
		swarmMail: sMail,
	}, nil
}

func (c *Client) newTransactor(key *ecdsa.PrivateKey, account common.Address) (*bind.TransactOpts, error) {
	nonce, err := c.c.PendingNonceAt(context.Background(), account)
	if err != nil {
		return nil, err
	}
	gasPrice, err := c.c.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	chainID, err := c.c.ChainID(context.Background())
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
		receipt, err := c.c.TransactionReceipt(ctx, tx.Hash())
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
		bn, err := c.c.BlockNumber(ctx)
		if err != nil {
			return err
		}

		nextBlock := receipt.BlockNumber.Uint64() + 1

		if bn >= nextBlock+additionalConfirmations {
			_, err = c.c.HeaderByNumber(ctx, new(big.Int).SetUint64(nextBlock))
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
