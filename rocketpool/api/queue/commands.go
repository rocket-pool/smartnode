package queue

import (
    "github.com/urfave/cli"
)


// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
    command.Subcommands = append(command.Subcommands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage the Rocket Pool deposit queue",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the deposit pool and minipool queue status",
                UsageText: "rocketpool api queue status",
                Action: func(c *cli.Context) error { return nil },
            },

            cli.Command{
                Name:      "process",
                Aliases:   []string{"p"},
                Usage:     "Process the deposit pool",
                UsageText: "rocketpool api queue process",
                Action: func(c *cli.Context) error { return nil },
            },

        },
    })
}

