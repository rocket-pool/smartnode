package claims

import (
	"github.com/urfave/cli"

	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "View and claim all available rewards and credits across the node",
		Subcommands: []cli.Command{
			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Display all available rewards and credits across the node without claiming",
				UsageText: "rocketpool claims status",
				Action: func(c *cli.Context) error {
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}
					return claimAll(c, true)
				},
			},
			{
				Name:      "claim-all",
				Aliases:   []string{"c"},
				Usage:     "Display all available rewards and credits and claim them",
				UsageText: "rocketpool claims claim-all [options]",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm all claims",
					},
					cli.StringFlag{
						Name:  "restake-amount",
						Usage: "The amount of RPL to automatically restake during periodic reward claiming (or 'all')",
					},
				},
				Action: func(c *cli.Context) error {
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}
					return claimAll(c, false)
				},
			},
		},
	})
}
