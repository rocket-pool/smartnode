package network

import (
	"github.com/urfave/cli"

	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage Rocket Pool network parameters",
		Subcommands: []cli.Command{

			{
				Name:      "stats",
				Aliases:   []string{"s"},
				Usage:     "Get stats about the Rocket Pool network and its tokens",
				UsageText: "rocketpool network stats",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getStats(c)

				},
			},

			{
				Name:      "timezone-map",
				Aliases:   []string{"t"},
				Usage:     "Shows a table of the timezones that node operators belong to",
				UsageText: "rocketpool network timezone-map",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getTimezones(c)

				},
			},

			{
				Name:      "node-fee",
				Aliases:   []string{"f"},
				Usage:     "Get the current network node commission rate",
				UsageText: "rocketpool network node-fee",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getNodeFee(c)

				},
			},

			{
				Name:      "rpl-price",
				Aliases:   []string{"p"},
				Usage:     "Get the current network RPL price in ETH",
				UsageText: "rocketpool network rpl-price",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getRplPrice(c)

				},
			},

			{
				Name:      "generate-rewards-tree",
				Aliases:   []string{"g"},
				Usage:     "Generate and save the rewards tree file for the provided interval.\nNote that this is an asynchronous process, so it will return before the file is generated.\nYou will need to use `rocketpool service logs api` to follow its progress.",
				UsageText: "rocketpool network generate-rewards-tree",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "execution-client-url, e",
						Usage: "The URL of a separate execution client you want to use for generation (ignore this flag to use your primary exeuction client). Use this if your primary client is not an archive node, and you need to provide a separate archive node URL.",
					},
					cli.Uint64Flag{
						Name:  "index",
						Usage: "The index of the rewards interval you want to generate the tree for",
					},
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm any questions about tree generation",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return generateRewardsTree(c)

				},
			},

			{
				Name:      "dao-proposals",
				Aliases:   []string{"d"},
				Usage:     "Get the currently active DAO proposals",
				UsageText: "rocketpool network dao-proposals",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getActiveDAOProposals(c)

				},
			},
		},
	})
}
