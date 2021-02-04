package tokens

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Get the nETH contract ETH balance
func GetNETHContractETHBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketTokenNETH, err := getRocketTokenNETH(rp)
    if err != nil {
        return nil, err
    }
    return contractETHBalance(rp, rocketTokenNETH, opts)
}


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


// Transfer nETH
func TransferNETH(rp *rocketpool.RocketPool, to common.Address, amount *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketTokenNETH, err := getRocketTokenNETH(rp)
    if err != nil {
        return nil, err
    }
    return transfer(rp.Client, rocketTokenNETH, "nETH", to, amount, opts)
}


// Burn nETH for ETH
func BurnNETH(rp *rocketpool.RocketPool, amount *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketTokenNETH, err := getRocketTokenNETH(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketTokenNETH.Transact(opts, "burn", amount)
    if err != nil {
        return nil, fmt.Errorf("Could not burn nETH: %w", err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketTokenNETHLock sync.Mutex
func getRocketTokenNETH(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketTokenNETHLock.Lock()
    defer rocketTokenNETHLock.Unlock()
    return rp.GetContract("rocketTokenNETH")
}

