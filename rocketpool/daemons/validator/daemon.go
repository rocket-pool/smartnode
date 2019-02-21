package validator

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/daemons/validator/beacon"
)


// Run daemon
func Run(c *cli.Context) error {

    // Error channel
    fatalErrorChannel := make(chan error)

    // Start beacon activity process
    go beacon.StartActivityProcess(c, fatalErrorChannel)

    // Block thread; return fatal errors
    select {
        case err := <-fatalErrorChannel:
            return err
    }
    return nil

}

