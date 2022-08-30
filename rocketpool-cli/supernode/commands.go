package supernode

import (
	"github.com/urfave/cli"

	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the node",
		Subcommands: []cli.Command{
			{
				Name:      "deposit",
				Aliases:   []string{"d"},
				Usage:     "Make a deposit and create a minipool under a supernode.",
				UsageText: "rocketpool node deposit [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "supernode-address, s",
						Usage: "The address of the supernode you want to make the minipool under.",
					},
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm deposit",
					},
					cli.StringFlag{
						Name:  "salt, l",
						Usage: "An optional seed to use when generating the new minipool's address. Use this if you want it to have a custom vanity address.",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("supernode-address") != "" {
						if _, err := cliutils.ValidateAddress("supernode-address", c.String("supernode-address")); err != nil {
							return err
						}
					}
					if c.String("salt") != "" {
						if _, err := cliutils.ValidateBigInt("salt", c.String("salt")); err != nil {
							return err
						}
					}

					// Run
					return nodeDeposit(c)

				},
			},
		},
	})
}
