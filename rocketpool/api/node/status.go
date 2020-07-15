package node

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/tokens"

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

    // Get node details
    if details, err := node.GetNodeDetails(rp, nodeAccount.Address); err != nil {
        return err
    } else {
        response.Registered = details.Exists
        response.Trusted = details.Trusted
        response.TimezoneLocation = details.TimezoneLocation
    }

    // Get node balances
    if balances, err := tokens.GetBalances(rp, nodeAccount.Address); err != nil {
        return err
    } else {
        response.EthBalance = balances.ETH.String()
        response.NethBalance = balances.NETH.String()
    }

    // Get node minipool details
    if minipoolCount, err := minipool.GetNodeMinipoolCount(rp, nodeAccount.Address); err != nil {
        return err
    } else {
        response.MinipoolCount = int(minipoolCount)
    }

    // Print response
    return api.PrintResponse(response)

}

