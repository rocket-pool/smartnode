package minipool

import (
    "context"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/rocket-pool/rocketpool-go/types"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canWithdrawMinipool(c *cli.Context, minipoolAddress common.Address) (*api.CanWithdrawMinipoolResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    ec, err := services.GetEthClient(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanWithdrawMinipoolResponse{}

    // Create minipool
    mp, err := minipool.NewMinipool(rp, minipoolAddress)
    if err != nil {
        return nil, err
    }

    // Validate minipool owner
    nodeAccount, _ := am.GetNodeAccount()
    if err := validateMinipoolOwner(mp, nodeAccount.Address); err != nil {
        return nil, err
    }

    // Data
    var wg errgroup.Group
    var currentBlock int64
    var statusBlock int64
    var withdrawalDelay int64

    // Check minipool status
    wg.Go(func() error {
        status, err := mp.GetStatus()
        if err == nil {
            response.InvalidStatus = (status != types.Withdrawable)
        }
        return err
    })

    // Get current block
    wg.Go(func() error {
        header, err := ec.HeaderByNumber(context.Background(), nil)
        if err == nil {
            currentBlock = header.Number.Int64()
        }
        return err
    })

    // Get minipool status block
    wg.Go(func() error {
        var err error
        statusBlock, err = mp.GetStatusBlock()
        return err
    })

    // Get withdrawal delay
    wg.Go(func() error {
        var err error
        withdrawalDelay, err = settings.GetMinipoolWithdrawalDelay(rp)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Check minipool withdrawal delay
    response.WithdrawalDelayActive = ((currentBlock - statusBlock) >= withdrawalDelay)

    // Update & return response
    response.CanWithdraw = !(response.InvalidStatus || response.WithdrawalDelayActive)
    return &response, nil

}


func withdrawMinipool(c *cli.Context, minipoolAddress common.Address) (*api.WithdrawMinipoolResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.WithdrawMinipoolResponse{}

    // Create minipool
    mp, err := minipool.NewMinipool(rp, minipoolAddress)
    if err != nil {
        return nil, err
    }

    // Get transactor
    opts, err := am.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Withdraw
    txReceipt, err := mp.Withdraw(opts)
    if err != nil {
        return nil, err
    }
    response.TxHash = txReceipt.TxHash

    // Return response
    return &response, nil

}

