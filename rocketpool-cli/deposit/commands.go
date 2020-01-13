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
                Usage:     "Get the current RPL requirement for a deposit",
                UsageText: "rocketpool deposit required",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return getRplRequired(c)

                },
            },

            // Reserve and complete a node deposit
            cli.Command{
                Name:      "make",
                Aliases:   []string{"m"},
                Usage:     "Make a deposit into Rocket Pool",
                UsageText: "rocketpool deposit make durationID",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var durationId string

                    // Validate arguments
                    if err := cliutils.ValidateArgs(c, 1, func(messages *[]string) {

                        // Get duration ID
                        durationId = c.Args().Get(0)

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return makeDeposit(c, durationId)

                },
            },

        },
    })
}

