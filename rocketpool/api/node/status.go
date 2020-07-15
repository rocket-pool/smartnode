package node

import (
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
    if !response.PasswordExists || !response.AccountExists {
        return api.PrintResponse(response)
    }

    // Get node account
    nodeAccount, _ := am.GetNodeAccount()
    response.AccountAddress = nodeAccount.Address.Hex()

    // Data
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

    // Get node minipool details
    wg.Go(func() error {
        minipoolCount, err := minipool.GetNodeMinipoolCount(rp, nodeAccount.Address)
        if err == nil {
            response.MinipoolCount = int(minipoolCount)
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil { return err }

    // Print response
    return api.PrintResponse(response)

}

