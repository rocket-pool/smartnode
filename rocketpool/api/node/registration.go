package node

import (
    "github.com/urfave/cli"
)


// Register a node with Rocket Pool
func registerNode(c *cli.Context) error {

    // Initialise ethereum client & node contracts
    _, _, err := initClient(c)
    if err != nil {
        return err
    }

    // Return
    return nil

}

