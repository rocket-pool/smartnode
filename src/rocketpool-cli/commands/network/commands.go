package network

import (
	"github.com/urfave/cli/v2"

	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/shared/utils"
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
					&cli.StringFlag{
						Name:    generateTreeEcFlag,
						Aliases: []string{"e"},
						Usage:   "The URL of a separate execution client you want to use for generation (ignore this flag to use your primary exeuction client). Use this if your primary client is not an archive node, and you need to provide a separate archive node URL.",
					},
					&cli.Uint64Flag{
						Name:  generateTreeIndexFlag,
						Usage: "The index of the rewards interval you want to generate the tree for",
					},
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

			{
				Name:    "initialize-voting",
				Aliases: []string{"iv"},
				Usage:   "Unlocks a node operator's voting power (only required for node operators who registered before governance structure was in place)",
				Action: func(c *cli.Context) error {
					// Run
					return initializeVoting(c)
				},
			},

			{
				Name:      "set-voting-delegate",
				Aliases:   []string{"svd"},
				Usage:     "Set the address you want to use when voting on Rocket Pool on-chain governance proposals, or the address you want to delegate your voting power to.",
				ArgsUsage: "address",
				Flags: []cli.Flag{
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					delegate := c.Args().Get(0)
					// Run
					return setVotingDelegate(c, delegate)
				},
			},
		},
	})
}
