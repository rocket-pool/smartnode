package fee

import (
    "strconv"

    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register deposit commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage user fees",
        Subcommands: []cli.Command{

            // Get the current user fee
            cli.Command{
                Name:      "get",
                Aliases:   []string{"g"},
                Usage:     "Get the current user fee percentage",
                UsageText: "rocketpool fee get",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return getUserFee(c)

                },
            },

            // Set the target user fee to vote for
            cli.Command{
                Name:      "set",
                Aliases:   []string{"s"},
                Usage:     "Set the target user fee percentage to vote for",
                UsageText: "rocketpool fee set percent",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var feePercent float64

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 1, func(messages *[]string) {
                        var err error

                        // Parse fee percentage
                        if feePercent, err = strconv.ParseFloat(c.Args().Get(0), 64); err != nil {
                            *messages = append(*messages, "Invalid fee percentage - must be a decimal number")
                        } else if feePercent < 0 || feePercent > 100 {
                            *messages = append(*messages, "Invalid fee percentage - must be between 0 and 100")
                        }

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return setTargetUserFee(c, feePercent)

                },
            },

        },
    })
}

