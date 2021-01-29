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
    rocketNodeETHToken, err := getRocketNodeETHToken(rp)
    if err != nil {
        return nil, err
    }
    return contractETHBalance(rp, rocketNodeETHToken, opts)
}


// Get nETH total supply
func GetNETHTotalSupply(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketNodeETHToken, err := getRocketNodeETHToken(rp)
    if err != nil {
        return nil, err
    }
    return totalSupply(rocketNodeETHToken, "nETH", opts)
}


// Get nETH balance
func GetNETHBalance(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (*big.Int, error) {
    rocketNodeETHToken, err := getRocketNodeETHToken(rp)
    if err != nil {
        return nil, err
    }
    return balanceOf(rocketNodeETHToken, "nETH", address, opts)
}


// Transfer nETH
func TransferNETH(rp *rocketpool.RocketPool, to common.Address, amount *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNodeETHToken, err := getRocketNodeETHToken(rp)
    if err != nil {
        return nil, err
    }
    return transfer(rp.Client, rocketNodeETHToken, "nETH", to, amount, opts)
}


// Burn nETH for ETH
func BurnNETH(rp *rocketpool.RocketPool, amount *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNodeETHToken, err := getRocketNodeETHToken(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketNodeETHToken.Transact(opts, "burn", amount)
    if err != nil {
        return nil, fmt.Errorf("Could not burn nETH: %w", err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketNodeETHTokenLock sync.Mutex
func getRocketNodeETHToken(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketNodeETHTokenLock.Lock()
    defer rocketNodeETHTokenLock.Unlock()
    return rp.GetContract("rocketNodeETHToken")
}

