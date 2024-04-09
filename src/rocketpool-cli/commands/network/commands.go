package network

import (
	"github.com/urfave/cli/v2"

	cliutils "github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/shared/utils"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, &cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage Rocket Pool network parameters",
		Subcommands: []*cli.Command{
			{
				Name:    "stats",
				Aliases: []string{"s"},
				Usage:   "Get stats about the Rocket Pool network and its tokens",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getStats(c)
				},
			},

			{
				Name:    "timezone-map",
				Aliases: []string{"t"},
				Usage:   "Shows a table of the timezones that node operators belong to",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getTimezones(c)
				},
			},

			{
				Name:    "node-fee",
				Aliases: []string{"f"},
				Usage:   "Get the current network node commission rate",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getNodeFee(c)
				},
			},

			{
				Name:    "rpl-price",
				Aliases: []string{"p"},
				Usage:   "Get the current network RPL price in ETH",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getRplPrice(c)
				},
			},

			{
				Name:    "generate-rewards-tree",
				Aliases: []string{"g"},
				Usage:   "Generate and save the rewards tree file for the provided interval.\nNote that this is an asynchronous process, so it will return before the file is generated.\nYou will need to use `rocketpool service logs api` to follow its progress.",
				Flags: []cli.Flag{
					generateTreeEcFlag,
					generateTreeIndexFlag,
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return generateRewardsTree(c)
				},
			},

			{
				Name:    "dao-proposals",
				Aliases: []string{"d"},
				Usage:   "Get the currently active DAO proposals",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getActiveDAOProposals(c)
				},
			},
		},
	})
}
