package queue

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
		Usage:   "Manage the Rocket Pool deposit queue",
		Commands: []*cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get the deposit pool and minipool queue status",
				UsageText: "rocketpool queue status",
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
				Name:      "process",
				Aliases:   []string{"p"},
				Usage:     "Process the deposit pool",
				UsageText: "rocketpool queue process",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return processQueue(c.Bool("yes"))

				},
			},

			{
				Name:      "assign-deposits",
				Aliases:   []string{"ad"},
				Usage:     "Assign deposits to queued validators",
				UsageText: "rocketpool queue assign-deposits",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm all prompts",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return assignDeposits(c.Bool("yes"))

				},
			},
		},
	})
}
