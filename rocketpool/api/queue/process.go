package queue

import (
    "math/big"

    "github.com/rocket-pool/rocketpool-go/deposit"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/settings/protocol"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canProcessQueue(c *cli.Context) (*api.CanProcessQueueResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanProcessQueueResponse{}

    // Data
    var wg errgroup.Group
    var nextMinipoolCapacity *big.Int
    var depositPoolBalance *big.Int

    // Check deposit assignments are enabled
    wg.Go(func() error {
        assignDepositsEnabled, err := protocol.GetAssignDepositsEnabled(rp, nil)
        if err == nil {
            response.AssignDepositsDisabled = !assignDepositsEnabled
        }
        return err
    })

    // Get next available minipool capacity
    wg.Go(func() error {
        var err error
        nextMinipoolCapacity, err = minipool.GetQueueNextCapacity(rp, nil)
        return err
    })

    // Get deposit pool balance
    wg.Go(func() error {
        var err error
        depositPoolBalance, err = deposit.GetBalance(rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Check next minipool capacity & deposit pool balance
    response.NoMinipoolsAvailable = (nextMinipoolCapacity.Cmp(big.NewInt(0)) == 0)
    response.InsufficientDepositBalance = (depositPoolBalance.Cmp(nextMinipoolCapacity) < 0)

    // Update & return response
    response.CanProcess = !(response.AssignDepositsDisabled || response.NoMinipoolsAvailable || response.InsufficientDepositBalance)
    return &response, nil

}


func processQueue(c *cli.Context) (*api.ProcessQueueResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.ProcessQueueResponse{}

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Process queue
    txReceipt, err := deposit.AssignDeposits(rp, opts)
    if err != nil {
        return nil, err
    }
    response.TxHash = txReceipt.TxHash

    // Return response
    return &response, nil

}

