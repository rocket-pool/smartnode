package tokens

import (
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Get rETH total supply
func GetRETHTotalSupply(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketETHToken, err := getRocketETHToken(rp)
    if err != nil {
        return nil, err
    }
    return totalSupply(rocketETHToken, "rETH", opts)
}


// Get contracts
var rocketETHTokenLock sync.Mutex
func getRocketETHToken(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketETHTokenLock.Lock()
    defer rocketETHTokenLock.Unlock()
    return rp.GetContract("rocketETHToken")
}

