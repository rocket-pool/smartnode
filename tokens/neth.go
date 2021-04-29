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

// Get nETH total supply
func GetNETHTotalSupply(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketTokenNETH, err := getRocketTokenNETH(rp)
    if err != nil {
        return nil, err
    }
    return totalSupply(rocketTokenNETH, "nETH", opts)
}


// Get nETH balance
func GetNETHBalance(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (*big.Int, error) {
    rocketTokenNETH, err := getRocketTokenNETH(rp)
    if err != nil {
        return nil, err
    }
    return balanceOf(rocketTokenNETH, "nETH", address, opts)
}


// Get nETH allowance
func GetNETHAllowance(rp *rocketpool.RocketPool, owner, spender common.Address, opts *bind.CallOpts) (*big.Int, error) {
    rocketTokenNETH, err := getRocketTokenNETH(rp)
    if err != nil {
        return nil, err
    }
    return allowance(rocketTokenNETH, "nETH", owner, spender, opts)
}


// Transfer nETH
func TransferNETH(rp *rocketpool.RocketPool, to common.Address, amount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
    rocketTokenNETH, err := getRocketTokenNETH(rp)
    if err != nil {
        return common.Hash{}, err
    }
    return transfer(rocketTokenNETH, "nETH", to, amount, opts)
}


// Approve a nETH spender
func ApproveNETH(rp *rocketpool.RocketPool, spender common.Address, amount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
    rocketTokenNETH, err := getRocketTokenNETH(rp)
    if err != nil {
        return common.Hash{}, err
    }
    return approve(rocketTokenNETH, "nETH", spender, amount, opts)
}


// Transfer nETH from a sender
func TransferFromNETH(rp *rocketpool.RocketPool, from, to common.Address, amount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
    rocketTokenNETH, err := getRocketTokenNETH(rp)
    if err != nil {
        return common.Hash{}, err
    }
    return transferFrom(rocketTokenNETH, "nETH", from, to, amount, opts)
}


//
// nETH functions
//


// Get the nETH contract ETH balance
func GetNETHContractETHBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketTokenNETH, err := getRocketTokenNETH(rp)
    if err != nil {
        return nil, err
    }
    return contractETHBalance(rp, rocketTokenNETH, opts)
}


// Burn nETH for ETH
func BurnNETH(rp *rocketpool.RocketPool, amount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
    rocketTokenNETH, err := getRocketTokenNETH(rp)
    if err != nil {
        return common.Hash{}, err
    }
    hash, err := rocketTokenNETH.Transact(opts, "burn", amount)
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not burn nETH: %w", err)
    }
    return hash, nil
}


//
// Contracts
//


// Get contracts
var rocketTokenNETHLock sync.Mutex
func getRocketTokenNETH(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketTokenNETHLock.Lock()
    defer rocketTokenNETHLock.Unlock()
    return rp.GetContract("rocketTokenNETH")
}

