package minipool

import (
    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register minipool subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
    command.Subcommands = append(command.Subcommands, cli.Command{
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
                    if err := cliutils.ValidateAPIArgs(c, 0, nil); err != nil {
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
                Usage:     "Withdraw deposit from a minipool",
                UsageText: "rocketpool minipool withdraw address",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var address string

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 1, func(messages *[]string) {

                        // Validate address
                        address = c.Args().Get(0)
                        if !common.IsHexAddress(address) {
                            *messages = append(*messages, "Invalid minipool address - must be a valid Ethereum address")
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

