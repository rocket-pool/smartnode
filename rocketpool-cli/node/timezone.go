package node

import (
    "errors"
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Set the node's timezone
func setNodeTimezone(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        CM: true,
        NodeContractAddress: true,
        LoadContracts: []string{"rocketNodeAPI"},
        WaitClientConn: true,
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Prompt user for timezone
    timezone := promptTimezone(p.Input, p.Output)

    // Set node timezone
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return err
    } else {
        fmt.Fprintln(p.Output, "Setting node timezone...")
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.CM.Addresses["rocketNodeAPI"], p.CM.Abis["rocketNodeAPI"], "setTimezoneLocation", timezone); err != nil {
            return errors.New("Error setting node timezone: " + err.Error())
        }
    }

    // Get node timezone
    nodeAccount, _ := p.AM.GetNodeAccount()
    nodeTimezone := new(string)
    if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, nodeTimezone, "getTimezoneLocation", nodeAccount.Address); err != nil {
        return errors.New("Error retrieving node timezone: " + err.Error())
    }

    // Log & return
    fmt.Fprintln(p.Output, "Node timezone successfully updated to:", *nodeTimezone)
    return nil

}

