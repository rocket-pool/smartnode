package node

import (
    "context"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/rocket-pool/rocketpool-go/types"
    "golang.org/x/sync/errgroup"
)


// Minipool count details
type minipoolCountDetails struct {
    Status types.MinipoolStatus
    RefundAvailable bool
    WithdrawalAvailable bool
    CloseAvailable bool
}


// Get all node minipool count details
func getNodeMinipoolCountDetails(rp *rocketpool.RocketPool, nodeAddress common.Address) ([]minipoolCountDetails, error) {

    // Data
    var wg1 errgroup.Group
    var addresses []common.Address
    var currentBlock uint64
    var withdrawalDelay uint64

    // Get minipool addresses
    wg1.Go(func() error {
        var err error
        addresses, err = minipool.GetNodeMinipoolAddresses(rp, nodeAddress, nil)
        return err
    })

    // Get current block
    wg1.Go(func() error {
        header, err := rp.Client.HeaderByNumber(context.Background(), nil)
        if err == nil {
            currentBlock = header.Number.Uint64()
        }
        return err
    })

    // Get withdrawal delay
    wg1.Go(func() error {
        var err error
        withdrawalDelay, err = settings.GetMinipoolWithdrawalDelay(rp, nil)
        return err
    })

    // Wait for data
    if err := wg1.Wait(); err != nil {
        return []minipoolCountDetails{}, err
    }

    // Data
    var wg2 errgroup.Group
    details := make([]minipoolCountDetails, len(addresses))

    // Load details
    for mi, address := range addresses {
        mi, address := mi, address
        wg2.Go(func() error {
            mpDetails, err := getMinipoolCountDetails(rp, address, currentBlock, withdrawalDelay)
            if err == nil { details[mi] = mpDetails }
            return err
        })
    }

    // Wait for data
    if err := wg2.Wait(); err != nil {
        return []minipoolCountDetails{}, err
    }

    // Return
    return details, nil

}


// Get a minipool's count details
func getMinipoolCountDetails(rp *rocketpool.RocketPool, minipoolAddress common.Address, currentBlock, withdrawalDelay uint64) (minipoolCountDetails, error) {

    // Create minipool
    mp, err := minipool.NewMinipool(rp, minipoolAddress)
    if err != nil {
        return minipoolCountDetails{}, err
    }

    // Data
    var wg errgroup.Group
    var status types.MinipoolStatus
    var statusBlock uint64
    var refundBalance *big.Int

    // Load data
    wg.Go(func() error {
        var err error
        status, err = mp.GetStatus(nil)
        return err
    })
    wg.Go(func() error {
        var err error
        statusBlock, err = mp.GetStatusBlock(nil)
        return err
    })
    wg.Go(func() error {
        var err error
        refundBalance, err = mp.GetNodeRefundBalance(nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return minipoolCountDetails{}, err
    }

    // Return
    return minipoolCountDetails{
        Status: status,
        RefundAvailable: (refundBalance.Cmp(big.NewInt(0)) > 0),
        WithdrawalAvailable: (status == types.Withdrawable && (currentBlock - statusBlock) >= withdrawalDelay),
        CloseAvailable: (status == types.Dissolved),
    }, nil

}

