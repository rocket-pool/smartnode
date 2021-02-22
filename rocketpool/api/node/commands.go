package node

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
        Usage:     "Manage the node",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the node's status",
                UsageText: "rocketpool api node status",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(getStatus(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-register",
                Usage:     "Check whether the node can be registered with Rocket Pool",
                UsageText: "rocketpool api node can-register",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(canRegisterNode(c))
                    return nil

                },
            },
            cli.Command{
                Name:      "register",
                Aliases:   []string{"r"},
                Usage:     "Register the node with Rocket Pool",
                UsageText: "rocketpool api node register timezone-location",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    timezoneLocation, err := cliutils.ValidateTimezoneLocation("timezone location", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(registerNode(c, timezoneLocation))
                    return nil

                },
            },

            cli.Command{
                Name:      "set-withdrawal-address",
                Aliases:   []string{"w"},
                Usage:     "Set the node's withdrawal address",
                UsageText: "rocketpool api node set-withdrawal-address address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    withdrawalAddress, err := cliutils.ValidateAddress("withdrawal address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(setWithdrawalAddress(c, withdrawalAddress))
                    return nil

                },
            },

            cli.Command{
                Name:      "set-timezone",
                Aliases:   []string{"t"},
                Usage:     "Set the node's timezone location",
                UsageText: "rocketpool api node set-timezone timezone-location",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    timezoneLocation, err := cliutils.ValidateTimezoneLocation("timezone location", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(setTimezoneLocation(c, timezoneLocation))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-swap-rpl",
                Usage:     "Check whether the node can swap old RPL for new RPL",
                UsageText: "rocketpool api node can-swap-rpl amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("swap amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canNodeSwapRpl(c, amountWei))
                    return nil

                },
            },
            cli.Command{
                Name:      "swap-rpl",
                Aliases:   []string{"p"},
                Usage:     "Swap old RPL for new RPL",
                UsageText: "rocketpool api node swap-rpl amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("swap amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(nodeSwapRpl(c, amountWei))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-stake-rpl",
                Usage:     "Check whether the node can stake RPL",
                UsageText: "rocketpool api node can-stake-rpl amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    amountWei, err := cliutils.ValidateDepositWeiAmount("stake amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canNodeStakeRpl(c, amountWei))
                    return nil

                },
            },
            cli.Command{
                Name:      "stake-rpl",
                Aliases:   []string{"k"},
                Usage:     "Stake RPL against the node",
                UsageText: "rocketpool api node stake-rpl amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    amountWei, err := cliutils.ValidateDepositWeiAmount("stake amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(nodeStakeRpl(c, amountWei))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-withdraw-rpl",
                Usage:     "Check whether the node can withdraw staked RPL",
                UsageText: "rocketpool api node can-withdraw-rpl amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    amountWei, err := cliutils.ValidateDepositWeiAmount("withdrawal amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canNodeWithdrawRpl(c, amountWei))
                    return nil

                },
            },
            cli.Command{
                Name:      "withdraw-rpl",
                Aliases:   []string{"i"},
                Usage:     "Withdraw RPL staked against the node",
                UsageText: "rocketpool api node withdraw-rpl amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    amountWei, err := cliutils.ValidateDepositWeiAmount("withdrawal amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(nodeWithdrawRpl(c, amountWei))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-deposit",
                Usage:     "Check whether the node can make a deposit",
                UsageText: "rocketpool api node can-deposit amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    amountWei, err := cliutils.ValidateDepositWeiAmount("deposit amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canNodeDeposit(c, amountWei))
                    return nil

                },
            },
            cli.Command{
                Name:      "deposit",
                Aliases:   []string{"d"},
                Usage:     "Make a deposit and create a minipool",
                UsageText: "rocketpool api node deposit amount min-fee",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    amountWei, err := cliutils.ValidateDepositWeiAmount("deposit amount", c.Args().Get(0))
                    if err != nil { return err }
                    minNodeFee, err := cliutils.ValidateFraction("minimum node fee", c.Args().Get(1))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(nodeDeposit(c, amountWei, minNodeFee))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-send",
                Usage:     "Check whether the node can send ETH or tokens to an address",
                UsageText: "rocketpool api node can-send amount token",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("send amount", c.Args().Get(0))
                    if err != nil { return err }
                    token, err := cliutils.ValidateTokenType("token type", c.Args().Get(1))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canNodeSend(c, amountWei, token))
                    return nil

                },
            },
            cli.Command{
                Name:      "send",
                Aliases:   []string{"n"},
                Usage:     "Send ETH or tokens from the node account to an address",
                UsageText: "rocketpool api node send amount token to",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 3); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("send amount", c.Args().Get(0))
                    if err != nil { return err }
                    token, err := cliutils.ValidateTokenType("token type", c.Args().Get(1))
                    if err != nil { return err }
                    toAddress, err := cliutils.ValidateAddress("to address", c.Args().Get(2))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(nodeSend(c, amountWei, token, toAddress))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-burn",
                Usage:     "Check whether the node can burn tokens for ETH",
                UsageText: "rocketpool api node can-burn amount token",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("burn amount", c.Args().Get(0))
                    if err != nil { return err }
                    token, err := cliutils.ValidateBurnableTokenType("token type", c.Args().Get(1))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canNodeBurn(c, amountWei, token))
                    return nil

                },
            },
            cli.Command{
                Name:      "burn",
                Aliases:   []string{"b"},
                Usage:     "Burn tokens for ETH",
                UsageText: "rocketpool api node burn amount token",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    amountWei, err := cliutils.ValidatePositiveWeiAmount("burn amount", c.Args().Get(0))
                    if err != nil { return err }
                    token, err := cliutils.ValidateBurnableTokenType("token type", c.Args().Get(1))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(nodeBurn(c, amountWei, token))
                    return nil

                },
            },

        },
    })
}

