package node

import (
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage the node",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the node's status",
                UsageText: "rocketpool node status",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return getStatus(c)

                },
            },

            cli.Command{
                Name:      "leader",
                Aliases:   []string{"l"},
                Usage:     "node leaderboard",
                UsageText: "rocketpool node leader",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return getLeader(c)

                },
            },

            cli.Command{
                Name:      "register",
                Aliases:   []string{"r"},
                Usage:     "Register the node with Rocket Pool",
                UsageText: "rocketpool node register [options]",
                Flags: []cli.Flag{
                    cli.StringFlag{
                        Name:  "timezone, t",
                        Usage: "The timezone location to register the node with (in the format 'Country/City')",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Validate flags
                    if c.String("timezone") != "" {
                        if _, err := cliutils.ValidateTimezoneLocation("timezone location", c.String("timezone")); err != nil { return err }
                    }

                    // Run
                    return registerNode(c)

                },
            },

            cli.Command{
                Name:      "set-withdrawal-address",
                Aliases:   []string{"w"},
                Usage:     "Set the node's withdrawal address",
                UsageText: "rocketpool node set-withdrawal-address [options] address",
                Flags: []cli.Flag{
                    cli.BoolFlag{
                        Name:  "yes, y",
                        Usage: "Automatically confirm setting withdrawal address",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    withdrawalAddress, err := cliutils.ValidateAddress("withdrawal address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    return setWithdrawalAddress(c, withdrawalAddress)

                },
            },

            cli.Command{
                Name:      "set-timezone",
                Aliases:   []string{"t"},
                Usage:     "Set the node's timezone location",
                UsageText: "rocketpool node set-timezone [options]",
                Flags: []cli.Flag{
                    cli.StringFlag{
                        Name:  "timezone, t",
                        Usage: "The timezone location to set for the node (in the format 'Country/City')",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Validate flags
                    if c.String("timezone") != "" {
                        if _, err := cliutils.ValidateTimezoneLocation("timezone location", c.String("timezone")); err != nil { return err }
                    }

                    // Run
                    return setTimezoneLocation(c)

                },
            },

            cli.Command{
                Name:      "swap-rpl",
                Aliases:   []string{"p"},
                Usage:     "Swap old RPL for new RPL",
                UsageText: "rocketpool node swap-rpl [options]",
                Flags: []cli.Flag{
                    cli.StringFlag{
                        Name:  "amount, a",
                        Usage: "The amount of old RPL to swap (or 'all')",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Validate flags
                    if c.String("amount") != "" && c.String("amount") != "all" {
                        if _, err := cliutils.ValidatePositiveEthAmount("swap amount", c.String("amount")); err != nil { return err }
                    }

                    // Run
                    return nodeSwapRpl(c)

                },
            },

            cli.Command{
                Name:      "stake-rpl",
                Aliases:   []string{"k"},
                Usage:     "Stake RPL against the node",
                UsageText: "rocketpool node stake-rpl [options]",
                Flags: []cli.Flag{
                    cli.StringFlag{
                        Name:  "amount, a",
                        Usage: "The amount of RPL to stake (or 'min', 'max', or 'all')",
                    },
                    cli.BoolFlag{
                        Name:  "yes, y",
                        Usage: "Automatically confirm RPL stake",
                    },
                    cli.BoolFlag{
                        Name:  "swap, s",
                        Usage: "Automatically confirm swapping old RPL before staking",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Validate flags
                    if c.String("amount") != "" && c.String("amount") != "min" && c.String("amount") != "max" && c.String("amount") != "all" {
                        if _, err := cliutils.ValidatePositiveEthAmount("stake amount", c.String("amount")); err != nil { return err }
                    }

                    // Run
                    return nodeStakeRpl(c)

                },
            },

            cli.Command{
                Name:      "withdraw-rpl",
                Aliases:   []string{"i"},
                Usage:     "Withdraw RPL staked against the node",
                UsageText: "rocketpool node withdraw-rpl [options]",
                Flags: []cli.Flag{
                    cli.StringFlag{
                        Name:  "amount, a",
                        Usage: "The amount of RPL to withdraw (or 'max')",
                    },
                    cli.BoolFlag{
                        Name:  "yes, y",
                        Usage: "Automatically confirm RPL withdrawal",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Validate flags
                    if c.String("amount") != "" && c.String("amount") != "max" {
                        if _, err := cliutils.ValidatePositiveEthAmount("withdrawal amount", c.String("amount")); err != nil { return err }
                    }

                    // Run
                    return nodeWithdrawRpl(c)

                },
            },

            cli.Command{
                Name:      "deposit",
                Aliases:   []string{"d"},
                Usage:     "Make a deposit and create a minipool",
                UsageText: "rocketpool node deposit [options]",
                Flags: []cli.Flag{
                    cli.StringFlag{
                        Name:  "amount, a",
                        Usage: "The amount of ETH to deposit (0, 16 or 32)",
                    },
                    cli.StringFlag{
                        Name:  "max-slippage, s",
                        Usage: "The maximum acceptable slippage in node commission rate for the deposit (or 'auto')",
                    },
                    cli.BoolFlag{
                        Name:  "yes, y",
                        Usage: "Automatically confirm deposit",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Validate flags
                    if c.String("amount") != "" {
                        if _, err := cliutils.ValidateDepositEthAmount("deposit amount", c.String("amount")); err != nil { return err }
                    }
                    if c.String("max-slippage") != "" && c.String("max-slippage") != "auto" {
                        if _, err := cliutils.ValidatePercentage("maximum commission rate slippage", c.String("max-slippage")); err != nil { return err }
                    }

                    // Run
                    return nodeDeposit(c)

                },
            },

            cli.Command{
                Name:      "send",
                Aliases:   []string{"n"},
                Usage:     "Send ETH or tokens from the node account to an address",
                UsageText: "rocketpool node send [options] amount token to",
                Flags: []cli.Flag{
                    cli.BoolFlag{
                        Name:  "yes, y",
                        Usage: "Automatically confirm token send",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 3); err != nil { return err }
                    amount, err := cliutils.ValidatePositiveEthAmount("send amount", c.Args().Get(0))
                    if err != nil { return err }
                    token, err := cliutils.ValidateTokenType("token type", c.Args().Get(1))
                    if err != nil { return err }
                    toAddress, err := cliutils.ValidateAddress("to address", c.Args().Get(2))
                    if err != nil { return err }

                    // Run
                    return nodeSend(c, amount, token, toAddress)

                },
            },

            cli.Command{
                Name:      "burn",
                Aliases:   []string{"b"},
                Usage:     "Burn tokens for ETH",
                UsageText: "rocketpool node burn amount token",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    amount, err := cliutils.ValidatePositiveEthAmount("burn amount", c.Args().Get(0))
                    if err != nil { return err }
                    token, err := cliutils.ValidateBurnableTokenType("token type", c.Args().Get(1))
                    if err != nil { return err }

                    // Run
                    return nodeBurn(c, amount, token)

                },
            },

        },
    })
}

