package main

import (
    "log"
    "os"

    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode/rocketpool-minipools/minipools"
    "github.com/rocket-pool/smartnode/shared/services"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Run application
func main() {

    // Initialise application
    app := cli.NewApp()

    // Set application info
    app.Name = "rocketpool-minipools"
    app.Usage = "Rocket Pool minipool management daemon"
    app.Version = "0.0.1"
    app.Authors = []cli.Author{
        cli.Author{
            Name:  "Darren Langley",
            Email: "darren@rocketpool.net",
        },
        cli.Author{
            Name:  "David Rugendyke",
            Email: "david@rocketpool.net",
        },
        cli.Author{
            Name:  "Jake Pospischil",
            Email: "jake@rocketpool.net",
        },
    }
    app.Copyright = "(c) 2019 Rocket Pool Pty Ltd"

    // Configure application
    cliutils.Configure(app)

    // Set application action
    app.Action = func(c *cli.Context) error {

        // Check argument count
        if len(c.Args()) != 4 {
            return cli.NewExitError("USAGE:" + "\n   " + "rocketpool-minipools rpPath rpNetwork powContainer beaconContainer", 1)
        }

        // Get arguments
        rpPath := c.Args().Get(0)
        rpNetwork := c.Args().Get(1)
        powContainer := c.Args().Get(2)
        beaconContainer := c.Args().Get(3)

        // Run process
        return run(c, rpPath, rpNetwork, powContainer, beaconContainer)

    }

    // Run application
    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }

}


// Run process
func run(c *cli.Context, rpPath string, rpNetwork string, powContainer string, beaconContainer string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        ClientConnection: true,
        ClientSync: true,
        CM: true,
        Docker: true,
        LoadContracts: []string{"utilAddressSetStorage"},
        LoadAbis: []string{"rocketMinipool"},
    })
    if err != nil {
        return err
    }

    // Start minipools management process
    go minipools.StartManagementProcess(p, rpPath, rpNetwork, powContainer, beaconContainer)

    // Block thread
    select {}

}

