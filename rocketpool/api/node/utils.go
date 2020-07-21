package node

import (
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/types"
    "golang.org/x/sync/errgroup"
)


// Minipool count details
type minipoolCountDetails struct {
    Status types.MinipoolStatus
    Refundable bool
}


// Get all node minipool count details
func getNodeMinipoolCountDetails(rp *rocketpool.RocketPool, nodeAddress common.Address) ([]minipoolCountDetails, error) {

    // Get minipool addresses
    addresses, err := minipool.GetNodeMinipoolAddresses(rp, nodeAddress)
    if err != nil {
        return []minipoolCountDetails{}, err
    }

    // Data
    var wg errgroup.Group
    details := make([]minipoolCountDetails, len(addresses))

    // Load details
    for mi, address := range addresses {
        mi, address := mi, address
        wg.Go(func() error {
            mpDetails, err := getMinipoolCountDetails(rp, address)
            if err == nil { details[mi] = mpDetails }
            return err
        })
    }

    // Wait for data
    if err := wg.Wait(); err != nil {
        return []minipoolCountDetails{}, err
    }

    // Return
    return details, nil

}


// Get a minipool's count details
func getMinipoolCountDetails(rp *rocketpool.RocketPool, minipoolAddress common.Address) (minipoolCountDetails, error) {

    // Create minipool
    mp, err := minipool.NewMinipool(rp, minipoolAddress)
    if err != nil {
        return minipoolCountDetails{}, err
    }

    // Data
    var wg errgroup.Group
    var status types.MinipoolStatus
    var refundBalance *big.Int

    // Load data
    wg.Go(func() error {
        var err error
        status, err = mp.GetStatus()
        return err
    })
    wg.Go(func() error {
        var err error
        refundBalance, err = mp.GetNodeRefundBalance()
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return minipoolCountDetails{}, err
    }

    // Return
    return minipoolCountDetails{
        Status: status,
        Refundable: (refundBalance.Cmp(big.NewInt(0)) > 0),
    }, nil

}

