package node

import (
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/rocket-pool/rocketpool-go/types"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func getStatus(c *cli.Context) (*api.NodeStatusResponse, error) {

    // Get services
    if err := services.RequireNodeAccount(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.NodeStatusResponse{}

    // Get node account
    nodeAccount, _ := am.GetNodeAccount()
    response.AccountAddress = nodeAccount.Address

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
            response.EthBalance = balances.ETH
            response.NethBalance = balances.NETH
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
                    case types.Initialized:  response.MinipoolCounts.Initialized++
                    case types.Prelaunch:    response.MinipoolCounts.Prelaunch++
                    case types.Staking:      response.MinipoolCounts.Staking++
                    case types.Exited:       response.MinipoolCounts.Exited++
                    case types.Withdrawable: response.MinipoolCounts.Withdrawable++
                    case types.Dissolved:    response.MinipoolCounts.Dissolved++
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

