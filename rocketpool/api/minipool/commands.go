package minipool

import (
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/utils/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
    command.Subcommands = append(command.Subcommands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage the node's minipools",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get a list of the node's minipools",
                UsageText: "rocketpool api minipool status",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(getStatus(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "leader",
                Aliases:   []string{"l"},
                Usage:     "validator leaderboard",
                UsageText: "rocketpool api minipool leader",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(getLeader(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-refund",
                Usage:     "Check whether the node can refund ETH from the minipool",
                UsageText: "rocketpool api minipool can-refund minipool-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canRefundMinipool(c, minipoolAddress))
                    return nil

                },
            },
            cli.Command{
                Name:      "refund",
                Aliases:   []string{"r"},
                Usage:     "Refund ETH belonging to the node from a minipool",
                UsageText: "rocketpool api minipool refund minipool-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(refundMinipool(c, minipoolAddress))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-dissolve",
                Usage:     "Check whether the minipool can be dissolved",
                UsageText: "rocketpool api minipool can-dissolve minipool-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canDissolveMinipool(c, minipoolAddress))
                    return nil

                },
            },
            cli.Command{
                Name:      "dissolve",
                Aliases:   []string{"d"},
                Usage:     "Dissolve an initialized or prelaunch minipool",
                UsageText: "rocketpool api minipool dissolve minipool-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(dissolveMinipool(c, minipoolAddress))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-exit",
                Usage:     "Check whether the minipool can be exited from the beacon chain",
                UsageText: "rocketpool api minipool can-exit minipool-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canExitMinipool(c, minipoolAddress))
                    return nil

                },
            },
            cli.Command{
                Name:      "exit",
                Aliases:   []string{"e"},
                Usage:     "Exit a staking minipool from the beacon chain",
                UsageText: "rocketpool api minipool exit minipool-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(exitMinipool(c, minipoolAddress))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-close",
                Usage:     "Check whether the minipool can be closed",
                UsageText: "rocketpool api minipool can-close minipool-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canCloseMinipool(c, minipoolAddress))
                    return nil

                },
            },
            cli.Command{
                Name:      "close",
                Aliases:   []string{"c"},
                Usage:     "Withdraw balance from a dissolved minipool and close it",
                UsageText: "rocketpool api minipool close minipool-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(closeMinipool(c, minipoolAddress))
                    return nil

                },
            },

        },
    })
}

