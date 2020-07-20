package minipool

import (
    "github.com/urfave/cli"
)


// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
    command.Subcommands = append(command.Subcommands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage the node's minipools",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get a list of the node's minipools",
                UsageText: "rocketpool api minipool status",
                Action: func(c *cli.Context) error { return nil },
            },

            cli.Command{
                Name:      "refund",
                Aliases:   []string{"r"},
                Usage:     "Refund ETH belonging to the node from a minipool",
                UsageText: "rocketpool api minipool refund minipool-address",
                Action: func(c *cli.Context) error { return nil },
            },

            cli.Command{
                Name:      "dissolve",
                Aliases:   []string{"d"},
                Usage:     "Dissolve an initialized or prelaunch minipool",
                UsageText: "rocketpool api minipool dissolve minipool-address",
                Action: func(c *cli.Context) error { return nil },
            },

            cli.Command{
                Name:      "exit",
                Aliases:   []string{"e"},
                Usage:     "Exit a staking minipool from the beacon chain",
                UsageText: "rocketpool api minipool exit minipool-address",
                Action: func(c *cli.Context) error { return nil },
            },

            cli.Command{
                Name:      "withdraw",
                Aliases:   []string{"w"},
                Usage:     "Withdraw final balance and rewards from a withdrawable minipool and close it",
                UsageText: "rocketpool api minipool withdraw minipool-address",
                Action: func(c *cli.Context) error { return nil },
            },

            cli.Command{
                Name:      "close",
                Aliases:   []string{"c"},
                Usage:     "Withdraw balance from a dissolved minipool and close it",
                UsageText: "rocketpool api minipool close minipool-address",
                Action: func(c *cli.Context) error { return nil },
            },

        },
    })
}

