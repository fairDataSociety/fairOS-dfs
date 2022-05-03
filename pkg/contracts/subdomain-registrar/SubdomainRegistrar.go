// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package subdomainregistrar

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

// SubdomainregistrarMetaData contains all meta data concerning the Subdomainregistrar contract.
var SubdomainregistrarMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractENS\",\"name\":\"_ensAddr\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"_node\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"label\",\"type\":\"bytes32\"}],\"name\":\"Log\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"ens\",\"outputs\":[{\"internalType\":\"contractENS\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"expiryTimes\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_label\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"rootNode\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// SubdomainregistrarABI is the input ABI used to generate the binding from.
// Deprecated: Use SubdomainregistrarMetaData.ABI instead.
var SubdomainregistrarABI = SubdomainregistrarMetaData.ABI

// Subdomainregistrar is an auto generated Go binding around an Ethereum contract.
type Subdomainregistrar struct {
	SubdomainregistrarCaller     // Read-only binding to the contract
	SubdomainregistrarTransactor // Write-only binding to the contract
	SubdomainregistrarFilterer   // Log filterer for contract events
}

// SubdomainregistrarCaller is an auto generated read-only Go binding around an Ethereum contract.
type SubdomainregistrarCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubdomainregistrarTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SubdomainregistrarTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubdomainregistrarFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SubdomainregistrarFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubdomainregistrarSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SubdomainregistrarSession struct {
	Contract     *Subdomainregistrar // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// SubdomainregistrarCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SubdomainregistrarCallerSession struct {
	Contract *SubdomainregistrarCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// SubdomainregistrarTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SubdomainregistrarTransactorSession struct {
	Contract     *SubdomainregistrarTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// SubdomainregistrarRaw is an auto generated low-level Go binding around an Ethereum contract.
type SubdomainregistrarRaw struct {
	Contract *Subdomainregistrar // Generic contract binding to access the raw methods on
}

// SubdomainregistrarCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SubdomainregistrarCallerRaw struct {
	Contract *SubdomainregistrarCaller // Generic read-only contract binding to access the raw methods on
}

// SubdomainregistrarTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SubdomainregistrarTransactorRaw struct {
	Contract *SubdomainregistrarTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSubdomainregistrar creates a new instance of Subdomainregistrar, bound to a specific deployed contract.
