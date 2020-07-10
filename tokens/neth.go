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


// Get nETH balance
func GetNETHBalance(rp *rocketpool.RocketPool, address common.Address) (*big.Int, error) {
    rocketNodeETHToken, err := getRocketNodeETHToken(rp)
    if err != nil {
        return nil, err
    }
    return balanceOf(rocketNodeETHToken, "nETH", address)
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
    txReceipt, err := contract.Transact(rp.Client, rocketNodeETHToken, opts, "burn", amount)
    if err != nil {
        return nil, fmt.Errorf("Could not burn nETH: %w", err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketNodeETHTokenLock sync.Mutex
func getRocketNodeETHToken(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketNodeETHTokenLock.Lock()
    defer rocketNodeETHTokenLock.Unlock()
    return rp.GetContract("rocketNodeETHToken")
}

