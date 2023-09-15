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
)

// SnapshotDelegationMetaData contains all meta data concerning the SnapshotDelegation contract.
var SnapshotDelegationMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegate\",\"type\":\"address\"}],\"name\":\"ClearDelegate\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegate\",\"type\":\"address\"}],\"name\":\"SetDelegate\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"clearDelegate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"delegation\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"delegate\",\"type\":\"address\"}],\"name\":\"setDelegate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// SnapshotDelegationABI is the input ABI used to generate the binding from.
// Deprecated: Use SnapshotDelegationMetaData.ABI instead.
var SnapshotDelegationABI = SnapshotDelegationMetaData.ABI

// SnapshotDelegation is an auto generated Go binding around an Ethereum contract.
type SnapshotDelegation struct {
	SnapshotDelegationCaller     // Read-only binding to the contract
	SnapshotDelegationTransactor // Write-only binding to the contract
	SnapshotDelegationFilterer   // Log filterer for contract events
}

// SnapshotDelegationCaller is an auto generated read-only Go binding around an Ethereum contract.
type SnapshotDelegationCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SnapshotDelegationTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SnapshotDelegationTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SnapshotDelegationFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SnapshotDelegationFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SnapshotDelegationSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SnapshotDelegationSession struct {
	Contract     *SnapshotDelegation // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// SnapshotDelegationCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SnapshotDelegationCallerSession struct {
	Contract *SnapshotDelegationCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// SnapshotDelegationTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SnapshotDelegationTransactorSession struct {
	Contract     *SnapshotDelegationTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// SnapshotDelegationRaw is an auto generated low-level Go binding around an Ethereum contract.
type SnapshotDelegationRaw struct {
	Contract *SnapshotDelegation // Generic contract binding to access the raw methods on
}

// SnapshotDelegationCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SnapshotDelegationCallerRaw struct {
	Contract *SnapshotDelegationCaller // Generic read-only contract binding to access the raw methods on
}

// SnapshotDelegationTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SnapshotDelegationTransactorRaw struct {
	Contract *SnapshotDelegationTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSnapshotDelegation creates a new instance of SnapshotDelegation, bound to a specific deployed contract.
