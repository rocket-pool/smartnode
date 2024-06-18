// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

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
	_ = abi.ConvertType
)

// RocketSignerRegistryMetaData contains all meta data concerning the RocketSignerRegistry contract.
var RocketSignerRegistryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"clearSigningDelegate\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"nodeToSigner\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"setSigningDelegate\",\"inputs\":[{\"name\":\"_signer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_v\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"_r\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"_s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"signerToNode\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"SigningDelegateSet\",\"inputs\":[{\"name\":\"nodeAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"signerAddress\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"StringsInsufficientHexLength\",\"inputs\":[{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}]",
}

// RocketSignerRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use RocketSignerRegistryMetaData.ABI instead.
var RocketSignerRegistryABI = RocketSignerRegistryMetaData.ABI

// RocketSignerRegistry is an auto generated Go binding around an Ethereum contract.
type RocketSignerRegistry struct {
	RocketSignerRegistryCaller     // Read-only binding to the contract
	RocketSignerRegistryTransactor // Write-only binding to the contract
	RocketSignerRegistryFilterer   // Log filterer for contract events
}

// RocketSignerRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type RocketSignerRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RocketSignerRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type RocketSignerRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RocketSignerRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RocketSignerRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RocketSignerRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RocketSignerRegistrySession struct {
	Contract     *RocketSignerRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts         // Call options to use throughout this session
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// RocketSignerRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RocketSignerRegistryCallerSession struct {
	Contract *RocketSignerRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts               // Call options to use throughout this session
}

// RocketSignerRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RocketSignerRegistryTransactorSession struct {
	Contract     *RocketSignerRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// RocketSignerRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type RocketSignerRegistryRaw struct {
	Contract *RocketSignerRegistry // Generic contract binding to access the raw methods on
}

// RocketSignerRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RocketSignerRegistryCallerRaw struct {
	Contract *RocketSignerRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// RocketSignerRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RocketSignerRegistryTransactorRaw struct {
	Contract *RocketSignerRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRocketSignerRegistry creates a new instance of RocketSignerRegistry, bound to a specific deployed contract.
func NewRocketSignerRegistry(address common.Address, backend bind.ContractBackend) (*RocketSignerRegistry, error) {
	contract, err := bindRocketSignerRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &RocketSignerRegistry{RocketSignerRegistryCaller: RocketSignerRegistryCaller{contract: contract}, RocketSignerRegistryTransactor: RocketSignerRegistryTransactor{contract: contract}, RocketSignerRegistryFilterer: RocketSignerRegistryFilterer{contract: contract}}, nil
}

// NewRocketSignerRegistryCaller creates a new read-only instance of RocketSignerRegistry, bound to a specific deployed contract.
func NewRocketSignerRegistryCaller(address common.Address, caller bind.ContractCaller) (*RocketSignerRegistryCaller, error) {
	contract, err := bindRocketSignerRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RocketSignerRegistryCaller{contract: contract}, nil
}

// NewRocketSignerRegistryTransactor creates a new write-only instance of RocketSignerRegistry, bound to a specific deployed contract.
func NewRocketSignerRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*RocketSignerRegistryTransactor, error) {
	contract, err := bindRocketSignerRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RocketSignerRegistryTransactor{contract: contract}, nil
}

// NewRocketSignerRegistryFilterer creates a new log filterer instance of RocketSignerRegistry, bound to a specific deployed contract.
func NewRocketSignerRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*RocketSignerRegistryFilterer, error) {
	contract, err := bindRocketSignerRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RocketSignerRegistryFilterer{contract: contract}, nil
}

// bindRocketSignerRegistry binds a generic wrapper to an already deployed contract.
func bindRocketSignerRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := RocketSignerRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RocketSignerRegistry *RocketSignerRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _RocketSignerRegistry.Contract.RocketSignerRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RocketSignerRegistry *RocketSignerRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RocketSignerRegistry.Contract.RocketSignerRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RocketSignerRegistry *RocketSignerRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RocketSignerRegistry.Contract.RocketSignerRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RocketSignerRegistry *RocketSignerRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _RocketSignerRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RocketSignerRegistry *RocketSignerRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RocketSignerRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RocketSignerRegistry *RocketSignerRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RocketSignerRegistry.Contract.contract.Transact(opts, method, params...)
}

