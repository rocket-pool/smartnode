package megapool

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
		Usage:   "Manage the node's megapool",
		Subcommands: []cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get the node's megapool status",
				UsageText: "rocketpool api megapool status",
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
				Name:      "can-repay-debt",
				Usage:     "Check if we can repay the megapool debt",
				UsageText: "rocketpool api megapool can-repay-debt amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Get amount
					amount, err := cliutils.ValidatePositiveWeiAmount(c.Args().Get(0), "amount")
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canRepayDebt(c, amount))
					return nil

				},
			},
			{
				Name:      "repay-debt",
				Aliases:   []string{"rd"},
				Usage:     "Repay the megapool debt",
				UsageText: "rocketpool api megapool repay-debt amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Get amount
					amount, err := cliutils.ValidatePositiveWeiAmount(c.Args().Get(0), "amount")
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(repayDebt(c, amount))
					return nil

				},
			},
		},
	})
}
