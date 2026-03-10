package claims

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
		Usage:   "View and claim all available rewards and credits across the node",
		Commands: []*cli.Command{
			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Display all available rewards and credits across the node without claiming",
				UsageText: "rocketpool claims status",
				Action: func(ctx context.Context, c *cli.Command) error {
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}
					return claimAll(c.String("restake-amount"), true, c.Bool("yes"))
				},
			},
			{
				Name:      "claim-all",
				Aliases:   []string{"c"},
				Usage:     "Display all available rewards and credits and claim them",
				UsageText: "rocketpool claims claim-all [options]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm all claims",
					},
					&cli.StringFlag{
						Name:  "restake-amount",
						Usage: "The amount of RPL to automatically restake during periodic reward claiming (or 'all')",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}
					return claimAll(c.String("restake-amount"), false, c.Bool("yes"))
				},
			},
		},
	})
}
