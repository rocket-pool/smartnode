package node

import (
    "math/big"

    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func getStatus(c *cli.Context) error {

    // Get services
    pm, err := services.GetPasswordManager(c)
    if err != nil { return err }
    am, err := services.GetAccountManager(c)
    if err != nil { return err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return err }

    // Response
    response := &types.NodeStatusResponse{}

    // Get account status
    response.PasswordExists = pm.PasswordExists()
    response.AccountExists = am.NodeAccountExists()
    if !(response.PasswordExists && response.AccountExists) {
        return api.PrintResponse(response)
    }

    // Get node account
    nodeAccount, _ := am.GetNodeAccount()
    response.AccountAddress = nodeAccount.Address.Hex()

    // Sync
    var wg errgroup.Group

    // Get node details
    wg.Go(func() error {
        details, err := node.GetNodeDetails(rp, nodeAccount.Address)
        if err == nil {
            response.Registered = details.Exists
            response.Trusted = details.Trusted
            response.TimezoneLocation = details.TimezoneLocation
        }
        return err
    })

    // Get node balances
    wg.Go(func() error {
        balances, err := tokens.GetBalances(rp, nodeAccount.Address)
        if err == nil {
            response.EthBalance = balances.ETH.String()
            response.NethBalance = balances.NETH.String()
        }
        return err
    })

    // Get node minipool counts
    wg.Go(func() error {

        // Get minipool addresses
        addresses, err := minipool.GetNodeMinipoolAddresses(rp, nodeAccount.Address)
        if err != nil { return err }

        // Update total count
        response.MinipoolCounts.Total = len(addresses)

        // Get minipool details
        var wg errgroup.Group
        for mi := 0; mi < len(addresses); mi++ {
            mi := mi
            wg.Go(func() error {

                // Create minipool
                mp, err := minipool.NewMinipool(rp, addresses[mi])
                if err != nil { return err }

                // Update status counts
                if status, err := mp.GetStatus(); err != nil {
                    return err
                } else {
                    switch status {
                        case minipool.Initialized:  response.MinipoolCounts.Initialized++
                        case minipool.Prelaunch:    response.MinipoolCounts.Prelaunch++
                        case minipool.Staking:      response.MinipoolCounts.Staking++
                        case minipool.Exited:       response.MinipoolCounts.Exited++
                        case minipool.Withdrawable: response.MinipoolCounts.Withdrawable++
                        case minipool.Dissolved:    response.MinipoolCounts.Dissolved++
                    }
                }

                // Update refundable count
                if refundBalance, err := mp.GetNodeRefundBalance(); err != nil {
                    return err
                } else if refundBalance.Cmp(big.NewInt(0)) > 0 {
                    response.MinipoolCounts.Refundable++
                }

                // Return
                return nil

            })
        }
        if err := wg.Wait(); err != nil { return err }

        // Return
        return nil

    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return api.PrintResponse(&types.NodeStatusResponse{
            Error: err.Error(),
        })
    }

    // Print response
    return api.PrintResponse(response)

}

