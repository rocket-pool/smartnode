package auction

import (
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/rocketpool-cli/flags"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, &cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage Rocket Pool RPL auctions",
		Subcommands: []*cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get RPL auction status",
				UsageText: "rocketpool auction status",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getStatus(c)

				},
			},

			{
				Name:      "lots",
				Aliases:   []string{"l"},
				Usage:     "Get RPL lots for auction",
				UsageText: "rocketpool auction lots",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getLots(c)

				},
			},

			{
				Name:      "create-lot",
				Aliases:   []string{"t"},
				Usage:     "Create a new lot",
				UsageText: "rocketpool auction create-lot",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return createLot(c)

				},
			},

			{
				Name:      "bid-lot",
				Aliases:   []string{"b"},
				Usage:     "Bid on a lot",
				UsageText: "rocketpool auction bid-lot [options]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    bidLotFlag,
						Aliases: []string{"l"},
						Usage:   "The ID of the lot to bid on",
					},
					&cli.StringFlag{
						Name:    bidAmountFlag,
						Aliases: []string{"a"},
						Usage:   "The amount of ETH to bid (or 'max')",
					},
					&cli.BoolFlag{
						Name:    flags.YesFlag,
						Aliases: []string{"y"},
						Usage:   "Automatically confirm bid",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("lot") != "" {
						if _, err := cliutils.ValidateUint("lot ID", c.String("lot")); err != nil {
							return err
						}
					}
					if c.String("amount") != "" && c.String("amount") != "max" {
						if _, err := cliutils.ValidatePositiveEthAmount("bid amount", c.String("amount")); err != nil {
							return err
						}
					}

					// Run
					return bidOnLot(c)
				},
			},

			{
				Name:      "claim-lot",
				Aliases:   []string{"c"},
				Usage:     "Claim RPL from one or more lots",
				UsageText: "rocketpool auction claim-lot [options]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    claimLotsFlag,
						Aliases: []string{"l"},
						Usage:   "A comma-separated list of lot indices to claim RPL from (or 'all' to claim from all available lots)",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return claimFromLot(c)
				},
			},

			{
				Name:      "recover-lot",
				Aliases:   []string{"r"},
				Usage:     "Recover unclaimed RPL from a lot (returning it to the auction contract)",
				UsageText: "rocketpool auction recover-lot [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "lot, l",
						Usage: "The lot to recover unclaimed RPL from (lot ID or 'all')",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("lot") != "" && c.String("lot") != "all" {
						if _, err := cliutils.ValidateUint("lot ID", c.String("lot")); err != nil {
							return err
						}
					}

					// Run
					return recoverRplFromLot(c)

				},
			},
		},
	})
}
