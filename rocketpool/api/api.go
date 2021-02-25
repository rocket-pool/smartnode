package api

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/rocketpool/api/auction"
    "github.com/rocket-pool/smartnode/rocketpool/api/minipool"
    "github.com/rocket-pool/smartnode/rocketpool/api/network"
    "github.com/rocket-pool/smartnode/rocketpool/api/node"
    "github.com/rocket-pool/smartnode/rocketpool/api/queue"
    "github.com/rocket-pool/smartnode/rocketpool/api/tndao"
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

    // Register subcommands
     auction.RegisterSubcommands(&command, "auction",  []string{"a"})
    minipool.RegisterSubcommands(&command, "minipool", []string{"m"})
     network.RegisterSubcommands(&command, "network",  []string{"e"})
        node.RegisterSubcommands(&command, "node",     []string{"n"})
       queue.RegisterSubcommands(&command, "queue",    []string{"q"})
       tndao.RegisterSubcommands(&command, "tndao",    []string{"t"})
      wallet.RegisterSubcommands(&command, "wallet",   []string{"w"})

    // Register CLI command
    app.Commands = append(app.Commands, command)

}

