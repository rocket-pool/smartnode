package smartnode

import (
    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode-cli/rocketpool/daemons/smartnode/node"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services"
)


// Run daemon
func Run(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        DB: true,
        AM: true,
        ClientSync: true,
        CM: true,
        NodeContract: true,
        LoadContracts: []string{"rocketNodeAPI", "rocketNodeSettings"},
        LoadAbis: []string{"rocketNodeContract"},
    })
    if err != nil {
        return err
    }

    // Start node checkin process
    go node.StartCheckinProcess(p)

    // Block thread
    select {}

}

