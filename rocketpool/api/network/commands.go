package network

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
        Usage:     "Manage Rocket Pool network parameters",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "node-fee",
                Aliases:   []string{"f"},
                Usage:     "Get the current network node commission rate",
                UsageText: "rocketpool api network node-fee",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(getNodeFee(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "rpl-price",
                Aliases:   []string{"p"},
                Usage:     "Get the current network RPL price in ETH",
                UsageText: "rocketpool api network rpl-price",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(getRplPrice(c))
                    return nil

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
                    api.PrintResponse(challenge(c, address))
                    return nil

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
                    api.PrintResponse(decide(c, address))
                    return nil

                },
            },

        },
    })
}

