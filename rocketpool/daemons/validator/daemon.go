package validator

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/daemons/validator/beacon"
    beaconchain "github.com/rocket-pool/smartnode-cli/rocketpool/services/beacon-chain"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/messaging"
)


// Run daemon
func Run(c *cli.Context) error {

    // Error channel
    fatalErrorChannel := make(chan error)

    // Initialise publisher
    publisher := messaging.NewPublisher()

    // Initialise beacon chain client
    beaconClient := beaconchain.NewClient(c.GlobalString("beacon"), publisher)

    // Start beacon processes
    go beacon.StartActivityProcess(publisher, fatalErrorChannel)
    go beacon.StartWithdrawalProcess(c, fatalErrorChannel)

    // Connect to beacon chain server
    beaconClient.Connect()

    // Block thread; return fatal errors
    select {
        case err := <-fatalErrorChannel:
            return err
    }
    return nil

}

