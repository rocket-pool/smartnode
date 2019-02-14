package smartnode

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/daemons/smartnode/node"
)


// Run daemon
func Run(c *cli.Context) error {

    // Error channels
    errorChannel := make(chan error)
    fatalErrorChannel := make(chan error)

    // Start node checkin process
    go node.StartCheckinProcess(c, errorChannel, fatalErrorChannel)

    // Block thread; log errors and return fatal errors
    for {
        select {
            case err := <-errorChannel:
                fmt.Println(err)
            case err := <-fatalErrorChannel:
                return err
        }
    }
    return nil

}

