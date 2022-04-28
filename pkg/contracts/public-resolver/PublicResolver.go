// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package publicresolver

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

// PublicresolverMetaData contains all meta data concerning the Publicresolver contract.
var PublicresolverMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractENS\",\"name\":\"_ensAddr\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"contentType\",\"type\":\"uint256\"}],\"name\":\"ABIChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"a\",\"type\":\"address\"}],\"name\":\"AddrChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"ContentChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"hash\",\"type\":\"bytes\"}],\"name\":\"MultihashChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"}],\"name\":\"NameChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"x\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"y\",\"type\":\"bytes32\"}],\"name\":\"PubkeyChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"indexedKey\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"}],\"name\":\"TextChanged\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_node\",\"type\":\"bytes32\"}],\"name\":\"addr\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_node\",\"type\":\"bytes32\"}],\"name\":\"content\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_node\",\"type\":\"bytes32\"}],\"name\":\"getAll\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"_addr\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"_content\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"_multihash\",\"type\":\"bytes\"},{\"internalType\":\"bytes32\",\"name\":\"_x\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_y\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"_name\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_node\",\"type\":\"bytes32\"}],\"name\":\"multihash\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_node\",\"type\":\"bytes32\"}],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_node\",\"type\":\"bytes32\"}],\"name\":\"pubkey\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"x\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"y\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_node\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"_contentType\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"setABI\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_node\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_addr\",\"type\":\"address\"}],\"name\":\"setAddr\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_node\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_addr\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"_content\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"_multihash\",\"type\":\"bytes\"},{\"internalType\":\"bytes32\",\"name\":\"_x\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_y\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"_name\",\"type\":\"string\"}],\"name\":\"setAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_node\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_hash\",\"type\":\"bytes32\"}],\"name\":\"setContent\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_node\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"_hash\",\"type\":\"bytes\"}],\"name\":\"setMultihash\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_node\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"_name\",\"type\":\"string\"}],\"name\":\"setName\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_node\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_x\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_y\",\"type\":\"bytes32\"}],\"name\":\"setPubkey\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_node\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"_key\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_value\",\"type\":\"string\"}],\"name\":\"setText\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_node\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"_key\",\"type\":\"string\"}],\"name\":\"text\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// PublicresolverABI is the input ABI used to generate the binding from.
// Deprecated: Use PublicresolverMetaData.ABI instead.
var PublicresolverABI = PublicresolverMetaData.ABI

// Publicresolver is an auto generated Go binding around an Ethereum contract.
type Publicresolver struct {
	PublicresolverCaller     // Read-only binding to the contract
	PublicresolverTransactor // Write-only binding to the contract
	PublicresolverFilterer   // Log filterer for contract events
}

