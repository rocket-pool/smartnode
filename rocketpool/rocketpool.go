package main

import (
    "log"
    "os"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/rocketpool/api"
    apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Run
func main() {

    // Initialise application
    app := cli.NewApp()

    // Set application info
    app.Name = "rocketpool"
    app.Usage = "Rocket Pool service"
    app.Version = "0.0.1"
    app.Authors = []cli.Author{
        cli.Author{
            Name:  "David Rugendyke",
            Email: "david@rocketpool.net",
        },
        cli.Author{
            Name:  "Jake Pospischil",
            Email: "jake@rocketpool.net",
        },
    }
    app.Copyright = "(c) 2020 Rocket Pool Pty Ltd"

    // Configure application
    cliutils.Configure(app)

    // Register commands
    api.RegisterCommands(app, "api", []string{"a"})

    // Get command being run
    var commandName string
    app.Before = func(c *cli.Context) error {
        commandName = c.Args().First()
        return nil
    }

    // Run application
    if err := app.Run(os.Args); err != nil {
        if commandName == "api" {
            apiutils.PrintErrorResponse(err)
        } else {
            log.Fatal(err)
        }
    }

}

