package minipool

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    //"github.com/ethereum/go-ethereum/core/types"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    //"github.com/rocket-pool/rocketpool-go/utils/contract"
)


// Contract access locks
var rocketMinipoolManagerLock sync.Mutex


// Get a node's minipool addresses
func GetNodeMinipoolAddresses(rp *rocketpool.RocketPool, nodeAddress common.Address) ([]common.Address, error) {

    // Get minipool count
    minipoolCount, err := GetNodeMinipoolCount(rp, nodeAddress)
    if err != nil {
        return []common.Address{}, err
    }

    // Data
    var wg errgroup.Group
    addresses := make([]common.Address, minipoolCount)

    // Load addresses
    for mi := int64(0); mi < minipoolCount; mi++ {
        mi := mi
        wg.Go(func() error {
            address, err := GetNodeMinipoolAt(rp, nodeAddress, mi)
            if err == nil { addresses[mi] = address }
            return err
        })
    }

    // Wait for data
    if err := wg.Wait(); err != nil {
        return []common.Address{}, err
    }

    // Return
    return addresses, nil

}


// Get a node's minipool count
func GetNodeMinipoolCount(rp *rocketpool.RocketPool, nodeAddress common.Address) (int64, error) {
    rocketMinipoolManager, err := getRocketMinipoolManager(rp)
    if err != nil {
        return 0, err
    }
    minipoolCount := new(*big.Int)
    if err := rocketMinipoolManager.Call(nil, minipoolCount, "getNodeMinipoolCount", nodeAddress); err != nil {
        return 0, fmt.Errorf("Could not get node %v minipool count: %w", nodeAddress.Hex(), err)
    }
    return (*minipoolCount).Int64(), nil
}


// Get a node's minipool address by index
func GetNodeMinipoolAt(rp *rocketpool.RocketPool, nodeAddress common.Address, index int64) (common.Address, error) {
    rocketMinipoolManager, err := getRocketMinipoolManager(rp)
    if err != nil {
        return common.Address{}, err
    }
    minipoolAddress := new(common.Address)
    if err := rocketMinipoolManager.Call(nil, minipoolAddress, "getNodeMinipoolAt", nodeAddress, big.NewInt(index)); err != nil {
        return common.Address{}, fmt.Errorf("Could not get node %v minipool %v address: %w", nodeAddress.Hex(), index, err)
    }
    return *minipoolAddress, nil
}


// Get contracts
func getRocketMinipoolManager(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketMinipoolManagerLock.Lock()
    defer rocketMinipoolManagerLock.Unlock()
    return rp.GetContract("rocketMinipoolManager")
}

