package eth

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

const (
	Erc20AbiString string = `[
		{
			"constant": true,
			"inputs": [],
			"name": "name",
			"outputs": [
			{
				"name": "",
				"type": "string"
			}
			],
			"payable": false,
			"type": "function"
		},
		{
			"constant": true,
			"inputs": [],
			"name": "decimals",
			"outputs": [
			{
				"name": "",
				"type": "uint8"
			}
			],
			"payable": false,
			"type": "function"
		},
		{
			"constant": true,
			"inputs": [
			{
				"name": "_owner",
				"type": "address"
			}
			],
			"name": "balanceOf",
			"outputs": [
			{
				"name": "balance",
				"type": "uint256"
			}
			],
			"payable": false,
			"type": "function"
		},
		{
			"constant": true,
			"inputs": [],
			"name": "symbol",
			"outputs": [
			{
				"name": "",
				"type": "string"
			}
			],
			"payable": false,
			"type": "function"
		},
		{
			"constant": false,
			"inputs": [
			{
				"name": "_from",
				"type": "address"
			},
			{
				"name": "_to",
				"type": "address"
			},
			{
				"name": "_value",
				"type": "uint256"
			}
			],
			"name": "transferFrom",
			"outputs": [
			{
				"name": "success",
				"type": "bool"
			}
			],
			"payable": false,
			"type": "function"
		},
	]`
)

// Global container for the parsed ABI above
var erc20Abi *abi.ABI

type Erc20Contract struct {
	Name     string
	Symbol   string
	Decimals uint64
	contract *rocketpool.Contract
}

// Creates a contract wrapper for the ERC20 at the given address
func NewErc20Contract(address common.Address, client rocketpool.ExecutionClient, opts *bind.CallOpts) (*Erc20Contract, error) {
	// Parse the ABI
	if erc20Abi == nil {
		abiParsed, err := abi.JSON(strings.NewReader(Erc20AbiString))
		if err != nil {
			return nil, fmt.Errorf("error parsing ERC20 ABI: %w", err)
		}
		erc20Abi = &abiParsed
	}

	// Create contract
	contract := &rocketpool.Contract{
		Contract: bind.NewBoundContract(address, *erc20Abi, client, client, client),
		Address:  &address,
		ABI:      erc20Abi,
		Client:   client,
	}

	// Create the wrapper
	wrapper := &Erc20Contract{
		contract: contract,
	}

	// Get the details
	name, err := wrapper.GetName(opts)
	if err != nil {
		return nil, err
	}
	wrapper.Name = name
	symbol, err := wrapper.GetSymbol(opts)
	if err != nil {
		return nil, err
	}
	wrapper.Symbol = symbol
	decimals, err := wrapper.GetDecimals(opts)
	if err != nil {
		return nil, err
	}
	wrapper.Decimals = decimals

	return wrapper, nil
}

// Get the token name
func (c *Erc20Contract) GetName(opts *bind.CallOpts) (string, error) {
	name := new(string)
	err := c.contract.Call(opts, name, "name")
	if err != nil {
		return "", fmt.Errorf("could not get ERC20 name: %w", err)
	}
	return *name, nil
}

// Get the token symbol
func (c *Erc20Contract) GetSymbol(opts *bind.CallOpts) (string, error) {
	symbol := new(string)
	err := c.contract.Call(opts, symbol, "symbol")
	if err != nil {
		return "", fmt.Errorf("could not get ERC20 symbol: %w", err)
	}
	return *symbol, nil
}

// Get the token decimals
func (c *Erc20Contract) GetDecimals(opts *bind.CallOpts) (uint64, error) {
	decimals := new(*big.Int)
	err := c.contract.Call(opts, decimals, "decimals")
	if err != nil {
		return 0, fmt.Errorf("could not get ERC20 decimals: %w", err)
	}
	return (*decimals).Uint64(), nil
}

// Get the token balance for an address
func (c *Erc20Contract) BalanceOf(address common.Address, opts *bind.CallOpts) (*big.Int, error) {
	balance := new(*big.Int)
	err := c.contract.Call(opts, balance, "balanceOf", address)
	if err != nil {
		return nil, fmt.Errorf("could not get ERC20 balance for address %s: %w", address.Hex(), err)
	}
	return *balance, nil
}

// Transfer an ERC20 from one address to another
func (c *Erc20Contract) TransferFrom(from common.Address, to common.Address, amount *big.Int, opts *bind.TransactOpts) (*types.Transaction, error) {
	tx, err := c.contract.Transact(opts, "transferFrom", from, to, amount)
	if err != nil {
		return nil, fmt.Errorf("could not transfer ERC20 from %s to %s: %w", from.Hex(), to.Hex(), err)
	}
	return tx, nil
}
