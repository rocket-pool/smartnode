package auction

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
		Usage:   "Manage Rocket Pool RPL auctions",
		Subcommands: []cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get RPL auction status",
				UsageText: "rocketpool api auction status",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getStatus(c))
					return nil

				},
			},

			{
				Name:      "lots",
				Aliases:   []string{"l"},
				Usage:     "Get RPL lots for auction",
				UsageText: "rocketpool api auction lots",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getLots(c))
					return nil

				},
			},

			{
				Name:      "create-lot",
				Aliases:   []string{"t"},
				Usage:     "Create a new lot",
				UsageText: "rocketpool api auction create-lot",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(createLot(c))
					return nil

				},
			},

			{
				Name:      "bid-lot",
				Aliases:   []string{"b"},
				Usage:     "Bid on a lot",
				UsageText: "rocketpool api auction bid-lot lot-id amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					lotIndex, err := cliutils.ValidateUint("lot ID", c.Args().Get(0))
					if err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("bid amount", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(bidOnLot(c, lotIndex, amountWei))
					return nil

				},
			},

			{
				Name:      "claim-lot",
				Aliases:   []string{"c"},
				Usage:     "Claim RPL from a lot",
				UsageText: "rocketpool api auction claim-lot lot-id",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					lotIndex, err := cliutils.ValidateUint("lot ID", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(claimFromLot(c, lotIndex))
					return nil

				},
			},

			{
				Name:      "recover-lot",
				Aliases:   []string{"r"},
				Usage:     "Recover unclaimed RPL from a lot (returning it to the auction contract)",
				UsageText: "rocketpool api auction recover-lot lot-id",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					lotIndex, err := cliutils.ValidateUint("lot ID", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(recoverRplFromLot(c, lotIndex))
					return nil

				},
			},
		},
	})
}
