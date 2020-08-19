package main

import (
    "fmt"
    "os"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/rocketpool/api"
    "github.com/rocket-pool/smartnode/rocketpool/node"
    "github.com/rocket-pool/smartnode/rocketpool/watchtower"
    apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
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

    // Set application flags
    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name:  "config",
            Usage: "Rocket Pool service global config absolute `path`",
            Value: "/.rocketpool/config.yml",
        },
        cli.StringFlag{
            Name:  "settings",
            Usage: "Rocket Pool service user config absolute `path`",
            Value: "/.rocketpool/settings.yml",
        },
        cli.StringFlag{
            Name:  "storageAddress",
            Usage: "Rocket Pool storage contract `address`",
        },
        cli.StringFlag{
            Name:  "password",
            Usage: "Rocket Pool wallet password file absolute `path`",
        },
        cli.StringFlag{
            Name:  "wallet",
            Usage: "Rocket Pool wallet file absolute `path`",
        },
        cli.StringFlag{
            Name:  "validatorKeychain",
            Usage: "Rocket Pool validator keychain absolute `path`",
        },
        cli.StringFlag{
            Name:  "eth1Provider",
            Usage: "Eth 1.0 provider `address`",
        },
        cli.StringFlag{
            Name:  "eth2Provider",
            Usage: "Eth 2.0 provider `address`",
        },
    }

    // Register commands
           api.RegisterCommands(app, "api",        []string{"a"})
          node.RegisterCommands(app, "node",       []string{"n"})
    watchtower.RegisterCommands(app, "watchtower", []string{"w"})

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
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }
    }

}

