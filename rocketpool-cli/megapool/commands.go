package megapool

import (
	"github.com/urfave/cli"

	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the node's megapool",
		Subcommands: []cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get the node's megapool status",
				UsageText: "rocketpool megapool status",
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
				Name:      "repay-debt",
				Usage:     "Repay megapool debt",
				UsageText: "rocketpool megapool repay-debt amount",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Get amount
					amount, err := cliutils.ValidatePositiveEthAmount("amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					return repayDebt(c, amount)
				},
			},
			{
				Name:      "exit-queue",
				Usage:     "Exit the megapool queue",
				UsageText: "rocketpool megapool exit-queue",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
					cli.StringFlag{
						Name:  "validator-index",
						Usage: "The validator index to exit",
					},
					cli.BoolFlag{
						Name:  "express",
						Usage: "Exit the validator from the express queue",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return exitQueue(c)
				},
			},
		},
	})
}
