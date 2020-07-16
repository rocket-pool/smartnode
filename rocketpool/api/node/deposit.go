package node

import (
    "context"
    "math/big"

    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func canNodeDeposit(c *cli.Context, amountWei *big.Int) error {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return err }
    am, err := services.GetAccountManager(c)
    if err != nil { return err }
    ec, err := services.GetEthClient(c)
    if err != nil { return err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return err }

    // Response
    response := &types.CanNodeDepositResponse{}

    // Get node account
    nodeAccount, _ := am.GetNodeAccount()

    // Sync
    var wg errgroup.Group

    // Check node balance
    wg.Go(func() error {
        ethBalanceWei, err := ec.BalanceAt(context.Background(), nodeAccount.Address, nil)
        if err == nil {
            response.InsufficientBalance = (amountWei.Cmp(ethBalanceWei) > 0)
        }
        return err
    })

    // Check deposit amount
    wg.Go(func() error {
        if amountWei.Cmp(big.NewInt(0)) > 0 {
            return nil
        }
        trusted, err := node.GetNodeTrusted(rp, nodeAccount.Address)
        if err == nil {
            response.InvalidAmount = !trusted
        }
        return err
    })

    // Check node deposits are enabled
    wg.Go(func() error {
        depositEnabled, err := settings.GetNodeRegistrationEnabled(rp)
        if err == nil {
            response.DepositDisabled = !depositEnabled
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return api.PrintResponse(&types.CanNodeDepositResponse{
            Error: err.Error(),
        })
    }

    // Update & print response
    response.CanDeposit = !(response.InsufficientBalance || response.InvalidAmount || response.DepositDisabled)
    return api.PrintResponse(response)

}


func nodeDeposit(c *cli.Context, amountWei *big.Int, minNodeFee float64) error {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return err }
    am, err := services.GetAccountManager(c)
    if err != nil { return err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return err }

    // Response
    response := &types.NodeDepositResponse{}

    // Get transactor
    opts, err := am.GetNodeAccountTransactor()
    if err != nil {
        return api.PrintResponse(&types.NodeDepositResponse{
            Error: err.Error(),
        })
    }
    opts.Value = amountWei

    // Deposit
    txReceipt, err := node.Deposit(rp, minNodeFee, opts)
    if err != nil {
        return api.PrintResponse(&types.NodeDepositResponse{
            Error: err.Error(),
        })
    }
    response.TxHash = txReceipt.TxHash.Hex()

    // Get created minipool address
    // TODO: implement
    _ = txReceipt

    // Print response
    return api.PrintResponse(response)

}

