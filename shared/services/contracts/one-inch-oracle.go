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
const OneInchOracleABI = "[{\"inputs\":[{\"internalType\":\"contractMultiWrapper\",\"name\":\"_multiWrapper\",\"type\":\"address\"},{\"internalType\":\"contractIOracle[]\",\"name\":\"existingOracles\",\"type\":\"address[]\"},{\"internalType\":\"enumOffchainOracle.OracleType[]\",\"name\":\"oracleTypes\",\"type\":\"uint8[]\"},{\"internalType\":\"contractIERC20[]\",\"name\":\"existingConnectors\",\"type\":\"address[]\"},{\"internalType\":\"contractIERC20\",\"name\":\"wBase\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"contractIERC20\",\"name\":\"connector\",\"type\":\"address\"}],\"name\":\"ConnectorAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"contractIERC20\",\"name\":\"connector\",\"type\":\"address\"}],\"name\":\"ConnectorRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"contractMultiWrapper\",\"name\":\"multiWrapper\",\"type\":\"address\"}],\"name\":\"MultiWrapperUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"contractIOracle\",\"name\":\"oracle\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"enumOffchainOracle.OracleType\",\"name\":\"oracleType\",\"type\":\"uint8\"}],\"name\":\"OracleAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"contractIOracle\",\"name\":\"oracle\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"enumOffchainOracle.OracleType\",\"name\":\"oracleType\",\"type\":\"uint8\"}],\"name\":\"OracleRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"connector\",\"type\":\"address\"}],\"name\":\"addConnector\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIOracle\",\"name\":\"oracle\",\"type\":\"address\"},{\"internalType\":\"enumOffchainOracle.OracleType\",\"name\":\"oracleKind\",\"type\":\"uint8\"}],\"name\":\"addOracle\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"connectors\",\"outputs\":[{\"internalType\":\"contractIERC20[]\",\"name\":\"allConnectors\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"srcToken\",\"type\":\"address\"},{\"internalType\":\"contractIERC20\",\"name\":\"dstToken\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"useWrappers\",\"type\":\"bool\"}],\"name\":\"getRate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"weightedRate\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"srcToken\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"useSrcWrappers\",\"type\":\"bool\"}],\"name\":\"getRateToEth\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"weightedRate\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"multiWrapper\",\"outputs\":[{\"internalType\":\"contractMultiWrapper\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"oracles\",\"outputs\":[{\"internalType\":\"contractIOracle[]\",\"name\":\"allOracles\",\"type\":\"address[]\"},{\"internalType\":\"enumOffchainOracle.OracleType[]\",\"name\":\"oracleTypes\",\"type\":\"uint8[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"connector\",\"type\":\"address\"}],\"name\":\"removeConnector\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIOracle\",\"name\":\"oracle\",\"type\":\"address\"},{\"internalType\":\"enumOffchainOracle.OracleType\",\"name\":\"oracleKind\",\"type\":\"uint8\"}],\"name\":\"removeOracle\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractMultiWrapper\",\"name\":\"_multiWrapper\",\"type\":\"address\"}],\"name\":\"setMultiWrapper\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

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

