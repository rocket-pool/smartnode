package node

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Withdraw resources from the node contract
func withdrawFromNode(c *cli.Context, amount float64, unit string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        CM: true,
        NodeContract: true,
        LoadContracts: []string{"rocketNodeAPI"},
        LoadAbis: []string{"rocketNodeContract"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get args
    amountWei := eth.EthToWei(amount)

    // Check deposit can be withdrawn from node
    response, err := node.CanWithdrawFromNode(p, amountWei, unit)
    if err != nil { return err }

    // Check response
    if response.InsufficientNodeBalance {
        api.PrintResponse(p.Output, response)
        return nil
    }

    // Withdraw from node
    response, err = node.WithdrawFromNode(p, amountWei, unit)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, response)
    return nil

}

