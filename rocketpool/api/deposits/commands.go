package deposits

import (
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode-cli/rocketpool/utils/cli"
)


// Register deposit commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage node deposits",
        Subcommands: []cli.Command{

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

                    // Run command
                    return reserveDeposit(c, durationId)

                },
            },

            // Cancel a node deposit

            // Finalise a node deposit

        },
    })
}

