package minipool

import (
    "github.com/urfave/cli"

    "github.com/ethereum/go-ethereum/common"

    cliutils "github.com/rocket-pool/smartnode-cli/rocketpool/utils/cli"
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
                    if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return getMinipoolStatus(c)

                },
            },

            // Withdraw node deposit from a minipool
            cli.Command{
                Name:      "withdraw",
                Aliases:   []string{"w"},
                Usage:     "Withdraw deposit from an initialized, withdrawn or timed out minipool",
                UsageText: "rocketpool minipool withdraw minipoolAddress" + "\n   " +
                           "- minipoolAddress must be a valid address",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var address string

                    // Validate arguments
                    if err := cliutils.ValidateArgs(c, 1, func(messages *[]string) {

                        // Get & validate address
                        if address = c.Args().Get(0); !common.IsHexAddress(address) {
                            *messages = append(*messages, "Invalid minipool address")
                        }

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return withdrawMinipool(c, address)

                },
            },

        },
    })
}

