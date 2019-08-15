package main

import (
    "log"
    "os"

    "github.com/urfave/cli"

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
            return cli.NewExitError("USAGE:" + "\n   " + "rocketpool-minipools rpPath imageName containerPrefix rpNetwork", 1)
        }

        // Get arguments
        rpPath := c.Args().Get(0)
        imageName := c.Args().Get(1)
        containerPrefix := c.Args().Get(2)
        rpNetwork := c.Args().Get(3)

        // Run process
        return run(c, rpPath, imageName, containerPrefix, rpNetwork)

    }

    // Run application
    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }

}


// Run process
func run(c *cli.Context, rpPath string, imageName string, containerPrefix string, rpNetwork string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        CM: true,
        Docker: true,
        LoadContracts: []string{"utilAddressSetStorage"},
        LoadAbis: []string{"rocketMinipool"},
        WaitPassword: true,
        WaitNodeAccount: true,
        WaitClientConn: true,
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }

    // Start minipools management process
    go minipools.StartManagementProcess(p, rpPath, imageName, containerPrefix, rpNetwork)

    // Block thread
    select {}

}

