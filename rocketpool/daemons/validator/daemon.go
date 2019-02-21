package validator

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/daemons/validator/beacon"
)


// Run daemon
func Run(c *cli.Context) error {

    // Error channels
    errorChannel := make(chan error)
    fatalErrorChannel := make(chan error)

    // Start beacon activity process
    go beacon.StartActivityProcess(c, errorChannel, fatalErrorChannel)

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