// NodeToSigner is a free data retrieval call binding the contract method 0x603ac290.
//
// Solidity: function nodeToSigner(address ) view returns(address)
func (_RocketSignerRegistry *RocketSignerRegistryCaller) NodeToSigner(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var out []interface{}
	err := _RocketSignerRegistry.contract.Call(opts, &out, "nodeToSigner", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// NodeToSigner is a free data retrieval call binding the contract method 0x603ac290.
//
// Solidity: function nodeToSigner(address ) view returns(address)
func (_RocketSignerRegistry *RocketSignerRegistrySession) NodeToSigner(arg0 common.Address) (common.Address, error) {
	return _RocketSignerRegistry.Contract.NodeToSigner(&_RocketSignerRegistry.CallOpts, arg0)
}

// NodeToSigner is a free data retrieval call binding the contract method 0x603ac290.
//
// Solidity: function nodeToSigner(address ) view returns(address)
func (_RocketSignerRegistry *RocketSignerRegistryCallerSession) NodeToSigner(arg0 common.Address) (common.Address, error) {
	return _RocketSignerRegistry.Contract.NodeToSigner(&_RocketSignerRegistry.CallOpts, arg0)
}

// SignerToNode is a free data retrieval call binding the contract method 0x8ec8af72.
//
// Solidity: function signerToNode(address ) view returns(address)
func (_RocketSignerRegistry *RocketSignerRegistryCaller) SignerToNode(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var out []interface{}
	err := _RocketSignerRegistry.contract.Call(opts, &out, "signerToNode", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SignerToNode is a free data retrieval call binding the contract method 0x8ec8af72.
//
// Solidity: function signerToNode(address ) view returns(address)
func (_RocketSignerRegistry *RocketSignerRegistrySession) SignerToNode(arg0 common.Address) (common.Address, error) {
	return _RocketSignerRegistry.Contract.SignerToNode(&_RocketSignerRegistry.CallOpts, arg0)
}

// SignerToNode is a free data retrieval call binding the contract method 0x8ec8af72.
//
// Solidity: function signerToNode(address ) view returns(address)
func (_RocketSignerRegistry *RocketSignerRegistryCallerSession) SignerToNode(arg0 common.Address) (common.Address, error) {
	return _RocketSignerRegistry.Contract.SignerToNode(&_RocketSignerRegistry.CallOpts, arg0)
}

// ClearSigningDelegate is a paid mutator transaction binding the contract method 0xbdab4704.
//
// Solidity: function clearSigningDelegate() returns()
func (_RocketSignerRegistry *RocketSignerRegistryTransactor) ClearSigningDelegate(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RocketSignerRegistry.contract.Transact(opts, "clearSigningDelegate")
}

// ClearSigningDelegate is a paid mutator transaction binding the contract method 0xbdab4704.
//
// Solidity: function clearSigningDelegate() returns()
func (_RocketSignerRegistry *RocketSignerRegistrySession) ClearSigningDelegate() (*types.Transaction, error) {
	return _RocketSignerRegistry.Contract.ClearSigningDelegate(&_RocketSignerRegistry.TransactOpts)
}

// ClearSigningDelegate is a paid mutator transaction binding the contract method 0xbdab4704.
//
// Solidity: function clearSigningDelegate() returns()
func (_RocketSignerRegistry *RocketSignerRegistryTransactorSession) ClearSigningDelegate() (*types.Transaction, error) {
	return _RocketSignerRegistry.Contract.ClearSigningDelegate(&_RocketSignerRegistry.TransactOpts)
}

// SetSigningDelegate is a paid mutator transaction binding the contract method 0x9bc23cc8.
//
// Solidity: function setSigningDelegate(address _signer, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_RocketSignerRegistry *RocketSignerRegistryTransactor) SetSigningDelegate(opts *bind.TransactOpts, _signer common.Address, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _RocketSignerRegistry.contract.Transact(opts, "setSigningDelegate", _signer, _v, _r, _s)
}

// SetSigningDelegate is a paid mutator transaction binding the contract method 0x9bc23cc8.
//
// Solidity: function setSigningDelegate(address _signer, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_RocketSignerRegistry *RocketSignerRegistrySession) SetSigningDelegate(_signer common.Address, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _RocketSignerRegistry.Contract.SetSigningDelegate(&_RocketSignerRegistry.TransactOpts, _signer, _v, _r, _s)
}

// SetSigningDelegate is a paid mutator transaction binding the contract method 0x9bc23cc8.
//
// Solidity: function setSigningDelegate(address _signer, uint8 _v, bytes32 _r, bytes32 _s) returns()
func (_RocketSignerRegistry *RocketSignerRegistryTransactorSession) SetSigningDelegate(_signer common.Address, _v uint8, _r [32]byte, _s [32]byte) (*types.Transaction, error) {
	return _RocketSignerRegistry.Contract.SetSigningDelegate(&_RocketSignerRegistry.TransactOpts, _signer, _v, _r, _s)
}

// RocketSignerRegistrySigningDelegateSetIterator is returned from FilterSigningDelegateSet and is used to iterate over the raw logs and unpacked data for SigningDelegateSet events raised by the RocketSignerRegistry contract.
type RocketSignerRegistrySigningDelegateSetIterator struct {
	Event *RocketSignerRegistrySigningDelegateSet // Event containing the contract specifics and raw log

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
func (it *RocketSignerRegistrySigningDelegateSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RocketSignerRegistrySigningDelegateSet)
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
		it.Event = new(RocketSignerRegistrySigningDelegateSet)
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
func (it *RocketSignerRegistrySigningDelegateSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RocketSignerRegistrySigningDelegateSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RocketSignerRegistrySigningDelegateSet represents a SigningDelegateSet event raised by the RocketSignerRegistry contract.
type RocketSignerRegistrySigningDelegateSet struct {
	NodeAddress   common.Address
	SignerAddress common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterSigningDelegateSet is a free log retrieval operation binding the contract event 0x3eecd8071b083c5b1ad587e0fd950d9d1caa005eb25ba8c7ac3061ebac0fae8f.
//
// Solidity: event SigningDelegateSet(address indexed nodeAddress, address signerAddress)
func (_RocketSignerRegistry *RocketSignerRegistryFilterer) FilterSigningDelegateSet(opts *bind.FilterOpts, nodeAddress []common.Address) (*RocketSignerRegistrySigningDelegateSetIterator, error) {

	var nodeAddressRule []interface{}
	for _, nodeAddressItem := range nodeAddress {
		nodeAddressRule = append(nodeAddressRule, nodeAddressItem)
	}

	logs, sub, err := _RocketSignerRegistry.contract.FilterLogs(opts, "SigningDelegateSet", nodeAddressRule)
	if err != nil {
		return nil, err
	}
	return &RocketSignerRegistrySigningDelegateSetIterator{contract: _RocketSignerRegistry.contract, event: "SigningDelegateSet", logs: logs, sub: sub}, nil
}

// WatchSigningDelegateSet is a free log subscription operation binding the contract event 0x3eecd8071b083c5b1ad587e0fd950d9d1caa005eb25ba8c7ac3061ebac0fae8f.
//
// Solidity: event SigningDelegateSet(address indexed nodeAddress, address signerAddress)
func (_RocketSignerRegistry *RocketSignerRegistryFilterer) WatchSigningDelegateSet(opts *bind.WatchOpts, sink chan<- *RocketSignerRegistrySigningDelegateSet, nodeAddress []common.Address) (event.Subscription, error) {

	var nodeAddressRule []interface{}
	for _, nodeAddressItem := range nodeAddress {
		nodeAddressRule = append(nodeAddressRule, nodeAddressItem)
	}

	logs, sub, err := _RocketSignerRegistry.contract.WatchLogs(opts, "SigningDelegateSet", nodeAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RocketSignerRegistrySigningDelegateSet)
				if err := _RocketSignerRegistry.contract.UnpackLog(event, "SigningDelegateSet", log); err != nil {
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

// ParseSigningDelegateSet is a log parse operation binding the contract event 0x3eecd8071b083c5b1ad587e0fd950d9d1caa005eb25ba8c7ac3061ebac0fae8f.
//
// Solidity: event SigningDelegateSet(address indexed nodeAddress, address signerAddress)
func (_RocketSignerRegistry *RocketSignerRegistryFilterer) ParseSigningDelegateSet(log types.Log) (*RocketSignerRegistrySigningDelegateSet, error) {
	event := new(RocketSignerRegistrySigningDelegateSet)
	if err := _RocketSignerRegistry.contract.UnpackLog(event, "SigningDelegateSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
