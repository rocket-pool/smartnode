package megapool

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
		Usage:   "Manage the node's megapool",
		Commands: []*cli.Command{
			{
				Name:      "deposit",
				Aliases:   []string{"d"},
				Usage:     "Make a deposit and create a new validator on the megapool. Optionally specify count to make multiple deposits.",
				UsageText: "rocketpool megapool deposit [options]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm deposit",
					},
					&cli.Int64Flag{
						Name:    "express-tickets",
						Aliases: []string{"e"},
						Usage:   "Number of express tickets to use",
						Value:   -1,
					},
					&cli.Uint64Flag{
						Name:    "count",
						Aliases: []string{"c"},
						Usage:   "Number of deposits to make",
						Value:   0,
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return nodeMegapoolDeposit(c.Uint64("count"), c.Int64("express-tickets"), c.Bool("yes"))

				},
			},
			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get the node's megapool status",
				UsageText: "rocketpool megapool status",
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
				Name:      "validators",
				Aliases:   []string{"v"},
				Usage:     "Get a list of the megapool's validators",
				UsageText: "rocketpool megapool validators",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getValidatorStatus()

				},
			},
			{
				Name:      "repay-debt",
				Aliases:   []string{"r"},
				Usage:     "Repay megapool debt",
				UsageText: "rocketpool megapool repay-debt",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm the action",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return repayDebt(c.Bool("yes"))
				},
			},
			{
				Name:      "reduce-bond",
				Aliases:   []string{"e"},
				Usage:     "Reduce the megapool bond",
				UsageText: "rocketpool megapool reduce-bond",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm the action",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return reduceBond(c.Bool("yes"))
				},
			},
			{
				Name:      "claim",
				Aliases:   []string{"c"},
				Usage:     "Claim any megapool rewards that were distributed but not yet claimed",
				UsageText: "rocketpool megapool claim",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm the action",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return claim(c.Bool("yes"))
				},
			},
			{
				Name:      "stake",
				Aliases:   []string{"k"},
				Usage:     "Stake a megapool validator",
				UsageText: "rocketpool megapool stake",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm the action",
					},
					&cli.Uint64Flag{
						Name:  "validator-id",
						Usage: "The validator id to stake",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					var validatorId uint64
					if !c.IsSet("validator-id") {
						var err error
						var found bool
						validatorId, found, err = getStakableValidator()
						if err != nil {
							return err
						}
						if !found {
							return nil
						}
					} else {
						validatorId = c.Uint64("validator-id")
					}

					// Run
					return stake(validatorId, c.Bool("yes"))
				},
			},
			{
				Name:      "exit-queue",
				Aliases:   []string{"x"},
				Usage:     "Exit the megapool queue",
				UsageText: "rocketpool megapool exit-queue",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
					&cli.StringFlag{
						Name:  "validator-id",
						Usage: "The validator id to exit",
					},
					&cli.BoolFlag{
						Name:  "express",
						Usage: "Exit the validator from the express queue",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return exitQueue(c.String("validator-id"), c.Bool("yes"))
				},
			},
			{
				Name:      "dissolve-validator",
				Aliases:   []string{"i"},
				Usage:     "Dissolve a megapool validator",
				UsageText: "rocketpool megapool dissolve-validator",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
					&cli.Uint64Flag{
						Name:  "validator-id",
						Usage: "The validator id to exit",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					var validatorId uint64
					if !c.IsSet("validator-id") {
						var err error
						var found bool
						validatorId, found, err = getDissolvableValidator()
						if err != nil {
							return err
						}
						if !found {
							return nil
						}
					} else {
						validatorId = c.Uint64("validator-id")
					}

					// Run
					return dissolveValidator(validatorId, c.Bool("yes"))
				},
			},
			{
				Name:      "exit-validator",
				Aliases:   []string{"t"},
				Usage:     "Request to exit a megapool validator",
				UsageText: "rocketpool megapool exit-validator",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
					&cli.Uint64Flag{
						Name:  "validator-id",
						Usage: "The validator id to exit",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					var validatorId uint64
					if !c.IsSet("validator-id") {
						var err error
						var found bool
						validatorId, found, err = getExitableValidator()
						if err != nil {
							return err
						}
						if !found {
							return nil
						}
					} else {
						validatorId = c.Uint64("validator-id")
					}

					// Run
					return exitValidator(validatorId, c.Bool("yes"))
				},
			},
			{
				Name:      "notify-validator-exit",
				Aliases:   []string{"n"},
				Usage:     "Notify that a validator exit is in progress",
				UsageText: "rocketpool megapool notify-validator-exit",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
					&cli.Uint64Flag{
						Name:  "validator-id",
						Usage: "The validator id for which the exit is being notified",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					var validatorId uint64
					if !c.IsSet("validator-id") {
						var err error
						var found bool
						validatorId, found, err = getExitedValidator()
						if err != nil {
							return err
						}
						if !found {
							return nil
						}
					} else {
						validatorId = c.Uint64("validator-id")
					}

					// Run
					return notifyValidatorExit(validatorId, c.Bool("yes"))
				},
			},
			{
				Name:      "notify-final-balance",
				Aliases:   []string{"f"},
				Usage:     "Notify that a validator exit has completed and the final balance has been withdrawn",
				UsageText: "rocketpool megapool notify-final-balance",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
					&cli.Uint64Flag{
						Name:  "validator-id",
						Usage: "The validator id for which the final balance is being notified",
					},
					&cli.Uint64Flag{
						Name:  "slot",
						Usage: "The withdrawal slot",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					validatorId, validatorIndex, found, err := getNotifiableValidator()
					if err != nil {
						return err
					}
					if !found {
						return nil
					}

					// Run
					return notifyFinalBalance(validatorId, validatorIndex, c.Uint64("slot"), c.Bool("yes"))
				},
			},
			{
				Name:      "distribute",
				Aliases:   []string{"b"},
				Usage:     "Distribute any accrued execution layer rewards sent to this megapool",
				UsageText: "rocketpool megapool distribute",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return distribute(c.Bool("yes"))
				},
			},
			{
				Name:      "set-use-latest-delegate",
				Aliases:   []string{"l"},
				Usage:     "Set the megapool to always use the latest delegate",
				UsageText: "rocketpool megapool set-use-latest-delegate",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "use-latest-delegate",
						Usage: "Enable (true) or disable (false) automatic using the latest delegate; omit to be prompted based on the current setting",
					},
					&cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					var useLatest *bool
					if c.IsSet("use-latest-delegate") {
						val := c.Bool("use-latest-delegate")
						useLatest = &val
					}

					return setUseLatestDelegateMegapool(useLatest, c.Bool("yes"))
				},
			},
		},
	})
}