// Connectors is a free data retrieval call binding the contract method 0x65050a68.
//
// Solidity: function connectors() view returns(address[] allConnectors)
func (_OneInchOracle *OneInchOracleCaller) Connectors(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _OneInchOracle.contract.Call(opts, &out, "connectors")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// Connectors is a free data retrieval call binding the contract method 0x65050a68.
//
// Solidity: function connectors() view returns(address[] allConnectors)
func (_OneInchOracle *OneInchOracleSession) Connectors() ([]common.Address, error) {
	return _OneInchOracle.Contract.Connectors(&_OneInchOracle.CallOpts)
}

// Connectors is a free data retrieval call binding the contract method 0x65050a68.
//
// Solidity: function connectors() view returns(address[] allConnectors)
func (_OneInchOracle *OneInchOracleCallerSession) Connectors() ([]common.Address, error) {
	return _OneInchOracle.Contract.Connectors(&_OneInchOracle.CallOpts)
}

// GetRate is a free data retrieval call binding the contract method 0x802431fb.
//
// Solidity: function getRate(address srcToken, address dstToken, bool useWrappers) view returns(uint256 weightedRate)
func (_OneInchOracle *OneInchOracleCaller) GetRate(opts *bind.CallOpts, srcToken common.Address, dstToken common.Address, useWrappers bool) (*big.Int, error) {
	var out []interface{}
	err := _OneInchOracle.contract.Call(opts, &out, "getRate", srcToken, dstToken, useWrappers)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRate is a free data retrieval call binding the contract method 0x802431fb.
//
// Solidity: function getRate(address srcToken, address dstToken, bool useWrappers) view returns(uint256 weightedRate)
func (_OneInchOracle *OneInchOracleSession) GetRate(srcToken common.Address, dstToken common.Address, useWrappers bool) (*big.Int, error) {
	return _OneInchOracle.Contract.GetRate(&_OneInchOracle.CallOpts, srcToken, dstToken, useWrappers)
}

// GetRate is a free data retrieval call binding the contract method 0x802431fb.
//
// Solidity: function getRate(address srcToken, address dstToken, bool useWrappers) view returns(uint256 weightedRate)
func (_OneInchOracle *OneInchOracleCallerSession) GetRate(srcToken common.Address, dstToken common.Address, useWrappers bool) (*big.Int, error) {
	return _OneInchOracle.Contract.GetRate(&_OneInchOracle.CallOpts, srcToken, dstToken, useWrappers)
}

// GetRateToEth is a free data retrieval call binding the contract method 0x7de4fd10.
//
// Solidity: function getRateToEth(address srcToken, bool useSrcWrappers) view returns(uint256 weightedRate)
func (_OneInchOracle *OneInchOracleCaller) GetRateToEth(opts *bind.CallOpts, srcToken common.Address, useSrcWrappers bool) (*big.Int, error) {
	var out []interface{}
	err := _OneInchOracle.contract.Call(opts, &out, "getRateToEth", srcToken, useSrcWrappers)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRateToEth is a free data retrieval call binding the contract method 0x7de4fd10.
//
// Solidity: function getRateToEth(address srcToken, bool useSrcWrappers) view returns(uint256 weightedRate)
func (_OneInchOracle *OneInchOracleSession) GetRateToEth(srcToken common.Address, useSrcWrappers bool) (*big.Int, error) {
	return _OneInchOracle.Contract.GetRateToEth(&_OneInchOracle.CallOpts, srcToken, useSrcWrappers)
}

// GetRateToEth is a free data retrieval call binding the contract method 0x7de4fd10.
//
// Solidity: function getRateToEth(address srcToken, bool useSrcWrappers) view returns(uint256 weightedRate)
func (_OneInchOracle *OneInchOracleCallerSession) GetRateToEth(srcToken common.Address, useSrcWrappers bool) (*big.Int, error) {
	return _OneInchOracle.Contract.GetRateToEth(&_OneInchOracle.CallOpts, srcToken, useSrcWrappers)
}

// MultiWrapper is a free data retrieval call binding the contract method 0xb77910dc.
//
// Solidity: function multiWrapper() view returns(address)
func (_OneInchOracle *OneInchOracleCaller) MultiWrapper(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OneInchOracle.contract.Call(opts, &out, "multiWrapper")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// MultiWrapper is a free data retrieval call binding the contract method 0xb77910dc.
//
// Solidity: function multiWrapper() view returns(address)
func (_OneInchOracle *OneInchOracleSession) MultiWrapper() (common.Address, error) {
	return _OneInchOracle.Contract.MultiWrapper(&_OneInchOracle.CallOpts)
}

// MultiWrapper is a free data retrieval call binding the contract method 0xb77910dc.
//
// Solidity: function multiWrapper() view returns(address)
func (_OneInchOracle *OneInchOracleCallerSession) MultiWrapper() (common.Address, error) {
	return _OneInchOracle.Contract.MultiWrapper(&_OneInchOracle.CallOpts)
}

// Oracles is a free data retrieval call binding the contract method 0x2857373a.
//
// Solidity: function oracles() view returns(address[] allOracles, uint8[] oracleTypes)
func (_OneInchOracle *OneInchOracleCaller) Oracles(opts *bind.CallOpts) (struct {
	AllOracles  []common.Address
	OracleTypes []uint8
}, error) {
	var out []interface{}
	err := _OneInchOracle.contract.Call(opts, &out, "oracles")

	outstruct := new(struct {
		AllOracles  []common.Address
		OracleTypes []uint8
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.AllOracles = *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)
	outstruct.OracleTypes = *abi.ConvertType(out[1], new([]uint8)).(*[]uint8)

	return *outstruct, err

}

// Oracles is a free data retrieval call binding the contract method 0x2857373a.
//
// Solidity: function oracles() view returns(address[] allOracles, uint8[] oracleTypes)
func (_OneInchOracle *OneInchOracleSession) Oracles() (struct {
	AllOracles  []common.Address
	OracleTypes []uint8
}, error) {
	return _OneInchOracle.Contract.Oracles(&_OneInchOracle.CallOpts)
}

// Oracles is a free data retrieval call binding the contract method 0x2857373a.
//
// Solidity: function oracles() view returns(address[] allOracles, uint8[] oracleTypes)
func (_OneInchOracle *OneInchOracleCallerSession) Oracles() (struct {
	AllOracles  []common.Address
	OracleTypes []uint8
}, error) {
	return _OneInchOracle.Contract.Oracles(&_OneInchOracle.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_OneInchOracle *OneInchOracleCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OneInchOracle.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_OneInchOracle *OneInchOracleSession) Owner() (common.Address, error) {
	return _OneInchOracle.Contract.Owner(&_OneInchOracle.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_OneInchOracle *OneInchOracleCallerSession) Owner() (common.Address, error) {
	return _OneInchOracle.Contract.Owner(&_OneInchOracle.CallOpts)
}

// AddConnector is a paid mutator transaction binding the contract method 0xaa16d4c0.
//
// Solidity: function addConnector(address connector) returns()
func (_OneInchOracle *OneInchOracleTransactor) AddConnector(opts *bind.TransactOpts, connector common.Address) (*types.Transaction, error) {
	return _OneInchOracle.contract.Transact(opts, "addConnector", connector)
}

// AddConnector is a paid mutator transaction binding the contract method 0xaa16d4c0.
//
// Solidity: function addConnector(address connector) returns()
func (_OneInchOracle *OneInchOracleSession) AddConnector(connector common.Address) (*types.Transaction, error) {
	return _OneInchOracle.Contract.AddConnector(&_OneInchOracle.TransactOpts, connector)
}

// AddConnector is a paid mutator transaction binding the contract method 0xaa16d4c0.
//
// Solidity: function addConnector(address connector) returns()
func (_OneInchOracle *OneInchOracleTransactorSession) AddConnector(connector common.Address) (*types.Transaction, error) {
	return _OneInchOracle.Contract.AddConnector(&_OneInchOracle.TransactOpts, connector)
}

// AddOracle is a paid mutator transaction binding the contract method 0x9d4d7b1c.
//
// Solidity: function addOracle(address oracle, uint8 oracleKind) returns()
func (_OneInchOracle *OneInchOracleTransactor) AddOracle(opts *bind.TransactOpts, oracle common.Address, oracleKind uint8) (*types.Transaction, error) {
	return _OneInchOracle.contract.Transact(opts, "addOracle", oracle, oracleKind)
}

// AddOracle is a paid mutator transaction binding the contract method 0x9d4d7b1c.
//
// Solidity: function addOracle(address oracle, uint8 oracleKind) returns()
func (_OneInchOracle *OneInchOracleSession) AddOracle(oracle common.Address, oracleKind uint8) (*types.Transaction, error) {
	return _OneInchOracle.Contract.AddOracle(&_OneInchOracle.TransactOpts, oracle, oracleKind)
}

// AddOracle is a paid mutator transaction binding the contract method 0x9d4d7b1c.
//
// Solidity: function addOracle(address oracle, uint8 oracleKind) returns()
func (_OneInchOracle *OneInchOracleTransactorSession) AddOracle(oracle common.Address, oracleKind uint8) (*types.Transaction, error) {
	return _OneInchOracle.Contract.AddOracle(&_OneInchOracle.TransactOpts, oracle, oracleKind)
}

// RemoveConnector is a paid mutator transaction binding the contract method 0x1a6c6a98.
//
// Solidity: function removeConnector(address connector) returns()
func (_OneInchOracle *OneInchOracleTransactor) RemoveConnector(opts *bind.TransactOpts, connector common.Address) (*types.Transaction, error) {
	return _OneInchOracle.contract.Transact(opts, "removeConnector", connector)
}

// RemoveConnector is a paid mutator transaction binding the contract method 0x1a6c6a98.
//
// Solidity: function removeConnector(address connector) returns()
func (_OneInchOracle *OneInchOracleSession) RemoveConnector(connector common.Address) (*types.Transaction, error) {
	return _OneInchOracle.Contract.RemoveConnector(&_OneInchOracle.TransactOpts, connector)
}

// RemoveConnector is a paid mutator transaction binding the contract method 0x1a6c6a98.
//
// Solidity: function removeConnector(address connector) returns()
func (_OneInchOracle *OneInchOracleTransactorSession) RemoveConnector(connector common.Address) (*types.Transaction, error) {
	return _OneInchOracle.Contract.RemoveConnector(&_OneInchOracle.TransactOpts, connector)
}

// RemoveOracle is a paid mutator transaction binding the contract method 0xf0b92e40.
//
// Solidity: function removeOracle(address oracle, uint8 oracleKind) returns()
func (_OneInchOracle *OneInchOracleTransactor) RemoveOracle(opts *bind.TransactOpts, oracle common.Address, oracleKind uint8) (*types.Transaction, error) {
	return _OneInchOracle.contract.Transact(opts, "removeOracle", oracle, oracleKind)
}

// RemoveOracle is a paid mutator transaction binding the contract method 0xf0b92e40.
//
// Solidity: function removeOracle(address oracle, uint8 oracleKind) returns()
func (_OneInchOracle *OneInchOracleSession) RemoveOracle(oracle common.Address, oracleKind uint8) (*types.Transaction, error) {
	return _OneInchOracle.Contract.RemoveOracle(&_OneInchOracle.TransactOpts, oracle, oracleKind)
}

// RemoveOracle is a paid mutator transaction binding the contract method 0xf0b92e40.
//
// Solidity: function removeOracle(address oracle, uint8 oracleKind) returns()
func (_OneInchOracle *OneInchOracleTransactorSession) RemoveOracle(oracle common.Address, oracleKind uint8) (*types.Transaction, error) {
	return _OneInchOracle.Contract.RemoveOracle(&_OneInchOracle.TransactOpts, oracle, oracleKind)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_OneInchOracle *OneInchOracleTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OneInchOracle.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_OneInchOracle *OneInchOracleSession) RenounceOwnership() (*types.Transaction, error) {
	return _OneInchOracle.Contract.RenounceOwnership(&_OneInchOracle.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_OneInchOracle *OneInchOracleTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _OneInchOracle.Contract.RenounceOwnership(&_OneInchOracle.TransactOpts)
}

// SetMultiWrapper is a paid mutator transaction binding the contract method 0xd0626518.
//
// Solidity: function setMultiWrapper(address _multiWrapper) returns()
func (_OneInchOracle *OneInchOracleTransactor) SetMultiWrapper(opts *bind.TransactOpts, _multiWrapper common.Address) (*types.Transaction, error) {
	return _OneInchOracle.contract.Transact(opts, "setMultiWrapper", _multiWrapper)
}

// SetMultiWrapper is a paid mutator transaction binding the contract method 0xd0626518.
//
// Solidity: function setMultiWrapper(address _multiWrapper) returns()
func (_OneInchOracle *OneInchOracleSession) SetMultiWrapper(_multiWrapper common.Address) (*types.Transaction, error) {
	return _OneInchOracle.Contract.SetMultiWrapper(&_OneInchOracle.TransactOpts, _multiWrapper)
}

// SetMultiWrapper is a paid mutator transaction binding the contract method 0xd0626518.
//
// Solidity: function setMultiWrapper(address _multiWrapper) returns()
func (_OneInchOracle *OneInchOracleTransactorSession) SetMultiWrapper(_multiWrapper common.Address) (*types.Transaction, error) {
	return _OneInchOracle.Contract.SetMultiWrapper(&_OneInchOracle.TransactOpts, _multiWrapper)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_OneInchOracle *OneInchOracleTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _OneInchOracle.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_OneInchOracle *OneInchOracleSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _OneInchOracle.Contract.TransferOwnership(&_OneInchOracle.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_OneInchOracle *OneInchOracleTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _OneInchOracle.Contract.TransferOwnership(&_OneInchOracle.TransactOpts, newOwner)
}

// OneInchOracleConnectorAddedIterator is returned from FilterConnectorAdded and is used to iterate over the raw logs and unpacked data for ConnectorAdded events raised by the OneInchOracle contract.
type OneInchOracleConnectorAddedIterator struct {
	Event *OneInchOracleConnectorAdded // Event containing the contract specifics and raw log

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
func (it *OneInchOracleConnectorAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OneInchOracleConnectorAdded)
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
		it.Event = new(OneInchOracleConnectorAdded)
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
func (it *OneInchOracleConnectorAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OneInchOracleConnectorAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OneInchOracleConnectorAdded represents a ConnectorAdded event raised by the OneInchOracle contract.
type OneInchOracleConnectorAdded struct {
	Connector common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterConnectorAdded is a free log retrieval operation binding the contract event 0xff88af5d962d47fd25d87755e8267a029fad5a91740c67d0dade2bdbe5268a1d.
//
// Solidity: event ConnectorAdded(address connector)
func (_OneInchOracle *OneInchOracleFilterer) FilterConnectorAdded(opts *bind.FilterOpts) (*OneInchOracleConnectorAddedIterator, error) {

	logs, sub, err := _OneInchOracle.contract.FilterLogs(opts, "ConnectorAdded")
	if err != nil {
		return nil, err
	}
	return &OneInchOracleConnectorAddedIterator{contract: _OneInchOracle.contract, event: "ConnectorAdded", logs: logs, sub: sub}, nil
}

// WatchConnectorAdded is a free log subscription operation binding the contract event 0xff88af5d962d47fd25d87755e8267a029fad5a91740c67d0dade2bdbe5268a1d.
//
// Solidity: event ConnectorAdded(address connector)
func (_OneInchOracle *OneInchOracleFilterer) WatchConnectorAdded(opts *bind.WatchOpts, sink chan<- *OneInchOracleConnectorAdded) (event.Subscription, error) {

	logs, sub, err := _OneInchOracle.contract.WatchLogs(opts, "ConnectorAdded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OneInchOracleConnectorAdded)
				if err := _OneInchOracle.contract.UnpackLog(event, "ConnectorAdded", log); err != nil {
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

// ParseConnectorAdded is a log parse operation binding the contract event 0xff88af5d962d47fd25d87755e8267a029fad5a91740c67d0dade2bdbe5268a1d.
//
// Solidity: event ConnectorAdded(address connector)
func (_OneInchOracle *OneInchOracleFilterer) ParseConnectorAdded(log types.Log) (*OneInchOracleConnectorAdded, error) {
	event := new(OneInchOracleConnectorAdded)
	if err := _OneInchOracle.contract.UnpackLog(event, "ConnectorAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OneInchOracleConnectorRemovedIterator is returned from FilterConnectorRemoved and is used to iterate over the raw logs and unpacked data for ConnectorRemoved events raised by the OneInchOracle contract.
type OneInchOracleConnectorRemovedIterator struct {
	Event *OneInchOracleConnectorRemoved // Event containing the contract specifics and raw log

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
func (it *OneInchOracleConnectorRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OneInchOracleConnectorRemoved)
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
		it.Event = new(OneInchOracleConnectorRemoved)
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
func (it *OneInchOracleConnectorRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OneInchOracleConnectorRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OneInchOracleConnectorRemoved represents a ConnectorRemoved event raised by the OneInchOracle contract.
type OneInchOracleConnectorRemoved struct {
	Connector common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterConnectorRemoved is a free log retrieval operation binding the contract event 0x6825b26a0827e9c2ceca01d6289ce4a40e629dc074ec48ea4727d1afbff359f5.
//
// Solidity: event ConnectorRemoved(address connector)
func (_OneInchOracle *OneInchOracleFilterer) FilterConnectorRemoved(opts *bind.FilterOpts) (*OneInchOracleConnectorRemovedIterator, error) {

	logs, sub, err := _OneInchOracle.contract.FilterLogs(opts, "ConnectorRemoved")
	if err != nil {
		return nil, err
	}
	return &OneInchOracleConnectorRemovedIterator{contract: _OneInchOracle.contract, event: "ConnectorRemoved", logs: logs, sub: sub}, nil
}

// WatchConnectorRemoved is a free log subscription operation binding the contract event 0x6825b26a0827e9c2ceca01d6289ce4a40e629dc074ec48ea4727d1afbff359f5.
//
// Solidity: event ConnectorRemoved(address connector)
func (_OneInchOracle *OneInchOracleFilterer) WatchConnectorRemoved(opts *bind.WatchOpts, sink chan<- *OneInchOracleConnectorRemoved) (event.Subscription, error) {

	logs, sub, err := _OneInchOracle.contract.WatchLogs(opts, "ConnectorRemoved")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OneInchOracleConnectorRemoved)
				if err := _OneInchOracle.contract.UnpackLog(event, "ConnectorRemoved", log); err != nil {
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

// ParseConnectorRemoved is a log parse operation binding the contract event 0x6825b26a0827e9c2ceca01d6289ce4a40e629dc074ec48ea4727d1afbff359f5.
//
// Solidity: event ConnectorRemoved(address connector)
func (_OneInchOracle *OneInchOracleFilterer) ParseConnectorRemoved(log types.Log) (*OneInchOracleConnectorRemoved, error) {
	event := new(OneInchOracleConnectorRemoved)
	if err := _OneInchOracle.contract.UnpackLog(event, "ConnectorRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OneInchOracleMultiWrapperUpdatedIterator is returned from FilterMultiWrapperUpdated and is used to iterate over the raw logs and unpacked data for MultiWrapperUpdated events raised by the OneInchOracle contract.
type OneInchOracleMultiWrapperUpdatedIterator struct {
	Event *OneInchOracleMultiWrapperUpdated // Event containing the contract specifics and raw log

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
func (it *OneInchOracleMultiWrapperUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OneInchOracleMultiWrapperUpdated)
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
		it.Event = new(OneInchOracleMultiWrapperUpdated)
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
func (it *OneInchOracleMultiWrapperUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OneInchOracleMultiWrapperUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OneInchOracleMultiWrapperUpdated represents a MultiWrapperUpdated event raised by the OneInchOracle contract.
type OneInchOracleMultiWrapperUpdated struct {
	MultiWrapper common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterMultiWrapperUpdated is a free log retrieval operation binding the contract event 0x1030152fe2062b574a830e6b9f13c65995990df31e4dc708d142533bb3ad0f52.
//
// Solidity: event MultiWrapperUpdated(address multiWrapper)
func (_OneInchOracle *OneInchOracleFilterer) FilterMultiWrapperUpdated(opts *bind.FilterOpts) (*OneInchOracleMultiWrapperUpdatedIterator, error) {

	logs, sub, err := _OneInchOracle.contract.FilterLogs(opts, "MultiWrapperUpdated")
	if err != nil {
		return nil, err
	}
	return &OneInchOracleMultiWrapperUpdatedIterator{contract: _OneInchOracle.contract, event: "MultiWrapperUpdated", logs: logs, sub: sub}, nil
}

// WatchMultiWrapperUpdated is a free log subscription operation binding the contract event 0x1030152fe2062b574a830e6b9f13c65995990df31e4dc708d142533bb3ad0f52.
//
// Solidity: event MultiWrapperUpdated(address multiWrapper)
func (_OneInchOracle *OneInchOracleFilterer) WatchMultiWrapperUpdated(opts *bind.WatchOpts, sink chan<- *OneInchOracleMultiWrapperUpdated) (event.Subscription, error) {

	logs, sub, err := _OneInchOracle.contract.WatchLogs(opts, "MultiWrapperUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OneInchOracleMultiWrapperUpdated)
				if err := _OneInchOracle.contract.UnpackLog(event, "MultiWrapperUpdated", log); err != nil {
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

// ParseMultiWrapperUpdated is a log parse operation binding the contract event 0x1030152fe2062b574a830e6b9f13c65995990df31e4dc708d142533bb3ad0f52.
//
// Solidity: event MultiWrapperUpdated(address multiWrapper)
func (_OneInchOracle *OneInchOracleFilterer) ParseMultiWrapperUpdated(log types.Log) (*OneInchOracleMultiWrapperUpdated, error) {
	event := new(OneInchOracleMultiWrapperUpdated)
	if err := _OneInchOracle.contract.UnpackLog(event, "MultiWrapperUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OneInchOracleOracleAddedIterator is returned from FilterOracleAdded and is used to iterate over the raw logs and unpacked data for OracleAdded events raised by the OneInchOracle contract.
type OneInchOracleOracleAddedIterator struct {
	Event *OneInchOracleOracleAdded // Event containing the contract specifics and raw log

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
func (it *OneInchOracleOracleAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OneInchOracleOracleAdded)
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
		it.Event = new(OneInchOracleOracleAdded)
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
func (it *OneInchOracleOracleAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OneInchOracleOracleAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OneInchOracleOracleAdded represents a OracleAdded event raised by the OneInchOracle contract.
type OneInchOracleOracleAdded struct {
	Oracle     common.Address
	OracleType uint8
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterOracleAdded is a free log retrieval operation binding the contract event 0x5874b2072ff37562df54063dd700c59d45f311bdf6f9cabb5a15f0ffb2e9f622.
//
// Solidity: event OracleAdded(address oracle, uint8 oracleType)
func (_OneInchOracle *OneInchOracleFilterer) FilterOracleAdded(opts *bind.FilterOpts) (*OneInchOracleOracleAddedIterator, error) {

	logs, sub, err := _OneInchOracle.contract.FilterLogs(opts, "OracleAdded")
	if err != nil {
		return nil, err
	}
	return &OneInchOracleOracleAddedIterator{contract: _OneInchOracle.contract, event: "OracleAdded", logs: logs, sub: sub}, nil
}

// WatchOracleAdded is a free log subscription operation binding the contract event 0x5874b2072ff37562df54063dd700c59d45f311bdf6f9cabb5a15f0ffb2e9f622.
//
// Solidity: event OracleAdded(address oracle, uint8 oracleType)
func (_OneInchOracle *OneInchOracleFilterer) WatchOracleAdded(opts *bind.WatchOpts, sink chan<- *OneInchOracleOracleAdded) (event.Subscription, error) {

	logs, sub, err := _OneInchOracle.contract.WatchLogs(opts, "OracleAdded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OneInchOracleOracleAdded)
				if err := _OneInchOracle.contract.UnpackLog(event, "OracleAdded", log); err != nil {
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

// ParseOracleAdded is a log parse operation binding the contract event 0x5874b2072ff37562df54063dd700c59d45f311bdf6f9cabb5a15f0ffb2e9f622.
//
// Solidity: event OracleAdded(address oracle, uint8 oracleType)
func (_OneInchOracle *OneInchOracleFilterer) ParseOracleAdded(log types.Log) (*OneInchOracleOracleAdded, error) {
	event := new(OneInchOracleOracleAdded)
	if err := _OneInchOracle.contract.UnpackLog(event, "OracleAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OneInchOracleOracleRemovedIterator is returned from FilterOracleRemoved and is used to iterate over the raw logs and unpacked data for OracleRemoved events raised by the OneInchOracle contract.
type OneInchOracleOracleRemovedIterator struct {
	Event *OneInchOracleOracleRemoved // Event containing the contract specifics and raw log

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
func (it *OneInchOracleOracleRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OneInchOracleOracleRemoved)
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
		it.Event = new(OneInchOracleOracleRemoved)
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
func (it *OneInchOracleOracleRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OneInchOracleOracleRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OneInchOracleOracleRemoved represents a OracleRemoved event raised by the OneInchOracle contract.
type OneInchOracleOracleRemoved struct {
	Oracle     common.Address
	OracleType uint8
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterOracleRemoved is a free log retrieval operation binding the contract event 0x7a7f56716fe703fb190529c336e57df71ab88188ba47e8d786bac684b61ab9a6.
//
// Solidity: event OracleRemoved(address oracle, uint8 oracleType)
func (_OneInchOracle *OneInchOracleFilterer) FilterOracleRemoved(opts *bind.FilterOpts) (*OneInchOracleOracleRemovedIterator, error) {

	logs, sub, err := _OneInchOracle.contract.FilterLogs(opts, "OracleRemoved")
	if err != nil {
		return nil, err
	}
	return &OneInchOracleOracleRemovedIterator{contract: _OneInchOracle.contract, event: "OracleRemoved", logs: logs, sub: sub}, nil
}

// WatchOracleRemoved is a free log subscription operation binding the contract event 0x7a7f56716fe703fb190529c336e57df71ab88188ba47e8d786bac684b61ab9a6.
//
// Solidity: event OracleRemoved(address oracle, uint8 oracleType)
func (_OneInchOracle *OneInchOracleFilterer) WatchOracleRemoved(opts *bind.WatchOpts, sink chan<- *OneInchOracleOracleRemoved) (event.Subscription, error) {

	logs, sub, err := _OneInchOracle.contract.WatchLogs(opts, "OracleRemoved")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OneInchOracleOracleRemoved)
				if err := _OneInchOracle.contract.UnpackLog(event, "OracleRemoved", log); err != nil {
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

// ParseOracleRemoved is a log parse operation binding the contract event 0x7a7f56716fe703fb190529c336e57df71ab88188ba47e8d786bac684b61ab9a6.
//
// Solidity: event OracleRemoved(address oracle, uint8 oracleType)
func (_OneInchOracle *OneInchOracleFilterer) ParseOracleRemoved(log types.Log) (*OneInchOracleOracleRemoved, error) {
	event := new(OneInchOracleOracleRemoved)
	if err := _OneInchOracle.contract.UnpackLog(event, "OracleRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OneInchOracleOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the OneInchOracle contract.
type OneInchOracleOwnershipTransferredIterator struct {
	Event *OneInchOracleOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *OneInchOracleOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OneInchOracleOwnershipTransferred)
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
		it.Event = new(OneInchOracleOwnershipTransferred)
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
func (it *OneInchOracleOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OneInchOracleOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OneInchOracleOwnershipTransferred represents a OwnershipTransferred event raised by the OneInchOracle contract.
type OneInchOracleOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_OneInchOracle *OneInchOracleFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*OneInchOracleOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _OneInchOracle.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &OneInchOracleOwnershipTransferredIterator{contract: _OneInchOracle.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_OneInchOracle *OneInchOracleFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *OneInchOracleOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _OneInchOracle.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OneInchOracleOwnershipTransferred)
				if err := _OneInchOracle.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_OneInchOracle *OneInchOracleFilterer) ParseOwnershipTransferred(log types.Log) (*OneInchOracleOwnershipTransferred, error) {
	event := new(OneInchOracleOwnershipTransferred)
	if err := _OneInchOracle.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
