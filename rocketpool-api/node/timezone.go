package node

import (
    "errors"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Set the node's timezone
func setNodeTimezone(c *cli.Context, timezone string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        CM: true,
        NodeContractAddress: true,
        LoadContracts: []string{"rocketNodeAPI"},
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Response
    response := api.NodeTimezoneResponse{}

    // Get node account
    nodeAccount, _ := p.AM.GetNodeAccount()
    response.AccountAddress = nodeAccount.Address

    // Set node timezone
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return err
    } else {
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.CM.Addresses["rocketNodeAPI"], p.CM.Abis["rocketNodeAPI"], "setTimezoneLocation", timezone); err != nil {
            return errors.New("Error setting node timezone: " + err.Error())
        } else {
            response.Success = true
        }
    }

    // Get node timezone
    nodeTimezone := new(string)
    if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, nodeTimezone, "getTimezoneLocation", nodeAccount.Address); err != nil {
        return errors.New("Error retrieving node timezone: " + err.Error())
    } else {
        response.Timezone = *nodeTimezone
    }

    // Print response & return
    api.PrintResponse(p.Output, response)
    return nil

}

