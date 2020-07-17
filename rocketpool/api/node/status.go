package node

import (
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func getStatus(c *cli.Context) (*api.NodeStatusResponse, error) {

    // Get services
    if err := services.RequireNodeAccount(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.NodeStatusResponse{}

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
        details, err := getNodeMinipoolCountDetails(rp, nodeAccount.Address)
        if err == nil {
            response.MinipoolCounts.Total = len(details)
            for _, mpDetails := range details {
                switch mpDetails.Status {
                    case minipool.Initialized:  response.MinipoolCounts.Initialized++
                    case minipool.Prelaunch:    response.MinipoolCounts.Prelaunch++
                    case minipool.Staking:      response.MinipoolCounts.Staking++
                    case minipool.Exited:       response.MinipoolCounts.Exited++
                    case minipool.Withdrawable: response.MinipoolCounts.Withdrawable++
                    case minipool.Dissolved:    response.MinipoolCounts.Dissolved++
                }
                if mpDetails.Refundable {
                    response.MinipoolCounts.Refundable++
                }
            }
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Return response
    return &response, nil

}

