package node

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Register the node with Rocket Pool
func registerNode(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        LoadContracts: []string{"rocketNodeAPI", "rocketNodeSettings"},
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Check node can be registered
    response, err := node.CanRegisterNode(p)
    if err != nil { return err }

    // Check response
    if response.HadExistingContract {
        fmt.Fprintln(p.Output, fmt.Sprintf("Node is already registered with Rocket Pool - current deposit contract is at %s", response.ContractAddress.Hex()))
        return nil
    }
    if response.RegistrationsDisabled {
        fmt.Fprintln(p.Output, "New node registrations are currently disabled in Rocket Pool.")
        return nil
    }
    if response.InsufficientAccountBalance {
        fmt.Fprintln(p.Output, fmt.Sprintf("Node account %s requires a minimum balance of %.2f ETH to operate in Rocket Pool", response.AccountAddress.Hex(), eth.WeiToEth(response.MinAccountBalanceEtherWei)))
        return nil
    }

    // Prompt for timezone
    timezone := promptTimezone(p.Input, p.Output)

    // Register node
    response, err = node.RegisterNode(p, timezone)
    if err != nil { return err }

    // Print output & return
    if response.Success {
        fmt.Fprintln(p.Output, fmt.Sprintf("Node registered successfully with Rocket Pool - new node deposit contract created at %s", response.ContractAddress.Hex()))
    }
    return nil

}

