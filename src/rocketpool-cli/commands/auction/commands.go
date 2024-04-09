package auction

import (
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/input"
	cliutils "github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/shared/utils"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, &cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage Rocket Pool RPL auctions",
		Subcommands: []*cli.Command{
			{
				Name:    "status",
				Aliases: []string{"s"},
				Usage:   "Get RPL auction status",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getStatus(c)
				},
			},

			{
				Name:    "lots",
				Aliases: []string{"l"},
				Usage:   "Get RPL lots for auction",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getLots(c)
				},
			},

			{
				Name:    "create-lot",
				Aliases: []string{"t"},
				Usage:   "Create a new lot",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return createLot(c)
				},
			},

			{
				Name:    "bid-lot",
				Aliases: []string{"b"},
				Usage:   "Bid on a lot",
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
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("lot") != "" {
						if _, err := input.ValidateUint("lot ID", c.String("lot")); err != nil {
							return err
						}
					}
					if c.String("amount") != "" && c.String("amount") != "max" {
						if _, err := input.ValidatePositiveEthAmount("bid amount", c.String("amount")); err != nil {
							return err
						}
					}

					// Run
					return bidOnLot(c)
				},
			},

			{
				Name:    "claim-lot",
				Aliases: []string{"c"},
				Usage:   "Claim RPL from one or more lots",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    claimLotsFlag,
						Aliases: []string{"l"},
						Usage:   "A comma-separated list of lot indices to claim RPL from (or 'all' to claim from all available lots)",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return claimFromLot(c)
				},
			},

			{
				Name:    "recover-lot",
				Aliases: []string{"r"},
				Usage:   "Recover unclaimed RPL from a lot (returning it to the auction contract)",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    recoverLotsFlag,
						Aliases: []string{"l"},
						Usage:   "A comma-separated list of lot indices to recover RPL from (or 'all' to recover from all available lots)",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return recoverRplFromLot(c)
				},
			},
		},
	})
}
