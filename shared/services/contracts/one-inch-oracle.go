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

// OneInchOracleABI is the input ABI used to generate the binding from.
const OneInchOracleABI = "[{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"srcToken\",\"type\":\"address\"},{\"internalType\":\"contractIERC20\",\"name\":\"dstToken\",\"type\":\"address\"}],\"name\":\"getRate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"weightedRate\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// OneInchOracle is an auto generated Go binding around an Ethereum contract.
type OneInchOracle struct {
	OneInchOracleCaller     // Read-only binding to the contract
	OneInchOracleTransactor // Write-only binding to the contract
	OneInchOracleFilterer   // Log filterer for contract events
}

// OneInchOracleCaller is an auto generated read-only Go binding around an Ethereum contract.
type OneInchOracleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OneInchOracleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OneInchOracleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OneInchOracleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OneInchOracleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OneInchOracleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OneInchOracleSession struct {
	Contract     *OneInchOracle    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OneInchOracleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OneInchOracleCallerSession struct {
	Contract *OneInchOracleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// OneInchOracleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OneInchOracleTransactorSession struct {
	Contract     *OneInchOracleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// OneInchOracleRaw is an auto generated low-level Go binding around an Ethereum contract.
type OneInchOracleRaw struct {
	Contract *OneInchOracle // Generic contract binding to access the raw methods on
}

// OneInchOracleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OneInchOracleCallerRaw struct {
	Contract *OneInchOracleCaller // Generic read-only contract binding to access the raw methods on
}

// OneInchOracleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OneInchOracleTransactorRaw struct {
	Contract *OneInchOracleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOneInchOracle creates a new instance of OneInchOracle, bound to a specific deployed contract.
func NewOneInchOracle(address common.Address, backend bind.ContractBackend) (*OneInchOracle, error) {
	contract, err := bindOneInchOracle(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OneInchOracle{OneInchOracleCaller: OneInchOracleCaller{contract: contract}, OneInchOracleTransactor: OneInchOracleTransactor{contract: contract}, OneInchOracleFilterer: OneInchOracleFilterer{contract: contract}}, nil
}

// NewOneInchOracleCaller creates a new read-only instance of OneInchOracle, bound to a specific deployed contract.
func NewOneInchOracleCaller(address common.Address, caller bind.ContractCaller) (*OneInchOracleCaller, error) {
	contract, err := bindOneInchOracle(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OneInchOracleCaller{contract: contract}, nil
}

// NewOneInchOracleTransactor creates a new write-only instance of OneInchOracle, bound to a specific deployed contract.
func NewOneInchOracleTransactor(address common.Address, transactor bind.ContractTransactor) (*OneInchOracleTransactor, error) {
	contract, err := bindOneInchOracle(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OneInchOracleTransactor{contract: contract}, nil
}

// NewOneInchOracleFilterer creates a new log filterer instance of OneInchOracle, bound to a specific deployed contract.
func NewOneInchOracleFilterer(address common.Address, filterer bind.ContractFilterer) (*OneInchOracleFilterer, error) {
	contract, err := bindOneInchOracle(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OneInchOracleFilterer{contract: contract}, nil
}

// bindOneInchOracle binds a generic wrapper to an already deployed contract.
func bindOneInchOracle(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(OneInchOracleABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OneInchOracle *OneInchOracleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OneInchOracle.Contract.OneInchOracleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OneInchOracle *OneInchOracleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OneInchOracle.Contract.OneInchOracleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OneInchOracle *OneInchOracleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OneInchOracle.Contract.OneInchOracleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OneInchOracle *OneInchOracleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OneInchOracle.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OneInchOracle *OneInchOracleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OneInchOracle.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OneInchOracle *OneInchOracleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OneInchOracle.Contract.contract.Transact(opts, method, params...)
}

// GetRate is a free data retrieval call binding the contract method 0x379b87ea.
//
// Solidity: function getRate(address srcToken, address dstToken) view returns(uint256 weightedRate)
func (_OneInchOracle *OneInchOracleCaller) GetRate(opts *bind.CallOpts, srcToken common.Address, dstToken common.Address) (*big.Int, error) {
	var out []interface{}
	err := _OneInchOracle.contract.Call(opts, &out, "getRate", srcToken, dstToken)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRate is a free data retrieval call binding the contract method 0x379b87ea.
//
// Solidity: function getRate(address srcToken, address dstToken) view returns(uint256 weightedRate)
func (_OneInchOracle *OneInchOracleSession) GetRate(srcToken common.Address, dstToken common.Address) (*big.Int, error) {
	return _OneInchOracle.Contract.GetRate(&_OneInchOracle.CallOpts, srcToken, dstToken)
}

// GetRate is a free data retrieval call binding the contract method 0x379b87ea.
//
// Solidity: function getRate(address srcToken, address dstToken) view returns(uint256 weightedRate)
func (_OneInchOracle *OneInchOracleCallerSession) GetRate(srcToken common.Address, dstToken common.Address) (*big.Int, error) {
	return _OneInchOracle.Contract.GetRate(&_OneInchOracle.CallOpts, srcToken, dstToken)
}
