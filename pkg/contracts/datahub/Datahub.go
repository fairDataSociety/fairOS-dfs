// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package datahub

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// DataHubActiveBid is an auto generated low-level Go binding around an user-defined struct.
type DataHubActiveBid struct {
	Seller      common.Address
	RequestHash [32]byte
}

// DataHubCategory is an auto generated low-level Go binding around an user-defined struct.
type DataHubCategory struct {
	SubIdxs []uint64
}

// DataHubSub is an auto generated low-level Go binding around an user-defined struct.
type DataHubSub struct {
	SubHash           [32]byte
	FdpSellerNameHash [32]byte
	Seller            common.Address
	SwarmLocation     [32]byte
	Price             *big.Int
	Active            bool
	Earned            *big.Int
	Bids              uint32
	Sells             uint32
	Reports           uint32
	DaysValid         uint16
}

// DataHubSubItem is an auto generated low-level Go binding around an user-defined struct.
type DataHubSubItem struct {
	SubHash           [32]byte
	UnlockKeyLocation [32]byte
	ValidTill         *big.Int
}

// DataHubSubRequest is an auto generated low-level Go binding around an user-defined struct.
type DataHubSubRequest struct {
	FdpBuyerNameHash [32]byte
	Buyer            common.Address
	SubHash          [32]byte
	RequestHash      [32]byte
}

// DatahubMetaData contains all meta data concerning the Datahub contract.
var DatahubMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"DEFAULT_ADMIN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ROLE_REPORTER\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"fdpBuyerNameHash\",\"type\":\"bytes32\"}],\"name\":\"bidSub\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"}],\"name\":\"enableSub\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"feesCollected\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"fundsBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"fundsTransfer\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getActiveBidAt\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"seller\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"requestHash\",\"type\":\"bytes32\"}],\"internalType\":\"structDataHub.ActiveBid\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getActiveBids\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"seller\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"requestHash\",\"type\":\"bytes32\"}],\"internalType\":\"structDataHub.ActiveBid[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getAllSubItems\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"unlockKeyLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"validTill\",\"type\":\"uint256\"}],\"internalType\":\"structDataHub.SubItem[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"category\",\"type\":\"bytes32\"}],\"name\":\"getCategory\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64[]\",\"name\":\"subIdxs\",\"type\":\"uint64[]\"}],\"internalType\":\"structDataHub.Category\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_fee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"getFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getListedSubs\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getPortableAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"}],\"name\":\"getSubBy\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"fdpSellerNameHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"seller\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"swarmLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"earned\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"bids\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"sells\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"reports\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"daysValid\",\"type\":\"uint16\"}],\"internalType\":\"structDataHub.Sub\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getSubByIndex\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"fdpSellerNameHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"seller\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"swarmLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"earned\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"bids\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"sells\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"reports\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"daysValid\",\"type\":\"uint16\"}],\"internalType\":\"structDataHub.Sub\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"forAddress\",\"type\":\"address\"}],\"name\":\"getSubInfoBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getSubItemAt\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"unlockKeyLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"validTill\",\"type\":\"uint256\"}],\"internalType\":\"structDataHub.SubItem\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"}],\"name\":\"getSubItemBy\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"unlockKeyLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"validTill\",\"type\":\"uint256\"}],\"internalType\":\"structDataHub.SubItem\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"start\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"}],\"name\":\"getSubItems\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"unlockKeyLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"validTill\",\"type\":\"uint256\"}],\"internalType\":\"structDataHub.SubItem[]\",\"name\":\"items\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"last\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getSubRequestAt\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"fdpBuyerNameHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"buyer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"requestHash\",\"type\":\"bytes32\"}],\"internalType\":\"structDataHub.SubRequest\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"requestHash\",\"type\":\"bytes32\"}],\"name\":\"getSubRequestByHash\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"fdpBuyerNameHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"buyer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"requestHash\",\"type\":\"bytes32\"}],\"internalType\":\"structDataHub.SubRequest\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getSubRequests\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"fdpBuyerNameHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"buyer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"requestHash\",\"type\":\"bytes32\"}],\"internalType\":\"structDataHub.SubRequest[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"}],\"name\":\"getSubSubscribers\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getSubs\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"fdpSellerNameHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"seller\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"swarmLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"earned\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"bids\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"sells\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"reports\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"daysValid\",\"type\":\"uint16\"}],\"internalType\":\"structDataHub.Sub[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getUserStats\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numSubRequests\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"numSubItems\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"numActiveBids\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"numListedSubs\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"inEscrow\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"fdpSellerNameHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"dataSwarmLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"category\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"podAddress\",\"type\":\"address\"},{\"internalType\":\"uint16\",\"name\":\"daysValid\",\"type\":\"uint16\"}],\"name\":\"listSub\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"marketFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minListingFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"release\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"requestHash\",\"type\":\"bytes32\"}],\"name\":\"removeUserActiveBid\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"}],\"name\":\"reportSub\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"requestHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"encryptedKeyLocation\",\"type\":\"bytes32\"}],\"name\":\"sellSub\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newFee\",\"type\":\"uint256\"}],\"name\":\"setFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newListingFee\",\"type\":\"uint256\"}],\"name\":\"setListingFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"setPortableAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"subscriptionIds\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"subscriptions\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"fdpSellerNameHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"seller\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"swarmLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"earned\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"bids\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"sells\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"reports\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"daysValid\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// DatahubABI is the input ABI used to generate the binding from.
// Deprecated: Use DatahubMetaData.ABI instead.
var DatahubABI = DatahubMetaData.ABI

// Datahub is an auto generated Go binding around an Ethereum contract.
type Datahub struct {
	DatahubCaller     // Read-only binding to the contract
	DatahubTransactor // Write-only binding to the contract
	DatahubFilterer   // Log filterer for contract events
}

