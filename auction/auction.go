package auction

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Get the total RPL balance of the auction contract
func GetTotalRPLBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return nil, err
    }
    totalRplBalance := new(*big.Int)
    if err := rocketAuctionManager.Call(opts, totalRplBalance, "getTotalRPLBalance"); err != nil {
        return nil, fmt.Errorf("Could not get auction contract total RPL balance: %w", err)
    }
    return *totalRplBalance, nil
}


// Get the allotted RPL balance of the auction contract
func GetAllottedRPLBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return nil, err
    }
    allottedRplBalance := new(*big.Int)
    if err := rocketAuctionManager.Call(opts, allottedRplBalance, "getAllottedRPLBalance"); err != nil {
        return nil, fmt.Errorf("Could not get auction contract allotted RPL balance: %w", err)
    }
    return *allottedRplBalance, nil
}


// Get the remaining RPL balance of the auction contract
func GetRemainingRPLBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return nil, err
    }
    remainingRplBalance := new(*big.Int)
    if err := rocketAuctionManager.Call(opts, remainingRplBalance, "getRemainingRPLBalance"); err != nil {
        return nil, fmt.Errorf("Could not get auction contract remaining RPL balance: %w", err)
    }
    return *remainingRplBalance, nil
}


// Get contracts
var rocketAuctionManagerLock sync.Mutex
func getRocketAuctionManager(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketAuctionManagerLock.Lock()
    defer rocketAuctionManagerLock.Unlock()
    return rp.GetContract("rocketAuctionManager")
}

