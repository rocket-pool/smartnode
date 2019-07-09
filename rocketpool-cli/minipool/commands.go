package minipool

import (
    "gopkg.in/urfave/cli.v1"

    "github.com/ethereum/go-ethereum/common"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
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
                UsageText: "rocketpool run minipool status",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return getMinipoolStatus(c)

                },
            },

            // Stop all running minipool containers
            cli.Command{
                Name:      "stop",
                Aliases:   []string{"t"},
                Usage:     "Stop all running minipool containers",
                UsageText: "rocketpool run minipool stop imageName",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateArgs(c, 1, nil); err != nil {
                        return err
                    }

                    // Get arguments
                    imageName := c.Args().Get(0)

                    // Run command
                    return stopMinipoolContainers(c, imageName)

                },
            },

            // Withdraw node deposit from a minipool
            cli.Command{
                Name:      "withdraw",
                Aliases:   []string{"w"},
                Usage:     "Withdraw deposit from an initialized, withdrawn or timed out minipool",
                UsageText: "rocketpool run minipool withdraw minipoolAddress" + "\n   " +
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

