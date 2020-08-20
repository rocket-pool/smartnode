package minipool

import (
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage the node's minipools",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get a list of the node's minipools",
                UsageText: "rocketpool minipool status",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return getStatus(c)

                },
            },

            cli.Command{
                Name:      "refund",
                Aliases:   []string{"r"},
                Usage:     "Refund ETH belonging to the node from a minipool",
                UsageText: "rocketpool minipool refund minipool-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    return refundMinipool(c, minipoolAddress)

                },
            },

            cli.Command{
                Name:      "dissolve",
                Aliases:   []string{"d"},
                Usage:     "Dissolve an initialized or prelaunch minipool",
                UsageText: "rocketpool minipool dissolve minipool-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    return dissolveMinipool(c, minipoolAddress)

                },
            },

            cli.Command{
                Name:      "exit",
                Aliases:   []string{"e"},
                Usage:     "Exit a staking minipool from the beacon chain",
                UsageText: "rocketpool minipool exit minipool-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    return exitMinipool(c, minipoolAddress)

                },
            },

            cli.Command{
                Name:      "withdraw",
                Aliases:   []string{"w"},
                Usage:     "Withdraw final balance and rewards from a withdrawable minipool and close it",
                UsageText: "rocketpool minipool withdraw minipool-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    return withdrawMinipool(c, minipoolAddress)

                },
            },

            cli.Command{
                Name:      "close",
                Aliases:   []string{"c"},
                Usage:     "Withdraw balance from a dissolved minipool and close it",
                UsageText: "rocketpool minipool close minipool-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    return closeMinipool(c, minipoolAddress)

                },
            },

        },
    })
}

