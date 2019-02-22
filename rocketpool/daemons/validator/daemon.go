package validator

import (
    "errors"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/daemons/validator/beacon"
    beaconchain "github.com/rocket-pool/smartnode-cli/rocketpool/services/beacon-chain"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/node"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/messaging"
)


// Run daemon
func Run(c *cli.Context) error {

    // Error channel
    fatalErrorChannel := make(chan error)

    // Connect to ethereum node
    client, err := ethclient.Dial(c.GlobalString("provider"))
    if err != nil {
        return errors.New("Error connecting to ethereum node: " + err.Error())
    }

    // Initialise publisher
    publisher := messaging.NewPublisher()

    // Initialise beacon chain client
    beaconClient := beaconchain.NewClient(c.GlobalString("beacon"), publisher)

    // Initialise validator manager
    vm, err := node.NewValidatorManager(c, client)
    if err != nil {
        return err
    }

    // Start beacon processes
    go beacon.StartActivityProcess(publisher, fatalErrorChannel)
    go beacon.StartWithdrawalProcess(c, client, vm, fatalErrorChannel)

    // Connect to beacon chain server
    beaconClient.Connect()

    // Start loading active validators
    vm.StartLoad()

    // Block thread; return fatal errors
    select {
        case err := <-fatalErrorChannel:
            return err
    }
    return nil

}

