package deposit

import (
    "gopkg.in/urfave/cli.v1"

    cliutils "github.com/rocket-pool/smartnode-cli/shared/utils/cli"
)


// Register deposit commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage node deposits",
        Subcommands: []cli.Command{

            // Get the node's current deposit status
            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the node's current deposit status",
                UsageText: "rocketpool deposit status",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return getDepositStatus(c)

                },
            },

            // Get the current deposit RPL requirement
            cli.Command{
                Name:      "required",
                Aliases:   []string{"q"},
                Usage:     "Get the current RPL requirement for a deposit",
                UsageText: "rocketpool deposit required durationID" + "\n   " +
                           "- durationID must be '3m', '6m' or '12m'",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var durationId string

                    // Validate arguments
                    if err := cliutils.ValidateArgs(c, 1, func(messages *[]string) {

                        // Parse duration ID
                        durationId = c.Args().Get(0)
                        switch durationId {
                            case "3m":
                            case "6m":
                            case "12m":
                            default:
                                *messages = append(*messages, "Invalid durationID - valid IDs are '3m', '6m' and '12m'")
                        }

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return getRplRequired(c, durationId)

                },
            },

            // Reserve a node deposit
            cli.Command{
                Name:      "reserve",
                Aliases:   []string{"r"},
                Usage:     "Reserve a deposit with a locked ETH:RPL ratio",
                UsageText: "rocketpool deposit reserve durationID" + "\n   " +
                           "- durationID must be '3m', '6m' or '12m'",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var durationId string

                    // Validate arguments
                    if err := cliutils.ValidateArgs(c, 1, func(messages *[]string) {

                        // Parse duration ID
                        durationId = c.Args().Get(0)
                        switch durationId {
                            case "3m":
                            case "6m":
                            case "12m":
                            default:
                                *messages = append(*messages, "Invalid durationID - valid IDs are '3m', '6m' and '12m'")
                        }

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return reserveDeposit(c, durationId)

                },
            },

            // Cancel a node deposit reservation
            cli.Command{
                Name:      "cancel",
                Aliases:   []string{"a"},
                Usage:     "Cancel a deposit reservation",
                UsageText: "rocketpool deposit cancel",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return cancelDeposit(c)

                },
            },

            // Complete a node deposit
            cli.Command{
                Name:      "complete",
                Aliases:   []string{"c"},
                Usage:     "Complete a deposit",
                UsageText: "rocketpool deposit complete",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return completeDeposit(c)

                },
            },

        },
    })
}

