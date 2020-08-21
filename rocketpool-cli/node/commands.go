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
                Name:      "register",
                Aliases:   []string{"r"},
                Usage:     "Register the node with Rocket Pool",
                UsageText: "rocketpool node register",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return registerNode(c)

                },
            },

            cli.Command{
                Name:      "set-timezone",
                Aliases:   []string{"t"},
                Usage:     "Set the node's timezone location",
                UsageText: "rocketpool node set-timezone",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return setTimezoneLocation(c)

                },
            },

            cli.Command{
                Name:      "deposit",
                Aliases:   []string{"d"},
                Usage:     "Make a deposit and create a minipool",
                UsageText: "rocketpool node deposit amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return nodeDeposit(c)

                },
            },

            cli.Command{
                Name:      "send",
                Aliases:   []string{"n"},
                Usage:     "Send ETH or tokens from the node account to an address",
                UsageText: "rocketpool node send amount token to",
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

