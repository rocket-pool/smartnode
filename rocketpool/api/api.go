package api

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/rocketpool/api/minipool"
    "github.com/rocket-pool/smartnode/rocketpool/api/network"
    "github.com/rocket-pool/smartnode/rocketpool/api/node"
    "github.com/rocket-pool/smartnode/rocketpool/api/queue"
    "github.com/rocket-pool/smartnode/rocketpool/api/wallet"
)


// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {

    // CLI command
    command := cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Run Rocket Pool API commands",
        Subcommands: []cli.Command{},
    }

    // Gas price & limit flags
    command.Flags = []cli.Flag{
        cli.StringFlag{
            Name:  "gasPrice, g",
            Usage: "Desired gas price in wei",
        },
        cli.StringFlag{
            Name:  "gasLimit, l",
            Usage: "Desired gas limit in wei",
        },
    }

    // Don't show help message for api errors because of JSON serialisation
    command.OnUsageError = func(context *cli.Context, err error, isSubcommand bool) error {
        return err
    }

    // Register subcommands
    minipool.RegisterSubcommands(&command, "minipool", []string{"m"})
     network.RegisterSubcommands(&command, "network",  []string{"e"})
        node.RegisterSubcommands(&command, "node",     []string{"n"})
       queue.RegisterSubcommands(&command, "queue",    []string{"q"})
      wallet.RegisterSubcommands(&command, "wallet",   []string{"w"})

    // Register CLI command
    app.Commands = append(app.Commands, command)

}

