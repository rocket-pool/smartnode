package validator

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/daemons/validator/beacon"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services"
)


// Run daemon
func Run(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        Publisher: true,
        Beacon: true,
        VM: true,
        LoadContracts: []string{"utilAddressSetStorage"},
        LoadAbis: []string{"rocketMinipool"},
    })
    if err != nil {
        return err
    }

    // Start beacon processes
    go beacon.StartActivityProcess(p)
    go beacon.StartWithdrawalProcess(p)

    // Start services
    p.Beacon.Connect()
    p.VM.StartLoad()

    // Block thread
    select {}

}

