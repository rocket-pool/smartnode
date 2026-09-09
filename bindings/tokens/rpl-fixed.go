package tokens

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
	"github.com/rocket-pool/smartnode/shared/units"
)

//
// Core ERC-20 functions
//

// Get fixed-supply RPL total supply
func GetFixedSupplyRPLTotalSupply(rp *rocketpool.RocketPool, opts *bind.CallOpts) (units.Wei, error) {
	rocketTokenFixedSupplyRPL, err := getRocketTokenRPLFixedSupply(rp, opts)
	if err != nil {
		return units.Wei{}, err
	}
	return totalSupply(rocketTokenFixedSupplyRPL, "fixed-supply RPL", opts)
}

// Get fixed-supply RPL balance
func GetFixedSupplyRPLBalance(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (units.Wei, error) {
	rocketTokenFixedSupplyRPL, err := getRocketTokenRPLFixedSupply(rp, opts)
	if err != nil {
		return units.Wei{}, err
	}
	b, err := balanceOf(rocketTokenFixedSupplyRPL, "fixed-supply RPL", address, opts)
	if err != nil {
		return units.Wei{}, err
	}
	return b, nil
}

// Get fixed-supply RPL allowance
func GetFixedSupplyRPLAllowance(rp *rocketpool.RocketPool, owner, spender common.Address, opts *bind.CallOpts) (units.Wei, error) {
	rocketTokenFixedSupplyRPL, err := getRocketTokenRPLFixedSupply(rp, opts)
	if err != nil {
		return units.Wei{}, err
	}
	return allowance(rocketTokenFixedSupplyRPL, "fixed-supply RPL", owner, spender, opts)
}

// Estimate the gas of TransferFixedSupplyRPL
func EstimateTransferFixedSupplyRPLGas(rp *rocketpool.RocketPool, to common.Address, amount units.Wei, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	rocketTokenFixedSupplyRPL, err := getRocketTokenRPLFixedSupply(rp, nil)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	return estimateTransferGas(rocketTokenFixedSupplyRPL, "fixed-supply RPL", to, amount, opts)
}

// Transfer fixed-supply RPL
func TransferFixedSupplyRPL(rp *rocketpool.RocketPool, to common.Address, amount units.Wei, opts *bind.TransactOpts) (common.Hash, error) {
	rocketTokenFixedSupplyRPL, err := getRocketTokenRPLFixedSupply(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	return transfer(rocketTokenFixedSupplyRPL, "fixed-supply RPL", to, amount, opts)
}

// Estimate the gas of ApproveFixedSupplyRPL
func EstimateApproveFixedSupplyRPLGas(rp *rocketpool.RocketPool, spender common.Address, amount units.Wei, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	rocketTokenFixedSupplyRPL, err := getRocketTokenRPLFixedSupply(rp, nil)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	return estimateApproveGas(rocketTokenFixedSupplyRPL, "fixed-supply RPL", spender, amount, opts)
}

// Approve a fixed-supply RPL spender
func ApproveFixedSupplyRPL(rp *rocketpool.RocketPool, spender common.Address, amount units.Wei, opts *bind.TransactOpts) (common.Hash, error) {
	rocketTokenFixedSupplyRPL, err := getRocketTokenRPLFixedSupply(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	return approve(rocketTokenFixedSupplyRPL, "fixed-supply RPL", spender, amount, opts)
}

// Estimate the gas of TransferFromFixedSupplyRPL
func EstimateTransferFromFixedSupplyRPLGas(rp *rocketpool.RocketPool, from, to common.Address, amount *big.Int, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	rocketTokenFixedSupplyRPL, err := getRocketTokenRPLFixedSupply(rp, nil)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	return estimateTransferFromGas(rocketTokenFixedSupplyRPL, "fixed-supply RPL", from, to, amount, opts)
}

// Transfer fixed-supply RPL from a sender
func TransferFromFixedSupplyRPL(rp *rocketpool.RocketPool, from, to common.Address, amount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketTokenFixedSupplyRPL, err := getRocketTokenRPLFixedSupply(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	return transferFrom(rocketTokenFixedSupplyRPL, "fixed-supply RPL", from, to, amount, opts)
}

//
// Contracts
//

// Get contracts
var rocketTokenFixedSupplyRPLLock sync.Mutex

func getRocketTokenRPLFixedSupply(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketTokenFixedSupplyRPLLock.Lock()
	defer rocketTokenFixedSupplyRPLLock.Unlock()
	return rp.GetContract("rocketTokenRPLFixedSupply", opts)
}
