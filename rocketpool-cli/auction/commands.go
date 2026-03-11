package auction

import (
	"context"

	"github.com/urfave/cli/v3"

	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register commands
func RegisterCommands(app *cli.Command, name string, aliases []string) {
	app.Commands = append(app.Commands, &cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage Rocket Pool RPL auctions",
		Commands: []*cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get RPL auction status",
				UsageText: "rocketpool auction status",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getStatus()

				},
			},

			{
				Name:      "lots",
				Aliases:   []string{"l"},
				Usage:     "Get RPL lots for auction",
				UsageText: "rocketpool auction lots",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getLots()

				},
			},

			{
				Name:      "create-lot",
				Aliases:   []string{"t"},
				Usage:     "Create a new lot",
				UsageText: "rocketpool auction create-lot",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return createLot(c.Bool("yes"))

				},
			},

			{
				Name:      "bid-lot",
				Aliases:   []string{"b"},
				Usage:     "Bid on a lot",
				UsageText: "rocketpool auction bid-lot [options]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "lot",
						Aliases: []string{"l"},
						Usage:   "The ID of the lot to bid on",
					},
					&cli.StringFlag{
						Name:    "amount",
						Aliases: []string{"a"},
						Usage:   "The amount of ETH to bid (or 'max')",
					},
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm bid",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

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
					return bidOnLot(c.String("lot"), c.String("amount"), c.Bool("yes"))

				},
			},

			{
				Name:      "claim-lot",
				Aliases:   []string{"c"},
				Usage:     "Claim RPL from a lot",
				UsageText: "rocketpool auction claim-lot [options]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "lot",
						Aliases: []string{"l"},
						Usage:   "The lot to claim RPL from (lot ID or 'all')",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

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
					return claimFromLot(c.String("lot"), c.Bool("yes"))

				},
			},

			{
				Name:      "recover-lot",
				Aliases:   []string{"r"},
				Usage:     "Recover unclaimed RPL from a lot (returning it to the auction contract)",
				UsageText: "rocketpool auction recover-lot [options]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "lot",
						Aliases: []string{"l"},
						Usage:   "The lot to recover unclaimed RPL from (lot ID or 'all')",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

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
					return recoverRplFromLot(c.String("lot"), c.Bool("yes"))

				},
			},
		},
	})
}
