package node

import (
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode-cli/rocketpool/utils/cli"
)


// Register node commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage node registration & state",
        Subcommands: []cli.Command{

            // Initialise the node with an account
            cli.Command{
                Name:      "init",
                Aliases:   []string{"i"},
                Usage:     "Initialize the node with an account",
                UsageText: "rocketpool node initialize",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := cliutils.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err
                    }

                    // Run command
                    return initNode(c)

                },
            },

            // Register the node with Rocket Pool
            cli.Command{
                Name:      "register",
                Aliases:   []string{"r"},
                Usage:     "Register the node on the Rocket Pool network",
                UsageText: "rocketpool node register",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := cliutils.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err
                    }

                    // Run command
                    return registerNode(c)

                },
            },

            // Set the node's timezone
            cli.Command{
                Name:      "timezone",
                Aliases:   []string{"t"},
                Usage:     "Set the node's timezone on the Rocket Pool network",
                UsageText: "rocketpool node timezone",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := cliutils.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err
                    }

                    // Run command
                    return setNodeTimezone(c)

                },
            },

        },
    })
}

