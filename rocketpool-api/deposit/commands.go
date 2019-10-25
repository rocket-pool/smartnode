package deposit

import (
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register deposit commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage node deposits",
        Subcommands: []cli.Command{

            // Get the current deposit RPL requirement
            cli.Command{
                Name:      "required",
                Aliases:   []string{"r"},
                Usage:     "Get the current RPL requirement information",
                UsageText: "rocketpool deposit required",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return getRplRequired(c)

                },
            },

            // Get the current deposit RPL requirement
            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the current deposit status information",
                UsageText: "rocketpool deposit status",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return getDepositStatus(c)

                },
            },

        },
    })
}

