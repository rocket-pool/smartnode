package main

import (
    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode-cli/rocketpool-minipool/minipool"
    "github.com/rocket-pool/smartnode-cli/shared/services"
)


// Run daemon
func Run(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        ClientSync: true,
        CM: true,
        Publisher: true,
        Beacon: true,
        VM: true,
        LoadContracts: []string{"rocketAdmin", "rocketNodeWatchtower", "rocketPool", "utilAddressSetStorage"},
        LoadAbis: []string{"rocketMinipool"},
    })
    if err != nil {
        return err
    }

    // Start minipool processes
    go minipool.StartActivityProcess(p)
    go minipool.StartWithdrawalProcess(p)

    // Start services
    p.Beacon.Connect()
    p.VM.StartLoad()

    // Block thread
    select {}

}

