package node

import (
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
    command.Subcommands = append(command.Subcommands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage the node",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the node's status",
                UsageText: "rocketpool api node status",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.CheckAPIArgCount(c, 0); err != nil { return err }

                    // Run
                    return getStatus(c)

                },
            },

            cli.Command{
                Name:      "register",
                Aliases:   []string{"r"},
                Usage:     "Register the node with Rocket Pool",
                UsageText: "rocketpool api node register timezone-location",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.CheckAPIArgCount(c, 1); err != nil { return err }
                    timezoneLocation := c.Args().Get(0)
                    if err := cliutils.ValidateTimezoneLocation("timezone location", timezoneLocation); err != nil { return err }

                    // Run
                    return registerNode(c, timezoneLocation)

                },
            },

            cli.Command{
                Name:      "set-timezone",
                Aliases:   []string{"t"},
                Usage:     "Set the node's timezone location",
                UsageText: "rocketpool api node set-timezone timezone-location",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.CheckAPIArgCount(c, 1); err != nil { return err }
                    timezoneLocation := c.Args().Get(0)
                    if err := cliutils.ValidateTimezoneLocation("timezone location", timezoneLocation); err != nil { return err }

                    // Run
                    return setTimezoneLocation(c, timezoneLocation)

                },
            },

            cli.Command{
                Name:      "deposit",
                Aliases:   []string{"d"},
                Usage:     "Make a deposit and create a minipool",
                UsageText: "rocketpool api node deposit amount min-fee",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.CheckAPIArgCount(c, 2); err != nil { return err }
                    amount := c.Args().Get(0)
                    minNodeFeeStr := c.Args().Get(1)
                    amountWei, err := cliutils.ValidateTokenAmount("deposit amount", amount)
                    if err != nil { return err }
                    minNodeFee, err := cliutils.ValidateFraction("minimum node fee", minNodeFeeStr)
                    if err != nil { return err }

                    // Run
                    return nodeDeposit(c, amountWei, minNodeFee)

                },
            },

            cli.Command{
                Name:      "send",
                Aliases:   []string{"n"},
                Usage:     "Send ETH or tokens from the node account to an address",
                UsageText: "rocketpool api node send amount token to",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.CheckAPIArgCount(c, 3); err != nil { return err }
                    amount := c.Args().Get(0)
                    token := c.Args().Get(1)
                    toAddressStr := c.Args().Get(2)
                    amountWei, err := cliutils.ValidateTokenAmount("send amount", amount)
                    if err != nil { return err }
                    if err := cliutils.ValidateTokenType("token type", token); err != nil { return err }
                    toAddress, err := cliutils.ValidateAddress("to address", toAddressStr)
                    if err != nil { return err }

                    // Run
                    return nodeSend(c, amountWei, token, toAddress)

                },
            },

            cli.Command{
                Name:      "burn",
                Aliases:   []string{"b"},
                Usage:     "Burn tokens for ETH",
                UsageText: "rocketpool api node burn amount token",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.CheckAPIArgCount(c, 2); err != nil { return err }
                    amount := c.Args().Get(0)
                    token := c.Args().Get(1)
                    amountWei, err := cliutils.ValidateTokenAmount("burn amount", amount)
                    if err != nil { return err }
                    if err := cliutils.ValidateBurnableType("token type", token); err != nil { return err }

                    // Run
                    return nodeBurn(c, amountWei, token)

                },
            },

        },
    })
}