func NewSubdomainregistrar(address common.Address, backend bind.ContractBackend) (*Subdomainregistrar, error) {
	contract, err := bindSubdomainregistrar(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Subdomainregistrar{SubdomainregistrarCaller: SubdomainregistrarCaller{contract: contract}, SubdomainregistrarTransactor: SubdomainregistrarTransactor{contract: contract}, SubdomainregistrarFilterer: SubdomainregistrarFilterer{contract: contract}}, nil
}

// NewSubdomainregistrarCaller creates a new read-only instance of Subdomainregistrar, bound to a specific deployed contract.
func NewSubdomainregistrarCaller(address common.Address, caller bind.ContractCaller) (*SubdomainregistrarCaller, error) {
	contract, err := bindSubdomainregistrar(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SubdomainregistrarCaller{contract: contract}, nil
}

// NewSubdomainregistrarTransactor creates a new write-only instance of Subdomainregistrar, bound to a specific deployed contract.
func NewSubdomainregistrarTransactor(address common.Address, transactor bind.ContractTransactor) (*SubdomainregistrarTransactor, error) {
	contract, err := bindSubdomainregistrar(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SubdomainregistrarTransactor{contract: contract}, nil
}

// NewSubdomainregistrarFilterer creates a new log filterer instance of Subdomainregistrar, bound to a specific deployed contract.
func NewSubdomainregistrarFilterer(address common.Address, filterer bind.ContractFilterer) (*SubdomainregistrarFilterer, error) {
	contract, err := bindSubdomainregistrar(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SubdomainregistrarFilterer{contract: contract}, nil
}

// bindSubdomainregistrar binds a generic wrapper to an already deployed contract.
func bindSubdomainregistrar(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SubdomainregistrarABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Subdomainregistrar *SubdomainregistrarRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Subdomainregistrar.Contract.SubdomainregistrarCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Subdomainregistrar *SubdomainregistrarRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Subdomainregistrar.Contract.SubdomainregistrarTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Subdomainregistrar *SubdomainregistrarRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Subdomainregistrar.Contract.SubdomainregistrarTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Subdomainregistrar *SubdomainregistrarCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Subdomainregistrar.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Subdomainregistrar *SubdomainregistrarTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Subdomainregistrar.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Subdomainregistrar *SubdomainregistrarTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Subdomainregistrar.Contract.contract.Transact(opts, method, params...)
}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_Subdomainregistrar *SubdomainregistrarCaller) Ens(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Subdomainregistrar.contract.Call(opts, &out, "ens")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_Subdomainregistrar *SubdomainregistrarSession) Ens() (common.Address, error) {
	return _Subdomainregistrar.Contract.Ens(&_Subdomainregistrar.CallOpts)
}

// Ens is a free data retrieval call binding the contract method 0x3f15457f.
//
// Solidity: function ens() view returns(address)
func (_Subdomainregistrar *SubdomainregistrarCallerSession) Ens() (common.Address, error) {
	return _Subdomainregistrar.Contract.Ens(&_Subdomainregistrar.CallOpts)
}

// ExpiryTimes is a free data retrieval call binding the contract method 0xaf9f26e4.
//
// Solidity: function expiryTimes(bytes32 ) view returns(uint256)
func (_Subdomainregistrar *SubdomainregistrarCaller) ExpiryTimes(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Subdomainregistrar.contract.Call(opts, &out, "expiryTimes", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ExpiryTimes is a free data retrieval call binding the contract method 0xaf9f26e4.
//
// Solidity: function expiryTimes(bytes32 ) view returns(uint256)
func (_Subdomainregistrar *SubdomainregistrarSession) ExpiryTimes(arg0 [32]byte) (*big.Int, error) {
	return _Subdomainregistrar.Contract.ExpiryTimes(&_Subdomainregistrar.CallOpts, arg0)
}

// ExpiryTimes is a free data retrieval call binding the contract method 0xaf9f26e4.
//
// Solidity: function expiryTimes(bytes32 ) view returns(uint256)
func (_Subdomainregistrar *SubdomainregistrarCallerSession) ExpiryTimes(arg0 [32]byte) (*big.Int, error) {
	return _Subdomainregistrar.Contract.ExpiryTimes(&_Subdomainregistrar.CallOpts, arg0)
}

// RootNode is a free data retrieval call binding the contract method 0xfaff50a8.
//
// Solidity: function rootNode() view returns(bytes32)
func (_Subdomainregistrar *SubdomainregistrarCaller) RootNode(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Subdomainregistrar.contract.Call(opts, &out, "rootNode")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// RootNode is a free data retrieval call binding the contract method 0xfaff50a8.
//
// Solidity: function rootNode() view returns(bytes32)
func (_Subdomainregistrar *SubdomainregistrarSession) RootNode() ([32]byte, error) {
	return _Subdomainregistrar.Contract.RootNode(&_Subdomainregistrar.CallOpts)
}

// RootNode is a free data retrieval call binding the contract method 0xfaff50a8.
//
// Solidity: function rootNode() view returns(bytes32)
func (_Subdomainregistrar *SubdomainregistrarCallerSession) RootNode() ([32]byte, error) {
	return _Subdomainregistrar.Contract.RootNode(&_Subdomainregistrar.CallOpts)
}

// Register is a paid mutator transaction binding the contract method 0xd22057a9.
//
// Solidity: function register(bytes32 _label, address _owner) returns()
func (_Subdomainregistrar *SubdomainregistrarTransactor) Register(opts *bind.TransactOpts, _label [32]byte, _owner common.Address) (*types.Transaction, error) {
	return _Subdomainregistrar.contract.Transact(opts, "register", _label, _owner)
}

// Register is a paid mutator transaction binding the contract method 0xd22057a9.
//
// Solidity: function register(bytes32 _label, address _owner) returns()
func (_Subdomainregistrar *SubdomainregistrarSession) Register(_label [32]byte, _owner common.Address) (*types.Transaction, error) {
	return _Subdomainregistrar.Contract.Register(&_Subdomainregistrar.TransactOpts, _label, _owner)
}

// Register is a paid mutator transaction binding the contract method 0xd22057a9.
//
// Solidity: function register(bytes32 _label, address _owner) returns()
func (_Subdomainregistrar *SubdomainregistrarTransactorSession) Register(_label [32]byte, _owner common.Address) (*types.Transaction, error) {
	return _Subdomainregistrar.Contract.Register(&_Subdomainregistrar.TransactOpts, _label, _owner)
}

// SubdomainregistrarLogIterator is returned from FilterLog and is used to iterate over the raw logs and unpacked data for Log events raised by the Subdomainregistrar contract.
type SubdomainregistrarLogIterator struct {
	Event *SubdomainregistrarLog // Event containing the contract specifics and raw log

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
func (it *SubdomainregistrarLogIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubdomainregistrarLog)
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
		it.Event = new(SubdomainregistrarLog)
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
func (it *SubdomainregistrarLogIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubdomainregistrarLogIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubdomainregistrarLog represents a Log event raised by the Subdomainregistrar contract.
type SubdomainregistrarLog struct {
	Owner common.Address
	Label [32]byte
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterLog is a free log retrieval operation binding the contract event 0x22c227465829cb03b9a4f59749925805a01244e82c68e891e42751abb94e6298.
//
// Solidity: event Log(address owner, bytes32 label)
func (_Subdomainregistrar *SubdomainregistrarFilterer) FilterLog(opts *bind.FilterOpts) (*SubdomainregistrarLogIterator, error) {

	logs, sub, err := _Subdomainregistrar.contract.FilterLogs(opts, "Log")
	if err != nil {
		return nil, err
	}
	return &SubdomainregistrarLogIterator{contract: _Subdomainregistrar.contract, event: "Log", logs: logs, sub: sub}, nil
}

// WatchLog is a free log subscription operation binding the contract event 0x22c227465829cb03b9a4f59749925805a01244e82c68e891e42751abb94e6298.
//
// Solidity: event Log(address owner, bytes32 label)
func (_Subdomainregistrar *SubdomainregistrarFilterer) WatchLog(opts *bind.WatchOpts, sink chan<- *SubdomainregistrarLog) (event.Subscription, error) {

	logs, sub, err := _Subdomainregistrar.contract.WatchLogs(opts, "Log")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubdomainregistrarLog)
				if err := _Subdomainregistrar.contract.UnpackLog(event, "Log", log); err != nil {
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

// ParseLog is a log parse operation binding the contract event 0x22c227465829cb03b9a4f59749925805a01244e82c68e891e42751abb94e6298.
//
// Solidity: event Log(address owner, bytes32 label)
func (_Subdomainregistrar *SubdomainregistrarFilterer) ParseLog(log types.Log) (*SubdomainregistrarLog, error) {
	event := new(SubdomainregistrarLog)
	if err := _Subdomainregistrar.contract.UnpackLog(event, "Log", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
