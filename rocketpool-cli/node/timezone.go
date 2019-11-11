package node

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
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

    // Prompt for timezone
    timezone := promptTimezone(p.Input, p.Output)

    // Set node timezone
    timezoneSet, err := node.SetNodeTimezone(p, timezone)
    if err != nil { return err }

    // Print output & return
    if timezoneSet.Success {
        fmt.Fprintln(p.Output, "Node timezone successfully updated to:", timezoneSet.Timezone)
    }
    return nil

}

