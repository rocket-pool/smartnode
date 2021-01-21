package tokens

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/contract"
)


// Get rETH total supply
func GetRETHTotalSupply(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketETHToken, err := getRocketETHToken(rp)
    if err != nil {
        return nil, err
    }
    return totalSupply(rocketETHToken, "rETH", opts)
}


// Get rETH balance
func GetRETHBalance(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (*big.Int, error) {
    rocketETHToken, err := getRocketETHToken(rp)
    if err != nil {
        return nil, err
    }
    return balanceOf(rocketETHToken, "rETH", address, opts)
}


// Transfer rETH
func TransferRETH(rp *rocketpool.RocketPool, to common.Address, amount *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketETHToken, err := getRocketETHToken(rp)
    if err != nil {
        return nil, err
    }
    return transfer(rp.Client, rocketETHToken, "rETH", to, amount, opts)
}


// Burn rETH for ETH
func BurnRETH(rp *rocketpool.RocketPool, amount *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketETHToken, err := getRocketETHToken(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := contract.Transact(rp.Client, rocketETHToken, opts, "burn", amount)
    if err != nil {
        return nil, fmt.Errorf("Could not burn rETH: %w", err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketETHTokenLock sync.Mutex
func getRocketETHToken(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketETHTokenLock.Lock()
    defer rocketETHTokenLock.Unlock()
    return rp.GetContract("rocketETHToken")
}