// DatahubCaller is an auto generated read-only Go binding around an Ethereum contract.
type DatahubCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DatahubTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DatahubTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DatahubFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DatahubFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DatahubSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DatahubSession struct {
	Contract     *Datahub          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DatahubCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DatahubCallerSession struct {
	Contract *DatahubCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// DatahubTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DatahubTransactorSession struct {
	Contract     *DatahubTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// DatahubRaw is an auto generated low-level Go binding around an Ethereum contract.
type DatahubRaw struct {
	Contract *Datahub // Generic contract binding to access the raw methods on
}

// DatahubCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DatahubCallerRaw struct {
	Contract *DatahubCaller // Generic read-only contract binding to access the raw methods on
}

// DatahubTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DatahubTransactorRaw struct {
	Contract *DatahubTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDatahub creates a new instance of Datahub, bound to a specific deployed contract.
func NewDatahub(address common.Address, backend bind.ContractBackend) (*Datahub, error) {
	contract, err := bindDatahub(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Datahub{DatahubCaller: DatahubCaller{contract: contract}, DatahubTransactor: DatahubTransactor{contract: contract}, DatahubFilterer: DatahubFilterer{contract: contract}}, nil
}

// NewDatahubCaller creates a new read-only instance of Datahub, bound to a specific deployed contract.
func NewDatahubCaller(address common.Address, caller bind.ContractCaller) (*DatahubCaller, error) {
	contract, err := bindDatahub(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DatahubCaller{contract: contract}, nil
}

// NewDatahubTransactor creates a new write-only instance of Datahub, bound to a specific deployed contract.
func NewDatahubTransactor(address common.Address, transactor bind.ContractTransactor) (*DatahubTransactor, error) {
	contract, err := bindDatahub(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DatahubTransactor{contract: contract}, nil
}

// NewDatahubFilterer creates a new log filterer instance of Datahub, bound to a specific deployed contract.
func NewDatahubFilterer(address common.Address, filterer bind.ContractFilterer) (*DatahubFilterer, error) {
	contract, err := bindDatahub(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DatahubFilterer{contract: contract}, nil
}

// bindDatahub binds a generic wrapper to an already deployed contract.
func bindDatahub(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(DatahubABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Datahub *DatahubRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Datahub.Contract.DatahubCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Datahub *DatahubRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Datahub.Contract.DatahubTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Datahub *DatahubRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Datahub.Contract.DatahubTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Datahub *DatahubCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Datahub.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Datahub *DatahubTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Datahub.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Datahub *DatahubTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Datahub.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Datahub *DatahubCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Datahub *DatahubSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _Datahub.Contract.DEFAULTADMINROLE(&_Datahub.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Datahub *DatahubCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _Datahub.Contract.DEFAULTADMINROLE(&_Datahub.CallOpts)
}

// ROLEREPORTER is a free data retrieval call binding the contract method 0x83102c2b.
//
// Solidity: function ROLE_REPORTER() view returns(bytes32)
func (_Datahub *DatahubCaller) ROLEREPORTER(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "ROLE_REPORTER")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ROLEREPORTER is a free data retrieval call binding the contract method 0x83102c2b.
//
// Solidity: function ROLE_REPORTER() view returns(bytes32)
func (_Datahub *DatahubSession) ROLEREPORTER() ([32]byte, error) {
	return _Datahub.Contract.ROLEREPORTER(&_Datahub.CallOpts)
}

// ROLEREPORTER is a free data retrieval call binding the contract method 0x83102c2b.
//
// Solidity: function ROLE_REPORTER() view returns(bytes32)
func (_Datahub *DatahubCallerSession) ROLEREPORTER() ([32]byte, error) {
	return _Datahub.Contract.ROLEREPORTER(&_Datahub.CallOpts)
}

// FeesCollected is a free data retrieval call binding the contract method 0xf071db5a.
//
// Solidity: function feesCollected() view returns(uint256)
func (_Datahub *DatahubCaller) FeesCollected(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "feesCollected")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FeesCollected is a free data retrieval call binding the contract method 0xf071db5a.
//
// Solidity: function feesCollected() view returns(uint256)
func (_Datahub *DatahubSession) FeesCollected() (*big.Int, error) {
	return _Datahub.Contract.FeesCollected(&_Datahub.CallOpts)
}

// FeesCollected is a free data retrieval call binding the contract method 0xf071db5a.
//
// Solidity: function feesCollected() view returns(uint256)
func (_Datahub *DatahubCallerSession) FeesCollected() (*big.Int, error) {
	return _Datahub.Contract.FeesCollected(&_Datahub.CallOpts)
}

// FundsBalance is a free data retrieval call binding the contract method 0x9454932c.
//
// Solidity: function fundsBalance() view returns(uint256)
func (_Datahub *DatahubCaller) FundsBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "fundsBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FundsBalance is a free data retrieval call binding the contract method 0x9454932c.
//
// Solidity: function fundsBalance() view returns(uint256)
func (_Datahub *DatahubSession) FundsBalance() (*big.Int, error) {
	return _Datahub.Contract.FundsBalance(&_Datahub.CallOpts)
}

// FundsBalance is a free data retrieval call binding the contract method 0x9454932c.
//
// Solidity: function fundsBalance() view returns(uint256)
func (_Datahub *DatahubCallerSession) FundsBalance() (*big.Int, error) {
	return _Datahub.Contract.FundsBalance(&_Datahub.CallOpts)
}

// GetActiveBidAt is a free data retrieval call binding the contract method 0x78ba33c6.
//
// Solidity: function getActiveBidAt(address addr, uint256 index) view returns((address,bytes32))
func (_Datahub *DatahubCaller) GetActiveBidAt(opts *bind.CallOpts, addr common.Address, index *big.Int) (DataHubActiveBid, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getActiveBidAt", addr, index)

	if err != nil {
		return *new(DataHubActiveBid), err
	}

	out0 := *abi.ConvertType(out[0], new(DataHubActiveBid)).(*DataHubActiveBid)

	return out0, err

}

// GetActiveBidAt is a free data retrieval call binding the contract method 0x78ba33c6.
//
// Solidity: function getActiveBidAt(address addr, uint256 index) view returns((address,bytes32))
func (_Datahub *DatahubSession) GetActiveBidAt(addr common.Address, index *big.Int) (DataHubActiveBid, error) {
	return _Datahub.Contract.GetActiveBidAt(&_Datahub.CallOpts, addr, index)
}

// GetActiveBidAt is a free data retrieval call binding the contract method 0x78ba33c6.
//
// Solidity: function getActiveBidAt(address addr, uint256 index) view returns((address,bytes32))
func (_Datahub *DatahubCallerSession) GetActiveBidAt(addr common.Address, index *big.Int) (DataHubActiveBid, error) {
	return _Datahub.Contract.GetActiveBidAt(&_Datahub.CallOpts, addr, index)
}

// GetActiveBids is a free data retrieval call binding the contract method 0xfbc4fc44.
//
// Solidity: function getActiveBids(address addr) view returns((address,bytes32)[])
func (_Datahub *DatahubCaller) GetActiveBids(opts *bind.CallOpts, addr common.Address) ([]DataHubActiveBid, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getActiveBids", addr)

	if err != nil {
		return *new([]DataHubActiveBid), err
	}

	out0 := *abi.ConvertType(out[0], new([]DataHubActiveBid)).(*[]DataHubActiveBid)

	return out0, err

}

// GetActiveBids is a free data retrieval call binding the contract method 0xfbc4fc44.
//
// Solidity: function getActiveBids(address addr) view returns((address,bytes32)[])
func (_Datahub *DatahubSession) GetActiveBids(addr common.Address) ([]DataHubActiveBid, error) {
	return _Datahub.Contract.GetActiveBids(&_Datahub.CallOpts, addr)
}

// GetActiveBids is a free data retrieval call binding the contract method 0xfbc4fc44.
//
// Solidity: function getActiveBids(address addr) view returns((address,bytes32)[])
func (_Datahub *DatahubCallerSession) GetActiveBids(addr common.Address) ([]DataHubActiveBid, error) {
	return _Datahub.Contract.GetActiveBids(&_Datahub.CallOpts, addr)
}

// GetAllSubItems is a free data retrieval call binding the contract method 0x224b6b8c.
//
// Solidity: function getAllSubItems(address addr) view returns((bytes32,bytes32,uint256)[])
func (_Datahub *DatahubCaller) GetAllSubItems(opts *bind.CallOpts, addr common.Address) ([]DataHubSubItem, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getAllSubItems", addr)

	if err != nil {
		return *new([]DataHubSubItem), err
	}

	out0 := *abi.ConvertType(out[0], new([]DataHubSubItem)).(*[]DataHubSubItem)

	return out0, err

}

// GetAllSubItems is a free data retrieval call binding the contract method 0x224b6b8c.
//
// Solidity: function getAllSubItems(address addr) view returns((bytes32,bytes32,uint256)[])
func (_Datahub *DatahubSession) GetAllSubItems(addr common.Address) ([]DataHubSubItem, error) {
	return _Datahub.Contract.GetAllSubItems(&_Datahub.CallOpts, addr)
}

// GetAllSubItems is a free data retrieval call binding the contract method 0x224b6b8c.
//
// Solidity: function getAllSubItems(address addr) view returns((bytes32,bytes32,uint256)[])
func (_Datahub *DatahubCallerSession) GetAllSubItems(addr common.Address) ([]DataHubSubItem, error) {
	return _Datahub.Contract.GetAllSubItems(&_Datahub.CallOpts, addr)
}

// GetCategory is a free data retrieval call binding the contract method 0x473b084c.
//
// Solidity: function getCategory(bytes32 category) view returns((uint64[]))
func (_Datahub *DatahubCaller) GetCategory(opts *bind.CallOpts, category [32]byte) (DataHubCategory, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getCategory", category)

	if err != nil {
		return *new(DataHubCategory), err
	}

	out0 := *abi.ConvertType(out[0], new(DataHubCategory)).(*DataHubCategory)

	return out0, err

}

// GetCategory is a free data retrieval call binding the contract method 0x473b084c.
//
// Solidity: function getCategory(bytes32 category) view returns((uint64[]))
func (_Datahub *DatahubSession) GetCategory(category [32]byte) (DataHubCategory, error) {
	return _Datahub.Contract.GetCategory(&_Datahub.CallOpts, category)
}

// GetCategory is a free data retrieval call binding the contract method 0x473b084c.
//
// Solidity: function getCategory(bytes32 category) view returns((uint64[]))
func (_Datahub *DatahubCallerSession) GetCategory(category [32]byte) (DataHubCategory, error) {
	return _Datahub.Contract.GetCategory(&_Datahub.CallOpts, category)
}

// GetFee is a free data retrieval call binding the contract method 0xd250185c.
//
// Solidity: function getFee(uint256 _fee, uint256 amount) pure returns(uint256)
func (_Datahub *DatahubCaller) GetFee(opts *bind.CallOpts, _fee *big.Int, amount *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getFee", _fee, amount)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetFee is a free data retrieval call binding the contract method 0xd250185c.
//
// Solidity: function getFee(uint256 _fee, uint256 amount) pure returns(uint256)
func (_Datahub *DatahubSession) GetFee(_fee *big.Int, amount *big.Int) (*big.Int, error) {
	return _Datahub.Contract.GetFee(&_Datahub.CallOpts, _fee, amount)
}

// GetFee is a free data retrieval call binding the contract method 0xd250185c.
//
// Solidity: function getFee(uint256 _fee, uint256 amount) pure returns(uint256)
func (_Datahub *DatahubCallerSession) GetFee(_fee *big.Int, amount *big.Int) (*big.Int, error) {
	return _Datahub.Contract.GetFee(&_Datahub.CallOpts, _fee, amount)
}

// GetListedSubs is a free data retrieval call binding the contract method 0xcddf64ea.
//
// Solidity: function getListedSubs(address addr) view returns(bytes32[])
func (_Datahub *DatahubCaller) GetListedSubs(opts *bind.CallOpts, addr common.Address) ([][32]byte, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getListedSubs", addr)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetListedSubs is a free data retrieval call binding the contract method 0xcddf64ea.
//
// Solidity: function getListedSubs(address addr) view returns(bytes32[])
func (_Datahub *DatahubSession) GetListedSubs(addr common.Address) ([][32]byte, error) {
	return _Datahub.Contract.GetListedSubs(&_Datahub.CallOpts, addr)
}

// GetListedSubs is a free data retrieval call binding the contract method 0xcddf64ea.
//
// Solidity: function getListedSubs(address addr) view returns(bytes32[])
func (_Datahub *DatahubCallerSession) GetListedSubs(addr common.Address) ([][32]byte, error) {
	return _Datahub.Contract.GetListedSubs(&_Datahub.CallOpts, addr)
}

// GetPortableAddress is a free data retrieval call binding the contract method 0xc3b4dde9.
//
// Solidity: function getPortableAddress(address addr) view returns(address)
func (_Datahub *DatahubCaller) GetPortableAddress(opts *bind.CallOpts, addr common.Address) (common.Address, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getPortableAddress", addr)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetPortableAddress is a free data retrieval call binding the contract method 0xc3b4dde9.
//
// Solidity: function getPortableAddress(address addr) view returns(address)
func (_Datahub *DatahubSession) GetPortableAddress(addr common.Address) (common.Address, error) {
	return _Datahub.Contract.GetPortableAddress(&_Datahub.CallOpts, addr)
}

// GetPortableAddress is a free data retrieval call binding the contract method 0xc3b4dde9.
//
// Solidity: function getPortableAddress(address addr) view returns(address)
func (_Datahub *DatahubCallerSession) GetPortableAddress(addr common.Address) (common.Address, error) {
	return _Datahub.Contract.GetPortableAddress(&_Datahub.CallOpts, addr)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Datahub *DatahubCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Datahub *DatahubSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Datahub.Contract.GetRoleAdmin(&_Datahub.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Datahub *DatahubCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Datahub.Contract.GetRoleAdmin(&_Datahub.CallOpts, role)
}

// GetSubBy is a free data retrieval call binding the contract method 0x1f9ef490.
//
// Solidity: function getSubBy(bytes32 subHash) view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32,uint16))
func (_Datahub *DatahubCaller) GetSubBy(opts *bind.CallOpts, subHash [32]byte) (DataHubSub, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getSubBy", subHash)

	if err != nil {
		return *new(DataHubSub), err
	}

	out0 := *abi.ConvertType(out[0], new(DataHubSub)).(*DataHubSub)

	return out0, err

}

// GetSubBy is a free data retrieval call binding the contract method 0x1f9ef490.
//
// Solidity: function getSubBy(bytes32 subHash) view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32,uint16))
func (_Datahub *DatahubSession) GetSubBy(subHash [32]byte) (DataHubSub, error) {
	return _Datahub.Contract.GetSubBy(&_Datahub.CallOpts, subHash)
}

// GetSubBy is a free data retrieval call binding the contract method 0x1f9ef490.
//
// Solidity: function getSubBy(bytes32 subHash) view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32,uint16))
func (_Datahub *DatahubCallerSession) GetSubBy(subHash [32]byte) (DataHubSub, error) {
	return _Datahub.Contract.GetSubBy(&_Datahub.CallOpts, subHash)
}

// GetSubByIndex is a free data retrieval call binding the contract method 0xeed5b6e5.
//
// Solidity: function getSubByIndex(uint256 index) view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32,uint16))
func (_Datahub *DatahubCaller) GetSubByIndex(opts *bind.CallOpts, index *big.Int) (DataHubSub, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getSubByIndex", index)

	if err != nil {
		return *new(DataHubSub), err
	}

	out0 := *abi.ConvertType(out[0], new(DataHubSub)).(*DataHubSub)

	return out0, err

}

// GetSubByIndex is a free data retrieval call binding the contract method 0xeed5b6e5.
//
// Solidity: function getSubByIndex(uint256 index) view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32,uint16))
func (_Datahub *DatahubSession) GetSubByIndex(index *big.Int) (DataHubSub, error) {
	return _Datahub.Contract.GetSubByIndex(&_Datahub.CallOpts, index)
}

// GetSubByIndex is a free data retrieval call binding the contract method 0xeed5b6e5.
//
// Solidity: function getSubByIndex(uint256 index) view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32,uint16))
func (_Datahub *DatahubCallerSession) GetSubByIndex(index *big.Int) (DataHubSub, error) {
	return _Datahub.Contract.GetSubByIndex(&_Datahub.CallOpts, index)
}

// GetSubInfoBalance is a free data retrieval call binding the contract method 0x254e287b.
//
// Solidity: function getSubInfoBalance(bytes32 subHash, address forAddress) view returns(uint256)
func (_Datahub *DatahubCaller) GetSubInfoBalance(opts *bind.CallOpts, subHash [32]byte, forAddress common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getSubInfoBalance", subHash, forAddress)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSubInfoBalance is a free data retrieval call binding the contract method 0x254e287b.
//
// Solidity: function getSubInfoBalance(bytes32 subHash, address forAddress) view returns(uint256)
func (_Datahub *DatahubSession) GetSubInfoBalance(subHash [32]byte, forAddress common.Address) (*big.Int, error) {
	return _Datahub.Contract.GetSubInfoBalance(&_Datahub.CallOpts, subHash, forAddress)
}

// GetSubInfoBalance is a free data retrieval call binding the contract method 0x254e287b.
//
// Solidity: function getSubInfoBalance(bytes32 subHash, address forAddress) view returns(uint256)
func (_Datahub *DatahubCallerSession) GetSubInfoBalance(subHash [32]byte, forAddress common.Address) (*big.Int, error) {
	return _Datahub.Contract.GetSubInfoBalance(&_Datahub.CallOpts, subHash, forAddress)
}

// GetSubItemAt is a free data retrieval call binding the contract method 0x80dd0d8e.
//
// Solidity: function getSubItemAt(address addr, uint256 index) view returns((bytes32,bytes32,uint256))
func (_Datahub *DatahubCaller) GetSubItemAt(opts *bind.CallOpts, addr common.Address, index *big.Int) (DataHubSubItem, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getSubItemAt", addr, index)

	if err != nil {
		return *new(DataHubSubItem), err
	}

	out0 := *abi.ConvertType(out[0], new(DataHubSubItem)).(*DataHubSubItem)

	return out0, err

}

// GetSubItemAt is a free data retrieval call binding the contract method 0x80dd0d8e.
//
// Solidity: function getSubItemAt(address addr, uint256 index) view returns((bytes32,bytes32,uint256))
func (_Datahub *DatahubSession) GetSubItemAt(addr common.Address, index *big.Int) (DataHubSubItem, error) {
	return _Datahub.Contract.GetSubItemAt(&_Datahub.CallOpts, addr, index)
}

// GetSubItemAt is a free data retrieval call binding the contract method 0x80dd0d8e.
//
// Solidity: function getSubItemAt(address addr, uint256 index) view returns((bytes32,bytes32,uint256))
func (_Datahub *DatahubCallerSession) GetSubItemAt(addr common.Address, index *big.Int) (DataHubSubItem, error) {
	return _Datahub.Contract.GetSubItemAt(&_Datahub.CallOpts, addr, index)
}

// GetSubItemBy is a free data retrieval call binding the contract method 0x9aad57bb.
//
// Solidity: function getSubItemBy(address addr, bytes32 subHash) view returns((bytes32,bytes32,uint256))
func (_Datahub *DatahubCaller) GetSubItemBy(opts *bind.CallOpts, addr common.Address, subHash [32]byte) (DataHubSubItem, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getSubItemBy", addr, subHash)

	if err != nil {
		return *new(DataHubSubItem), err
	}

	out0 := *abi.ConvertType(out[0], new(DataHubSubItem)).(*DataHubSubItem)

	return out0, err

}

// GetSubItemBy is a free data retrieval call binding the contract method 0x9aad57bb.
//
// Solidity: function getSubItemBy(address addr, bytes32 subHash) view returns((bytes32,bytes32,uint256))
func (_Datahub *DatahubSession) GetSubItemBy(addr common.Address, subHash [32]byte) (DataHubSubItem, error) {
	return _Datahub.Contract.GetSubItemBy(&_Datahub.CallOpts, addr, subHash)
}

// GetSubItemBy is a free data retrieval call binding the contract method 0x9aad57bb.
//
// Solidity: function getSubItemBy(address addr, bytes32 subHash) view returns((bytes32,bytes32,uint256))
func (_Datahub *DatahubCallerSession) GetSubItemBy(addr common.Address, subHash [32]byte) (DataHubSubItem, error) {
	return _Datahub.Contract.GetSubItemBy(&_Datahub.CallOpts, addr, subHash)
}

// GetSubItems is a free data retrieval call binding the contract method 0xd3fbc74c.
//
// Solidity: function getSubItems(address addr, uint256 start, uint256 length) view returns((bytes32,bytes32,uint256)[] items, uint256 last)
func (_Datahub *DatahubCaller) GetSubItems(opts *bind.CallOpts, addr common.Address, start *big.Int, length *big.Int) (struct {
	Items []DataHubSubItem
	Last  *big.Int
}, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getSubItems", addr, start, length)

	outstruct := new(struct {
		Items []DataHubSubItem
		Last  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Items = *abi.ConvertType(out[0], new([]DataHubSubItem)).(*[]DataHubSubItem)
	outstruct.Last = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetSubItems is a free data retrieval call binding the contract method 0xd3fbc74c.
//
// Solidity: function getSubItems(address addr, uint256 start, uint256 length) view returns((bytes32,bytes32,uint256)[] items, uint256 last)
func (_Datahub *DatahubSession) GetSubItems(addr common.Address, start *big.Int, length *big.Int) (struct {
	Items []DataHubSubItem
	Last  *big.Int
}, error) {
	return _Datahub.Contract.GetSubItems(&_Datahub.CallOpts, addr, start, length)
}

// GetSubItems is a free data retrieval call binding the contract method 0xd3fbc74c.
//
// Solidity: function getSubItems(address addr, uint256 start, uint256 length) view returns((bytes32,bytes32,uint256)[] items, uint256 last)
func (_Datahub *DatahubCallerSession) GetSubItems(addr common.Address, start *big.Int, length *big.Int) (struct {
	Items []DataHubSubItem
	Last  *big.Int
}, error) {
	return _Datahub.Contract.GetSubItems(&_Datahub.CallOpts, addr, start, length)
}

// GetSubRequestAt is a free data retrieval call binding the contract method 0x84053229.
//
// Solidity: function getSubRequestAt(address addr, uint256 index) view returns((bytes32,address,bytes32,bytes32))
func (_Datahub *DatahubCaller) GetSubRequestAt(opts *bind.CallOpts, addr common.Address, index *big.Int) (DataHubSubRequest, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getSubRequestAt", addr, index)

	if err != nil {
		return *new(DataHubSubRequest), err
	}

	out0 := *abi.ConvertType(out[0], new(DataHubSubRequest)).(*DataHubSubRequest)

	return out0, err

}

// GetSubRequestAt is a free data retrieval call binding the contract method 0x84053229.
//
// Solidity: function getSubRequestAt(address addr, uint256 index) view returns((bytes32,address,bytes32,bytes32))
func (_Datahub *DatahubSession) GetSubRequestAt(addr common.Address, index *big.Int) (DataHubSubRequest, error) {
	return _Datahub.Contract.GetSubRequestAt(&_Datahub.CallOpts, addr, index)
}

// GetSubRequestAt is a free data retrieval call binding the contract method 0x84053229.
//
// Solidity: function getSubRequestAt(address addr, uint256 index) view returns((bytes32,address,bytes32,bytes32))
func (_Datahub *DatahubCallerSession) GetSubRequestAt(addr common.Address, index *big.Int) (DataHubSubRequest, error) {
	return _Datahub.Contract.GetSubRequestAt(&_Datahub.CallOpts, addr, index)
}

// GetSubRequestByHash is a free data retrieval call binding the contract method 0x9bde82dc.
//
// Solidity: function getSubRequestByHash(address addr, bytes32 requestHash) view returns((bytes32,address,bytes32,bytes32))
func (_Datahub *DatahubCaller) GetSubRequestByHash(opts *bind.CallOpts, addr common.Address, requestHash [32]byte) (DataHubSubRequest, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getSubRequestByHash", addr, requestHash)

	if err != nil {
		return *new(DataHubSubRequest), err
	}

	out0 := *abi.ConvertType(out[0], new(DataHubSubRequest)).(*DataHubSubRequest)

	return out0, err

}

// GetSubRequestByHash is a free data retrieval call binding the contract method 0x9bde82dc.
//
// Solidity: function getSubRequestByHash(address addr, bytes32 requestHash) view returns((bytes32,address,bytes32,bytes32))
func (_Datahub *DatahubSession) GetSubRequestByHash(addr common.Address, requestHash [32]byte) (DataHubSubRequest, error) {
	return _Datahub.Contract.GetSubRequestByHash(&_Datahub.CallOpts, addr, requestHash)
}

// GetSubRequestByHash is a free data retrieval call binding the contract method 0x9bde82dc.
//
// Solidity: function getSubRequestByHash(address addr, bytes32 requestHash) view returns((bytes32,address,bytes32,bytes32))
func (_Datahub *DatahubCallerSession) GetSubRequestByHash(addr common.Address, requestHash [32]byte) (DataHubSubRequest, error) {
	return _Datahub.Contract.GetSubRequestByHash(&_Datahub.CallOpts, addr, requestHash)
}

// GetSubRequests is a free data retrieval call binding the contract method 0x92b58bc2.
//
// Solidity: function getSubRequests(address addr) view returns((bytes32,address,bytes32,bytes32)[])
func (_Datahub *DatahubCaller) GetSubRequests(opts *bind.CallOpts, addr common.Address) ([]DataHubSubRequest, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getSubRequests", addr)

	if err != nil {
		return *new([]DataHubSubRequest), err
	}

	out0 := *abi.ConvertType(out[0], new([]DataHubSubRequest)).(*[]DataHubSubRequest)

	return out0, err

}

// GetSubRequests is a free data retrieval call binding the contract method 0x92b58bc2.
//
// Solidity: function getSubRequests(address addr) view returns((bytes32,address,bytes32,bytes32)[])
func (_Datahub *DatahubSession) GetSubRequests(addr common.Address) ([]DataHubSubRequest, error) {
	return _Datahub.Contract.GetSubRequests(&_Datahub.CallOpts, addr)
}

// GetSubRequests is a free data retrieval call binding the contract method 0x92b58bc2.
//
// Solidity: function getSubRequests(address addr) view returns((bytes32,address,bytes32,bytes32)[])
func (_Datahub *DatahubCallerSession) GetSubRequests(addr common.Address) ([]DataHubSubRequest, error) {
	return _Datahub.Contract.GetSubRequests(&_Datahub.CallOpts, addr)
}

// GetSubSubscribers is a free data retrieval call binding the contract method 0x7de2e5e8.
//
// Solidity: function getSubSubscribers(bytes32 subHash) view returns(address[])
func (_Datahub *DatahubCaller) GetSubSubscribers(opts *bind.CallOpts, subHash [32]byte) ([]common.Address, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getSubSubscribers", subHash)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetSubSubscribers is a free data retrieval call binding the contract method 0x7de2e5e8.
//
// Solidity: function getSubSubscribers(bytes32 subHash) view returns(address[])
func (_Datahub *DatahubSession) GetSubSubscribers(subHash [32]byte) ([]common.Address, error) {
	return _Datahub.Contract.GetSubSubscribers(&_Datahub.CallOpts, subHash)
}

// GetSubSubscribers is a free data retrieval call binding the contract method 0x7de2e5e8.
//
// Solidity: function getSubSubscribers(bytes32 subHash) view returns(address[])
func (_Datahub *DatahubCallerSession) GetSubSubscribers(subHash [32]byte) ([]common.Address, error) {
	return _Datahub.Contract.GetSubSubscribers(&_Datahub.CallOpts, subHash)
}

// GetSubs is a free data retrieval call binding the contract method 0xb8fb1bac.
//
// Solidity: function getSubs() view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32,uint16)[])
func (_Datahub *DatahubCaller) GetSubs(opts *bind.CallOpts) ([]DataHubSub, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getSubs")

	if err != nil {
		return *new([]DataHubSub), err
	}

	out0 := *abi.ConvertType(out[0], new([]DataHubSub)).(*[]DataHubSub)

	return out0, err

}

// GetSubs is a free data retrieval call binding the contract method 0xb8fb1bac.
//
// Solidity: function getSubs() view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32,uint16)[])
func (_Datahub *DatahubSession) GetSubs() ([]DataHubSub, error) {
	return _Datahub.Contract.GetSubs(&_Datahub.CallOpts)
}

// GetSubs is a free data retrieval call binding the contract method 0xb8fb1bac.
//
// Solidity: function getSubs() view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32,uint16)[])
func (_Datahub *DatahubCallerSession) GetSubs() ([]DataHubSub, error) {
	return _Datahub.Contract.GetSubs(&_Datahub.CallOpts)
}

// GetUserStats is a free data retrieval call binding the contract method 0x4e43603a.
//
// Solidity: function getUserStats(address addr) view returns(uint256 numSubRequests, uint256 numSubItems, uint256 numActiveBids, uint256 numListedSubs)
func (_Datahub *DatahubCaller) GetUserStats(opts *bind.CallOpts, addr common.Address) (struct {
	NumSubRequests *big.Int
	NumSubItems    *big.Int
	NumActiveBids  *big.Int
	NumListedSubs  *big.Int
}, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "getUserStats", addr)

	outstruct := new(struct {
		NumSubRequests *big.Int
		NumSubItems    *big.Int
		NumActiveBids  *big.Int
		NumListedSubs  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.NumSubRequests = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.NumSubItems = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.NumActiveBids = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.NumListedSubs = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetUserStats is a free data retrieval call binding the contract method 0x4e43603a.
//
// Solidity: function getUserStats(address addr) view returns(uint256 numSubRequests, uint256 numSubItems, uint256 numActiveBids, uint256 numListedSubs)
func (_Datahub *DatahubSession) GetUserStats(addr common.Address) (struct {
	NumSubRequests *big.Int
	NumSubItems    *big.Int
	NumActiveBids  *big.Int
	NumListedSubs  *big.Int
}, error) {
	return _Datahub.Contract.GetUserStats(&_Datahub.CallOpts, addr)
}

// GetUserStats is a free data retrieval call binding the contract method 0x4e43603a.
//
// Solidity: function getUserStats(address addr) view returns(uint256 numSubRequests, uint256 numSubItems, uint256 numActiveBids, uint256 numListedSubs)
func (_Datahub *DatahubCallerSession) GetUserStats(addr common.Address) (struct {
	NumSubRequests *big.Int
	NumSubItems    *big.Int
	NumActiveBids  *big.Int
	NumListedSubs  *big.Int
}, error) {
	return _Datahub.Contract.GetUserStats(&_Datahub.CallOpts, addr)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Datahub *DatahubCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Datahub *DatahubSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Datahub.Contract.HasRole(&_Datahub.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Datahub *DatahubCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Datahub.Contract.HasRole(&_Datahub.CallOpts, role, account)
}

// InEscrow is a free data retrieval call binding the contract method 0xb7391341.
//
// Solidity: function inEscrow() view returns(uint256)
func (_Datahub *DatahubCaller) InEscrow(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "inEscrow")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// InEscrow is a free data retrieval call binding the contract method 0xb7391341.
//
// Solidity: function inEscrow() view returns(uint256)
func (_Datahub *DatahubSession) InEscrow() (*big.Int, error) {
	return _Datahub.Contract.InEscrow(&_Datahub.CallOpts)
}

// InEscrow is a free data retrieval call binding the contract method 0xb7391341.
//
// Solidity: function inEscrow() view returns(uint256)
func (_Datahub *DatahubCallerSession) InEscrow() (*big.Int, error) {
	return _Datahub.Contract.InEscrow(&_Datahub.CallOpts)
}

// MarketFee is a free data retrieval call binding the contract method 0x0ccf2156.
//
// Solidity: function marketFee() view returns(uint256)
func (_Datahub *DatahubCaller) MarketFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "marketFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MarketFee is a free data retrieval call binding the contract method 0x0ccf2156.
//
// Solidity: function marketFee() view returns(uint256)
func (_Datahub *DatahubSession) MarketFee() (*big.Int, error) {
	return _Datahub.Contract.MarketFee(&_Datahub.CallOpts)
}

// MarketFee is a free data retrieval call binding the contract method 0x0ccf2156.
//
// Solidity: function marketFee() view returns(uint256)
func (_Datahub *DatahubCallerSession) MarketFee() (*big.Int, error) {
	return _Datahub.Contract.MarketFee(&_Datahub.CallOpts)
}

// MinListingFee is a free data retrieval call binding the contract method 0x703a54b5.
//
// Solidity: function minListingFee() view returns(uint256)
func (_Datahub *DatahubCaller) MinListingFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "minListingFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinListingFee is a free data retrieval call binding the contract method 0x703a54b5.
//
// Solidity: function minListingFee() view returns(uint256)
func (_Datahub *DatahubSession) MinListingFee() (*big.Int, error) {
	return _Datahub.Contract.MinListingFee(&_Datahub.CallOpts)
}

// MinListingFee is a free data retrieval call binding the contract method 0x703a54b5.
//
// Solidity: function minListingFee() view returns(uint256)
func (_Datahub *DatahubCallerSession) MinListingFee() (*big.Int, error) {
	return _Datahub.Contract.MinListingFee(&_Datahub.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Datahub *DatahubCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Datahub *DatahubSession) Owner() (common.Address, error) {
	return _Datahub.Contract.Owner(&_Datahub.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Datahub *DatahubCallerSession) Owner() (common.Address, error) {
	return _Datahub.Contract.Owner(&_Datahub.CallOpts)
}

// SubscriptionIds is a free data retrieval call binding the contract method 0x0e499994.
//
// Solidity: function subscriptionIds(bytes32 ) view returns(uint256)
func (_Datahub *DatahubCaller) SubscriptionIds(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "subscriptionIds", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SubscriptionIds is a free data retrieval call binding the contract method 0x0e499994.
//
// Solidity: function subscriptionIds(bytes32 ) view returns(uint256)
func (_Datahub *DatahubSession) SubscriptionIds(arg0 [32]byte) (*big.Int, error) {
	return _Datahub.Contract.SubscriptionIds(&_Datahub.CallOpts, arg0)
}

// SubscriptionIds is a free data retrieval call binding the contract method 0x0e499994.
//
// Solidity: function subscriptionIds(bytes32 ) view returns(uint256)
func (_Datahub *DatahubCallerSession) SubscriptionIds(arg0 [32]byte) (*big.Int, error) {
	return _Datahub.Contract.SubscriptionIds(&_Datahub.CallOpts, arg0)
}

// Subscriptions is a free data retrieval call binding the contract method 0x2d5bbf60.
//
// Solidity: function subscriptions(uint256 ) view returns(bytes32 subHash, bytes32 fdpSellerNameHash, address seller, bytes32 swarmLocation, uint256 price, bool active, uint256 earned, uint32 bids, uint32 sells, uint32 reports, uint16 daysValid)
func (_Datahub *DatahubCaller) Subscriptions(opts *bind.CallOpts, arg0 *big.Int) (struct {
	SubHash           [32]byte
	FdpSellerNameHash [32]byte
	Seller            common.Address
	SwarmLocation     [32]byte
	Price             *big.Int
	Active            bool
	Earned            *big.Int
	Bids              uint32
	Sells             uint32
	Reports           uint32
	DaysValid         uint16
}, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "subscriptions", arg0)

	outstruct := new(struct {
		SubHash           [32]byte
		FdpSellerNameHash [32]byte
		Seller            common.Address
		SwarmLocation     [32]byte
		Price             *big.Int
		Active            bool
		Earned            *big.Int
		Bids              uint32
		Sells             uint32
		Reports           uint32
		DaysValid         uint16
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.SubHash = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.FdpSellerNameHash = *abi.ConvertType(out[1], new([32]byte)).(*[32]byte)
	outstruct.Seller = *abi.ConvertType(out[2], new(common.Address)).(*common.Address)
	outstruct.SwarmLocation = *abi.ConvertType(out[3], new([32]byte)).(*[32]byte)
	outstruct.Price = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.Active = *abi.ConvertType(out[5], new(bool)).(*bool)
	outstruct.Earned = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)
	outstruct.Bids = *abi.ConvertType(out[7], new(uint32)).(*uint32)
	outstruct.Sells = *abi.ConvertType(out[8], new(uint32)).(*uint32)
	outstruct.Reports = *abi.ConvertType(out[9], new(uint32)).(*uint32)
	outstruct.DaysValid = *abi.ConvertType(out[10], new(uint16)).(*uint16)

	return *outstruct, err

}

// Subscriptions is a free data retrieval call binding the contract method 0x2d5bbf60.
//
// Solidity: function subscriptions(uint256 ) view returns(bytes32 subHash, bytes32 fdpSellerNameHash, address seller, bytes32 swarmLocation, uint256 price, bool active, uint256 earned, uint32 bids, uint32 sells, uint32 reports, uint16 daysValid)
func (_Datahub *DatahubSession) Subscriptions(arg0 *big.Int) (struct {
	SubHash           [32]byte
	FdpSellerNameHash [32]byte
	Seller            common.Address
	SwarmLocation     [32]byte
	Price             *big.Int
	Active            bool
	Earned            *big.Int
	Bids              uint32
	Sells             uint32
	Reports           uint32
	DaysValid         uint16
}, error) {
	return _Datahub.Contract.Subscriptions(&_Datahub.CallOpts, arg0)
}

// Subscriptions is a free data retrieval call binding the contract method 0x2d5bbf60.
//
// Solidity: function subscriptions(uint256 ) view returns(bytes32 subHash, bytes32 fdpSellerNameHash, address seller, bytes32 swarmLocation, uint256 price, bool active, uint256 earned, uint32 bids, uint32 sells, uint32 reports, uint16 daysValid)
func (_Datahub *DatahubCallerSession) Subscriptions(arg0 *big.Int) (struct {
	SubHash           [32]byte
	FdpSellerNameHash [32]byte
	Seller            common.Address
	SwarmLocation     [32]byte
	Price             *big.Int
	Active            bool
	Earned            *big.Int
	Bids              uint32
	Sells             uint32
	Reports           uint32
	DaysValid         uint16
}, error) {
	return _Datahub.Contract.Subscriptions(&_Datahub.CallOpts, arg0)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Datahub *DatahubCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Datahub.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Datahub *DatahubSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Datahub.Contract.SupportsInterface(&_Datahub.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Datahub *DatahubCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Datahub.Contract.SupportsInterface(&_Datahub.CallOpts, interfaceId)
}

// BidSub is a paid mutator transaction binding the contract method 0xe91dbcb0.
//
// Solidity: function bidSub(bytes32 subHash, bytes32 fdpBuyerNameHash) payable returns()
func (_Datahub *DatahubTransactor) BidSub(opts *bind.TransactOpts, subHash [32]byte, fdpBuyerNameHash [32]byte) (*types.Transaction, error) {
	return _Datahub.contract.Transact(opts, "bidSub", subHash, fdpBuyerNameHash)
}

// BidSub is a paid mutator transaction binding the contract method 0xe91dbcb0.
//
// Solidity: function bidSub(bytes32 subHash, bytes32 fdpBuyerNameHash) payable returns()
func (_Datahub *DatahubSession) BidSub(subHash [32]byte, fdpBuyerNameHash [32]byte) (*types.Transaction, error) {
	return _Datahub.Contract.BidSub(&_Datahub.TransactOpts, subHash, fdpBuyerNameHash)
}

// BidSub is a paid mutator transaction binding the contract method 0xe91dbcb0.
//
// Solidity: function bidSub(bytes32 subHash, bytes32 fdpBuyerNameHash) payable returns()
func (_Datahub *DatahubTransactorSession) BidSub(subHash [32]byte, fdpBuyerNameHash [32]byte) (*types.Transaction, error) {
	return _Datahub.Contract.BidSub(&_Datahub.TransactOpts, subHash, fdpBuyerNameHash)
}

// EnableSub is a paid mutator transaction binding the contract method 0x88ac2917.
//
// Solidity: function enableSub(bytes32 subHash, bool active) returns()
func (_Datahub *DatahubTransactor) EnableSub(opts *bind.TransactOpts, subHash [32]byte, active bool) (*types.Transaction, error) {
	return _Datahub.contract.Transact(opts, "enableSub", subHash, active)
}

// EnableSub is a paid mutator transaction binding the contract method 0x88ac2917.
//
// Solidity: function enableSub(bytes32 subHash, bool active) returns()
func (_Datahub *DatahubSession) EnableSub(subHash [32]byte, active bool) (*types.Transaction, error) {
	return _Datahub.Contract.EnableSub(&_Datahub.TransactOpts, subHash, active)
}

// EnableSub is a paid mutator transaction binding the contract method 0x88ac2917.
//
// Solidity: function enableSub(bytes32 subHash, bool active) returns()
func (_Datahub *DatahubTransactorSession) EnableSub(subHash [32]byte, active bool) (*types.Transaction, error) {
	return _Datahub.Contract.EnableSub(&_Datahub.TransactOpts, subHash, active)
}

// FundsTransfer is a paid mutator transaction binding the contract method 0x567556a4.
//
// Solidity: function fundsTransfer() payable returns()
func (_Datahub *DatahubTransactor) FundsTransfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Datahub.contract.Transact(opts, "fundsTransfer")
}

// FundsTransfer is a paid mutator transaction binding the contract method 0x567556a4.
//
// Solidity: function fundsTransfer() payable returns()
func (_Datahub *DatahubSession) FundsTransfer() (*types.Transaction, error) {
	return _Datahub.Contract.FundsTransfer(&_Datahub.TransactOpts)
}

// FundsTransfer is a paid mutator transaction binding the contract method 0x567556a4.
//
// Solidity: function fundsTransfer() payable returns()
func (_Datahub *DatahubTransactorSession) FundsTransfer() (*types.Transaction, error) {
	return _Datahub.Contract.FundsTransfer(&_Datahub.TransactOpts)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Datahub *DatahubTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Datahub.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Datahub *DatahubSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Datahub.Contract.GrantRole(&_Datahub.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Datahub *DatahubTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Datahub.Contract.GrantRole(&_Datahub.TransactOpts, role, account)
}

// ListSub is a paid mutator transaction binding the contract method 0x202cff8a.
//
// Solidity: function listSub(bytes32 fdpSellerNameHash, bytes32 dataSwarmLocation, uint256 price, bytes32 category, address podAddress, uint16 daysValid) payable returns()
func (_Datahub *DatahubTransactor) ListSub(opts *bind.TransactOpts, fdpSellerNameHash [32]byte, dataSwarmLocation [32]byte, price *big.Int, category [32]byte, podAddress common.Address, daysValid uint16) (*types.Transaction, error) {
	return _Datahub.contract.Transact(opts, "listSub", fdpSellerNameHash, dataSwarmLocation, price, category, podAddress, daysValid)
}

// ListSub is a paid mutator transaction binding the contract method 0x202cff8a.
//
// Solidity: function listSub(bytes32 fdpSellerNameHash, bytes32 dataSwarmLocation, uint256 price, bytes32 category, address podAddress, uint16 daysValid) payable returns()
func (_Datahub *DatahubSession) ListSub(fdpSellerNameHash [32]byte, dataSwarmLocation [32]byte, price *big.Int, category [32]byte, podAddress common.Address, daysValid uint16) (*types.Transaction, error) {
	return _Datahub.Contract.ListSub(&_Datahub.TransactOpts, fdpSellerNameHash, dataSwarmLocation, price, category, podAddress, daysValid)
}

// ListSub is a paid mutator transaction binding the contract method 0x202cff8a.
//
// Solidity: function listSub(bytes32 fdpSellerNameHash, bytes32 dataSwarmLocation, uint256 price, bytes32 category, address podAddress, uint16 daysValid) payable returns()
func (_Datahub *DatahubTransactorSession) ListSub(fdpSellerNameHash [32]byte, dataSwarmLocation [32]byte, price *big.Int, category [32]byte, podAddress common.Address, daysValid uint16) (*types.Transaction, error) {
	return _Datahub.Contract.ListSub(&_Datahub.TransactOpts, fdpSellerNameHash, dataSwarmLocation, price, category, podAddress, daysValid)
}

// Release is a paid mutator transaction binding the contract method 0x0357371d.
//
// Solidity: function release(address token, uint256 amount) returns()
func (_Datahub *DatahubTransactor) Release(opts *bind.TransactOpts, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Datahub.contract.Transact(opts, "release", token, amount)
}

// Release is a paid mutator transaction binding the contract method 0x0357371d.
//
// Solidity: function release(address token, uint256 amount) returns()
func (_Datahub *DatahubSession) Release(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Datahub.Contract.Release(&_Datahub.TransactOpts, token, amount)
}

// Release is a paid mutator transaction binding the contract method 0x0357371d.
//
// Solidity: function release(address token, uint256 amount) returns()
func (_Datahub *DatahubTransactorSession) Release(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Datahub.Contract.Release(&_Datahub.TransactOpts, token, amount)
}

// RemoveUserActiveBid is a paid mutator transaction binding the contract method 0x0260f912.
//
// Solidity: function removeUserActiveBid(bytes32 requestHash) returns()
func (_Datahub *DatahubTransactor) RemoveUserActiveBid(opts *bind.TransactOpts, requestHash [32]byte) (*types.Transaction, error) {
	return _Datahub.contract.Transact(opts, "removeUserActiveBid", requestHash)
}

// RemoveUserActiveBid is a paid mutator transaction binding the contract method 0x0260f912.
//
// Solidity: function removeUserActiveBid(bytes32 requestHash) returns()
func (_Datahub *DatahubSession) RemoveUserActiveBid(requestHash [32]byte) (*types.Transaction, error) {
	return _Datahub.Contract.RemoveUserActiveBid(&_Datahub.TransactOpts, requestHash)
}

// RemoveUserActiveBid is a paid mutator transaction binding the contract method 0x0260f912.
//
// Solidity: function removeUserActiveBid(bytes32 requestHash) returns()
func (_Datahub *DatahubTransactorSession) RemoveUserActiveBid(requestHash [32]byte) (*types.Transaction, error) {
	return _Datahub.Contract.RemoveUserActiveBid(&_Datahub.TransactOpts, requestHash)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Datahub *DatahubTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Datahub.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Datahub *DatahubSession) RenounceOwnership() (*types.Transaction, error) {
	return _Datahub.Contract.RenounceOwnership(&_Datahub.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Datahub *DatahubTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Datahub.Contract.RenounceOwnership(&_Datahub.TransactOpts)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Datahub *DatahubTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Datahub.contract.Transact(opts, "renounceRole", role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Datahub *DatahubSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Datahub.Contract.RenounceRole(&_Datahub.TransactOpts, role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Datahub *DatahubTransactorSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Datahub.Contract.RenounceRole(&_Datahub.TransactOpts, role, account)
}

// ReportSub is a paid mutator transaction binding the contract method 0xd76ac1d1.
//
// Solidity: function reportSub(bytes32 subHash) returns()
func (_Datahub *DatahubTransactor) ReportSub(opts *bind.TransactOpts, subHash [32]byte) (*types.Transaction, error) {
	return _Datahub.contract.Transact(opts, "reportSub", subHash)
}

// ReportSub is a paid mutator transaction binding the contract method 0xd76ac1d1.
//
// Solidity: function reportSub(bytes32 subHash) returns()
func (_Datahub *DatahubSession) ReportSub(subHash [32]byte) (*types.Transaction, error) {
	return _Datahub.Contract.ReportSub(&_Datahub.TransactOpts, subHash)
}

// ReportSub is a paid mutator transaction binding the contract method 0xd76ac1d1.
//
// Solidity: function reportSub(bytes32 subHash) returns()
func (_Datahub *DatahubTransactorSession) ReportSub(subHash [32]byte) (*types.Transaction, error) {
	return _Datahub.Contract.ReportSub(&_Datahub.TransactOpts, subHash)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Datahub *DatahubTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Datahub.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Datahub *DatahubSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Datahub.Contract.RevokeRole(&_Datahub.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Datahub *DatahubTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Datahub.Contract.RevokeRole(&_Datahub.TransactOpts, role, account)
}

// SellSub is a paid mutator transaction binding the contract method 0x3ca684e3.
//
// Solidity: function sellSub(bytes32 requestHash, bytes32 encryptedKeyLocation) payable returns()
func (_Datahub *DatahubTransactor) SellSub(opts *bind.TransactOpts, requestHash [32]byte, encryptedKeyLocation [32]byte) (*types.Transaction, error) {
	return _Datahub.contract.Transact(opts, "sellSub", requestHash, encryptedKeyLocation)
}

// SellSub is a paid mutator transaction binding the contract method 0x3ca684e3.
//
// Solidity: function sellSub(bytes32 requestHash, bytes32 encryptedKeyLocation) payable returns()
func (_Datahub *DatahubSession) SellSub(requestHash [32]byte, encryptedKeyLocation [32]byte) (*types.Transaction, error) {
	return _Datahub.Contract.SellSub(&_Datahub.TransactOpts, requestHash, encryptedKeyLocation)
}

// SellSub is a paid mutator transaction binding the contract method 0x3ca684e3.
//
// Solidity: function sellSub(bytes32 requestHash, bytes32 encryptedKeyLocation) payable returns()
func (_Datahub *DatahubTransactorSession) SellSub(requestHash [32]byte, encryptedKeyLocation [32]byte) (*types.Transaction, error) {
	return _Datahub.Contract.SellSub(&_Datahub.TransactOpts, requestHash, encryptedKeyLocation)
}

// SetFee is a paid mutator transaction binding the contract method 0x69fe0e2d.
//
// Solidity: function setFee(uint256 newFee) returns()
func (_Datahub *DatahubTransactor) SetFee(opts *bind.TransactOpts, newFee *big.Int) (*types.Transaction, error) {
	return _Datahub.contract.Transact(opts, "setFee", newFee)
}

// SetFee is a paid mutator transaction binding the contract method 0x69fe0e2d.
//
// Solidity: function setFee(uint256 newFee) returns()
func (_Datahub *DatahubSession) SetFee(newFee *big.Int) (*types.Transaction, error) {
	return _Datahub.Contract.SetFee(&_Datahub.TransactOpts, newFee)
}

// SetFee is a paid mutator transaction binding the contract method 0x69fe0e2d.
//
// Solidity: function setFee(uint256 newFee) returns()
func (_Datahub *DatahubTransactorSession) SetFee(newFee *big.Int) (*types.Transaction, error) {
	return _Datahub.Contract.SetFee(&_Datahub.TransactOpts, newFee)
}

// SetListingFee is a paid mutator transaction binding the contract method 0x131dbd09.
//
// Solidity: function setListingFee(uint256 newListingFee) returns()
func (_Datahub *DatahubTransactor) SetListingFee(opts *bind.TransactOpts, newListingFee *big.Int) (*types.Transaction, error) {
	return _Datahub.contract.Transact(opts, "setListingFee", newListingFee)
}

// SetListingFee is a paid mutator transaction binding the contract method 0x131dbd09.
//
// Solidity: function setListingFee(uint256 newListingFee) returns()
func (_Datahub *DatahubSession) SetListingFee(newListingFee *big.Int) (*types.Transaction, error) {
	return _Datahub.Contract.SetListingFee(&_Datahub.TransactOpts, newListingFee)
}

// SetListingFee is a paid mutator transaction binding the contract method 0x131dbd09.
//
// Solidity: function setListingFee(uint256 newListingFee) returns()
func (_Datahub *DatahubTransactorSession) SetListingFee(newListingFee *big.Int) (*types.Transaction, error) {
	return _Datahub.Contract.SetListingFee(&_Datahub.TransactOpts, newListingFee)
}

// SetPortableAddress is a paid mutator transaction binding the contract method 0xc6d05aee.
//
// Solidity: function setPortableAddress(address addr) returns()
func (_Datahub *DatahubTransactor) SetPortableAddress(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _Datahub.contract.Transact(opts, "setPortableAddress", addr)
}

// SetPortableAddress is a paid mutator transaction binding the contract method 0xc6d05aee.
//
// Solidity: function setPortableAddress(address addr) returns()
func (_Datahub *DatahubSession) SetPortableAddress(addr common.Address) (*types.Transaction, error) {
	return _Datahub.Contract.SetPortableAddress(&_Datahub.TransactOpts, addr)
}

// SetPortableAddress is a paid mutator transaction binding the contract method 0xc6d05aee.
//
// Solidity: function setPortableAddress(address addr) returns()
func (_Datahub *DatahubTransactorSession) SetPortableAddress(addr common.Address) (*types.Transaction, error) {
	return _Datahub.Contract.SetPortableAddress(&_Datahub.TransactOpts, addr)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Datahub *DatahubTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Datahub.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Datahub *DatahubSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Datahub.Contract.TransferOwnership(&_Datahub.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Datahub *DatahubTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Datahub.Contract.TransferOwnership(&_Datahub.TransactOpts, newOwner)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Datahub *DatahubTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Datahub.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Datahub *DatahubSession) Receive() (*types.Transaction, error) {
	return _Datahub.Contract.Receive(&_Datahub.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Datahub *DatahubTransactorSession) Receive() (*types.Transaction, error) {
	return _Datahub.Contract.Receive(&_Datahub.TransactOpts)
}

// DatahubOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Datahub contract.
type DatahubOwnershipTransferredIterator struct {
	Event *DatahubOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DatahubOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DatahubOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DatahubOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DatahubOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DatahubOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DatahubOwnershipTransferred represents a OwnershipTransferred event raised by the Datahub contract.
type DatahubOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Datahub *DatahubFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*DatahubOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Datahub.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &DatahubOwnershipTransferredIterator{contract: _Datahub.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Datahub *DatahubFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *DatahubOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Datahub.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DatahubOwnershipTransferred)
				if err := _Datahub.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Datahub *DatahubFilterer) ParseOwnershipTransferred(log types.Log) (*DatahubOwnershipTransferred, error) {
	event := new(DatahubOwnershipTransferred)
	if err := _Datahub.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DatahubRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the Datahub contract.
type DatahubRoleAdminChangedIterator struct {
	Event *DatahubRoleAdminChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DatahubRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DatahubRoleAdminChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DatahubRoleAdminChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DatahubRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DatahubRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DatahubRoleAdminChanged represents a RoleAdminChanged event raised by the Datahub contract.
type DatahubRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Datahub *DatahubFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*DatahubRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _Datahub.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &DatahubRoleAdminChangedIterator{contract: _Datahub.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Datahub *DatahubFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *DatahubRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _Datahub.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DatahubRoleAdminChanged)
				if err := _Datahub.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Datahub *DatahubFilterer) ParseRoleAdminChanged(log types.Log) (*DatahubRoleAdminChanged, error) {
	event := new(DatahubRoleAdminChanged)
	if err := _Datahub.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DatahubRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the Datahub contract.
type DatahubRoleGrantedIterator struct {
	Event *DatahubRoleGranted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DatahubRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DatahubRoleGranted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DatahubRoleGranted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DatahubRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DatahubRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DatahubRoleGranted represents a RoleGranted event raised by the Datahub contract.
type DatahubRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Datahub *DatahubFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*DatahubRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Datahub.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &DatahubRoleGrantedIterator{contract: _Datahub.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Datahub *DatahubFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *DatahubRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Datahub.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DatahubRoleGranted)
				if err := _Datahub.contract.UnpackLog(event, "RoleGranted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Datahub *DatahubFilterer) ParseRoleGranted(log types.Log) (*DatahubRoleGranted, error) {
	event := new(DatahubRoleGranted)
	if err := _Datahub.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DatahubRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the Datahub contract.
type DatahubRoleRevokedIterator struct {
	Event *DatahubRoleRevoked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DatahubRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DatahubRoleRevoked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DatahubRoleRevoked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DatahubRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DatahubRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DatahubRoleRevoked represents a RoleRevoked event raised by the Datahub contract.
type DatahubRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Datahub *DatahubFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*DatahubRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Datahub.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &DatahubRoleRevokedIterator{contract: _Datahub.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Datahub *DatahubFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *DatahubRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Datahub.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DatahubRoleRevoked)
				if err := _Datahub.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Datahub *DatahubFilterer) ParseRoleRevoked(log types.Log) (*DatahubRoleRevoked, error) {
	event := new(DatahubRoleRevoked)
	if err := _Datahub.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
