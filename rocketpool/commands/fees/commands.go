package fees

import (
    "fmt"
    "strconv"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/commands"
)


// Register fee commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage user fees",
        Subcommands: []cli.Command{

            // Display fee
            cli.Command{
                Name:      "display",
                Aliases:   []string{"d"},
                Usage:     "Display the current user fee percentage",
                UsageText: "rocketpool fee display",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := commands.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err
                    }

                    // Run command
                    fmt.Println("User fee:")
                    return nil

                },
            },

            // Vote on fee
            cli.Command{
                Name:      "vote",
                Aliases:   []string{"v"},
                Usage:     "Vote on the user fee percentage to be charged",
                UsageText: "rocketpool fee vote [fee percentage]" + "\n   " +
                           "- fee percentage must be a decimal number between 0 and 100",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var feePercent float64

                    // Validate arguments
                    err := commands.ValidateArgs(c, 1, func(messages *[]string) {
                        var err error

                        // Parse fee percentage
                        feePercent, err = strconv.ParseFloat(c.Args().Get(0), 64)
                        if err != nil {
                            *messages = append(*messages, "Invalid fee percentage - must be a decimal number")
                        }
                        if feePercent < 0 || feePercent > 100 {
                            *messages = append(*messages, "Invalid fee percentage - must be between 0 and 100")
                        }

                    })
                    if err != nil {
                        return err
                    }

                    // Run command
                    fmt.Println("Voting:", feePercent)
                    return nil

                },
            },

        },
    })
}

