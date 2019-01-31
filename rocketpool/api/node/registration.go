package node

import (
    "github.com/urfave/cli"
)


// Register a node with Rocket Pool
func registerNode(c *cli.Context) error {

    // Initialise ethereum client & load node contracts
    _, err := loadContracts(c)
    if err != nil {
        return err
    }

    // Return
    return nil

}