func NewSnapshotDelegation(address common.Address, backend bind.ContractBackend) (*SnapshotDelegation, error) {
	contract, err := bindSnapshotDelegation(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SnapshotDelegation{SnapshotDelegationCaller: SnapshotDelegationCaller{contract: contract}, SnapshotDelegationTransactor: SnapshotDelegationTransactor{contract: contract}, SnapshotDelegationFilterer: SnapshotDelegationFilterer{contract: contract}}, nil
}

// NewSnapshotDelegationCaller creates a new read-only instance of SnapshotDelegation, bound to a specific deployed contract.
func NewSnapshotDelegationCaller(address common.Address, caller bind.ContractCaller) (*SnapshotDelegationCaller, error) {
	contract, err := bindSnapshotDelegation(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SnapshotDelegationCaller{contract: contract}, nil
}

// NewSnapshotDelegationTransactor creates a new write-only instance of SnapshotDelegation, bound to a specific deployed contract.
func NewSnapshotDelegationTransactor(address common.Address, transactor bind.ContractTransactor) (*SnapshotDelegationTransactor, error) {
	contract, err := bindSnapshotDelegation(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SnapshotDelegationTransactor{contract: contract}, nil
}

// NewSnapshotDelegationFilterer creates a new log filterer instance of SnapshotDelegation, bound to a specific deployed contract.
func NewSnapshotDelegationFilterer(address common.Address, filterer bind.ContractFilterer) (*SnapshotDelegationFilterer, error) {
	contract, err := bindSnapshotDelegation(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SnapshotDelegationFilterer{contract: contract}, nil
}

// bindSnapshotDelegation binds a generic wrapper to an already deployed contract.
func bindSnapshotDelegation(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SnapshotDelegationABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SnapshotDelegation *SnapshotDelegationRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SnapshotDelegation.Contract.SnapshotDelegationCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SnapshotDelegation *SnapshotDelegationRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SnapshotDelegation.Contract.SnapshotDelegationTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SnapshotDelegation *SnapshotDelegationRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SnapshotDelegation.Contract.SnapshotDelegationTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SnapshotDelegation *SnapshotDelegationCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SnapshotDelegation.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SnapshotDelegation *SnapshotDelegationTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SnapshotDelegation.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SnapshotDelegation *SnapshotDelegationTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SnapshotDelegation.Contract.contract.Transact(opts, method, params...)
}

// Delegation is a free data retrieval call binding the contract method 0x74c6c454.
//
// Solidity: function delegation(address , bytes32 ) view returns(address)
func (_SnapshotDelegation *SnapshotDelegationCaller) Delegation(opts *bind.CallOpts, arg0 common.Address, arg1 [32]byte) (common.Address, error) {
	var out []interface{}
	err := _SnapshotDelegation.contract.Call(opts, &out, "delegation", arg0, arg1)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Delegation is a free data retrieval call binding the contract method 0x74c6c454.
//
// Solidity: function delegation(address , bytes32 ) view returns(address)
func (_SnapshotDelegation *SnapshotDelegationSession) Delegation(arg0 common.Address, arg1 [32]byte) (common.Address, error) {
	return _SnapshotDelegation.Contract.Delegation(&_SnapshotDelegation.CallOpts, arg0, arg1)
}

// Delegation is a free data retrieval call binding the contract method 0x74c6c454.
//
// Solidity: function delegation(address , bytes32 ) view returns(address)
func (_SnapshotDelegation *SnapshotDelegationCallerSession) Delegation(arg0 common.Address, arg1 [32]byte) (common.Address, error) {
	return _SnapshotDelegation.Contract.Delegation(&_SnapshotDelegation.CallOpts, arg0, arg1)
}

// ClearDelegate is a paid mutator transaction binding the contract method 0xf0bedbe2.
//
// Solidity: function clearDelegate(bytes32 id) returns()
func (_SnapshotDelegation *SnapshotDelegationTransactor) ClearDelegate(opts *bind.TransactOpts, id [32]byte) (*types.Transaction, error) {
	return _SnapshotDelegation.contract.Transact(opts, "clearDelegate", id)
}

// ClearDelegate is a paid mutator transaction binding the contract method 0xf0bedbe2.
//
// Solidity: function clearDelegate(bytes32 id) returns()
func (_SnapshotDelegation *SnapshotDelegationSession) ClearDelegate(id [32]byte) (*types.Transaction, error) {
	return _SnapshotDelegation.Contract.ClearDelegate(&_SnapshotDelegation.TransactOpts, id)
}

// ClearDelegate is a paid mutator transaction binding the contract method 0xf0bedbe2.
//
// Solidity: function clearDelegate(bytes32 id) returns()
func (_SnapshotDelegation *SnapshotDelegationTransactorSession) ClearDelegate(id [32]byte) (*types.Transaction, error) {
	return _SnapshotDelegation.Contract.ClearDelegate(&_SnapshotDelegation.TransactOpts, id)
}

// SetDelegate is a paid mutator transaction binding the contract method 0xbd86e508.
//
// Solidity: function setDelegate(bytes32 id, address delegate) returns()
func (_SnapshotDelegation *SnapshotDelegationTransactor) SetDelegate(opts *bind.TransactOpts, id [32]byte, delegate common.Address) (*types.Transaction, error) {
	return _SnapshotDelegation.contract.Transact(opts, "setDelegate", id, delegate)
}

// SetDelegate is a paid mutator transaction binding the contract method 0xbd86e508.
//
// Solidity: function setDelegate(bytes32 id, address delegate) returns()
func (_SnapshotDelegation *SnapshotDelegationSession) SetDelegate(id [32]byte, delegate common.Address) (*types.Transaction, error) {
	return _SnapshotDelegation.Contract.SetDelegate(&_SnapshotDelegation.TransactOpts, id, delegate)
}

// SetDelegate is a paid mutator transaction binding the contract method 0xbd86e508.
//
// Solidity: function setDelegate(bytes32 id, address delegate) returns()
func (_SnapshotDelegation *SnapshotDelegationTransactorSession) SetDelegate(id [32]byte, delegate common.Address) (*types.Transaction, error) {
	return _SnapshotDelegation.Contract.SetDelegate(&_SnapshotDelegation.TransactOpts, id, delegate)
}

// SnapshotDelegationClearDelegateIterator is returned from FilterClearDelegate and is used to iterate over the raw logs and unpacked data for ClearDelegate events raised by the SnapshotDelegation contract.
type SnapshotDelegationClearDelegateIterator struct {
	Event *SnapshotDelegationClearDelegate // Event containing the contract specifics and raw log

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
func (it *SnapshotDelegationClearDelegateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SnapshotDelegationClearDelegate)
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
		it.Event = new(SnapshotDelegationClearDelegate)
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
func (it *SnapshotDelegationClearDelegateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SnapshotDelegationClearDelegateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SnapshotDelegationClearDelegate represents a ClearDelegate event raised by the SnapshotDelegation contract.
type SnapshotDelegationClearDelegate struct {
	Delegator common.Address
	Id        [32]byte
	Delegate  common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterClearDelegate is a free log retrieval operation binding the contract event 0x9c4f00c4291262731946e308dc2979a56bd22cce8f95906b975065e96cd5a064.
//
// Solidity: event ClearDelegate(address indexed delegator, bytes32 indexed id, address indexed delegate)
func (_SnapshotDelegation *SnapshotDelegationFilterer) FilterClearDelegate(opts *bind.FilterOpts, delegator []common.Address, id [][32]byte, delegate []common.Address) (*SnapshotDelegationClearDelegateIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var delegateRule []interface{}
	for _, delegateItem := range delegate {
		delegateRule = append(delegateRule, delegateItem)
	}

	logs, sub, err := _SnapshotDelegation.contract.FilterLogs(opts, "ClearDelegate", delegatorRule, idRule, delegateRule)
	if err != nil {
		return nil, err
	}
	return &SnapshotDelegationClearDelegateIterator{contract: _SnapshotDelegation.contract, event: "ClearDelegate", logs: logs, sub: sub}, nil
}

// WatchClearDelegate is a free log subscription operation binding the contract event 0x9c4f00c4291262731946e308dc2979a56bd22cce8f95906b975065e96cd5a064.
//
// Solidity: event ClearDelegate(address indexed delegator, bytes32 indexed id, address indexed delegate)
func (_SnapshotDelegation *SnapshotDelegationFilterer) WatchClearDelegate(opts *bind.WatchOpts, sink chan<- *SnapshotDelegationClearDelegate, delegator []common.Address, id [][32]byte, delegate []common.Address) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var delegateRule []interface{}
	for _, delegateItem := range delegate {
		delegateRule = append(delegateRule, delegateItem)
	}

	logs, sub, err := _SnapshotDelegation.contract.WatchLogs(opts, "ClearDelegate", delegatorRule, idRule, delegateRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SnapshotDelegationClearDelegate)
				if err := _SnapshotDelegation.contract.UnpackLog(event, "ClearDelegate", log); err != nil {
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

// ParseClearDelegate is a log parse operation binding the contract event 0x9c4f00c4291262731946e308dc2979a56bd22cce8f95906b975065e96cd5a064.
//
// Solidity: event ClearDelegate(address indexed delegator, bytes32 indexed id, address indexed delegate)
func (_SnapshotDelegation *SnapshotDelegationFilterer) ParseClearDelegate(log types.Log) (*SnapshotDelegationClearDelegate, error) {
	event := new(SnapshotDelegationClearDelegate)
	if err := _SnapshotDelegation.contract.UnpackLog(event, "ClearDelegate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SnapshotDelegationSetDelegateIterator is returned from FilterSetDelegate and is used to iterate over the raw logs and unpacked data for SetDelegate events raised by the SnapshotDelegation contract.
type SnapshotDelegationSetDelegateIterator struct {
	Event *SnapshotDelegationSetDelegate // Event containing the contract specifics and raw log

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
func (it *SnapshotDelegationSetDelegateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SnapshotDelegationSetDelegate)
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
		it.Event = new(SnapshotDelegationSetDelegate)
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
func (it *SnapshotDelegationSetDelegateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SnapshotDelegationSetDelegateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SnapshotDelegationSetDelegate represents a SetDelegate event raised by the SnapshotDelegation contract.
type SnapshotDelegationSetDelegate struct {
	Delegator common.Address
	Id        [32]byte
	Delegate  common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterSetDelegate is a free log retrieval operation binding the contract event 0xa9a7fd460f56bddb880a465a9c3e9730389c70bc53108148f16d55a87a6c468e.
//
// Solidity: event SetDelegate(address indexed delegator, bytes32 indexed id, address indexed delegate)
func (_SnapshotDelegation *SnapshotDelegationFilterer) FilterSetDelegate(opts *bind.FilterOpts, delegator []common.Address, id [][32]byte, delegate []common.Address) (*SnapshotDelegationSetDelegateIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var delegateRule []interface{}
	for _, delegateItem := range delegate {
		delegateRule = append(delegateRule, delegateItem)
	}

	logs, sub, err := _SnapshotDelegation.contract.FilterLogs(opts, "SetDelegate", delegatorRule, idRule, delegateRule)
	if err != nil {
		return nil, err
	}
	return &SnapshotDelegationSetDelegateIterator{contract: _SnapshotDelegation.contract, event: "SetDelegate", logs: logs, sub: sub}, nil
}

// WatchSetDelegate is a free log subscription operation binding the contract event 0xa9a7fd460f56bddb880a465a9c3e9730389c70bc53108148f16d55a87a6c468e.
//
// Solidity: event SetDelegate(address indexed delegator, bytes32 indexed id, address indexed delegate)
func (_SnapshotDelegation *SnapshotDelegationFilterer) WatchSetDelegate(opts *bind.WatchOpts, sink chan<- *SnapshotDelegationSetDelegate, delegator []common.Address, id [][32]byte, delegate []common.Address) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var delegateRule []interface{}
	for _, delegateItem := range delegate {
		delegateRule = append(delegateRule, delegateItem)
	}

	logs, sub, err := _SnapshotDelegation.contract.WatchLogs(opts, "SetDelegate", delegatorRule, idRule, delegateRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SnapshotDelegationSetDelegate)
				if err := _SnapshotDelegation.contract.UnpackLog(event, "SetDelegate", log); err != nil {
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

// ParseSetDelegate is a log parse operation binding the contract event 0xa9a7fd460f56bddb880a465a9c3e9730389c70bc53108148f16d55a87a6c468e.
//
// Solidity: event SetDelegate(address indexed delegator, bytes32 indexed id, address indexed delegate)
func (_SnapshotDelegation *SnapshotDelegationFilterer) ParseSetDelegate(log types.Log) (*SnapshotDelegationSetDelegate, error) {
	event := new(SnapshotDelegationSetDelegate)
	if err := _SnapshotDelegation.contract.UnpackLog(event, "SetDelegate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
