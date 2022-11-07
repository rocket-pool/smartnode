package tokens

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

//
// Core ERC-20 functions
//

// Get RPL total supply
func GetRPLTotalSupply(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketTokenRPL, err := getRocketTokenRPL(rp, opts)
	if err != nil {
		return nil, err
	}
	return totalSupply(rocketTokenRPL, "RPL", opts)
}

// Get RPL balance
func GetRPLBalance(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketTokenRPL, err := getRocketTokenRPL(rp, opts)
	if err != nil {
		return nil, err
	}
	return balanceOf(rocketTokenRPL, "RPL", address, opts)
}

// Get RPL allowance
func GetRPLAllowance(rp *rocketpool.RocketPool, owner, spender common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketTokenRPL, err := getRocketTokenRPL(rp, opts)
	if err != nil {
		return nil, err
	}
	return allowance(rocketTokenRPL, "RPL", owner, spender, opts)
}

// Estimate the gas of TransferRPL
func EstimateTransferRPLGas(rp *rocketpool.RocketPool, to common.Address, amount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketTokenRPL, err := getRocketTokenRPL(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return estimateTransferGas(rocketTokenRPL, "RPL", to, amount, opts)
}

// Transfer RPL
func TransferRPL(rp *rocketpool.RocketPool, to common.Address, amount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketTokenRPL, err := getRocketTokenRPL(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	return transfer(rocketTokenRPL, "RPL", to, amount, opts)
}

// Estimate the gas of ApproveRPL
func EstimateApproveRPLGas(rp *rocketpool.RocketPool, spender common.Address, amount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketTokenRPL, err := getRocketTokenRPL(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return estimateApproveGas(rocketTokenRPL, "RPL", spender, amount, opts)
}

// Approve an RPL spender
func ApproveRPL(rp *rocketpool.RocketPool, spender common.Address, amount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketTokenRPL, err := getRocketTokenRPL(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	return approve(rocketTokenRPL, "RPL", spender, amount, opts)
}

// Estimate the gas of TransferFromRPL
func EstimateTransferFromRPLGas(rp *rocketpool.RocketPool, from, to common.Address, amount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketTokenRPL, err := getRocketTokenRPL(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return estimateTransferFromGas(rocketTokenRPL, "RPL", from, to, amount, opts)
}

// Transfer RPL from a sender
func TransferFromRPL(rp *rocketpool.RocketPool, from, to common.Address, amount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketTokenRPL, err := getRocketTokenRPL(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	return transferFrom(rocketTokenRPL, "RPL", from, to, amount, opts)
}

//
// RPL functions
//

// Estimate the gas of MintInflationRPL
func EstimateMintInflationRPLGas(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketTokenRPL, err := getRocketTokenRPL(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketTokenRPL.GetTransactionGasInfo(opts, "inflationMintTokens")
}

// Mint new RPL tokens from inflation
func MintInflationRPL(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketTokenRPL, err := getRocketTokenRPL(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketTokenRPL.Transact(opts, "inflationMintTokens")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not mint RPL tokens from inflation: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of SwapFixedSupplyRPLForRPL
func EstimateSwapFixedSupplyRPLForRPLGas(rp *rocketpool.RocketPool, amount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketTokenRPL, err := getRocketTokenRPL(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketTokenRPL.GetTransactionGasInfo(opts, "swapTokens", amount)
}

// Swap fixed-supply RPL for new RPL tokens
func SwapFixedSupplyRPLForRPL(rp *rocketpool.RocketPool, amount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketTokenRPL, err := getRocketTokenRPL(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketTokenRPL.Transact(opts, "swapTokens", amount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not swap fixed-supply RPL for new RPL: %w", err)
	}
	return tx.Hash(), nil
}

// Get the RPL inflation interval rate
func GetRPLInflationIntervalRate(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketTokenRPL, err := getRocketTokenRPL(rp, opts)
	if err != nil {
		return nil, err
	}
	rate := new(*big.Int)
	if err := rocketTokenRPL.Call(opts, rate, "getInflationIntervalRate"); err != nil {
		return nil, fmt.Errorf("Could not get RPL inflation interval rate: %w", err)
	}
	return *rate, nil
}

//
// Contracts
//

// Get contracts
var rocketTokenRPLLock sync.Mutex

func getRocketTokenRPL(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketTokenRPLLock.Lock()
	defer rocketTokenRPLLock.Unlock()
	return rp.GetContract("rocketTokenRPL", opts)
}
