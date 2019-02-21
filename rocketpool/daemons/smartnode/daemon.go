package smartnode

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/daemons/smartnode/node"
)


// Run daemon
func Run(c *cli.Context) error {

    // Error channel
    fatalErrorChannel := make(chan error)

    // Start node checkin process
    go node.StartCheckinProcess(c, fatalErrorChannel)

    // Block thread; return fatal errors
    select {
        case err := <-fatalErrorChannel:
            return err
    }
    return nil

}

