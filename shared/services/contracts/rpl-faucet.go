// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

import (
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
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// RPLFaucetABI is the input ABI used to generate the binding from.
const RPLFaucetABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"maxWithdrawalPerPeriod\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"withdrawalFee\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"withdrawalPeriod\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_rplTokenAddress\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"created\",\"type\":\"uint256\"}],\"name\":\"Withdrawal\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"withdrawTo\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getAllowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_address\",\"type\":\"address\"}],\"name\":\"getAllowanceFor\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getWithdrawalPeriodStart\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_withdrawalPeriod\",\"type\":\"uint256\"}],\"name\":\"setWithdrawalPeriod\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_maxWithdrawalPerPeriod\",\"type\":\"uint256\"}],\"name\":\"setMaxWithdrawalPerPeriod\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_withdrawalFee\",\"type\":\"uint256\"}],\"name\":\"setWithdrawalFee\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// RPLFaucet is an auto generated Go binding around an Ethereum contract.
type RPLFaucet struct {
	RPLFaucetCaller     // Read-only binding to the contract
	RPLFaucetTransactor // Write-only binding to the contract
	RPLFaucetFilterer   // Log filterer for contract events
}

// RPLFaucetCaller is an auto generated read-only Go binding around an Ethereum contract.
type RPLFaucetCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RPLFaucetTransactor is an auto generated write-only Go binding around an Ethereum contract.
type RPLFaucetTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RPLFaucetFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RPLFaucetFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RPLFaucetSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RPLFaucetSession struct {
	Contract     *RPLFaucet        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RPLFaucetCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RPLFaucetCallerSession struct {
	Contract *RPLFaucetCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// RPLFaucetTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RPLFaucetTransactorSession struct {
	Contract     *RPLFaucetTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// RPLFaucetRaw is an auto generated low-level Go binding around an Ethereum contract.
type RPLFaucetRaw struct {
	Contract *RPLFaucet // Generic contract binding to access the raw methods on
}

// RPLFaucetCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RPLFaucetCallerRaw struct {
	Contract *RPLFaucetCaller // Generic read-only contract binding to access the raw methods on
}

// RPLFaucetTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RPLFaucetTransactorRaw struct {
	Contract *RPLFaucetTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRPLFaucet creates a new instance of RPLFaucet, bound to a specific deployed contract.
func NewRPLFaucet(address common.Address, backend bind.ContractBackend) (*RPLFaucet, error) {
	contract, err := bindRPLFaucet(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &RPLFaucet{RPLFaucetCaller: RPLFaucetCaller{contract: contract}, RPLFaucetTransactor: RPLFaucetTransactor{contract: contract}, RPLFaucetFilterer: RPLFaucetFilterer{contract: contract}}, nil
}

// NewRPLFaucetCaller creates a new read-only instance of RPLFaucet, bound to a specific deployed contract.
func NewRPLFaucetCaller(address common.Address, caller bind.ContractCaller) (*RPLFaucetCaller, error) {
	contract, err := bindRPLFaucet(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RPLFaucetCaller{contract: contract}, nil
}

// NewRPLFaucetTransactor creates a new write-only instance of RPLFaucet, bound to a specific deployed contract.
func NewRPLFaucetTransactor(address common.Address, transactor bind.ContractTransactor) (*RPLFaucetTransactor, error) {
	contract, err := bindRPLFaucet(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RPLFaucetTransactor{contract: contract}, nil
}

// NewRPLFaucetFilterer creates a new log filterer instance of RPLFaucet, bound to a specific deployed contract.
func NewRPLFaucetFilterer(address common.Address, filterer bind.ContractFilterer) (*RPLFaucetFilterer, error) {
	contract, err := bindRPLFaucet(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RPLFaucetFilterer{contract: contract}, nil
}

// bindRPLFaucet binds a generic wrapper to an already deployed contract.
func bindRPLFaucet(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(RPLFaucetABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RPLFaucet *RPLFaucetRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _RPLFaucet.Contract.RPLFaucetCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RPLFaucet *RPLFaucetRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RPLFaucet.Contract.RPLFaucetTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RPLFaucet *RPLFaucetRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RPLFaucet.Contract.RPLFaucetTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RPLFaucet *RPLFaucetCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _RPLFaucet.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RPLFaucet *RPLFaucetTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RPLFaucet.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RPLFaucet *RPLFaucetTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RPLFaucet.Contract.contract.Transact(opts, method, params...)
}

// GetAllowance is a free data retrieval call binding the contract method 0x973e9b8b.
//
// Solidity: function getAllowance() view returns(uint256)
func (_RPLFaucet *RPLFaucetCaller) GetAllowance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _RPLFaucet.contract.Call(opts, &out, "getAllowance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAllowance is a free data retrieval call binding the contract method 0x973e9b8b.
//
// Solidity: function getAllowance() view returns(uint256)
func (_RPLFaucet *RPLFaucetSession) GetAllowance() (*big.Int, error) {
	return _RPLFaucet.Contract.GetAllowance(&_RPLFaucet.CallOpts)
}

// GetAllowance is a free data retrieval call binding the contract method 0x973e9b8b.
//
// Solidity: function getAllowance() view returns(uint256)
func (_RPLFaucet *RPLFaucetCallerSession) GetAllowance() (*big.Int, error) {
	return _RPLFaucet.Contract.GetAllowance(&_RPLFaucet.CallOpts)
}

// GetAllowanceFor is a free data retrieval call binding the contract method 0x7639a24b.
//
// Solidity: function getAllowanceFor(address _address) view returns(uint256)
func (_RPLFaucet *RPLFaucetCaller) GetAllowanceFor(opts *bind.CallOpts, _address common.Address) (*big.Int, error) {
	var out []interface{}
	err := _RPLFaucet.contract.Call(opts, &out, "getAllowanceFor", _address)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAllowanceFor is a free data retrieval call binding the contract method 0x7639a24b.
//
// Solidity: function getAllowanceFor(address _address) view returns(uint256)
func (_RPLFaucet *RPLFaucetSession) GetAllowanceFor(_address common.Address) (*big.Int, error) {
	return _RPLFaucet.Contract.GetAllowanceFor(&_RPLFaucet.CallOpts, _address)
}

// GetAllowanceFor is a free data retrieval call binding the contract method 0x7639a24b.
//
// Solidity: function getAllowanceFor(address _address) view returns(uint256)
func (_RPLFaucet *RPLFaucetCallerSession) GetAllowanceFor(_address common.Address) (*big.Int, error) {
	return _RPLFaucet.Contract.GetAllowanceFor(&_RPLFaucet.CallOpts, _address)
}

// GetBalance is a free data retrieval call binding the contract method 0x12065fe0.
//
// Solidity: function getBalance() view returns(uint256)
func (_RPLFaucet *RPLFaucetCaller) GetBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _RPLFaucet.contract.Call(opts, &out, "getBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetBalance is a free data retrieval call binding the contract method 0x12065fe0.
//
// Solidity: function getBalance() view returns(uint256)
func (_RPLFaucet *RPLFaucetSession) GetBalance() (*big.Int, error) {
	return _RPLFaucet.Contract.GetBalance(&_RPLFaucet.CallOpts)
}

// GetBalance is a free data retrieval call binding the contract method 0x12065fe0.
//
// Solidity: function getBalance() view returns(uint256)
func (_RPLFaucet *RPLFaucetCallerSession) GetBalance() (*big.Int, error) {
	return _RPLFaucet.Contract.GetBalance(&_RPLFaucet.CallOpts)
}

// GetWithdrawalPeriodStart is a free data retrieval call binding the contract method 0xfc65bc4f.
//
// Solidity: function getWithdrawalPeriodStart() view returns(uint256)
func (_RPLFaucet *RPLFaucetCaller) GetWithdrawalPeriodStart(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _RPLFaucet.contract.Call(opts, &out, "getWithdrawalPeriodStart")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetWithdrawalPeriodStart is a free data retrieval call binding the contract method 0xfc65bc4f.
//
// Solidity: function getWithdrawalPeriodStart() view returns(uint256)
func (_RPLFaucet *RPLFaucetSession) GetWithdrawalPeriodStart() (*big.Int, error) {
	return _RPLFaucet.Contract.GetWithdrawalPeriodStart(&_RPLFaucet.CallOpts)
}

// GetWithdrawalPeriodStart is a free data retrieval call binding the contract method 0xfc65bc4f.
//
// Solidity: function getWithdrawalPeriodStart() view returns(uint256)
func (_RPLFaucet *RPLFaucetCallerSession) GetWithdrawalPeriodStart() (*big.Int, error) {
	return _RPLFaucet.Contract.GetWithdrawalPeriodStart(&_RPLFaucet.CallOpts)
}

// MaxWithdrawalPerPeriod is a free data retrieval call binding the contract method 0x203bf056.
//
// Solidity: function maxWithdrawalPerPeriod() view returns(uint256)
func (_RPLFaucet *RPLFaucetCaller) MaxWithdrawalPerPeriod(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _RPLFaucet.contract.Call(opts, &out, "maxWithdrawalPerPeriod")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxWithdrawalPerPeriod is a free data retrieval call binding the contract method 0x203bf056.
//
// Solidity: function maxWithdrawalPerPeriod() view returns(uint256)
func (_RPLFaucet *RPLFaucetSession) MaxWithdrawalPerPeriod() (*big.Int, error) {
	return _RPLFaucet.Contract.MaxWithdrawalPerPeriod(&_RPLFaucet.CallOpts)
}

// MaxWithdrawalPerPeriod is a free data retrieval call binding the contract method 0x203bf056.
//
// Solidity: function maxWithdrawalPerPeriod() view returns(uint256)
func (_RPLFaucet *RPLFaucetCallerSession) MaxWithdrawalPerPeriod() (*big.Int, error) {
	return _RPLFaucet.Contract.MaxWithdrawalPerPeriod(&_RPLFaucet.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_RPLFaucet *RPLFaucetCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _RPLFaucet.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_RPLFaucet *RPLFaucetSession) Owner() (common.Address, error) {
	return _RPLFaucet.Contract.Owner(&_RPLFaucet.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_RPLFaucet *RPLFaucetCallerSession) Owner() (common.Address, error) {
	return _RPLFaucet.Contract.Owner(&_RPLFaucet.CallOpts)
}

// WithdrawalFee is a free data retrieval call binding the contract method 0x8bc7e8c4.
//
// Solidity: function withdrawalFee() view returns(uint256)
func (_RPLFaucet *RPLFaucetCaller) WithdrawalFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _RPLFaucet.contract.Call(opts, &out, "withdrawalFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// WithdrawalFee is a free data retrieval call binding the contract method 0x8bc7e8c4.
//
// Solidity: function withdrawalFee() view returns(uint256)
func (_RPLFaucet *RPLFaucetSession) WithdrawalFee() (*big.Int, error) {
	return _RPLFaucet.Contract.WithdrawalFee(&_RPLFaucet.CallOpts)
}

// WithdrawalFee is a free data retrieval call binding the contract method 0x8bc7e8c4.
//
// Solidity: function withdrawalFee() view returns(uint256)
func (_RPLFaucet *RPLFaucetCallerSession) WithdrawalFee() (*big.Int, error) {
	return _RPLFaucet.Contract.WithdrawalFee(&_RPLFaucet.CallOpts)
}

// WithdrawalPeriod is a free data retrieval call binding the contract method 0xbca7093d.
//
// Solidity: function withdrawalPeriod() view returns(uint256)
func (_RPLFaucet *RPLFaucetCaller) WithdrawalPeriod(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _RPLFaucet.contract.Call(opts, &out, "withdrawalPeriod")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// WithdrawalPeriod is a free data retrieval call binding the contract method 0xbca7093d.
//
// Solidity: function withdrawalPeriod() view returns(uint256)
func (_RPLFaucet *RPLFaucetSession) WithdrawalPeriod() (*big.Int, error) {
	return _RPLFaucet.Contract.WithdrawalPeriod(&_RPLFaucet.CallOpts)
}

// WithdrawalPeriod is a free data retrieval call binding the contract method 0xbca7093d.
//
// Solidity: function withdrawalPeriod() view returns(uint256)
func (_RPLFaucet *RPLFaucetCallerSession) WithdrawalPeriod() (*big.Int, error) {
	return _RPLFaucet.Contract.WithdrawalPeriod(&_RPLFaucet.CallOpts)
}

// SetMaxWithdrawalPerPeriod is a paid mutator transaction binding the contract method 0xc0ac9128.
//
// Solidity: function setMaxWithdrawalPerPeriod(uint256 _maxWithdrawalPerPeriod) returns()
func (_RPLFaucet *RPLFaucetTransactor) SetMaxWithdrawalPerPeriod(opts *bind.TransactOpts, _maxWithdrawalPerPeriod *big.Int) (*types.Transaction, error) {
	return _RPLFaucet.contract.Transact(opts, "setMaxWithdrawalPerPeriod", _maxWithdrawalPerPeriod)
}

// SetMaxWithdrawalPerPeriod is a paid mutator transaction binding the contract method 0xc0ac9128.
//
// Solidity: function setMaxWithdrawalPerPeriod(uint256 _maxWithdrawalPerPeriod) returns()
func (_RPLFaucet *RPLFaucetSession) SetMaxWithdrawalPerPeriod(_maxWithdrawalPerPeriod *big.Int) (*types.Transaction, error) {
	return _RPLFaucet.Contract.SetMaxWithdrawalPerPeriod(&_RPLFaucet.TransactOpts, _maxWithdrawalPerPeriod)
}

// SetMaxWithdrawalPerPeriod is a paid mutator transaction binding the contract method 0xc0ac9128.
//
// Solidity: function setMaxWithdrawalPerPeriod(uint256 _maxWithdrawalPerPeriod) returns()
func (_RPLFaucet *RPLFaucetTransactorSession) SetMaxWithdrawalPerPeriod(_maxWithdrawalPerPeriod *big.Int) (*types.Transaction, error) {
	return _RPLFaucet.Contract.SetMaxWithdrawalPerPeriod(&_RPLFaucet.TransactOpts, _maxWithdrawalPerPeriod)
}

// SetWithdrawalFee is a paid mutator transaction binding the contract method 0xac1e5025.
//
// Solidity: function setWithdrawalFee(uint256 _withdrawalFee) returns()
func (_RPLFaucet *RPLFaucetTransactor) SetWithdrawalFee(opts *bind.TransactOpts, _withdrawalFee *big.Int) (*types.Transaction, error) {
	return _RPLFaucet.contract.Transact(opts, "setWithdrawalFee", _withdrawalFee)
}

// SetWithdrawalFee is a paid mutator transaction binding the contract method 0xac1e5025.
//
// Solidity: function setWithdrawalFee(uint256 _withdrawalFee) returns()
func (_RPLFaucet *RPLFaucetSession) SetWithdrawalFee(_withdrawalFee *big.Int) (*types.Transaction, error) {
	return _RPLFaucet.Contract.SetWithdrawalFee(&_RPLFaucet.TransactOpts, _withdrawalFee)
}

// SetWithdrawalFee is a paid mutator transaction binding the contract method 0xac1e5025.
//
// Solidity: function setWithdrawalFee(uint256 _withdrawalFee) returns()
func (_RPLFaucet *RPLFaucetTransactorSession) SetWithdrawalFee(_withdrawalFee *big.Int) (*types.Transaction, error) {
	return _RPLFaucet.Contract.SetWithdrawalFee(&_RPLFaucet.TransactOpts, _withdrawalFee)
}

// SetWithdrawalPeriod is a paid mutator transaction binding the contract method 0x973b294f.
//
// Solidity: function setWithdrawalPeriod(uint256 _withdrawalPeriod) returns()
func (_RPLFaucet *RPLFaucetTransactor) SetWithdrawalPeriod(opts *bind.TransactOpts, _withdrawalPeriod *big.Int) (*types.Transaction, error) {
	return _RPLFaucet.contract.Transact(opts, "setWithdrawalPeriod", _withdrawalPeriod)
}

// SetWithdrawalPeriod is a paid mutator transaction binding the contract method 0x973b294f.
//
// Solidity: function setWithdrawalPeriod(uint256 _withdrawalPeriod) returns()
func (_RPLFaucet *RPLFaucetSession) SetWithdrawalPeriod(_withdrawalPeriod *big.Int) (*types.Transaction, error) {
	return _RPLFaucet.Contract.SetWithdrawalPeriod(&_RPLFaucet.TransactOpts, _withdrawalPeriod)
}

// SetWithdrawalPeriod is a paid mutator transaction binding the contract method 0x973b294f.
//
// Solidity: function setWithdrawalPeriod(uint256 _withdrawalPeriod) returns()
func (_RPLFaucet *RPLFaucetTransactorSession) SetWithdrawalPeriod(_withdrawalPeriod *big.Int) (*types.Transaction, error) {
	return _RPLFaucet.Contract.SetWithdrawalPeriod(&_RPLFaucet.TransactOpts, _withdrawalPeriod)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 _amount) payable returns(bool)
func (_RPLFaucet *RPLFaucetTransactor) Withdraw(opts *bind.TransactOpts, _amount *big.Int) (*types.Transaction, error) {
	return _RPLFaucet.contract.Transact(opts, "withdraw", _amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 _amount) payable returns(bool)
func (_RPLFaucet *RPLFaucetSession) Withdraw(_amount *big.Int) (*types.Transaction, error) {
	return _RPLFaucet.Contract.Withdraw(&_RPLFaucet.TransactOpts, _amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 _amount) payable returns(bool)
func (_RPLFaucet *RPLFaucetTransactorSession) Withdraw(_amount *big.Int) (*types.Transaction, error) {
	return _RPLFaucet.Contract.Withdraw(&_RPLFaucet.TransactOpts, _amount)
}

// WithdrawTo is a paid mutator transaction binding the contract method 0x205c2878.
//
// Solidity: function withdrawTo(address _to, uint256 _amount) payable returns(bool)
func (_RPLFaucet *RPLFaucetTransactor) WithdrawTo(opts *bind.TransactOpts, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _RPLFaucet.contract.Transact(opts, "withdrawTo", _to, _amount)
}

// WithdrawTo is a paid mutator transaction binding the contract method 0x205c2878.
//
// Solidity: function withdrawTo(address _to, uint256 _amount) payable returns(bool)
func (_RPLFaucet *RPLFaucetSession) WithdrawTo(_to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _RPLFaucet.Contract.WithdrawTo(&_RPLFaucet.TransactOpts, _to, _amount)
}

// WithdrawTo is a paid mutator transaction binding the contract method 0x205c2878.
//
// Solidity: function withdrawTo(address _to, uint256 _amount) payable returns(bool)
func (_RPLFaucet *RPLFaucetTransactorSession) WithdrawTo(_to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _RPLFaucet.Contract.WithdrawTo(&_RPLFaucet.TransactOpts, _to, _amount)
}

// RPLFaucetWithdrawalIterator is returned from FilterWithdrawal and is used to iterate over the raw logs and unpacked data for Withdrawal events raised by the RPLFaucet contract.
type RPLFaucetWithdrawalIterator struct {
	Event *RPLFaucetWithdrawal // Event containing the contract specifics and raw log

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
func (it *RPLFaucetWithdrawalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RPLFaucetWithdrawal)
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
		it.Event = new(RPLFaucetWithdrawal)
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
func (it *RPLFaucetWithdrawalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RPLFaucetWithdrawalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RPLFaucetWithdrawal represents a Withdrawal event raised by the RPLFaucet contract.
type RPLFaucetWithdrawal struct {
	To      common.Address
	Value   *big.Int
	Created *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterWithdrawal is a free log retrieval operation binding the contract event 0xdf273cb619d95419a9cd0ec88123a0538c85064229baa6363788f743fff90deb.
//
// Solidity: event Withdrawal(address indexed to, uint256 value, uint256 created)
func (_RPLFaucet *RPLFaucetFilterer) FilterWithdrawal(opts *bind.FilterOpts, to []common.Address) (*RPLFaucetWithdrawalIterator, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _RPLFaucet.contract.FilterLogs(opts, "Withdrawal", toRule)
	if err != nil {
		return nil, err
	}
	return &RPLFaucetWithdrawalIterator{contract: _RPLFaucet.contract, event: "Withdrawal", logs: logs, sub: sub}, nil
}

// WatchWithdrawal is a free log subscription operation binding the contract event 0xdf273cb619d95419a9cd0ec88123a0538c85064229baa6363788f743fff90deb.
//
// Solidity: event Withdrawal(address indexed to, uint256 value, uint256 created)
func (_RPLFaucet *RPLFaucetFilterer) WatchWithdrawal(opts *bind.WatchOpts, sink chan<- *RPLFaucetWithdrawal, to []common.Address) (event.Subscription, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _RPLFaucet.contract.WatchLogs(opts, "Withdrawal", toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RPLFaucetWithdrawal)
				if err := _RPLFaucet.contract.UnpackLog(event, "Withdrawal", log); err != nil {
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

// ParseWithdrawal is a log parse operation binding the contract event 0xdf273cb619d95419a9cd0ec88123a0538c85064229baa6363788f743fff90deb.
//
// Solidity: event Withdrawal(address indexed to, uint256 value, uint256 created)
func (_RPLFaucet *RPLFaucetFilterer) ParseWithdrawal(log types.Log) (*RPLFaucetWithdrawal, error) {
	event := new(RPLFaucetWithdrawal)
	if err := _RPLFaucet.contract.UnpackLog(event, "Withdrawal", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
