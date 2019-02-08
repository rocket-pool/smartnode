package deposit

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/api/deposit/actions"
    cliutils "github.com/rocket-pool/smartnode-cli/rocketpool/utils/cli"
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
                    err := cliutils.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err
                    }

                    // Run action
                    return actions.GetDepositStatus(c)

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
                    err := cliutils.ValidateArgs(c, 1, func(messages *[]string) {

                        // Parse duration ID
                        durationId = c.Args().Get(0)
                        switch durationId {
                            case "3m":
                            case "6m":
                            case "12m":
                            default:
                                *messages = append(*messages, "Invalid durationID - valid IDs are '3m', '6m' and '12m'")
                        }

                    })
                    if err != nil {
                        return err
                    }

                    // Run action
                    return actions.ReserveDeposit(c, durationId)

                },
            },

            // Cancel a node deposit reservation
            cli.Command{
                Name:      "cancel",
                Aliases:   []string{"c"},
                Usage:     "Cancel a deposit reservation",
                UsageText: "rocketpool deposit cancel",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := cliutils.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err
                    }

                    // Run action
                    return actions.CancelDeposit(c)

                },
            },

            // Complete a node deposit

        },
    })
}

