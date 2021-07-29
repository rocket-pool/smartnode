package network

import (
	"github.com/urfave/cli"

	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage Rocket Pool network parameters",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "node-fee",
                Aliases:   []string{"f"},
                Usage:     "Get the current network node commission rate",
                UsageText: "rocketpool network node-fee",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return getNodeFee(c)

                },
            },

            cli.Command{
                Name:      "rpl-price",
                Aliases:   []string{"p"},
                Usage:     "Get the current network RPL price in ETH",
                UsageText: "rocketpool network rpl-price",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return getRplPrice(c)

                },
            },

            cli.Command{
                Name:      "challenge",
                Aliases:   []string{"c"},
                Usage:     "Get the current network RPL price in ETH",
                UsageText: "rocketpool api network rpl-price",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    address, err := cliutils.ValidateAddress("address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    return challenge(c, address)

                },
            },

            cli.Command{
                Name:      "decide",
                Aliases:   []string{"d"},
                Usage:     "Get the current network RPL price in ETH",
                UsageText: "rocketpool api network rpl-price",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    address, err := cliutils.ValidateAddress("address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    return decide(c, address)

                },
            },

        },
    })
}

