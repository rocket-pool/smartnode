package faucet

import (
	"github.com/urfave/cli"

	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Access the legacy RPL faucet",
		Subcommands: []cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get the faucet's status",
				UsageText: "rocketpool faucet status",
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
				Name:      "withdraw-rpl",
				Aliases:   []string{"w"},
				Usage:     "Withdraw legacy RPL from the faucet",
				UsageText: "rocketpool faucet withdraw-rpl",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return withdrawRpl(c)

				},
			},
		},
	})
}
