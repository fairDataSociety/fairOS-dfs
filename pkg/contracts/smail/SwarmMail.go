// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package SwarmMail

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

// SwarmMailActiveBid is an auto generated low-level Go binding around an user-defined struct.
type SwarmMailActiveBid struct {
	Seller      common.Address
	RequestHash [32]byte
}

// SwarmMailCategory is an auto generated low-level Go binding around an user-defined struct.
type SwarmMailCategory struct {
	SubIdxs []*big.Int
}

// SwarmMailEmail is an auto generated low-level Go binding around an user-defined struct.
type SwarmMailEmail struct {
	IsEncryption  bool
	Time          *big.Int
	From          common.Address
	To            common.Address
	SwarmLocation [32]byte
	Signed        bool
}

// SwarmMailSub is an auto generated low-level Go binding around an user-defined struct.
type SwarmMailSub struct {
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
}

// SwarmMailSubItem is an auto generated low-level Go binding around an user-defined struct.
type SwarmMailSubItem struct {
	SubHash           [32]byte
	UnlockKeyLocation [32]byte
	ValidTill         *big.Int
}

// SwarmMailSubRequest is an auto generated low-level Go binding around an user-defined struct.
type SwarmMailSubRequest struct {
	FdpBuyerNameHash [32]byte
	Buyer            common.Address
	SubHash          [32]byte
	RequestHash      [32]byte
}

// SwarmMailMetaData contains all meta data concerning the SwarmMail contract.
var SwarmMailMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"DEFAULT_ADMIN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"fdpBuyerNameHash\",\"type\":\"bytes32\"}],\"name\":\"bidSub\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"}],\"name\":\"enableSub\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"feesCollected\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"fundsBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"fundsTransfer\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getActiveBidAt\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"seller\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"requestHash\",\"type\":\"bytes32\"}],\"internalType\":\"structSwarmMail.ActiveBid\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getActiveBids\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"seller\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"requestHash\",\"type\":\"bytes32\"}],\"internalType\":\"structSwarmMail.ActiveBid[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getAllSubItems\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"unlockKeyLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"validTill\",\"type\":\"uint256\"}],\"internalType\":\"structSwarmMail.SubItem[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getBoxCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numInboxItems\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"numSentItems\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"numSubRequests\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"numSubItems\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"numActiveBids\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"category\",\"type\":\"bytes32\"}],\"name\":\"getCategory\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256[]\",\"name\":\"subIdxs\",\"type\":\"uint256[]\"}],\"internalType\":\"structSwarmMail.Category\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_fee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"getFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getInbox\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"isEncryption\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"swarmLocation\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"signed\",\"type\":\"bool\"}],\"internalType\":\"structSwarmMail.Email[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getInboxAt\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"isEncryption\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"swarmLocation\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"signed\",\"type\":\"bool\"}],\"internalType\":\"structSwarmMail.Email\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getListedSubs\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getPublicKeys\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"registered\",\"type\":\"bool\"},{\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"smail\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getSent\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"isEncryption\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"swarmLocation\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"signed\",\"type\":\"bool\"}],\"internalType\":\"structSwarmMail.Email[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getSentAt\",\"outputs\":[{\"components\":[{\"internalType\":\"bool\",\"name\":\"isEncryption\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"swarmLocation\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"signed\",\"type\":\"bool\"}],\"internalType\":\"structSwarmMail.Email\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getSub\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"fdpSellerNameHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"seller\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"swarmLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"earned\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"bids\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"sells\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"reports\",\"type\":\"uint32\"}],\"internalType\":\"structSwarmMail.Sub\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"}],\"name\":\"getSubBy\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"fdpSellerNameHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"seller\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"swarmLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"earned\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"bids\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"sells\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"reports\",\"type\":\"uint32\"}],\"internalType\":\"structSwarmMail.Sub\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"forAddress\",\"type\":\"address\"}],\"name\":\"getSubInfoBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getSubItemAt\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"unlockKeyLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"validTill\",\"type\":\"uint256\"}],\"internalType\":\"structSwarmMail.SubItem\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"}],\"name\":\"getSubItemBy\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"unlockKeyLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"validTill\",\"type\":\"uint256\"}],\"internalType\":\"structSwarmMail.SubItem\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"start\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"}],\"name\":\"getSubItems\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"unlockKeyLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"validTill\",\"type\":\"uint256\"}],\"internalType\":\"structSwarmMail.SubItem[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getSubItemsCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getSubRequestAt\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"fdpBuyerNameHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"buyer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"requestHash\",\"type\":\"bytes32\"}],\"internalType\":\"structSwarmMail.SubRequest\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"requestHash\",\"type\":\"bytes32\"}],\"name\":\"getSubRequestByHash\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"fdpBuyerNameHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"buyer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"requestHash\",\"type\":\"bytes32\"}],\"internalType\":\"structSwarmMail.SubRequest\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getSubRequests\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"fdpBuyerNameHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"buyer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"requestHash\",\"type\":\"bytes32\"}],\"internalType\":\"structSwarmMail.SubRequest[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"}],\"name\":\"getSubSubscribers\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getSubs\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"fdpSellerNameHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"seller\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"swarmLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"earned\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"bids\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"sells\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"reports\",\"type\":\"uint32\"}],\"internalType\":\"structSwarmMail.Sub[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"inEscrow\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"fdpSellerNameHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"dataSwarmLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"category\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"podAddress\",\"type\":\"address\"}],\"name\":\"listSub\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"marketFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minListingFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"smail\",\"type\":\"bytes32\"}],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"types\",\"type\":\"uint256\"},{\"internalType\":\"bytes32[]\",\"name\":\"swarmLocations\",\"type\":\"bytes32[]\"}],\"name\":\"removeEmails\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"swarmLocation\",\"type\":\"bytes32\"}],\"name\":\"removeInboxEmail\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"swarmLocation\",\"type\":\"bytes32\"}],\"name\":\"removeSentEmail\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"removeSubItem\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"requestHash\",\"type\":\"bytes32\"}],\"name\":\"removeUserActiveBid\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"requestHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"encryptedKeyLocation\",\"type\":\"bytes32\"}],\"name\":\"sellSub\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"toAddress\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"isEncryption\",\"type\":\"bool\"},{\"internalType\":\"bytes32\",\"name\":\"swarmLocation\",\"type\":\"bytes32\"}],\"name\":\"sendEmail\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newFee\",\"type\":\"uint256\"}],\"name\":\"setFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newListingFee\",\"type\":\"uint256\"}],\"name\":\"setListingFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"swarmLocation\",\"type\":\"bytes32\"}],\"name\":\"signEmail\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"subscriptionIds\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"subscriptions\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"subHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"fdpSellerNameHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"seller\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"swarmLocation\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"earned\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"bids\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"sells\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"reports\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// SwarmMailABI is the input ABI used to generate the binding from.
// Deprecated: Use SwarmMailMetaData.ABI instead.
var SwarmMailABI = SwarmMailMetaData.ABI

// SwarmMail is an auto generated Go binding around an Ethereum contract.
type SwarmMail struct {
	SwarmMailCaller     // Read-only binding to the contract
	SwarmMailTransactor // Write-only binding to the contract
	SwarmMailFilterer   // Log filterer for contract events
}

