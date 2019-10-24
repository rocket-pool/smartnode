package node

import (
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register node commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage node registration & state",
        Subcommands: []cli.Command{

            // Get the node's status
            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the node's status information",
                UsageText: "rocketpool node status",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return getNodeStatus(c)

                },
            },

        },
    })
}

