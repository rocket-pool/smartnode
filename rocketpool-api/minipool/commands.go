package minipool

import (
    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register minipool commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage node minipools and users",
        Subcommands: []cli.Command{

            // Get the node's minipool statuses
            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the node's current minipool statuses",
                UsageText: "rocketpool minipool status",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return getMinipoolStatus(c)

                },
            },

        },
    })
}