// PublicresolverCaller is an auto generated read-only Go binding around an Ethereum contract.
type PublicresolverCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PublicresolverTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PublicresolverTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PublicresolverFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PublicresolverFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PublicresolverSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PublicresolverSession struct {
	Contract     *Publicresolver   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PublicresolverCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PublicresolverCallerSession struct {
	Contract *PublicresolverCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// PublicresolverTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PublicresolverTransactorSession struct {
	Contract     *PublicresolverTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// PublicresolverRaw is an auto generated low-level Go binding around an Ethereum contract.
type PublicresolverRaw struct {
	Contract *Publicresolver // Generic contract binding to access the raw methods on
}

// PublicresolverCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PublicresolverCallerRaw struct {
	Contract *PublicresolverCaller // Generic read-only contract binding to access the raw methods on
}

// PublicresolverTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PublicresolverTransactorRaw struct {
	Contract *PublicresolverTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPublicresolver creates a new instance of Publicresolver, bound to a specific deployed contract.
func NewPublicresolver(address common.Address, backend bind.ContractBackend) (*Publicresolver, error) {
	contract, err := bindPublicresolver(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Publicresolver{PublicresolverCaller: PublicresolverCaller{contract: contract}, PublicresolverTransactor: PublicresolverTransactor{contract: contract}, PublicresolverFilterer: PublicresolverFilterer{contract: contract}}, nil
}

// NewPublicresolverCaller creates a new read-only instance of Publicresolver, bound to a specific deployed contract.
func NewPublicresolverCaller(address common.Address, caller bind.ContractCaller) (*PublicresolverCaller, error) {
	contract, err := bindPublicresolver(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PublicresolverCaller{contract: contract}, nil
}

// NewPublicresolverTransactor creates a new write-only instance of Publicresolver, bound to a specific deployed contract.
func NewPublicresolverTransactor(address common.Address, transactor bind.ContractTransactor) (*PublicresolverTransactor, error) {
	contract, err := bindPublicresolver(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PublicresolverTransactor{contract: contract}, nil
}

// NewPublicresolverFilterer creates a new log filterer instance of Publicresolver, bound to a specific deployed contract.
func NewPublicresolverFilterer(address common.Address, filterer bind.ContractFilterer) (*PublicresolverFilterer, error) {
	contract, err := bindPublicresolver(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PublicresolverFilterer{contract: contract}, nil
}

// bindPublicresolver binds a generic wrapper to an already deployed contract.
func bindPublicresolver(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PublicresolverABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Publicresolver *PublicresolverRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Publicresolver.Contract.PublicresolverCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Publicresolver *PublicresolverRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Publicresolver.Contract.PublicresolverTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Publicresolver *PublicresolverRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Publicresolver.Contract.PublicresolverTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Publicresolver *PublicresolverCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Publicresolver.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Publicresolver *PublicresolverTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Publicresolver.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Publicresolver *PublicresolverTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Publicresolver.Contract.contract.Transact(opts, method, params...)
}

// Addr is a free data retrieval call binding the contract method 0x3b3b57de.
//
// Solidity: function addr(bytes32 _node) view returns(address)
func (_Publicresolver *PublicresolverCaller) Addr(opts *bind.CallOpts, _node [32]byte) (common.Address, error) {
	var out []interface{}
	err := _Publicresolver.contract.Call(opts, &out, "addr", _node)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Addr is a free data retrieval call binding the contract method 0x3b3b57de.
//
// Solidity: function addr(bytes32 _node) view returns(address)
func (_Publicresolver *PublicresolverSession) Addr(_node [32]byte) (common.Address, error) {
	return _Publicresolver.Contract.Addr(&_Publicresolver.CallOpts, _node)
}

// Addr is a free data retrieval call binding the contract method 0x3b3b57de.
//
// Solidity: function addr(bytes32 _node) view returns(address)
func (_Publicresolver *PublicresolverCallerSession) Addr(_node [32]byte) (common.Address, error) {
	return _Publicresolver.Contract.Addr(&_Publicresolver.CallOpts, _node)
}

// Content is a free data retrieval call binding the contract method 0x2dff6941.
//
// Solidity: function content(bytes32 _node) view returns(bytes32)
func (_Publicresolver *PublicresolverCaller) Content(opts *bind.CallOpts, _node [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Publicresolver.contract.Call(opts, &out, "content", _node)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// Content is a free data retrieval call binding the contract method 0x2dff6941.
//
// Solidity: function content(bytes32 _node) view returns(bytes32)
func (_Publicresolver *PublicresolverSession) Content(_node [32]byte) ([32]byte, error) {
	return _Publicresolver.Contract.Content(&_Publicresolver.CallOpts, _node)
}

// Content is a free data retrieval call binding the contract method 0x2dff6941.
//
// Solidity: function content(bytes32 _node) view returns(bytes32)
func (_Publicresolver *PublicresolverCallerSession) Content(_node [32]byte) ([32]byte, error) {
	return _Publicresolver.Contract.Content(&_Publicresolver.CallOpts, _node)
}

// GetAll is a free data retrieval call binding the contract method 0xed80e1f7.
//
// Solidity: function getAll(bytes32 _node) view returns(address _addr, bytes32 _content, bytes _multihash, bytes32 _x, bytes32 _y, string _name)
func (_Publicresolver *PublicresolverCaller) GetAll(opts *bind.CallOpts, _node [32]byte) (struct {
	Addr      common.Address
	Content   [32]byte
	Multihash []byte
	X         [32]byte
	Y         [32]byte
	Name      string
}, error) {
	var out []interface{}
	err := _Publicresolver.contract.Call(opts, &out, "getAll", _node)

	outstruct := new(struct {
		Addr      common.Address
		Content   [32]byte
		Multihash []byte
		X         [32]byte
		Y         [32]byte
		Name      string
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Addr = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Content = *abi.ConvertType(out[1], new([32]byte)).(*[32]byte)
	outstruct.Multihash = *abi.ConvertType(out[2], new([]byte)).(*[]byte)
	outstruct.X = *abi.ConvertType(out[3], new([32]byte)).(*[32]byte)
	outstruct.Y = *abi.ConvertType(out[4], new([32]byte)).(*[32]byte)
	outstruct.Name = *abi.ConvertType(out[5], new(string)).(*string)

	return *outstruct, err

}

// GetAll is a free data retrieval call binding the contract method 0xed80e1f7.
//
// Solidity: function getAll(bytes32 _node) view returns(address _addr, bytes32 _content, bytes _multihash, bytes32 _x, bytes32 _y, string _name)
func (_Publicresolver *PublicresolverSession) GetAll(_node [32]byte) (struct {
	Addr      common.Address
	Content   [32]byte
	Multihash []byte
	X         [32]byte
	Y         [32]byte
	Name      string
}, error) {
	return _Publicresolver.Contract.GetAll(&_Publicresolver.CallOpts, _node)
}

// GetAll is a free data retrieval call binding the contract method 0xed80e1f7.
//
// Solidity: function getAll(bytes32 _node) view returns(address _addr, bytes32 _content, bytes _multihash, bytes32 _x, bytes32 _y, string _name)
func (_Publicresolver *PublicresolverCallerSession) GetAll(_node [32]byte) (struct {
	Addr      common.Address
	Content   [32]byte
	Multihash []byte
	X         [32]byte
	Y         [32]byte
	Name      string
}, error) {
	return _Publicresolver.Contract.GetAll(&_Publicresolver.CallOpts, _node)
}

// Multihash is a free data retrieval call binding the contract method 0xe89401a1.
//
// Solidity: function multihash(bytes32 _node) view returns(bytes)
func (_Publicresolver *PublicresolverCaller) Multihash(opts *bind.CallOpts, _node [32]byte) ([]byte, error) {
	var out []interface{}
	err := _Publicresolver.contract.Call(opts, &out, "multihash", _node)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// Multihash is a free data retrieval call binding the contract method 0xe89401a1.
//
// Solidity: function multihash(bytes32 _node) view returns(bytes)
func (_Publicresolver *PublicresolverSession) Multihash(_node [32]byte) ([]byte, error) {
	return _Publicresolver.Contract.Multihash(&_Publicresolver.CallOpts, _node)
}

// Multihash is a free data retrieval call binding the contract method 0xe89401a1.
//
// Solidity: function multihash(bytes32 _node) view returns(bytes)
func (_Publicresolver *PublicresolverCallerSession) Multihash(_node [32]byte) ([]byte, error) {
	return _Publicresolver.Contract.Multihash(&_Publicresolver.CallOpts, _node)
}

// Name is a free data retrieval call binding the contract method 0x691f3431.
//
// Solidity: function name(bytes32 _node) view returns(string)
func (_Publicresolver *PublicresolverCaller) Name(opts *bind.CallOpts, _node [32]byte) (string, error) {
	var out []interface{}
	err := _Publicresolver.contract.Call(opts, &out, "name", _node)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x691f3431.
//
// Solidity: function name(bytes32 _node) view returns(string)
func (_Publicresolver *PublicresolverSession) Name(_node [32]byte) (string, error) {
	return _Publicresolver.Contract.Name(&_Publicresolver.CallOpts, _node)
}

// Name is a free data retrieval call binding the contract method 0x691f3431.
//
// Solidity: function name(bytes32 _node) view returns(string)
func (_Publicresolver *PublicresolverCallerSession) Name(_node [32]byte) (string, error) {
	return _Publicresolver.Contract.Name(&_Publicresolver.CallOpts, _node)
}

// Pubkey is a free data retrieval call binding the contract method 0xc8690233.
//
// Solidity: function pubkey(bytes32 _node) view returns(bytes32 x, bytes32 y)
func (_Publicresolver *PublicresolverCaller) Pubkey(opts *bind.CallOpts, _node [32]byte) (struct {
	X [32]byte
	Y [32]byte
}, error) {
	var out []interface{}
	err := _Publicresolver.contract.Call(opts, &out, "pubkey", _node)

	outstruct := new(struct {
		X [32]byte
		Y [32]byte
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.X = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.Y = *abi.ConvertType(out[1], new([32]byte)).(*[32]byte)

	return *outstruct, err

}

// Pubkey is a free data retrieval call binding the contract method 0xc8690233.
//
// Solidity: function pubkey(bytes32 _node) view returns(bytes32 x, bytes32 y)
func (_Publicresolver *PublicresolverSession) Pubkey(_node [32]byte) (struct {
	X [32]byte
	Y [32]byte
}, error) {
	return _Publicresolver.Contract.Pubkey(&_Publicresolver.CallOpts, _node)
}

// Pubkey is a free data retrieval call binding the contract method 0xc8690233.
//
// Solidity: function pubkey(bytes32 _node) view returns(bytes32 x, bytes32 y)
func (_Publicresolver *PublicresolverCallerSession) Pubkey(_node [32]byte) (struct {
	X [32]byte
	Y [32]byte
}, error) {
	return _Publicresolver.Contract.Pubkey(&_Publicresolver.CallOpts, _node)
}

// Text is a free data retrieval call binding the contract method 0x59d1d43c.
//
// Solidity: function text(bytes32 _node, string _key) view returns(string)
func (_Publicresolver *PublicresolverCaller) Text(opts *bind.CallOpts, _node [32]byte, _key string) (string, error) {
	var out []interface{}
	err := _Publicresolver.contract.Call(opts, &out, "text", _node, _key)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Text is a free data retrieval call binding the contract method 0x59d1d43c.
//
// Solidity: function text(bytes32 _node, string _key) view returns(string)
func (_Publicresolver *PublicresolverSession) Text(_node [32]byte, _key string) (string, error) {
	return _Publicresolver.Contract.Text(&_Publicresolver.CallOpts, _node, _key)
}

// Text is a free data retrieval call binding the contract method 0x59d1d43c.
//
// Solidity: function text(bytes32 _node, string _key) view returns(string)
func (_Publicresolver *PublicresolverCallerSession) Text(_node [32]byte, _key string) (string, error) {
	return _Publicresolver.Contract.Text(&_Publicresolver.CallOpts, _node, _key)
}

// SetABI is a paid mutator transaction binding the contract method 0x623195b0.
//
// Solidity: function setABI(bytes32 _node, uint256 _contentType, bytes _data) returns()
func (_Publicresolver *PublicresolverTransactor) SetABI(opts *bind.TransactOpts, _node [32]byte, _contentType *big.Int, _data []byte) (*types.Transaction, error) {
	return _Publicresolver.contract.Transact(opts, "setABI", _node, _contentType, _data)
}

// SetABI is a paid mutator transaction binding the contract method 0x623195b0.
//
// Solidity: function setABI(bytes32 _node, uint256 _contentType, bytes _data) returns()
func (_Publicresolver *PublicresolverSession) SetABI(_node [32]byte, _contentType *big.Int, _data []byte) (*types.Transaction, error) {
	return _Publicresolver.Contract.SetABI(&_Publicresolver.TransactOpts, _node, _contentType, _data)
}

// SetABI is a paid mutator transaction binding the contract method 0x623195b0.
//
// Solidity: function setABI(bytes32 _node, uint256 _contentType, bytes _data) returns()
func (_Publicresolver *PublicresolverTransactorSession) SetABI(_node [32]byte, _contentType *big.Int, _data []byte) (*types.Transaction, error) {
	return _Publicresolver.Contract.SetABI(&_Publicresolver.TransactOpts, _node, _contentType, _data)
}

// SetAddr is a paid mutator transaction binding the contract method 0xd5fa2b00.
//
// Solidity: function setAddr(bytes32 _node, address _addr) returns()
func (_Publicresolver *PublicresolverTransactor) SetAddr(opts *bind.TransactOpts, _node [32]byte, _addr common.Address) (*types.Transaction, error) {
	return _Publicresolver.contract.Transact(opts, "setAddr", _node, _addr)
}

// SetAddr is a paid mutator transaction binding the contract method 0xd5fa2b00.
//
// Solidity: function setAddr(bytes32 _node, address _addr) returns()
func (_Publicresolver *PublicresolverSession) SetAddr(_node [32]byte, _addr common.Address) (*types.Transaction, error) {
	return _Publicresolver.Contract.SetAddr(&_Publicresolver.TransactOpts, _node, _addr)
}

// SetAddr is a paid mutator transaction binding the contract method 0xd5fa2b00.
//
// Solidity: function setAddr(bytes32 _node, address _addr) returns()
func (_Publicresolver *PublicresolverTransactorSession) SetAddr(_node [32]byte, _addr common.Address) (*types.Transaction, error) {
	return _Publicresolver.Contract.SetAddr(&_Publicresolver.TransactOpts, _node, _addr)
}

// SetAll is a paid mutator transaction binding the contract method 0x9f3a206d.
//
// Solidity: function setAll(bytes32 _node, address _addr, bytes32 _content, bytes _multihash, bytes32 _x, bytes32 _y, string _name) returns()
func (_Publicresolver *PublicresolverTransactor) SetAll(opts *bind.TransactOpts, _node [32]byte, _addr common.Address, _content [32]byte, _multihash []byte, _x [32]byte, _y [32]byte, _name string) (*types.Transaction, error) {
	return _Publicresolver.contract.Transact(opts, "setAll", _node, _addr, _content, _multihash, _x, _y, _name)
}

// SetAll is a paid mutator transaction binding the contract method 0x9f3a206d.
//
// Solidity: function setAll(bytes32 _node, address _addr, bytes32 _content, bytes _multihash, bytes32 _x, bytes32 _y, string _name) returns()
func (_Publicresolver *PublicresolverSession) SetAll(_node [32]byte, _addr common.Address, _content [32]byte, _multihash []byte, _x [32]byte, _y [32]byte, _name string) (*types.Transaction, error) {
	return _Publicresolver.Contract.SetAll(&_Publicresolver.TransactOpts, _node, _addr, _content, _multihash, _x, _y, _name)
}

// SetAll is a paid mutator transaction binding the contract method 0x9f3a206d.
//
// Solidity: function setAll(bytes32 _node, address _addr, bytes32 _content, bytes _multihash, bytes32 _x, bytes32 _y, string _name) returns()
func (_Publicresolver *PublicresolverTransactorSession) SetAll(_node [32]byte, _addr common.Address, _content [32]byte, _multihash []byte, _x [32]byte, _y [32]byte, _name string) (*types.Transaction, error) {
	return _Publicresolver.Contract.SetAll(&_Publicresolver.TransactOpts, _node, _addr, _content, _multihash, _x, _y, _name)
}

// SetContent is a paid mutator transaction binding the contract method 0xc3d014d6.
//
// Solidity: function setContent(bytes32 _node, bytes32 _hash) returns()
func (_Publicresolver *PublicresolverTransactor) SetContent(opts *bind.TransactOpts, _node [32]byte, _hash [32]byte) (*types.Transaction, error) {
	return _Publicresolver.contract.Transact(opts, "setContent", _node, _hash)
}

// SetContent is a paid mutator transaction binding the contract method 0xc3d014d6.
//
// Solidity: function setContent(bytes32 _node, bytes32 _hash) returns()
func (_Publicresolver *PublicresolverSession) SetContent(_node [32]byte, _hash [32]byte) (*types.Transaction, error) {
	return _Publicresolver.Contract.SetContent(&_Publicresolver.TransactOpts, _node, _hash)
}

// SetContent is a paid mutator transaction binding the contract method 0xc3d014d6.
//
// Solidity: function setContent(bytes32 _node, bytes32 _hash) returns()
func (_Publicresolver *PublicresolverTransactorSession) SetContent(_node [32]byte, _hash [32]byte) (*types.Transaction, error) {
	return _Publicresolver.Contract.SetContent(&_Publicresolver.TransactOpts, _node, _hash)
}

// SetMultihash is a paid mutator transaction binding the contract method 0xaa4cb547.
//
// Solidity: function setMultihash(bytes32 _node, bytes _hash) returns()
func (_Publicresolver *PublicresolverTransactor) SetMultihash(opts *bind.TransactOpts, _node [32]byte, _hash []byte) (*types.Transaction, error) {
	return _Publicresolver.contract.Transact(opts, "setMultihash", _node, _hash)
}

// SetMultihash is a paid mutator transaction binding the contract method 0xaa4cb547.
//
// Solidity: function setMultihash(bytes32 _node, bytes _hash) returns()
func (_Publicresolver *PublicresolverSession) SetMultihash(_node [32]byte, _hash []byte) (*types.Transaction, error) {
	return _Publicresolver.Contract.SetMultihash(&_Publicresolver.TransactOpts, _node, _hash)
}

// SetMultihash is a paid mutator transaction binding the contract method 0xaa4cb547.
//
// Solidity: function setMultihash(bytes32 _node, bytes _hash) returns()
func (_Publicresolver *PublicresolverTransactorSession) SetMultihash(_node [32]byte, _hash []byte) (*types.Transaction, error) {
	return _Publicresolver.Contract.SetMultihash(&_Publicresolver.TransactOpts, _node, _hash)
}

// SetName is a paid mutator transaction binding the contract method 0x77372213.
//
// Solidity: function setName(bytes32 _node, string _name) returns()
func (_Publicresolver *PublicresolverTransactor) SetName(opts *bind.TransactOpts, _node [32]byte, _name string) (*types.Transaction, error) {
	return _Publicresolver.contract.Transact(opts, "setName", _node, _name)
}

// SetName is a paid mutator transaction binding the contract method 0x77372213.
//
// Solidity: function setName(bytes32 _node, string _name) returns()
func (_Publicresolver *PublicresolverSession) SetName(_node [32]byte, _name string) (*types.Transaction, error) {
	return _Publicresolver.Contract.SetName(&_Publicresolver.TransactOpts, _node, _name)
}

// SetName is a paid mutator transaction binding the contract method 0x77372213.
//
// Solidity: function setName(bytes32 _node, string _name) returns()
func (_Publicresolver *PublicresolverTransactorSession) SetName(_node [32]byte, _name string) (*types.Transaction, error) {
	return _Publicresolver.Contract.SetName(&_Publicresolver.TransactOpts, _node, _name)
}

// SetPubkey is a paid mutator transaction binding the contract method 0x29cd62ea.
//
// Solidity: function setPubkey(bytes32 _node, bytes32 _x, bytes32 _y) returns()
func (_Publicresolver *PublicresolverTransactor) SetPubkey(opts *bind.TransactOpts, _node [32]byte, _x [32]byte, _y [32]byte) (*types.Transaction, error) {
	return _Publicresolver.contract.Transact(opts, "setPubkey", _node, _x, _y)
}

// SetPubkey is a paid mutator transaction binding the contract method 0x29cd62ea.
//
// Solidity: function setPubkey(bytes32 _node, bytes32 _x, bytes32 _y) returns()
func (_Publicresolver *PublicresolverSession) SetPubkey(_node [32]byte, _x [32]byte, _y [32]byte) (*types.Transaction, error) {
	return _Publicresolver.Contract.SetPubkey(&_Publicresolver.TransactOpts, _node, _x, _y)
}

// SetPubkey is a paid mutator transaction binding the contract method 0x29cd62ea.
//
// Solidity: function setPubkey(bytes32 _node, bytes32 _x, bytes32 _y) returns()
func (_Publicresolver *PublicresolverTransactorSession) SetPubkey(_node [32]byte, _x [32]byte, _y [32]byte) (*types.Transaction, error) {
	return _Publicresolver.Contract.SetPubkey(&_Publicresolver.TransactOpts, _node, _x, _y)
}

// SetText is a paid mutator transaction binding the contract method 0x10f13a8c.
//
// Solidity: function setText(bytes32 _node, string _key, string _value) returns()
func (_Publicresolver *PublicresolverTransactor) SetText(opts *bind.TransactOpts, _node [32]byte, _key string, _value string) (*types.Transaction, error) {
	return _Publicresolver.contract.Transact(opts, "setText", _node, _key, _value)
}

// SetText is a paid mutator transaction binding the contract method 0x10f13a8c.
//
// Solidity: function setText(bytes32 _node, string _key, string _value) returns()
func (_Publicresolver *PublicresolverSession) SetText(_node [32]byte, _key string, _value string) (*types.Transaction, error) {
	return _Publicresolver.Contract.SetText(&_Publicresolver.TransactOpts, _node, _key, _value)
}

// SetText is a paid mutator transaction binding the contract method 0x10f13a8c.
//
// Solidity: function setText(bytes32 _node, string _key, string _value) returns()
func (_Publicresolver *PublicresolverTransactorSession) SetText(_node [32]byte, _key string, _value string) (*types.Transaction, error) {
	return _Publicresolver.Contract.SetText(&_Publicresolver.TransactOpts, _node, _key, _value)
}

// PublicresolverABIChangedIterator is returned from FilterABIChanged and is used to iterate over the raw logs and unpacked data for ABIChanged events raised by the Publicresolver contract.
type PublicresolverABIChangedIterator struct {
	Event *PublicresolverABIChanged // Event containing the contract specifics and raw log

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
func (it *PublicresolverABIChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PublicresolverABIChanged)
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
		it.Event = new(PublicresolverABIChanged)
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
func (it *PublicresolverABIChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PublicresolverABIChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PublicresolverABIChanged represents a ABIChanged event raised by the Publicresolver contract.
type PublicresolverABIChanged struct {
	Node        [32]byte
	ContentType *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterABIChanged is a free log retrieval operation binding the contract event 0xaa121bbeef5f32f5961a2a28966e769023910fc9479059ee3495d4c1a696efe3.
//
// Solidity: event ABIChanged(bytes32 indexed node, uint256 indexed contentType)
func (_Publicresolver *PublicresolverFilterer) FilterABIChanged(opts *bind.FilterOpts, node [][32]byte, contentType []*big.Int) (*PublicresolverABIChangedIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}
	var contentTypeRule []interface{}
	for _, contentTypeItem := range contentType {
		contentTypeRule = append(contentTypeRule, contentTypeItem)
	}

	logs, sub, err := _Publicresolver.contract.FilterLogs(opts, "ABIChanged", nodeRule, contentTypeRule)
	if err != nil {
		return nil, err
	}
	return &PublicresolverABIChangedIterator{contract: _Publicresolver.contract, event: "ABIChanged", logs: logs, sub: sub}, nil
}

// WatchABIChanged is a free log subscription operation binding the contract event 0xaa121bbeef5f32f5961a2a28966e769023910fc9479059ee3495d4c1a696efe3.
//
// Solidity: event ABIChanged(bytes32 indexed node, uint256 indexed contentType)
func (_Publicresolver *PublicresolverFilterer) WatchABIChanged(opts *bind.WatchOpts, sink chan<- *PublicresolverABIChanged, node [][32]byte, contentType []*big.Int) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}
	var contentTypeRule []interface{}
	for _, contentTypeItem := range contentType {
		contentTypeRule = append(contentTypeRule, contentTypeItem)
	}

	logs, sub, err := _Publicresolver.contract.WatchLogs(opts, "ABIChanged", nodeRule, contentTypeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PublicresolverABIChanged)
				if err := _Publicresolver.contract.UnpackLog(event, "ABIChanged", log); err != nil {
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

// ParseABIChanged is a log parse operation binding the contract event 0xaa121bbeef5f32f5961a2a28966e769023910fc9479059ee3495d4c1a696efe3.
//
// Solidity: event ABIChanged(bytes32 indexed node, uint256 indexed contentType)
func (_Publicresolver *PublicresolverFilterer) ParseABIChanged(log types.Log) (*PublicresolverABIChanged, error) {
	event := new(PublicresolverABIChanged)
	if err := _Publicresolver.contract.UnpackLog(event, "ABIChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PublicresolverAddrChangedIterator is returned from FilterAddrChanged and is used to iterate over the raw logs and unpacked data for AddrChanged events raised by the Publicresolver contract.
type PublicresolverAddrChangedIterator struct {
	Event *PublicresolverAddrChanged // Event containing the contract specifics and raw log

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
func (it *PublicresolverAddrChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PublicresolverAddrChanged)
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
		it.Event = new(PublicresolverAddrChanged)
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
func (it *PublicresolverAddrChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PublicresolverAddrChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PublicresolverAddrChanged represents a AddrChanged event raised by the Publicresolver contract.
type PublicresolverAddrChanged struct {
	Node [32]byte
	A    common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAddrChanged is a free log retrieval operation binding the contract event 0x52d7d861f09ab3d26239d492e8968629f95e9e318cf0b73bfddc441522a15fd2.
//
// Solidity: event AddrChanged(bytes32 indexed node, address a)
func (_Publicresolver *PublicresolverFilterer) FilterAddrChanged(opts *bind.FilterOpts, node [][32]byte) (*PublicresolverAddrChangedIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _Publicresolver.contract.FilterLogs(opts, "AddrChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return &PublicresolverAddrChangedIterator{contract: _Publicresolver.contract, event: "AddrChanged", logs: logs, sub: sub}, nil
}

// WatchAddrChanged is a free log subscription operation binding the contract event 0x52d7d861f09ab3d26239d492e8968629f95e9e318cf0b73bfddc441522a15fd2.
//
// Solidity: event AddrChanged(bytes32 indexed node, address a)
func (_Publicresolver *PublicresolverFilterer) WatchAddrChanged(opts *bind.WatchOpts, sink chan<- *PublicresolverAddrChanged, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _Publicresolver.contract.WatchLogs(opts, "AddrChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PublicresolverAddrChanged)
				if err := _Publicresolver.contract.UnpackLog(event, "AddrChanged", log); err != nil {
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

// ParseAddrChanged is a log parse operation binding the contract event 0x52d7d861f09ab3d26239d492e8968629f95e9e318cf0b73bfddc441522a15fd2.
//
// Solidity: event AddrChanged(bytes32 indexed node, address a)
func (_Publicresolver *PublicresolverFilterer) ParseAddrChanged(log types.Log) (*PublicresolverAddrChanged, error) {
	event := new(PublicresolverAddrChanged)
	if err := _Publicresolver.contract.UnpackLog(event, "AddrChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PublicresolverContentChangedIterator is returned from FilterContentChanged and is used to iterate over the raw logs and unpacked data for ContentChanged events raised by the Publicresolver contract.
type PublicresolverContentChangedIterator struct {
	Event *PublicresolverContentChanged // Event containing the contract specifics and raw log

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
func (it *PublicresolverContentChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PublicresolverContentChanged)
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
		it.Event = new(PublicresolverContentChanged)
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
func (it *PublicresolverContentChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PublicresolverContentChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PublicresolverContentChanged represents a ContentChanged event raised by the Publicresolver contract.
type PublicresolverContentChanged struct {
	Node [32]byte
	Hash [32]byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterContentChanged is a free log retrieval operation binding the contract event 0x0424b6fe0d9c3bdbece0e7879dc241bb0c22e900be8b6c168b4ee08bd9bf83bc.
//
// Solidity: event ContentChanged(bytes32 indexed node, bytes32 hash)
func (_Publicresolver *PublicresolverFilterer) FilterContentChanged(opts *bind.FilterOpts, node [][32]byte) (*PublicresolverContentChangedIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _Publicresolver.contract.FilterLogs(opts, "ContentChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return &PublicresolverContentChangedIterator{contract: _Publicresolver.contract, event: "ContentChanged", logs: logs, sub: sub}, nil
}

// WatchContentChanged is a free log subscription operation binding the contract event 0x0424b6fe0d9c3bdbece0e7879dc241bb0c22e900be8b6c168b4ee08bd9bf83bc.
//
// Solidity: event ContentChanged(bytes32 indexed node, bytes32 hash)
func (_Publicresolver *PublicresolverFilterer) WatchContentChanged(opts *bind.WatchOpts, sink chan<- *PublicresolverContentChanged, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _Publicresolver.contract.WatchLogs(opts, "ContentChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PublicresolverContentChanged)
				if err := _Publicresolver.contract.UnpackLog(event, "ContentChanged", log); err != nil {
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

// ParseContentChanged is a log parse operation binding the contract event 0x0424b6fe0d9c3bdbece0e7879dc241bb0c22e900be8b6c168b4ee08bd9bf83bc.
//
// Solidity: event ContentChanged(bytes32 indexed node, bytes32 hash)
func (_Publicresolver *PublicresolverFilterer) ParseContentChanged(log types.Log) (*PublicresolverContentChanged, error) {
	event := new(PublicresolverContentChanged)
	if err := _Publicresolver.contract.UnpackLog(event, "ContentChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PublicresolverMultihashChangedIterator is returned from FilterMultihashChanged and is used to iterate over the raw logs and unpacked data for MultihashChanged events raised by the Publicresolver contract.
type PublicresolverMultihashChangedIterator struct {
	Event *PublicresolverMultihashChanged // Event containing the contract specifics and raw log

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
func (it *PublicresolverMultihashChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PublicresolverMultihashChanged)
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
		it.Event = new(PublicresolverMultihashChanged)
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
func (it *PublicresolverMultihashChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PublicresolverMultihashChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PublicresolverMultihashChanged represents a MultihashChanged event raised by the Publicresolver contract.
type PublicresolverMultihashChanged struct {
	Node [32]byte
	Hash []byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterMultihashChanged is a free log retrieval operation binding the contract event 0xc0b0fc07269fc2749adada3221c095a1d2187b2d075b51c915857b520f3a5021.
//
// Solidity: event MultihashChanged(bytes32 indexed node, bytes hash)
func (_Publicresolver *PublicresolverFilterer) FilterMultihashChanged(opts *bind.FilterOpts, node [][32]byte) (*PublicresolverMultihashChangedIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _Publicresolver.contract.FilterLogs(opts, "MultihashChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return &PublicresolverMultihashChangedIterator{contract: _Publicresolver.contract, event: "MultihashChanged", logs: logs, sub: sub}, nil
}

// WatchMultihashChanged is a free log subscription operation binding the contract event 0xc0b0fc07269fc2749adada3221c095a1d2187b2d075b51c915857b520f3a5021.
//
// Solidity: event MultihashChanged(bytes32 indexed node, bytes hash)
func (_Publicresolver *PublicresolverFilterer) WatchMultihashChanged(opts *bind.WatchOpts, sink chan<- *PublicresolverMultihashChanged, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _Publicresolver.contract.WatchLogs(opts, "MultihashChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PublicresolverMultihashChanged)
				if err := _Publicresolver.contract.UnpackLog(event, "MultihashChanged", log); err != nil {
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

// ParseMultihashChanged is a log parse operation binding the contract event 0xc0b0fc07269fc2749adada3221c095a1d2187b2d075b51c915857b520f3a5021.
//
// Solidity: event MultihashChanged(bytes32 indexed node, bytes hash)
func (_Publicresolver *PublicresolverFilterer) ParseMultihashChanged(log types.Log) (*PublicresolverMultihashChanged, error) {
	event := new(PublicresolverMultihashChanged)
	if err := _Publicresolver.contract.UnpackLog(event, "MultihashChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PublicresolverNameChangedIterator is returned from FilterNameChanged and is used to iterate over the raw logs and unpacked data for NameChanged events raised by the Publicresolver contract.
type PublicresolverNameChangedIterator struct {
	Event *PublicresolverNameChanged // Event containing the contract specifics and raw log

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
func (it *PublicresolverNameChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PublicresolverNameChanged)
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
		it.Event = new(PublicresolverNameChanged)
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
func (it *PublicresolverNameChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PublicresolverNameChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PublicresolverNameChanged represents a NameChanged event raised by the Publicresolver contract.
type PublicresolverNameChanged struct {
	Node [32]byte
	Name string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterNameChanged is a free log retrieval operation binding the contract event 0xb7d29e911041e8d9b843369e890bcb72c9388692ba48b65ac54e7214c4c348f7.
//
// Solidity: event NameChanged(bytes32 indexed node, string name)
func (_Publicresolver *PublicresolverFilterer) FilterNameChanged(opts *bind.FilterOpts, node [][32]byte) (*PublicresolverNameChangedIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _Publicresolver.contract.FilterLogs(opts, "NameChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return &PublicresolverNameChangedIterator{contract: _Publicresolver.contract, event: "NameChanged", logs: logs, sub: sub}, nil
}

// WatchNameChanged is a free log subscription operation binding the contract event 0xb7d29e911041e8d9b843369e890bcb72c9388692ba48b65ac54e7214c4c348f7.
//
// Solidity: event NameChanged(bytes32 indexed node, string name)
func (_Publicresolver *PublicresolverFilterer) WatchNameChanged(opts *bind.WatchOpts, sink chan<- *PublicresolverNameChanged, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _Publicresolver.contract.WatchLogs(opts, "NameChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PublicresolverNameChanged)
				if err := _Publicresolver.contract.UnpackLog(event, "NameChanged", log); err != nil {
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

// ParseNameChanged is a log parse operation binding the contract event 0xb7d29e911041e8d9b843369e890bcb72c9388692ba48b65ac54e7214c4c348f7.
//
// Solidity: event NameChanged(bytes32 indexed node, string name)
func (_Publicresolver *PublicresolverFilterer) ParseNameChanged(log types.Log) (*PublicresolverNameChanged, error) {
	event := new(PublicresolverNameChanged)
	if err := _Publicresolver.contract.UnpackLog(event, "NameChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PublicresolverPubkeyChangedIterator is returned from FilterPubkeyChanged and is used to iterate over the raw logs and unpacked data for PubkeyChanged events raised by the Publicresolver contract.
type PublicresolverPubkeyChangedIterator struct {
	Event *PublicresolverPubkeyChanged // Event containing the contract specifics and raw log

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
func (it *PublicresolverPubkeyChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PublicresolverPubkeyChanged)
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
		it.Event = new(PublicresolverPubkeyChanged)
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
func (it *PublicresolverPubkeyChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PublicresolverPubkeyChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PublicresolverPubkeyChanged represents a PubkeyChanged event raised by the Publicresolver contract.
type PublicresolverPubkeyChanged struct {
	Node [32]byte
	X    [32]byte
	Y    [32]byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterPubkeyChanged is a free log retrieval operation binding the contract event 0x1d6f5e03d3f63eb58751986629a5439baee5079ff04f345becb66e23eb154e46.
//
// Solidity: event PubkeyChanged(bytes32 indexed node, bytes32 x, bytes32 y)
func (_Publicresolver *PublicresolverFilterer) FilterPubkeyChanged(opts *bind.FilterOpts, node [][32]byte) (*PublicresolverPubkeyChangedIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _Publicresolver.contract.FilterLogs(opts, "PubkeyChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return &PublicresolverPubkeyChangedIterator{contract: _Publicresolver.contract, event: "PubkeyChanged", logs: logs, sub: sub}, nil
}

// WatchPubkeyChanged is a free log subscription operation binding the contract event 0x1d6f5e03d3f63eb58751986629a5439baee5079ff04f345becb66e23eb154e46.
//
// Solidity: event PubkeyChanged(bytes32 indexed node, bytes32 x, bytes32 y)
func (_Publicresolver *PublicresolverFilterer) WatchPubkeyChanged(opts *bind.WatchOpts, sink chan<- *PublicresolverPubkeyChanged, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _Publicresolver.contract.WatchLogs(opts, "PubkeyChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PublicresolverPubkeyChanged)
				if err := _Publicresolver.contract.UnpackLog(event, "PubkeyChanged", log); err != nil {
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

// ParsePubkeyChanged is a log parse operation binding the contract event 0x1d6f5e03d3f63eb58751986629a5439baee5079ff04f345becb66e23eb154e46.
//
// Solidity: event PubkeyChanged(bytes32 indexed node, bytes32 x, bytes32 y)
func (_Publicresolver *PublicresolverFilterer) ParsePubkeyChanged(log types.Log) (*PublicresolverPubkeyChanged, error) {
	event := new(PublicresolverPubkeyChanged)
	if err := _Publicresolver.contract.UnpackLog(event, "PubkeyChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PublicresolverTextChangedIterator is returned from FilterTextChanged and is used to iterate over the raw logs and unpacked data for TextChanged events raised by the Publicresolver contract.
type PublicresolverTextChangedIterator struct {
	Event *PublicresolverTextChanged // Event containing the contract specifics and raw log

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
func (it *PublicresolverTextChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PublicresolverTextChanged)
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
		it.Event = new(PublicresolverTextChanged)
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
func (it *PublicresolverTextChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PublicresolverTextChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PublicresolverTextChanged represents a TextChanged event raised by the Publicresolver contract.
type PublicresolverTextChanged struct {
	Node       [32]byte
	IndexedKey string
	Key        string
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterTextChanged is a free log retrieval operation binding the contract event 0xd8c9334b1a9c2f9da342a0a2b32629c1a229b6445dad78947f674b44444a7550.
//
// Solidity: event TextChanged(bytes32 indexed node, string indexedKey, string key)
func (_Publicresolver *PublicresolverFilterer) FilterTextChanged(opts *bind.FilterOpts, node [][32]byte) (*PublicresolverTextChangedIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _Publicresolver.contract.FilterLogs(opts, "TextChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return &PublicresolverTextChangedIterator{contract: _Publicresolver.contract, event: "TextChanged", logs: logs, sub: sub}, nil
}

// WatchTextChanged is a free log subscription operation binding the contract event 0xd8c9334b1a9c2f9da342a0a2b32629c1a229b6445dad78947f674b44444a7550.
//
// Solidity: event TextChanged(bytes32 indexed node, string indexedKey, string key)
func (_Publicresolver *PublicresolverFilterer) WatchTextChanged(opts *bind.WatchOpts, sink chan<- *PublicresolverTextChanged, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _Publicresolver.contract.WatchLogs(opts, "TextChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PublicresolverTextChanged)
				if err := _Publicresolver.contract.UnpackLog(event, "TextChanged", log); err != nil {
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

// ParseTextChanged is a log parse operation binding the contract event 0xd8c9334b1a9c2f9da342a0a2b32629c1a229b6445dad78947f674b44444a7550.
//
// Solidity: event TextChanged(bytes32 indexed node, string indexedKey, string key)
func (_Publicresolver *PublicresolverFilterer) ParseTextChanged(log types.Log) (*PublicresolverTextChanged, error) {
	event := new(PublicresolverTextChanged)
	if err := _Publicresolver.contract.UnpackLog(event, "TextChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
