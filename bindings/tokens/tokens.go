package tokens

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
	"github.com/rocket-pool/smartnode/shared/units"
)

// Token balances
type Balances struct {
	ETH            units.Wei `json:"eth"`
	RETH           units.Wei `json:"reth"`
	RPL            units.Wei `json:"rpl"`
	FixedSupplyRPL units.Wei `json:"fixedSupplyRpl"`
}

// Get token balances of an address
func GetBalances(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (Balances, error) {

	// Get call options block number
	var blockNumber *big.Int
	if opts != nil {
		blockNumber = opts.BlockNumber
	}

	// Data
	var wg errgroup.Group
	var ethBalance units.Wei
	var rethBalance units.Wei
	var rplBalance units.Wei
	var fixedSupplyRplBalance units.Wei

	// Load data
	wg.Go(func() error {
		var err error
		ethBalance, err = rp.Client.BalanceAt(context.Background(), address, blockNumber)
		return err
	})
	wg.Go(func() error {
		var err error
		rethBalance, err = GetRETHBalance(rp, address, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		rplBalance, err = GetRPLBalance(rp, address, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		fixedSupplyRplBalance, err = GetFixedSupplyRPLBalance(rp, address, opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return Balances{}, err
	}

	// Return
	return Balances{
		ETH:            ethBalance,
		RETH:           rethBalance,
		RPL:            rplBalance,
		FixedSupplyRPL: fixedSupplyRplBalance,
	}, nil

}

// Get a token contract's ETH balance
func contractETHBalance(rp *rocketpool.RocketPool, tokenContract *rocketpool.Contract, opts *bind.CallOpts) (units.Wei, error) {
	var blockNumber *big.Int
	if opts != nil {
		blockNumber = opts.BlockNumber
	}
	return rp.Client.BalanceAt(context.Background(), *(tokenContract.Address), blockNumber)
}

// Get a token's total supply
func totalSupply(tokenContract *rocketpool.Contract, tokenName string, opts *bind.CallOpts) (units.Wei, error) {
	totalSupply := units.Wei{}
	if err := tokenContract.Call(opts, &totalSupply, "totalSupply"); err != nil {
		return units.Wei{}, fmt.Errorf("error getting %s total supply: %w", tokenName, err)
	}
	return totalSupply, nil
}

// Get a token balance
func balanceOf(tokenContract *rocketpool.Contract, tokenName string, address common.Address, opts *bind.CallOpts) (units.Wei, error) {
	var balance units.Wei
	if err := tokenContract.Call(opts, &balance, "balanceOf", address); err != nil {
		return units.Wei{}, fmt.Errorf("error getting %s balance of %s: %w", tokenName, address.Hex(), err)
	}
	return balance, nil
}

// Get a spender's allowance for an address
func allowance(tokenContract *rocketpool.Contract, tokenName string, owner, spender common.Address, opts *bind.CallOpts) (units.Wei, error) {
	allowance := new(units.Wei)
	if err := tokenContract.Call(opts, allowance, "allowance", owner, spender); err != nil {
		return units.Wei{}, fmt.Errorf("error getting %s allowance of %s for %s: %w", tokenName, spender.Hex(), owner.Hex(), err)
	}
	return *allowance, nil
}

// Estimate the gas of transfer
func estimateTransferGas(tokenContract *rocketpool.Contract, tokenName string, to common.Address, amount units.Wei, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return tokenContract.GetTransactionGasInfo(opts, "transfer", to, amount)
}

// Transfer tokens to an address
func transfer(tokenContract *rocketpool.Contract, tokenName string, to common.Address, amount units.Wei, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := tokenContract.Transact(opts, "transfer", to, amount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error transferring %s to %s: %w", tokenName, to.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of approve
func estimateApproveGas(tokenContract *rocketpool.Contract, tokenName string, spender common.Address, amount units.Wei, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return tokenContract.GetTransactionGasInfo(opts, "approve", spender, amount)
}

// Approve a token allowance for a spender
func approve(tokenContract *rocketpool.Contract, tokenName string, spender common.Address, amount units.Wei, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := tokenContract.Transact(opts, "approve", spender, amount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error approving %s allowance for %s: %w", tokenName, spender.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of transferFrom
func estimateTransferFromGas(tokenContract *rocketpool.Contract, tokenName string, from, to common.Address, amount *big.Int, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return tokenContract.GetTransactionGasInfo(opts, "transferFrom", from, to, amount)
}

// Transfer tokens from a sender to an address
func transferFrom(tokenContract *rocketpool.Contract, tokenName string, from, to common.Address, amount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := tokenContract.Transact(opts, "transferFrom", from, to, amount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error transferring %s from %s to %s: %w", tokenName, from.Hex(), to.Hex(), err)
	}
	return tx.Hash(), nil
}
