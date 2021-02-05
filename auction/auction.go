package auction

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
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


// Get the number of lots for auction
func GetLotCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return 0, err
    }
    lotCount := new(*big.Int)
    if err := rocketAuctionManager.Call(opts, lotCount, "getLotCount"); err != nil {
        return 0, fmt.Errorf("Could not get lot count: %w", err)
    }
    return (*lotCount).Uint64(), nil
}


// Lot details
func GetLotExists(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return false, err
    }
    lotExists := new(bool)
    if err := rocketAuctionManager.Call(opts, lotExists, "getLotExists"); err != nil {
        return false, fmt.Errorf("Could not get lot exists status: %w", err)
    }
    return *lotExists, nil
}
func GetLotStartBlock(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return 0, err
    }
    lotStartBlock := new(*big.Int)
    if err := rocketAuctionManager.Call(opts, lotStartBlock, "getLotStartBlock"); err != nil {
        return 0, fmt.Errorf("Could not get lot start block: %w", err)
    }
    return (*lotStartBlock).Uint64(), nil
}
func GetLotEndBlock(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return 0, err
    }
    lotEndBlock := new(*big.Int)
    if err := rocketAuctionManager.Call(opts, lotEndBlock, "getLotEndBlock"); err != nil {
        return 0, fmt.Errorf("Could not get lot end block: %w", err)
    }
    return (*lotEndBlock).Uint64(), nil
}
func GetLotStartPrice(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return nil, err
    }
    lotStartPrice := new(*big.Int)
    if err := rocketAuctionManager.Call(opts, lotStartPrice, "getLotStartPrice"); err != nil {
        return nil, fmt.Errorf("Could not get lot start price: %w", err)
    }
    return *lotStartPrice, nil
}
func GetLotReservePrice(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return nil, err
    }
    lotReservePrice := new(*big.Int)
    if err := rocketAuctionManager.Call(opts, lotReservePrice, "getLotReservePrice"); err != nil {
        return nil, fmt.Errorf("Could not get lot reserve price: %w", err)
    }
    return *lotReservePrice, nil
}
func GetLotTotalRPLAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return nil, err
    }
    lotTotalRplAmount := new(*big.Int)
    if err := rocketAuctionManager.Call(opts, lotTotalRplAmount, "getLotTotalRPLAmount"); err != nil {
        return nil, fmt.Errorf("Could not get lot total RPL amount: %w", err)
    }
    return *lotTotalRplAmount, nil
}
func GetLotTotalBidAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return nil, err
    }
    lotTotalBidAmount := new(*big.Int)
    if err := rocketAuctionManager.Call(opts, lotTotalBidAmount, "getLotTotalBidAmount"); err != nil {
        return nil, fmt.Errorf("Could not get lot total ETH bid amount: %w", err)
    }
    return *lotTotalBidAmount, nil
}
func GetLotRPLRecovered(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return false, err
    }
    lotRplRecovered := new(bool)
    if err := rocketAuctionManager.Call(opts, lotRplRecovered, "getLotRPLRecovered"); err != nil {
        return false, fmt.Errorf("Could not get lot RPL recovered status: %w", err)
    }
    return *lotRplRecovered, nil
}
func GetLotPriceByTotalBids(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return nil, err
    }
    lotPriceByTotalBids := new(*big.Int)
    if err := rocketAuctionManager.Call(opts, lotPriceByTotalBids, "getLotPriceByTotalBids"); err != nil {
        return nil, fmt.Errorf("Could not get lot price by total bids: %w", err)
    }
    return *lotPriceByTotalBids, nil
}
func GetLotCurrentPrice(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return nil, err
    }
    lotCurrentPrice := new(*big.Int)
    if err := rocketAuctionManager.Call(opts, lotCurrentPrice, "getLotCurrentPrice"); err != nil {
        return nil, fmt.Errorf("Could not get lot current price: %w", err)
    }
    return *lotCurrentPrice, nil
}
func GetLotClaimedRPLAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return nil, err
    }
    lotClaimedRplAmount := new(*big.Int)
    if err := rocketAuctionManager.Call(opts, lotClaimedRplAmount, "getLotClaimedRPLAmount"); err != nil {
        return nil, fmt.Errorf("Could not get lot claimed RPL amount: %w", err)
    }
    return *lotClaimedRplAmount, nil
}
func GetLotRemainingRPLAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return nil, err
    }
    lotRemainingRplAmount := new(*big.Int)
    if err := rocketAuctionManager.Call(opts, lotRemainingRplAmount, "getLotRemainingRPLAmount"); err != nil {
        return nil, fmt.Errorf("Could not get lot remaining RPL amount: %w", err)
    }
    return *lotRemainingRplAmount, nil
}
func GetLotIsCleared(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return false, err
    }
    lotIsCleared := new(bool)
    if err := rocketAuctionManager.Call(opts, lotIsCleared, "getLotIsCleared"); err != nil {
        return false, fmt.Errorf("Could not get lot cleared status: %w", err)
    }
    return *lotIsCleared, nil
}


// Get the ETH amount bid on a lot by an address
func GetLotAddressBidAmount(rp *rocketpool.RocketPool, bidder common.Address, opts *bind.CallOpts) (*big.Int, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return nil, err
    }
    lot := new(*big.Int)
    if err := rocketAuctionManager.Call(opts, lot, "getLotAddressBidAmount", bidder); err != nil {
        return nil, fmt.Errorf("Could not get lot address ETH bid amount: %w", err)
    }
    return *lot, nil
}


// Get the price of a lot at a specific block
func GetLotPriceAtBlock(rp *rocketpool.RocketPool, blockNumber uint64, opts *bind.CallOpts) (*big.Int, error) {
    rocketAuctionManager, err := getRocketAuctionManager(rp)
    if err != nil {
        return nil, err
    }
    lotPriceAtBlock := new(*big.Int)
    if err := rocketAuctionManager.Call(opts, lotPriceAtBlock, "getLotPriceAtBlock", big.NewInt(int64(blockNumber))); err != nil {
        return nil, fmt.Errorf("Could not get lot price at block: %w", err)
    }
    return *lotPriceAtBlock, nil
}


// Get contracts
var rocketAuctionManagerLock sync.Mutex
func getRocketAuctionManager(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketAuctionManagerLock.Lock()
    defer rocketAuctionManagerLock.Unlock()
    return rp.GetContract("rocketAuctionManager")
}

