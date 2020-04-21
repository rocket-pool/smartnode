package cli

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/rocketpool/cli/deposit"
    "github.com/rocket-pool/smartnode/rocketpool/cli/fee"
    "github.com/rocket-pool/smartnode/rocketpool/cli/minipool"
    "github.com/rocket-pool/smartnode/rocketpool/cli/node"
    "github.com/rocket-pool/smartnode/rocketpool/cli/queue"
)


// Register CLI commands
func RegisterCommands(app *cli.App, name string, aliases []string) {

    // CLI command
    command := cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Run Rocket Pool CLI command",
        Subcommands: []cli.Command{},
    }

    // Register subcommands
     deposit.RegisterSubcommands(&command, "deposit",  []string{"d"})
         fee.RegisterSubcommands(&command, "fee",      []string{"f"})
    minipool.RegisterSubcommands(&command, "minipool", []string{"m"})
        node.RegisterSubcommands(&command, "node",     []string{"n"})
       queue.RegisterSubcommands(&command, "queue",    []string{"q"})

    // Register CLI command
    app.Commands = append(app.Commands, command)

}

