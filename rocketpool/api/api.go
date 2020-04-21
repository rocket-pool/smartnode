package api

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/rocketpool/api/deposit"
    "github.com/rocket-pool/smartnode/rocketpool/api/exchange"
    "github.com/rocket-pool/smartnode/rocketpool/api/fee"
    "github.com/rocket-pool/smartnode/rocketpool/api/minipool"
    "github.com/rocket-pool/smartnode/rocketpool/api/node"
    "github.com/rocket-pool/smartnode/rocketpool/api/queue"
)


// Register API commands
func RegisterCommands(app *cli.App, name string, aliases []string) {

    // CLI command
    command := cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Run Rocket Pool API command",
        Subcommands: []cli.Command{},
    }

    // Register subcommands
     deposit.RegisterSubcommands(&command, "deposit",  []string{"d"})
    exchange.RegisterSubcommands(&command, "exchange", []string{"x"})
         fee.RegisterSubcommands(&command, "fee",      []string{"f"})
    minipool.RegisterSubcommands(&command, "minipool", []string{"m"})
        node.RegisterSubcommands(&command, "node",     []string{"n"})
       queue.RegisterSubcommands(&command, "queue",    []string{"q"})

    // Register CLI command
    app.Commands = append(app.Commands, command)

}

