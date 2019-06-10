package main

import (
    "log"
    "os"

    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode-cli/rocketpool-watchtower/watchtower"
    "github.com/rocket-pool/smartnode-cli/shared/services"
    cliutils "github.com/rocket-pool/smartnode-cli/shared/utils/cli"
)


// Run application
func main() {

    // Initialise application
    app := cli.NewApp()

    // Set application info
    app.Name = "rocketpool-watchtower"
    app.Usage = "Rocket Pool watchtower activity daemon"
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
        return run(c)
    }

    // Run application
    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }

}


// Run process
func run(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        ClientSync: true,
        CM: true,
        Publisher: true,
        Beacon: true,
        LoadContracts: []string{"rocketNodeAPI", "rocketNodeWatchtower", "rocketPool"},
        LoadAbis: []string{"rocketMinipool"},
    })
    if err != nil {
        return err
    }

    // Start minipool processes
    go watchtower.StartWatchtowerProcess(p)

    // Start services
    p.Beacon.Connect()

    // Block thread
    select {}

}

