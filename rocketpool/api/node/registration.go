package node

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
)


// Register the node with Rocket Pool
func registerNode(c *cli.Context) error {

    // Initialise ethereum client & Rocket Pool contract manager
    _, contractManager, err := rocketpool.InitClient(c.GlobalString("powProvider"), c.GlobalString("storageAddress"))
    if err != nil {
        return err
    }

    // Load Rocket Pool node contracts
    err = contractManager.LoadContracts([]string{"rocketNodeAPI"})
    if err != nil {
        return err
    }

    // Return
    return nil

}

