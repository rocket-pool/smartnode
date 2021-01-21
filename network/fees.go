package network

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


// Get the current network node commission rate
func GetNodeFee(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
    rocketNetworkFees, err := getRocketNetworkFees(rp)
    if err != nil {
        return 0, err
    }
    nodeFee := new(*big.Int)
    if err := rocketNetworkFees.Call(opts, nodeFee, "getNodeFee"); err != nil {
        return 0, fmt.Errorf("Could not get network node fee: %w", err)
    }
    return eth.WeiToEth(*nodeFee), nil
}


// Get contracts
var rocketNetworkFeesLock sync.Mutex
func getRocketNetworkFees(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketNetworkFeesLock.Lock()
    defer rocketNetworkFeesLock.Unlock()
    return rp.GetContract("rocketNetworkFees")
}

