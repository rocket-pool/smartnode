package fee

import (
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode-cli/rocketpool/utils/cli"
)


// Register deposit commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage user fees",
        Subcommands: []cli.Command{

            // Display the current user fee
            cli.Command{
                Name:      "display",
                Aliases:   []string{"d"},
                Usage:     "Display the current user fee percentage",
                UsageText: "rocketpool fee display",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := cliutils.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err
                    }

                    // Run command
                    return displayUserFee(c)

                },
            },

        },
    })
}