// SwarmMailCaller is an auto generated read-only Go binding around an Ethereum contract.
type SwarmMailCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwarmMailTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SwarmMailTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwarmMailFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SwarmMailFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwarmMailSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SwarmMailSession struct {
	Contract     *SwarmMail        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SwarmMailCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SwarmMailCallerSession struct {
	Contract *SwarmMailCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// SwarmMailTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SwarmMailTransactorSession struct {
	Contract     *SwarmMailTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// SwarmMailRaw is an auto generated low-level Go binding around an Ethereum contract.
type SwarmMailRaw struct {
	Contract *SwarmMail // Generic contract binding to access the raw methods on
}

// SwarmMailCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SwarmMailCallerRaw struct {
	Contract *SwarmMailCaller // Generic read-only contract binding to access the raw methods on
}

// SwarmMailTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SwarmMailTransactorRaw struct {
	Contract *SwarmMailTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSwarmMail creates a new instance of SwarmMail, bound to a specific deployed contract.
func NewSwarmMail(address common.Address, backend bind.ContractBackend) (*SwarmMail, error) {
	contract, err := bindSwarmMail(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SwarmMail{SwarmMailCaller: SwarmMailCaller{contract: contract}, SwarmMailTransactor: SwarmMailTransactor{contract: contract}, SwarmMailFilterer: SwarmMailFilterer{contract: contract}}, nil
}

// NewSwarmMailCaller creates a new read-only instance of SwarmMail, bound to a specific deployed contract.
func NewSwarmMailCaller(address common.Address, caller bind.ContractCaller) (*SwarmMailCaller, error) {
	contract, err := bindSwarmMail(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SwarmMailCaller{contract: contract}, nil
}

// NewSwarmMailTransactor creates a new write-only instance of SwarmMail, bound to a specific deployed contract.
func NewSwarmMailTransactor(address common.Address, transactor bind.ContractTransactor) (*SwarmMailTransactor, error) {
	contract, err := bindSwarmMail(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SwarmMailTransactor{contract: contract}, nil
}

// NewSwarmMailFilterer creates a new log filterer instance of SwarmMail, bound to a specific deployed contract.
func NewSwarmMailFilterer(address common.Address, filterer bind.ContractFilterer) (*SwarmMailFilterer, error) {
	contract, err := bindSwarmMail(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SwarmMailFilterer{contract: contract}, nil
}

// bindSwarmMail binds a generic wrapper to an already deployed contract.
func bindSwarmMail(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SwarmMailABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SwarmMail *SwarmMailRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SwarmMail.Contract.SwarmMailCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SwarmMail *SwarmMailRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SwarmMail.Contract.SwarmMailTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SwarmMail *SwarmMailRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SwarmMail.Contract.SwarmMailTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SwarmMail *SwarmMailCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SwarmMail.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SwarmMail *SwarmMailTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SwarmMail.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SwarmMail *SwarmMailTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SwarmMail.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_SwarmMail *SwarmMailCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_SwarmMail *SwarmMailSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _SwarmMail.Contract.DEFAULTADMINROLE(&_SwarmMail.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_SwarmMail *SwarmMailCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _SwarmMail.Contract.DEFAULTADMINROLE(&_SwarmMail.CallOpts)
}

// FeesCollected is a free data retrieval call binding the contract method 0xf071db5a.
//
// Solidity: function feesCollected() view returns(uint256)
func (_SwarmMail *SwarmMailCaller) FeesCollected(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "feesCollected")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FeesCollected is a free data retrieval call binding the contract method 0xf071db5a.
//
// Solidity: function feesCollected() view returns(uint256)
func (_SwarmMail *SwarmMailSession) FeesCollected() (*big.Int, error) {
	return _SwarmMail.Contract.FeesCollected(&_SwarmMail.CallOpts)
}

// FeesCollected is a free data retrieval call binding the contract method 0xf071db5a.
//
// Solidity: function feesCollected() view returns(uint256)
func (_SwarmMail *SwarmMailCallerSession) FeesCollected() (*big.Int, error) {
	return _SwarmMail.Contract.FeesCollected(&_SwarmMail.CallOpts)
}

// FundsBalance is a free data retrieval call binding the contract method 0x9454932c.
//
// Solidity: function fundsBalance() view returns(uint256)
func (_SwarmMail *SwarmMailCaller) FundsBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "fundsBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FundsBalance is a free data retrieval call binding the contract method 0x9454932c.
//
// Solidity: function fundsBalance() view returns(uint256)
func (_SwarmMail *SwarmMailSession) FundsBalance() (*big.Int, error) {
	return _SwarmMail.Contract.FundsBalance(&_SwarmMail.CallOpts)
}

// FundsBalance is a free data retrieval call binding the contract method 0x9454932c.
//
// Solidity: function fundsBalance() view returns(uint256)
func (_SwarmMail *SwarmMailCallerSession) FundsBalance() (*big.Int, error) {
	return _SwarmMail.Contract.FundsBalance(&_SwarmMail.CallOpts)
}

// GetActiveBidAt is a free data retrieval call binding the contract method 0x78ba33c6.
//
// Solidity: function getActiveBidAt(address addr, uint256 index) view returns((address,bytes32))
func (_SwarmMail *SwarmMailCaller) GetActiveBidAt(opts *bind.CallOpts, addr common.Address, index *big.Int) (SwarmMailActiveBid, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getActiveBidAt", addr, index)

	if err != nil {
		return *new(SwarmMailActiveBid), err
	}

	out0 := *abi.ConvertType(out[0], new(SwarmMailActiveBid)).(*SwarmMailActiveBid)

	return out0, err

}

// GetActiveBidAt is a free data retrieval call binding the contract method 0x78ba33c6.
//
// Solidity: function getActiveBidAt(address addr, uint256 index) view returns((address,bytes32))
func (_SwarmMail *SwarmMailSession) GetActiveBidAt(addr common.Address, index *big.Int) (SwarmMailActiveBid, error) {
	return _SwarmMail.Contract.GetActiveBidAt(&_SwarmMail.CallOpts, addr, index)
}

// GetActiveBidAt is a free data retrieval call binding the contract method 0x78ba33c6.
//
// Solidity: function getActiveBidAt(address addr, uint256 index) view returns((address,bytes32))
func (_SwarmMail *SwarmMailCallerSession) GetActiveBidAt(addr common.Address, index *big.Int) (SwarmMailActiveBid, error) {
	return _SwarmMail.Contract.GetActiveBidAt(&_SwarmMail.CallOpts, addr, index)
}

// GetActiveBids is a free data retrieval call binding the contract method 0xfbc4fc44.
//
// Solidity: function getActiveBids(address addr) view returns((address,bytes32)[])
func (_SwarmMail *SwarmMailCaller) GetActiveBids(opts *bind.CallOpts, addr common.Address) ([]SwarmMailActiveBid, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getActiveBids", addr)

	if err != nil {
		return *new([]SwarmMailActiveBid), err
	}

	out0 := *abi.ConvertType(out[0], new([]SwarmMailActiveBid)).(*[]SwarmMailActiveBid)

	return out0, err

}

// GetActiveBids is a free data retrieval call binding the contract method 0xfbc4fc44.
//
// Solidity: function getActiveBids(address addr) view returns((address,bytes32)[])
func (_SwarmMail *SwarmMailSession) GetActiveBids(addr common.Address) ([]SwarmMailActiveBid, error) {
	return _SwarmMail.Contract.GetActiveBids(&_SwarmMail.CallOpts, addr)
}

// GetActiveBids is a free data retrieval call binding the contract method 0xfbc4fc44.
//
// Solidity: function getActiveBids(address addr) view returns((address,bytes32)[])
func (_SwarmMail *SwarmMailCallerSession) GetActiveBids(addr common.Address) ([]SwarmMailActiveBid, error) {
	return _SwarmMail.Contract.GetActiveBids(&_SwarmMail.CallOpts, addr)
}

// GetAllSubItems is a free data retrieval call binding the contract method 0x224b6b8c.
//
// Solidity: function getAllSubItems(address addr) view returns((bytes32,bytes32,uint256)[])
func (_SwarmMail *SwarmMailCaller) GetAllSubItems(opts *bind.CallOpts, addr common.Address) ([]SwarmMailSubItem, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getAllSubItems", addr)

	if err != nil {
		return *new([]SwarmMailSubItem), err
	}

	out0 := *abi.ConvertType(out[0], new([]SwarmMailSubItem)).(*[]SwarmMailSubItem)

	return out0, err

}

// GetAllSubItems is a free data retrieval call binding the contract method 0x224b6b8c.
//
// Solidity: function getAllSubItems(address addr) view returns((bytes32,bytes32,uint256)[])
func (_SwarmMail *SwarmMailSession) GetAllSubItems(addr common.Address) ([]SwarmMailSubItem, error) {
	return _SwarmMail.Contract.GetAllSubItems(&_SwarmMail.CallOpts, addr)
}

// GetAllSubItems is a free data retrieval call binding the contract method 0x224b6b8c.
//
// Solidity: function getAllSubItems(address addr) view returns((bytes32,bytes32,uint256)[])
func (_SwarmMail *SwarmMailCallerSession) GetAllSubItems(addr common.Address) ([]SwarmMailSubItem, error) {
	return _SwarmMail.Contract.GetAllSubItems(&_SwarmMail.CallOpts, addr)
}

// GetBoxCount is a free data retrieval call binding the contract method 0xa88b5c4f.
//
// Solidity: function getBoxCount(address addr) view returns(uint256 numInboxItems, uint256 numSentItems, uint256 numSubRequests, uint256 numSubItems, uint256 numActiveBids)
func (_SwarmMail *SwarmMailCaller) GetBoxCount(opts *bind.CallOpts, addr common.Address) (struct {
	NumInboxItems  *big.Int
	NumSentItems   *big.Int
	NumSubRequests *big.Int
	NumSubItems    *big.Int
	NumActiveBids  *big.Int
}, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getBoxCount", addr)

	outstruct := new(struct {
		NumInboxItems  *big.Int
		NumSentItems   *big.Int
		NumSubRequests *big.Int
		NumSubItems    *big.Int
		NumActiveBids  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.NumInboxItems = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.NumSentItems = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.NumSubRequests = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.NumSubItems = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.NumActiveBids = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetBoxCount is a free data retrieval call binding the contract method 0xa88b5c4f.
//
// Solidity: function getBoxCount(address addr) view returns(uint256 numInboxItems, uint256 numSentItems, uint256 numSubRequests, uint256 numSubItems, uint256 numActiveBids)
func (_SwarmMail *SwarmMailSession) GetBoxCount(addr common.Address) (struct {
	NumInboxItems  *big.Int
	NumSentItems   *big.Int
	NumSubRequests *big.Int
	NumSubItems    *big.Int
	NumActiveBids  *big.Int
}, error) {
	return _SwarmMail.Contract.GetBoxCount(&_SwarmMail.CallOpts, addr)
}

// GetBoxCount is a free data retrieval call binding the contract method 0xa88b5c4f.
//
// Solidity: function getBoxCount(address addr) view returns(uint256 numInboxItems, uint256 numSentItems, uint256 numSubRequests, uint256 numSubItems, uint256 numActiveBids)
func (_SwarmMail *SwarmMailCallerSession) GetBoxCount(addr common.Address) (struct {
	NumInboxItems  *big.Int
	NumSentItems   *big.Int
	NumSubRequests *big.Int
	NumSubItems    *big.Int
	NumActiveBids  *big.Int
}, error) {
	return _SwarmMail.Contract.GetBoxCount(&_SwarmMail.CallOpts, addr)
}

// GetCategory is a free data retrieval call binding the contract method 0x473b084c.
//
// Solidity: function getCategory(bytes32 category) view returns((uint256[]))
func (_SwarmMail *SwarmMailCaller) GetCategory(opts *bind.CallOpts, category [32]byte) (SwarmMailCategory, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getCategory", category)

	if err != nil {
		return *new(SwarmMailCategory), err
	}

	out0 := *abi.ConvertType(out[0], new(SwarmMailCategory)).(*SwarmMailCategory)

	return out0, err

}

// GetCategory is a free data retrieval call binding the contract method 0x473b084c.
//
// Solidity: function getCategory(bytes32 category) view returns((uint256[]))
func (_SwarmMail *SwarmMailSession) GetCategory(category [32]byte) (SwarmMailCategory, error) {
	return _SwarmMail.Contract.GetCategory(&_SwarmMail.CallOpts, category)
}

// GetCategory is a free data retrieval call binding the contract method 0x473b084c.
//
// Solidity: function getCategory(bytes32 category) view returns((uint256[]))
func (_SwarmMail *SwarmMailCallerSession) GetCategory(category [32]byte) (SwarmMailCategory, error) {
	return _SwarmMail.Contract.GetCategory(&_SwarmMail.CallOpts, category)
}

// GetFee is a free data retrieval call binding the contract method 0xd250185c.
//
// Solidity: function getFee(uint256 _fee, uint256 amount) pure returns(uint256)
func (_SwarmMail *SwarmMailCaller) GetFee(opts *bind.CallOpts, _fee *big.Int, amount *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getFee", _fee, amount)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetFee is a free data retrieval call binding the contract method 0xd250185c.
//
// Solidity: function getFee(uint256 _fee, uint256 amount) pure returns(uint256)
func (_SwarmMail *SwarmMailSession) GetFee(_fee *big.Int, amount *big.Int) (*big.Int, error) {
	return _SwarmMail.Contract.GetFee(&_SwarmMail.CallOpts, _fee, amount)
}

// GetFee is a free data retrieval call binding the contract method 0xd250185c.
//
// Solidity: function getFee(uint256 _fee, uint256 amount) pure returns(uint256)
func (_SwarmMail *SwarmMailCallerSession) GetFee(_fee *big.Int, amount *big.Int) (*big.Int, error) {
	return _SwarmMail.Contract.GetFee(&_SwarmMail.CallOpts, _fee, amount)
}

// GetInbox is a free data retrieval call binding the contract method 0x02201681.
//
// Solidity: function getInbox(address addr) view returns((bool,uint256,address,address,bytes32,bool)[])
func (_SwarmMail *SwarmMailCaller) GetInbox(opts *bind.CallOpts, addr common.Address) ([]SwarmMailEmail, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getInbox", addr)

	if err != nil {
		return *new([]SwarmMailEmail), err
	}

	out0 := *abi.ConvertType(out[0], new([]SwarmMailEmail)).(*[]SwarmMailEmail)

	return out0, err

}

// GetInbox is a free data retrieval call binding the contract method 0x02201681.
//
// Solidity: function getInbox(address addr) view returns((bool,uint256,address,address,bytes32,bool)[])
func (_SwarmMail *SwarmMailSession) GetInbox(addr common.Address) ([]SwarmMailEmail, error) {
	return _SwarmMail.Contract.GetInbox(&_SwarmMail.CallOpts, addr)
}

// GetInbox is a free data retrieval call binding the contract method 0x02201681.
//
// Solidity: function getInbox(address addr) view returns((bool,uint256,address,address,bytes32,bool)[])
func (_SwarmMail *SwarmMailCallerSession) GetInbox(addr common.Address) ([]SwarmMailEmail, error) {
	return _SwarmMail.Contract.GetInbox(&_SwarmMail.CallOpts, addr)
}

// GetInboxAt is a free data retrieval call binding the contract method 0xed354d5e.
//
// Solidity: function getInboxAt(address addr, uint256 index) view returns((bool,uint256,address,address,bytes32,bool))
func (_SwarmMail *SwarmMailCaller) GetInboxAt(opts *bind.CallOpts, addr common.Address, index *big.Int) (SwarmMailEmail, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getInboxAt", addr, index)

	if err != nil {
		return *new(SwarmMailEmail), err
	}

	out0 := *abi.ConvertType(out[0], new(SwarmMailEmail)).(*SwarmMailEmail)

	return out0, err

}

// GetInboxAt is a free data retrieval call binding the contract method 0xed354d5e.
//
// Solidity: function getInboxAt(address addr, uint256 index) view returns((bool,uint256,address,address,bytes32,bool))
func (_SwarmMail *SwarmMailSession) GetInboxAt(addr common.Address, index *big.Int) (SwarmMailEmail, error) {
	return _SwarmMail.Contract.GetInboxAt(&_SwarmMail.CallOpts, addr, index)
}

// GetInboxAt is a free data retrieval call binding the contract method 0xed354d5e.
//
// Solidity: function getInboxAt(address addr, uint256 index) view returns((bool,uint256,address,address,bytes32,bool))
func (_SwarmMail *SwarmMailCallerSession) GetInboxAt(addr common.Address, index *big.Int) (SwarmMailEmail, error) {
	return _SwarmMail.Contract.GetInboxAt(&_SwarmMail.CallOpts, addr, index)
}

// GetListedSubs is a free data retrieval call binding the contract method 0xcddf64ea.
//
// Solidity: function getListedSubs(address addr) view returns(bytes32[])
func (_SwarmMail *SwarmMailCaller) GetListedSubs(opts *bind.CallOpts, addr common.Address) ([][32]byte, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getListedSubs", addr)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetListedSubs is a free data retrieval call binding the contract method 0xcddf64ea.
//
// Solidity: function getListedSubs(address addr) view returns(bytes32[])
func (_SwarmMail *SwarmMailSession) GetListedSubs(addr common.Address) ([][32]byte, error) {
	return _SwarmMail.Contract.GetListedSubs(&_SwarmMail.CallOpts, addr)
}

// GetListedSubs is a free data retrieval call binding the contract method 0xcddf64ea.
//
// Solidity: function getListedSubs(address addr) view returns(bytes32[])
func (_SwarmMail *SwarmMailCallerSession) GetListedSubs(addr common.Address) ([][32]byte, error) {
	return _SwarmMail.Contract.GetListedSubs(&_SwarmMail.CallOpts, addr)
}

// GetPublicKeys is a free data retrieval call binding the contract method 0x5fcbb7d6.
//
// Solidity: function getPublicKeys(address addr) view returns(bool registered, bytes32 key, bytes32 smail)
func (_SwarmMail *SwarmMailCaller) GetPublicKeys(opts *bind.CallOpts, addr common.Address) (struct {
	Registered bool
	Key        [32]byte
	Smail      [32]byte
}, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getPublicKeys", addr)

	outstruct := new(struct {
		Registered bool
		Key        [32]byte
		Smail      [32]byte
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Registered = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.Key = *abi.ConvertType(out[1], new([32]byte)).(*[32]byte)
	outstruct.Smail = *abi.ConvertType(out[2], new([32]byte)).(*[32]byte)

	return *outstruct, err

}

// GetPublicKeys is a free data retrieval call binding the contract method 0x5fcbb7d6.
//
// Solidity: function getPublicKeys(address addr) view returns(bool registered, bytes32 key, bytes32 smail)
func (_SwarmMail *SwarmMailSession) GetPublicKeys(addr common.Address) (struct {
	Registered bool
	Key        [32]byte
	Smail      [32]byte
}, error) {
	return _SwarmMail.Contract.GetPublicKeys(&_SwarmMail.CallOpts, addr)
}

// GetPublicKeys is a free data retrieval call binding the contract method 0x5fcbb7d6.
//
// Solidity: function getPublicKeys(address addr) view returns(bool registered, bytes32 key, bytes32 smail)
func (_SwarmMail *SwarmMailCallerSession) GetPublicKeys(addr common.Address) (struct {
	Registered bool
	Key        [32]byte
	Smail      [32]byte
}, error) {
	return _SwarmMail.Contract.GetPublicKeys(&_SwarmMail.CallOpts, addr)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_SwarmMail *SwarmMailCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_SwarmMail *SwarmMailSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _SwarmMail.Contract.GetRoleAdmin(&_SwarmMail.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_SwarmMail *SwarmMailCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _SwarmMail.Contract.GetRoleAdmin(&_SwarmMail.CallOpts, role)
}

// GetSent is a free data retrieval call binding the contract method 0xd75d691d.
//
// Solidity: function getSent(address addr) view returns((bool,uint256,address,address,bytes32,bool)[])
func (_SwarmMail *SwarmMailCaller) GetSent(opts *bind.CallOpts, addr common.Address) ([]SwarmMailEmail, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getSent", addr)

	if err != nil {
		return *new([]SwarmMailEmail), err
	}

	out0 := *abi.ConvertType(out[0], new([]SwarmMailEmail)).(*[]SwarmMailEmail)

	return out0, err

}

// GetSent is a free data retrieval call binding the contract method 0xd75d691d.
//
// Solidity: function getSent(address addr) view returns((bool,uint256,address,address,bytes32,bool)[])
func (_SwarmMail *SwarmMailSession) GetSent(addr common.Address) ([]SwarmMailEmail, error) {
	return _SwarmMail.Contract.GetSent(&_SwarmMail.CallOpts, addr)
}

// GetSent is a free data retrieval call binding the contract method 0xd75d691d.
//
// Solidity: function getSent(address addr) view returns((bool,uint256,address,address,bytes32,bool)[])
func (_SwarmMail *SwarmMailCallerSession) GetSent(addr common.Address) ([]SwarmMailEmail, error) {
	return _SwarmMail.Contract.GetSent(&_SwarmMail.CallOpts, addr)
}

// GetSentAt is a free data retrieval call binding the contract method 0x9d9a4f94.
//
// Solidity: function getSentAt(address addr, uint256 index) view returns((bool,uint256,address,address,bytes32,bool))
func (_SwarmMail *SwarmMailCaller) GetSentAt(opts *bind.CallOpts, addr common.Address, index *big.Int) (SwarmMailEmail, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getSentAt", addr, index)

	if err != nil {
		return *new(SwarmMailEmail), err
	}

	out0 := *abi.ConvertType(out[0], new(SwarmMailEmail)).(*SwarmMailEmail)

	return out0, err

}

// GetSentAt is a free data retrieval call binding the contract method 0x9d9a4f94.
//
// Solidity: function getSentAt(address addr, uint256 index) view returns((bool,uint256,address,address,bytes32,bool))
func (_SwarmMail *SwarmMailSession) GetSentAt(addr common.Address, index *big.Int) (SwarmMailEmail, error) {
	return _SwarmMail.Contract.GetSentAt(&_SwarmMail.CallOpts, addr, index)
}

// GetSentAt is a free data retrieval call binding the contract method 0x9d9a4f94.
//
// Solidity: function getSentAt(address addr, uint256 index) view returns((bool,uint256,address,address,bytes32,bool))
func (_SwarmMail *SwarmMailCallerSession) GetSentAt(addr common.Address, index *big.Int) (SwarmMailEmail, error) {
	return _SwarmMail.Contract.GetSentAt(&_SwarmMail.CallOpts, addr, index)
}

// GetSub is a free data retrieval call binding the contract method 0xb1bf9a22.
//
// Solidity: function getSub(uint256 index) view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32))
func (_SwarmMail *SwarmMailCaller) GetSub(opts *bind.CallOpts, index *big.Int) (SwarmMailSub, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getSub", index)

	if err != nil {
		return *new(SwarmMailSub), err
	}

	out0 := *abi.ConvertType(out[0], new(SwarmMailSub)).(*SwarmMailSub)

	return out0, err

}

// GetSub is a free data retrieval call binding the contract method 0xb1bf9a22.
//
// Solidity: function getSub(uint256 index) view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32))
func (_SwarmMail *SwarmMailSession) GetSub(index *big.Int) (SwarmMailSub, error) {
	return _SwarmMail.Contract.GetSub(&_SwarmMail.CallOpts, index)
}

// GetSub is a free data retrieval call binding the contract method 0xb1bf9a22.
//
// Solidity: function getSub(uint256 index) view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32))
func (_SwarmMail *SwarmMailCallerSession) GetSub(index *big.Int) (SwarmMailSub, error) {
	return _SwarmMail.Contract.GetSub(&_SwarmMail.CallOpts, index)
}

// GetSubBy is a free data retrieval call binding the contract method 0x1f9ef490.
//
// Solidity: function getSubBy(bytes32 subHash) view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32))
func (_SwarmMail *SwarmMailCaller) GetSubBy(opts *bind.CallOpts, subHash [32]byte) (SwarmMailSub, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getSubBy", subHash)

	if err != nil {
		return *new(SwarmMailSub), err
	}

	out0 := *abi.ConvertType(out[0], new(SwarmMailSub)).(*SwarmMailSub)

	return out0, err

}

// GetSubBy is a free data retrieval call binding the contract method 0x1f9ef490.
//
// Solidity: function getSubBy(bytes32 subHash) view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32))
func (_SwarmMail *SwarmMailSession) GetSubBy(subHash [32]byte) (SwarmMailSub, error) {
	return _SwarmMail.Contract.GetSubBy(&_SwarmMail.CallOpts, subHash)
}

// GetSubBy is a free data retrieval call binding the contract method 0x1f9ef490.
//
// Solidity: function getSubBy(bytes32 subHash) view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32))
func (_SwarmMail *SwarmMailCallerSession) GetSubBy(subHash [32]byte) (SwarmMailSub, error) {
	return _SwarmMail.Contract.GetSubBy(&_SwarmMail.CallOpts, subHash)
}

// GetSubInfoBalance is a free data retrieval call binding the contract method 0x254e287b.
//
// Solidity: function getSubInfoBalance(bytes32 subHash, address forAddress) view returns(uint256)
func (_SwarmMail *SwarmMailCaller) GetSubInfoBalance(opts *bind.CallOpts, subHash [32]byte, forAddress common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getSubInfoBalance", subHash, forAddress)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSubInfoBalance is a free data retrieval call binding the contract method 0x254e287b.
//
// Solidity: function getSubInfoBalance(bytes32 subHash, address forAddress) view returns(uint256)
func (_SwarmMail *SwarmMailSession) GetSubInfoBalance(subHash [32]byte, forAddress common.Address) (*big.Int, error) {
	return _SwarmMail.Contract.GetSubInfoBalance(&_SwarmMail.CallOpts, subHash, forAddress)
}

// GetSubInfoBalance is a free data retrieval call binding the contract method 0x254e287b.
//
// Solidity: function getSubInfoBalance(bytes32 subHash, address forAddress) view returns(uint256)
func (_SwarmMail *SwarmMailCallerSession) GetSubInfoBalance(subHash [32]byte, forAddress common.Address) (*big.Int, error) {
	return _SwarmMail.Contract.GetSubInfoBalance(&_SwarmMail.CallOpts, subHash, forAddress)
}

// GetSubItemAt is a free data retrieval call binding the contract method 0x80dd0d8e.
//
// Solidity: function getSubItemAt(address addr, uint256 index) view returns((bytes32,bytes32,uint256))
func (_SwarmMail *SwarmMailCaller) GetSubItemAt(opts *bind.CallOpts, addr common.Address, index *big.Int) (SwarmMailSubItem, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getSubItemAt", addr, index)

	if err != nil {
		return *new(SwarmMailSubItem), err
	}

	out0 := *abi.ConvertType(out[0], new(SwarmMailSubItem)).(*SwarmMailSubItem)

	return out0, err

}

// GetSubItemAt is a free data retrieval call binding the contract method 0x80dd0d8e.
//
// Solidity: function getSubItemAt(address addr, uint256 index) view returns((bytes32,bytes32,uint256))
func (_SwarmMail *SwarmMailSession) GetSubItemAt(addr common.Address, index *big.Int) (SwarmMailSubItem, error) {
	return _SwarmMail.Contract.GetSubItemAt(&_SwarmMail.CallOpts, addr, index)
}

// GetSubItemAt is a free data retrieval call binding the contract method 0x80dd0d8e.
//
// Solidity: function getSubItemAt(address addr, uint256 index) view returns((bytes32,bytes32,uint256))
func (_SwarmMail *SwarmMailCallerSession) GetSubItemAt(addr common.Address, index *big.Int) (SwarmMailSubItem, error) {
	return _SwarmMail.Contract.GetSubItemAt(&_SwarmMail.CallOpts, addr, index)
}

// GetSubItemBy is a free data retrieval call binding the contract method 0x9aad57bb.
//
// Solidity: function getSubItemBy(address addr, bytes32 subHash) view returns((bytes32,bytes32,uint256))
func (_SwarmMail *SwarmMailCaller) GetSubItemBy(opts *bind.CallOpts, addr common.Address, subHash [32]byte) (SwarmMailSubItem, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getSubItemBy", addr, subHash)

	if err != nil {
		return *new(SwarmMailSubItem), err
	}

	out0 := *abi.ConvertType(out[0], new(SwarmMailSubItem)).(*SwarmMailSubItem)

	return out0, err

}

// GetSubItemBy is a free data retrieval call binding the contract method 0x9aad57bb.
//
// Solidity: function getSubItemBy(address addr, bytes32 subHash) view returns((bytes32,bytes32,uint256))
func (_SwarmMail *SwarmMailSession) GetSubItemBy(addr common.Address, subHash [32]byte) (SwarmMailSubItem, error) {
	return _SwarmMail.Contract.GetSubItemBy(&_SwarmMail.CallOpts, addr, subHash)
}

// GetSubItemBy is a free data retrieval call binding the contract method 0x9aad57bb.
//
// Solidity: function getSubItemBy(address addr, bytes32 subHash) view returns((bytes32,bytes32,uint256))
func (_SwarmMail *SwarmMailCallerSession) GetSubItemBy(addr common.Address, subHash [32]byte) (SwarmMailSubItem, error) {
	return _SwarmMail.Contract.GetSubItemBy(&_SwarmMail.CallOpts, addr, subHash)
}

// GetSubItems is a free data retrieval call binding the contract method 0xd3fbc74c.
//
// Solidity: function getSubItems(address addr, uint256 start, uint256 length) view returns((bytes32,bytes32,uint256)[])
func (_SwarmMail *SwarmMailCaller) GetSubItems(opts *bind.CallOpts, addr common.Address, start *big.Int, length *big.Int) ([]SwarmMailSubItem, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getSubItems", addr, start, length)

	if err != nil {
		return *new([]SwarmMailSubItem), err
	}

	out0 := *abi.ConvertType(out[0], new([]SwarmMailSubItem)).(*[]SwarmMailSubItem)

	return out0, err

}

// GetSubItems is a free data retrieval call binding the contract method 0xd3fbc74c.
//
// Solidity: function getSubItems(address addr, uint256 start, uint256 length) view returns((bytes32,bytes32,uint256)[])
func (_SwarmMail *SwarmMailSession) GetSubItems(addr common.Address, start *big.Int, length *big.Int) ([]SwarmMailSubItem, error) {
	return _SwarmMail.Contract.GetSubItems(&_SwarmMail.CallOpts, addr, start, length)
}

// GetSubItems is a free data retrieval call binding the contract method 0xd3fbc74c.
//
// Solidity: function getSubItems(address addr, uint256 start, uint256 length) view returns((bytes32,bytes32,uint256)[])
func (_SwarmMail *SwarmMailCallerSession) GetSubItems(addr common.Address, start *big.Int, length *big.Int) ([]SwarmMailSubItem, error) {
	return _SwarmMail.Contract.GetSubItems(&_SwarmMail.CallOpts, addr, start, length)
}

// GetSubItemsCount is a free data retrieval call binding the contract method 0x51b6d3c9.
//
// Solidity: function getSubItemsCount(address addr) view returns(uint256)
func (_SwarmMail *SwarmMailCaller) GetSubItemsCount(opts *bind.CallOpts, addr common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getSubItemsCount", addr)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSubItemsCount is a free data retrieval call binding the contract method 0x51b6d3c9.
//
// Solidity: function getSubItemsCount(address addr) view returns(uint256)
func (_SwarmMail *SwarmMailSession) GetSubItemsCount(addr common.Address) (*big.Int, error) {
	return _SwarmMail.Contract.GetSubItemsCount(&_SwarmMail.CallOpts, addr)
}

// GetSubItemsCount is a free data retrieval call binding the contract method 0x51b6d3c9.
//
// Solidity: function getSubItemsCount(address addr) view returns(uint256)
func (_SwarmMail *SwarmMailCallerSession) GetSubItemsCount(addr common.Address) (*big.Int, error) {
	return _SwarmMail.Contract.GetSubItemsCount(&_SwarmMail.CallOpts, addr)
}

// GetSubRequestAt is a free data retrieval call binding the contract method 0x84053229.
//
// Solidity: function getSubRequestAt(address addr, uint256 index) view returns((bytes32,address,bytes32,bytes32))
func (_SwarmMail *SwarmMailCaller) GetSubRequestAt(opts *bind.CallOpts, addr common.Address, index *big.Int) (SwarmMailSubRequest, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getSubRequestAt", addr, index)

	if err != nil {
		return *new(SwarmMailSubRequest), err
	}

	out0 := *abi.ConvertType(out[0], new(SwarmMailSubRequest)).(*SwarmMailSubRequest)

	return out0, err

}

// GetSubRequestAt is a free data retrieval call binding the contract method 0x84053229.
//
// Solidity: function getSubRequestAt(address addr, uint256 index) view returns((bytes32,address,bytes32,bytes32))
func (_SwarmMail *SwarmMailSession) GetSubRequestAt(addr common.Address, index *big.Int) (SwarmMailSubRequest, error) {
	return _SwarmMail.Contract.GetSubRequestAt(&_SwarmMail.CallOpts, addr, index)
}

// GetSubRequestAt is a free data retrieval call binding the contract method 0x84053229.
//
// Solidity: function getSubRequestAt(address addr, uint256 index) view returns((bytes32,address,bytes32,bytes32))
func (_SwarmMail *SwarmMailCallerSession) GetSubRequestAt(addr common.Address, index *big.Int) (SwarmMailSubRequest, error) {
	return _SwarmMail.Contract.GetSubRequestAt(&_SwarmMail.CallOpts, addr, index)
}

// GetSubRequestByHash is a free data retrieval call binding the contract method 0x9bde82dc.
//
// Solidity: function getSubRequestByHash(address addr, bytes32 requestHash) view returns((bytes32,address,bytes32,bytes32))
func (_SwarmMail *SwarmMailCaller) GetSubRequestByHash(opts *bind.CallOpts, addr common.Address, requestHash [32]byte) (SwarmMailSubRequest, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getSubRequestByHash", addr, requestHash)

	if err != nil {
		return *new(SwarmMailSubRequest), err
	}

	out0 := *abi.ConvertType(out[0], new(SwarmMailSubRequest)).(*SwarmMailSubRequest)

	return out0, err

}

// GetSubRequestByHash is a free data retrieval call binding the contract method 0x9bde82dc.
//
// Solidity: function getSubRequestByHash(address addr, bytes32 requestHash) view returns((bytes32,address,bytes32,bytes32))
func (_SwarmMail *SwarmMailSession) GetSubRequestByHash(addr common.Address, requestHash [32]byte) (SwarmMailSubRequest, error) {
	return _SwarmMail.Contract.GetSubRequestByHash(&_SwarmMail.CallOpts, addr, requestHash)
}

// GetSubRequestByHash is a free data retrieval call binding the contract method 0x9bde82dc.
//
// Solidity: function getSubRequestByHash(address addr, bytes32 requestHash) view returns((bytes32,address,bytes32,bytes32))
func (_SwarmMail *SwarmMailCallerSession) GetSubRequestByHash(addr common.Address, requestHash [32]byte) (SwarmMailSubRequest, error) {
	return _SwarmMail.Contract.GetSubRequestByHash(&_SwarmMail.CallOpts, addr, requestHash)
}

// GetSubRequests is a free data retrieval call binding the contract method 0x92b58bc2.
//
// Solidity: function getSubRequests(address addr) view returns((bytes32,address,bytes32,bytes32)[])
func (_SwarmMail *SwarmMailCaller) GetSubRequests(opts *bind.CallOpts, addr common.Address) ([]SwarmMailSubRequest, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getSubRequests", addr)

	if err != nil {
		return *new([]SwarmMailSubRequest), err
	}

	out0 := *abi.ConvertType(out[0], new([]SwarmMailSubRequest)).(*[]SwarmMailSubRequest)

	return out0, err

}

// GetSubRequests is a free data retrieval call binding the contract method 0x92b58bc2.
//
// Solidity: function getSubRequests(address addr) view returns((bytes32,address,bytes32,bytes32)[])
func (_SwarmMail *SwarmMailSession) GetSubRequests(addr common.Address) ([]SwarmMailSubRequest, error) {
	return _SwarmMail.Contract.GetSubRequests(&_SwarmMail.CallOpts, addr)
}

// GetSubRequests is a free data retrieval call binding the contract method 0x92b58bc2.
//
// Solidity: function getSubRequests(address addr) view returns((bytes32,address,bytes32,bytes32)[])
func (_SwarmMail *SwarmMailCallerSession) GetSubRequests(addr common.Address) ([]SwarmMailSubRequest, error) {
	return _SwarmMail.Contract.GetSubRequests(&_SwarmMail.CallOpts, addr)
}

// GetSubSubscribers is a free data retrieval call binding the contract method 0x7de2e5e8.
//
// Solidity: function getSubSubscribers(bytes32 subHash) view returns(address[])
func (_SwarmMail *SwarmMailCaller) GetSubSubscribers(opts *bind.CallOpts, subHash [32]byte) ([]common.Address, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getSubSubscribers", subHash)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetSubSubscribers is a free data retrieval call binding the contract method 0x7de2e5e8.
//
// Solidity: function getSubSubscribers(bytes32 subHash) view returns(address[])
func (_SwarmMail *SwarmMailSession) GetSubSubscribers(subHash [32]byte) ([]common.Address, error) {
	return _SwarmMail.Contract.GetSubSubscribers(&_SwarmMail.CallOpts, subHash)
}

// GetSubSubscribers is a free data retrieval call binding the contract method 0x7de2e5e8.
//
// Solidity: function getSubSubscribers(bytes32 subHash) view returns(address[])
func (_SwarmMail *SwarmMailCallerSession) GetSubSubscribers(subHash [32]byte) ([]common.Address, error) {
	return _SwarmMail.Contract.GetSubSubscribers(&_SwarmMail.CallOpts, subHash)
}

// GetSubs is a free data retrieval call binding the contract method 0xb8fb1bac.
//
// Solidity: function getSubs() view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32)[])
func (_SwarmMail *SwarmMailCaller) GetSubs(opts *bind.CallOpts) ([]SwarmMailSub, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "getSubs")

	if err != nil {
		return *new([]SwarmMailSub), err
	}

	out0 := *abi.ConvertType(out[0], new([]SwarmMailSub)).(*[]SwarmMailSub)

	return out0, err

}

// GetSubs is a free data retrieval call binding the contract method 0xb8fb1bac.
//
// Solidity: function getSubs() view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32)[])
func (_SwarmMail *SwarmMailSession) GetSubs() ([]SwarmMailSub, error) {
	return _SwarmMail.Contract.GetSubs(&_SwarmMail.CallOpts)
}

// GetSubs is a free data retrieval call binding the contract method 0xb8fb1bac.
//
// Solidity: function getSubs() view returns((bytes32,bytes32,address,bytes32,uint256,bool,uint256,uint32,uint32,uint32)[])
func (_SwarmMail *SwarmMailCallerSession) GetSubs() ([]SwarmMailSub, error) {
	return _SwarmMail.Contract.GetSubs(&_SwarmMail.CallOpts)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_SwarmMail *SwarmMailCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_SwarmMail *SwarmMailSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _SwarmMail.Contract.HasRole(&_SwarmMail.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_SwarmMail *SwarmMailCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _SwarmMail.Contract.HasRole(&_SwarmMail.CallOpts, role, account)
}

// InEscrow is a free data retrieval call binding the contract method 0xb7391341.
//
// Solidity: function inEscrow() view returns(uint256)
func (_SwarmMail *SwarmMailCaller) InEscrow(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "inEscrow")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// InEscrow is a free data retrieval call binding the contract method 0xb7391341.
//
// Solidity: function inEscrow() view returns(uint256)
func (_SwarmMail *SwarmMailSession) InEscrow() (*big.Int, error) {
	return _SwarmMail.Contract.InEscrow(&_SwarmMail.CallOpts)
}

// InEscrow is a free data retrieval call binding the contract method 0xb7391341.
//
// Solidity: function inEscrow() view returns(uint256)
func (_SwarmMail *SwarmMailCallerSession) InEscrow() (*big.Int, error) {
	return _SwarmMail.Contract.InEscrow(&_SwarmMail.CallOpts)
}

// MarketFee is a free data retrieval call binding the contract method 0x0ccf2156.
//
// Solidity: function marketFee() view returns(uint256)
func (_SwarmMail *SwarmMailCaller) MarketFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "marketFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MarketFee is a free data retrieval call binding the contract method 0x0ccf2156.
//
// Solidity: function marketFee() view returns(uint256)
func (_SwarmMail *SwarmMailSession) MarketFee() (*big.Int, error) {
	return _SwarmMail.Contract.MarketFee(&_SwarmMail.CallOpts)
}

// MarketFee is a free data retrieval call binding the contract method 0x0ccf2156.
//
// Solidity: function marketFee() view returns(uint256)
func (_SwarmMail *SwarmMailCallerSession) MarketFee() (*big.Int, error) {
	return _SwarmMail.Contract.MarketFee(&_SwarmMail.CallOpts)
}

// MinListingFee is a free data retrieval call binding the contract method 0x703a54b5.
//
// Solidity: function minListingFee() view returns(uint256)
func (_SwarmMail *SwarmMailCaller) MinListingFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "minListingFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinListingFee is a free data retrieval call binding the contract method 0x703a54b5.
//
// Solidity: function minListingFee() view returns(uint256)
func (_SwarmMail *SwarmMailSession) MinListingFee() (*big.Int, error) {
	return _SwarmMail.Contract.MinListingFee(&_SwarmMail.CallOpts)
}

// MinListingFee is a free data retrieval call binding the contract method 0x703a54b5.
//
// Solidity: function minListingFee() view returns(uint256)
func (_SwarmMail *SwarmMailCallerSession) MinListingFee() (*big.Int, error) {
	return _SwarmMail.Contract.MinListingFee(&_SwarmMail.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SwarmMail *SwarmMailCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SwarmMail *SwarmMailSession) Owner() (common.Address, error) {
	return _SwarmMail.Contract.Owner(&_SwarmMail.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SwarmMail *SwarmMailCallerSession) Owner() (common.Address, error) {
	return _SwarmMail.Contract.Owner(&_SwarmMail.CallOpts)
}

// SubscriptionIds is a free data retrieval call binding the contract method 0x0e499994.
//
// Solidity: function subscriptionIds(bytes32 ) view returns(uint256)
func (_SwarmMail *SwarmMailCaller) SubscriptionIds(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "subscriptionIds", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SubscriptionIds is a free data retrieval call binding the contract method 0x0e499994.
//
// Solidity: function subscriptionIds(bytes32 ) view returns(uint256)
func (_SwarmMail *SwarmMailSession) SubscriptionIds(arg0 [32]byte) (*big.Int, error) {
	return _SwarmMail.Contract.SubscriptionIds(&_SwarmMail.CallOpts, arg0)
}

// SubscriptionIds is a free data retrieval call binding the contract method 0x0e499994.
//
// Solidity: function subscriptionIds(bytes32 ) view returns(uint256)
func (_SwarmMail *SwarmMailCallerSession) SubscriptionIds(arg0 [32]byte) (*big.Int, error) {
	return _SwarmMail.Contract.SubscriptionIds(&_SwarmMail.CallOpts, arg0)
}

// Subscriptions is a free data retrieval call binding the contract method 0x2d5bbf60.
//
// Solidity: function subscriptions(uint256 ) view returns(bytes32 subHash, bytes32 fdpSellerNameHash, address seller, bytes32 swarmLocation, uint256 price, bool active, uint256 earned, uint32 bids, uint32 sells, uint32 reports)
func (_SwarmMail *SwarmMailCaller) Subscriptions(opts *bind.CallOpts, arg0 *big.Int) (struct {
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
}, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "subscriptions", arg0)

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

	return *outstruct, err

}

// Subscriptions is a free data retrieval call binding the contract method 0x2d5bbf60.
//
// Solidity: function subscriptions(uint256 ) view returns(bytes32 subHash, bytes32 fdpSellerNameHash, address seller, bytes32 swarmLocation, uint256 price, bool active, uint256 earned, uint32 bids, uint32 sells, uint32 reports)
func (_SwarmMail *SwarmMailSession) Subscriptions(arg0 *big.Int) (struct {
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
}, error) {
	return _SwarmMail.Contract.Subscriptions(&_SwarmMail.CallOpts, arg0)
}

// Subscriptions is a free data retrieval call binding the contract method 0x2d5bbf60.
//
// Solidity: function subscriptions(uint256 ) view returns(bytes32 subHash, bytes32 fdpSellerNameHash, address seller, bytes32 swarmLocation, uint256 price, bool active, uint256 earned, uint32 bids, uint32 sells, uint32 reports)
func (_SwarmMail *SwarmMailCallerSession) Subscriptions(arg0 *big.Int) (struct {
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
}, error) {
	return _SwarmMail.Contract.Subscriptions(&_SwarmMail.CallOpts, arg0)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_SwarmMail *SwarmMailCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _SwarmMail.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_SwarmMail *SwarmMailSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _SwarmMail.Contract.SupportsInterface(&_SwarmMail.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_SwarmMail *SwarmMailCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _SwarmMail.Contract.SupportsInterface(&_SwarmMail.CallOpts, interfaceId)
}

// BidSub is a paid mutator transaction binding the contract method 0xe91dbcb0.
//
// Solidity: function bidSub(bytes32 subHash, bytes32 fdpBuyerNameHash) payable returns()
func (_SwarmMail *SwarmMailTransactor) BidSub(opts *bind.TransactOpts, subHash [32]byte, fdpBuyerNameHash [32]byte) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "bidSub", subHash, fdpBuyerNameHash)
}

// BidSub is a paid mutator transaction binding the contract method 0xe91dbcb0.
//
// Solidity: function bidSub(bytes32 subHash, bytes32 fdpBuyerNameHash) payable returns()
func (_SwarmMail *SwarmMailSession) BidSub(subHash [32]byte, fdpBuyerNameHash [32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.BidSub(&_SwarmMail.TransactOpts, subHash, fdpBuyerNameHash)
}

// BidSub is a paid mutator transaction binding the contract method 0xe91dbcb0.
//
// Solidity: function bidSub(bytes32 subHash, bytes32 fdpBuyerNameHash) payable returns()
func (_SwarmMail *SwarmMailTransactorSession) BidSub(subHash [32]byte, fdpBuyerNameHash [32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.BidSub(&_SwarmMail.TransactOpts, subHash, fdpBuyerNameHash)
}

// EnableSub is a paid mutator transaction binding the contract method 0x88ac2917.
//
// Solidity: function enableSub(bytes32 subHash, bool active) returns()
func (_SwarmMail *SwarmMailTransactor) EnableSub(opts *bind.TransactOpts, subHash [32]byte, active bool) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "enableSub", subHash, active)
}

// EnableSub is a paid mutator transaction binding the contract method 0x88ac2917.
//
// Solidity: function enableSub(bytes32 subHash, bool active) returns()
func (_SwarmMail *SwarmMailSession) EnableSub(subHash [32]byte, active bool) (*types.Transaction, error) {
	return _SwarmMail.Contract.EnableSub(&_SwarmMail.TransactOpts, subHash, active)
}

// EnableSub is a paid mutator transaction binding the contract method 0x88ac2917.
//
// Solidity: function enableSub(bytes32 subHash, bool active) returns()
func (_SwarmMail *SwarmMailTransactorSession) EnableSub(subHash [32]byte, active bool) (*types.Transaction, error) {
	return _SwarmMail.Contract.EnableSub(&_SwarmMail.TransactOpts, subHash, active)
}

// FundsTransfer is a paid mutator transaction binding the contract method 0x567556a4.
//
// Solidity: function fundsTransfer() payable returns()
func (_SwarmMail *SwarmMailTransactor) FundsTransfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "fundsTransfer")
}

// FundsTransfer is a paid mutator transaction binding the contract method 0x567556a4.
//
// Solidity: function fundsTransfer() payable returns()
func (_SwarmMail *SwarmMailSession) FundsTransfer() (*types.Transaction, error) {
	return _SwarmMail.Contract.FundsTransfer(&_SwarmMail.TransactOpts)
}

// FundsTransfer is a paid mutator transaction binding the contract method 0x567556a4.
//
// Solidity: function fundsTransfer() payable returns()
func (_SwarmMail *SwarmMailTransactorSession) FundsTransfer() (*types.Transaction, error) {
	return _SwarmMail.Contract.FundsTransfer(&_SwarmMail.TransactOpts)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_SwarmMail *SwarmMailTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_SwarmMail *SwarmMailSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _SwarmMail.Contract.GrantRole(&_SwarmMail.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_SwarmMail *SwarmMailTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _SwarmMail.Contract.GrantRole(&_SwarmMail.TransactOpts, role, account)
}

// ListSub is a paid mutator transaction binding the contract method 0x1273b932.
//
// Solidity: function listSub(bytes32 fdpSellerNameHash, bytes32 dataSwarmLocation, uint256 price, bytes32 category, address podAddress) payable returns()
func (_SwarmMail *SwarmMailTransactor) ListSub(opts *bind.TransactOpts, fdpSellerNameHash [32]byte, dataSwarmLocation [32]byte, price *big.Int, category [32]byte, podAddress common.Address) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "listSub", fdpSellerNameHash, dataSwarmLocation, price, category, podAddress)
}

// ListSub is a paid mutator transaction binding the contract method 0x1273b932.
//
// Solidity: function listSub(bytes32 fdpSellerNameHash, bytes32 dataSwarmLocation, uint256 price, bytes32 category, address podAddress) payable returns()
func (_SwarmMail *SwarmMailSession) ListSub(fdpSellerNameHash [32]byte, dataSwarmLocation [32]byte, price *big.Int, category [32]byte, podAddress common.Address) (*types.Transaction, error) {
	return _SwarmMail.Contract.ListSub(&_SwarmMail.TransactOpts, fdpSellerNameHash, dataSwarmLocation, price, category, podAddress)
}

// ListSub is a paid mutator transaction binding the contract method 0x1273b932.
//
// Solidity: function listSub(bytes32 fdpSellerNameHash, bytes32 dataSwarmLocation, uint256 price, bytes32 category, address podAddress) payable returns()
func (_SwarmMail *SwarmMailTransactorSession) ListSub(fdpSellerNameHash [32]byte, dataSwarmLocation [32]byte, price *big.Int, category [32]byte, podAddress common.Address) (*types.Transaction, error) {
	return _SwarmMail.Contract.ListSub(&_SwarmMail.TransactOpts, fdpSellerNameHash, dataSwarmLocation, price, category, podAddress)
}

// Register is a paid mutator transaction binding the contract method 0x2f926732.
//
// Solidity: function register(bytes32 key, bytes32 smail) returns()
func (_SwarmMail *SwarmMailTransactor) Register(opts *bind.TransactOpts, key [32]byte, smail [32]byte) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "register", key, smail)
}

// Register is a paid mutator transaction binding the contract method 0x2f926732.
//
// Solidity: function register(bytes32 key, bytes32 smail) returns()
func (_SwarmMail *SwarmMailSession) Register(key [32]byte, smail [32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.Register(&_SwarmMail.TransactOpts, key, smail)
}

// Register is a paid mutator transaction binding the contract method 0x2f926732.
//
// Solidity: function register(bytes32 key, bytes32 smail) returns()
func (_SwarmMail *SwarmMailTransactorSession) Register(key [32]byte, smail [32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.Register(&_SwarmMail.TransactOpts, key, smail)
}

// RemoveEmails is a paid mutator transaction binding the contract method 0xb663ab5f.
//
// Solidity: function removeEmails(uint256 types, bytes32[] swarmLocations) returns()
func (_SwarmMail *SwarmMailTransactor) RemoveEmails(opts *bind.TransactOpts, types *big.Int, swarmLocations [][32]byte) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "removeEmails", types, swarmLocations)
}

// RemoveEmails is a paid mutator transaction binding the contract method 0xb663ab5f.
//
// Solidity: function removeEmails(uint256 types, bytes32[] swarmLocations) returns()
func (_SwarmMail *SwarmMailSession) RemoveEmails(types *big.Int, swarmLocations [][32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.RemoveEmails(&_SwarmMail.TransactOpts, types, swarmLocations)
}

// RemoveEmails is a paid mutator transaction binding the contract method 0xb663ab5f.
//
// Solidity: function removeEmails(uint256 types, bytes32[] swarmLocations) returns()
func (_SwarmMail *SwarmMailTransactorSession) RemoveEmails(types *big.Int, swarmLocations [][32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.RemoveEmails(&_SwarmMail.TransactOpts, types, swarmLocations)
}

// RemoveInboxEmail is a paid mutator transaction binding the contract method 0xc34ba5f6.
//
// Solidity: function removeInboxEmail(bytes32 swarmLocation) returns()
func (_SwarmMail *SwarmMailTransactor) RemoveInboxEmail(opts *bind.TransactOpts, swarmLocation [32]byte) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "removeInboxEmail", swarmLocation)
}

// RemoveInboxEmail is a paid mutator transaction binding the contract method 0xc34ba5f6.
//
// Solidity: function removeInboxEmail(bytes32 swarmLocation) returns()
func (_SwarmMail *SwarmMailSession) RemoveInboxEmail(swarmLocation [32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.RemoveInboxEmail(&_SwarmMail.TransactOpts, swarmLocation)
}

// RemoveInboxEmail is a paid mutator transaction binding the contract method 0xc34ba5f6.
//
// Solidity: function removeInboxEmail(bytes32 swarmLocation) returns()
func (_SwarmMail *SwarmMailTransactorSession) RemoveInboxEmail(swarmLocation [32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.RemoveInboxEmail(&_SwarmMail.TransactOpts, swarmLocation)
}

// RemoveSentEmail is a paid mutator transaction binding the contract method 0xc9bdc1c5.
//
// Solidity: function removeSentEmail(bytes32 swarmLocation) returns()
func (_SwarmMail *SwarmMailTransactor) RemoveSentEmail(opts *bind.TransactOpts, swarmLocation [32]byte) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "removeSentEmail", swarmLocation)
}

// RemoveSentEmail is a paid mutator transaction binding the contract method 0xc9bdc1c5.
//
// Solidity: function removeSentEmail(bytes32 swarmLocation) returns()
func (_SwarmMail *SwarmMailSession) RemoveSentEmail(swarmLocation [32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.RemoveSentEmail(&_SwarmMail.TransactOpts, swarmLocation)
}

// RemoveSentEmail is a paid mutator transaction binding the contract method 0xc9bdc1c5.
//
// Solidity: function removeSentEmail(bytes32 swarmLocation) returns()
func (_SwarmMail *SwarmMailTransactorSession) RemoveSentEmail(swarmLocation [32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.RemoveSentEmail(&_SwarmMail.TransactOpts, swarmLocation)
}

// RemoveSubItem is a paid mutator transaction binding the contract method 0x9673a9e9.
//
// Solidity: function removeSubItem(uint256 index) returns()
func (_SwarmMail *SwarmMailTransactor) RemoveSubItem(opts *bind.TransactOpts, index *big.Int) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "removeSubItem", index)
}

// RemoveSubItem is a paid mutator transaction binding the contract method 0x9673a9e9.
//
// Solidity: function removeSubItem(uint256 index) returns()
func (_SwarmMail *SwarmMailSession) RemoveSubItem(index *big.Int) (*types.Transaction, error) {
	return _SwarmMail.Contract.RemoveSubItem(&_SwarmMail.TransactOpts, index)
}

// RemoveSubItem is a paid mutator transaction binding the contract method 0x9673a9e9.
//
// Solidity: function removeSubItem(uint256 index) returns()
func (_SwarmMail *SwarmMailTransactorSession) RemoveSubItem(index *big.Int) (*types.Transaction, error) {
	return _SwarmMail.Contract.RemoveSubItem(&_SwarmMail.TransactOpts, index)
}

// RemoveUserActiveBid is a paid mutator transaction binding the contract method 0x0260f912.
//
// Solidity: function removeUserActiveBid(bytes32 requestHash) returns()
func (_SwarmMail *SwarmMailTransactor) RemoveUserActiveBid(opts *bind.TransactOpts, requestHash [32]byte) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "removeUserActiveBid", requestHash)
}

// RemoveUserActiveBid is a paid mutator transaction binding the contract method 0x0260f912.
//
// Solidity: function removeUserActiveBid(bytes32 requestHash) returns()
func (_SwarmMail *SwarmMailSession) RemoveUserActiveBid(requestHash [32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.RemoveUserActiveBid(&_SwarmMail.TransactOpts, requestHash)
}

// RemoveUserActiveBid is a paid mutator transaction binding the contract method 0x0260f912.
//
// Solidity: function removeUserActiveBid(bytes32 requestHash) returns()
func (_SwarmMail *SwarmMailTransactorSession) RemoveUserActiveBid(requestHash [32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.RemoveUserActiveBid(&_SwarmMail.TransactOpts, requestHash)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SwarmMail *SwarmMailTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SwarmMail *SwarmMailSession) RenounceOwnership() (*types.Transaction, error) {
	return _SwarmMail.Contract.RenounceOwnership(&_SwarmMail.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SwarmMail *SwarmMailTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _SwarmMail.Contract.RenounceOwnership(&_SwarmMail.TransactOpts)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_SwarmMail *SwarmMailTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "renounceRole", role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_SwarmMail *SwarmMailSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _SwarmMail.Contract.RenounceRole(&_SwarmMail.TransactOpts, role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_SwarmMail *SwarmMailTransactorSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _SwarmMail.Contract.RenounceRole(&_SwarmMail.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_SwarmMail *SwarmMailTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_SwarmMail *SwarmMailSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _SwarmMail.Contract.RevokeRole(&_SwarmMail.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_SwarmMail *SwarmMailTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _SwarmMail.Contract.RevokeRole(&_SwarmMail.TransactOpts, role, account)
}

// SellSub is a paid mutator transaction binding the contract method 0x3ca684e3.
//
// Solidity: function sellSub(bytes32 requestHash, bytes32 encryptedKeyLocation) payable returns()
func (_SwarmMail *SwarmMailTransactor) SellSub(opts *bind.TransactOpts, requestHash [32]byte, encryptedKeyLocation [32]byte) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "sellSub", requestHash, encryptedKeyLocation)
}

// SellSub is a paid mutator transaction binding the contract method 0x3ca684e3.
//
// Solidity: function sellSub(bytes32 requestHash, bytes32 encryptedKeyLocation) payable returns()
func (_SwarmMail *SwarmMailSession) SellSub(requestHash [32]byte, encryptedKeyLocation [32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.SellSub(&_SwarmMail.TransactOpts, requestHash, encryptedKeyLocation)
}

// SellSub is a paid mutator transaction binding the contract method 0x3ca684e3.
//
// Solidity: function sellSub(bytes32 requestHash, bytes32 encryptedKeyLocation) payable returns()
func (_SwarmMail *SwarmMailTransactorSession) SellSub(requestHash [32]byte, encryptedKeyLocation [32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.SellSub(&_SwarmMail.TransactOpts, requestHash, encryptedKeyLocation)
}

// SendEmail is a paid mutator transaction binding the contract method 0xd2465fab.
//
// Solidity: function sendEmail(address toAddress, bool isEncryption, bytes32 swarmLocation) payable returns()
func (_SwarmMail *SwarmMailTransactor) SendEmail(opts *bind.TransactOpts, toAddress common.Address, isEncryption bool, swarmLocation [32]byte) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "sendEmail", toAddress, isEncryption, swarmLocation)
}

// SendEmail is a paid mutator transaction binding the contract method 0xd2465fab.
//
// Solidity: function sendEmail(address toAddress, bool isEncryption, bytes32 swarmLocation) payable returns()
func (_SwarmMail *SwarmMailSession) SendEmail(toAddress common.Address, isEncryption bool, swarmLocation [32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.SendEmail(&_SwarmMail.TransactOpts, toAddress, isEncryption, swarmLocation)
}

// SendEmail is a paid mutator transaction binding the contract method 0xd2465fab.
//
// Solidity: function sendEmail(address toAddress, bool isEncryption, bytes32 swarmLocation) payable returns()
func (_SwarmMail *SwarmMailTransactorSession) SendEmail(toAddress common.Address, isEncryption bool, swarmLocation [32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.SendEmail(&_SwarmMail.TransactOpts, toAddress, isEncryption, swarmLocation)
}

// SetFee is a paid mutator transaction binding the contract method 0x69fe0e2d.
//
// Solidity: function setFee(uint256 newFee) returns()
func (_SwarmMail *SwarmMailTransactor) SetFee(opts *bind.TransactOpts, newFee *big.Int) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "setFee", newFee)
}

// SetFee is a paid mutator transaction binding the contract method 0x69fe0e2d.
//
// Solidity: function setFee(uint256 newFee) returns()
func (_SwarmMail *SwarmMailSession) SetFee(newFee *big.Int) (*types.Transaction, error) {
	return _SwarmMail.Contract.SetFee(&_SwarmMail.TransactOpts, newFee)
}

// SetFee is a paid mutator transaction binding the contract method 0x69fe0e2d.
//
// Solidity: function setFee(uint256 newFee) returns()
func (_SwarmMail *SwarmMailTransactorSession) SetFee(newFee *big.Int) (*types.Transaction, error) {
	return _SwarmMail.Contract.SetFee(&_SwarmMail.TransactOpts, newFee)
}

// SetListingFee is a paid mutator transaction binding the contract method 0x131dbd09.
//
// Solidity: function setListingFee(uint256 newListingFee) returns()
func (_SwarmMail *SwarmMailTransactor) SetListingFee(opts *bind.TransactOpts, newListingFee *big.Int) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "setListingFee", newListingFee)
}

// SetListingFee is a paid mutator transaction binding the contract method 0x131dbd09.
//
// Solidity: function setListingFee(uint256 newListingFee) returns()
func (_SwarmMail *SwarmMailSession) SetListingFee(newListingFee *big.Int) (*types.Transaction, error) {
	return _SwarmMail.Contract.SetListingFee(&_SwarmMail.TransactOpts, newListingFee)
}

// SetListingFee is a paid mutator transaction binding the contract method 0x131dbd09.
//
// Solidity: function setListingFee(uint256 newListingFee) returns()
func (_SwarmMail *SwarmMailTransactorSession) SetListingFee(newListingFee *big.Int) (*types.Transaction, error) {
	return _SwarmMail.Contract.SetListingFee(&_SwarmMail.TransactOpts, newListingFee)
}

// SignEmail is a paid mutator transaction binding the contract method 0x87134952.
//
// Solidity: function signEmail(bytes32 swarmLocation) returns()
func (_SwarmMail *SwarmMailTransactor) SignEmail(opts *bind.TransactOpts, swarmLocation [32]byte) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "signEmail", swarmLocation)
}

// SignEmail is a paid mutator transaction binding the contract method 0x87134952.
//
// Solidity: function signEmail(bytes32 swarmLocation) returns()
func (_SwarmMail *SwarmMailSession) SignEmail(swarmLocation [32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.SignEmail(&_SwarmMail.TransactOpts, swarmLocation)
}

// SignEmail is a paid mutator transaction binding the contract method 0x87134952.
//
// Solidity: function signEmail(bytes32 swarmLocation) returns()
func (_SwarmMail *SwarmMailTransactorSession) SignEmail(swarmLocation [32]byte) (*types.Transaction, error) {
	return _SwarmMail.Contract.SignEmail(&_SwarmMail.TransactOpts, swarmLocation)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SwarmMail *SwarmMailTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _SwarmMail.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SwarmMail *SwarmMailSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SwarmMail.Contract.TransferOwnership(&_SwarmMail.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SwarmMail *SwarmMailTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SwarmMail.Contract.TransferOwnership(&_SwarmMail.TransactOpts, newOwner)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_SwarmMail *SwarmMailTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SwarmMail.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_SwarmMail *SwarmMailSession) Receive() (*types.Transaction, error) {
	return _SwarmMail.Contract.Receive(&_SwarmMail.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_SwarmMail *SwarmMailTransactorSession) Receive() (*types.Transaction, error) {
	return _SwarmMail.Contract.Receive(&_SwarmMail.TransactOpts)
}

// SwarmMailOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the SwarmMail contract.
type SwarmMailOwnershipTransferredIterator struct {
	Event *SwarmMailOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *SwarmMailOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwarmMailOwnershipTransferred)
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
		it.Event = new(SwarmMailOwnershipTransferred)
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
func (it *SwarmMailOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwarmMailOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwarmMailOwnershipTransferred represents a OwnershipTransferred event raised by the SwarmMail contract.
type SwarmMailOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SwarmMail *SwarmMailFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*SwarmMailOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SwarmMail.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SwarmMailOwnershipTransferredIterator{contract: _SwarmMail.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SwarmMail *SwarmMailFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *SwarmMailOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SwarmMail.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwarmMailOwnershipTransferred)
				if err := _SwarmMail.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_SwarmMail *SwarmMailFilterer) ParseOwnershipTransferred(log types.Log) (*SwarmMailOwnershipTransferred, error) {
	event := new(SwarmMailOwnershipTransferred)
	if err := _SwarmMail.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwarmMailRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the SwarmMail contract.
type SwarmMailRoleAdminChangedIterator struct {
	Event *SwarmMailRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *SwarmMailRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwarmMailRoleAdminChanged)
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
		it.Event = new(SwarmMailRoleAdminChanged)
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
func (it *SwarmMailRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwarmMailRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwarmMailRoleAdminChanged represents a RoleAdminChanged event raised by the SwarmMail contract.
type SwarmMailRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_SwarmMail *SwarmMailFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*SwarmMailRoleAdminChangedIterator, error) {

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

	logs, sub, err := _SwarmMail.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &SwarmMailRoleAdminChangedIterator{contract: _SwarmMail.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_SwarmMail *SwarmMailFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *SwarmMailRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

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

	logs, sub, err := _SwarmMail.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwarmMailRoleAdminChanged)
				if err := _SwarmMail.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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
func (_SwarmMail *SwarmMailFilterer) ParseRoleAdminChanged(log types.Log) (*SwarmMailRoleAdminChanged, error) {
	event := new(SwarmMailRoleAdminChanged)
	if err := _SwarmMail.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwarmMailRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the SwarmMail contract.
type SwarmMailRoleGrantedIterator struct {
	Event *SwarmMailRoleGranted // Event containing the contract specifics and raw log

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
func (it *SwarmMailRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwarmMailRoleGranted)
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
		it.Event = new(SwarmMailRoleGranted)
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
func (it *SwarmMailRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwarmMailRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwarmMailRoleGranted represents a RoleGranted event raised by the SwarmMail contract.
type SwarmMailRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_SwarmMail *SwarmMailFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*SwarmMailRoleGrantedIterator, error) {

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

	logs, sub, err := _SwarmMail.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &SwarmMailRoleGrantedIterator{contract: _SwarmMail.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_SwarmMail *SwarmMailFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *SwarmMailRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _SwarmMail.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwarmMailRoleGranted)
				if err := _SwarmMail.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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
func (_SwarmMail *SwarmMailFilterer) ParseRoleGranted(log types.Log) (*SwarmMailRoleGranted, error) {
	event := new(SwarmMailRoleGranted)
	if err := _SwarmMail.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SwarmMailRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the SwarmMail contract.
type SwarmMailRoleRevokedIterator struct {
	Event *SwarmMailRoleRevoked // Event containing the contract specifics and raw log

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
func (it *SwarmMailRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwarmMailRoleRevoked)
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
		it.Event = new(SwarmMailRoleRevoked)
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
func (it *SwarmMailRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwarmMailRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwarmMailRoleRevoked represents a RoleRevoked event raised by the SwarmMail contract.
type SwarmMailRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_SwarmMail *SwarmMailFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*SwarmMailRoleRevokedIterator, error) {

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

	logs, sub, err := _SwarmMail.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &SwarmMailRoleRevokedIterator{contract: _SwarmMail.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_SwarmMail *SwarmMailFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *SwarmMailRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _SwarmMail.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwarmMailRoleRevoked)
				if err := _SwarmMail.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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
func (_SwarmMail *SwarmMailFilterer) ParseRoleRevoked(log types.Log) (*SwarmMailRoleRevoked, error) {
	event := new(SwarmMailRoleRevoked)
	if err := _SwarmMail.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
