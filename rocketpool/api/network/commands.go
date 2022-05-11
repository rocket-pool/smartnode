package network

import (
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/utils/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
	command.Subcommands = append(command.Subcommands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage Rocket Pool network parameters",
		Subcommands: []cli.Command{

			{
				Name:      "node-fee",
				Aliases:   []string{"f"},
				Usage:     "Get the current network node commission rate",
				UsageText: "rocketpool api network node-fee",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getNodeFee(c))
					return nil

				},
			},

			{
				Name:      "rpl-price",
				Aliases:   []string{"p"},
				Usage:     "Get the current network RPL price in ETH",
				UsageText: "rocketpool api network rpl-price",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getRplPrice(c))
					return nil

				},
			},

			{
				Name:      "stats",
				Aliases:   []string{"s"},
				Usage:     "Get stats about the Rocket Pool network and its tokens",
				UsageText: "rocketpool api network stats",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getStats(c))
					return nil

				},
			},

			{
				Name:      "timezone-map",
				Aliases:   []string{"t"},
				Usage:     "Get the table of node operators by timezone",
				UsageText: "rocketpool api network stats",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getTimezones(c))
					return nil

				},
			},

			{
				Name:      "merge-update-status",
				Aliases:   []string{"u"},
				Usage:     "Check if the contract upgrades for the merge have been deployed yet",
				UsageText: "rocketpool api network merge-update-status",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(mergeUpdateStatus(c))
					return nil

				},
			},

			{
				Name:      "can-generate-rewards-tree",
				Usage:     "Check if the rewards tree for the provided interval can be generated",
				UsageText: "rocketpool api network can-generate-rewards-tree index",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					index, err := cliutils.ValidateUint("index", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canGenerateRewardsTree(c, index))
					return nil

				},
			},

			{
				Name:      "generate-rewards-tree",
				Usage:     "Set a request marker for the watchtower to generate the rewards tree for the given interval",
				UsageText: "rocketpool api network generate-rewards-tree index",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					index, err := cliutils.ValidateUint("index", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(generateRewardsTree(c, index))
					return nil

				},
			},
		},
	})
}
