package node

import (
    "strconv"
    "strings"

    "gopkg.in/urfave/cli.v1"

    cliutils "github.com/rocket-pool/smartnode-cli/shared/utils/cli"
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
                Usage:     "Get the node's status",
                UsageText: "rocketpool node status",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return getNodeStatus(c)

                },
            },

            // Initialise the node with an account
            cli.Command{
                Name:      "init",
                Aliases:   []string{"i"},
                Usage:     "Initialize the node with an account",
                UsageText: "rocketpool node initialize",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
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
                    if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return registerNode(c)

                },
            },

            // Withdraw resources from the node
            cli.Command{
                Name:      "withdraw",
                Aliases:   []string{"w"},
                Usage:     "Withdraw resources from the node",
                UsageText: "rocketpool node withdraw amount unit" + "\n   " +
                           "- amount must be a positive decimal number" + "\n   " +
                           "- valid units are 'eth' and 'rpl'",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var amount float64
                    var unit string

                    // Validate arguments
                    if err := cliutils.ValidateArgs(c, 2, func(messages *[]string) {
                        var err error

                        // Parse amount
                        if amount, err = strconv.ParseFloat(c.Args().Get(0), 64); err != nil || amount <= 0 {
                            *messages = append(*messages, "Invalid amount - must be a positive decimal number")
                        }

                        // Parse unit
                        unit = strings.ToUpper(c.Args().Get(1))
                        switch unit {
                            case "ETH":
                            case "RPL":
                            default:
                                *messages = append(*messages, "Invalid unit - valid units are 'ETH' and 'RPL'")
                        }

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return withdrawFromNode(c, amount, unit)

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
                    if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return setNodeTimezone(c)

                },
            },

        },
    })
}

