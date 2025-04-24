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
				Name:      "deploy",
				Usage:     "Deploy a megapool contract for your node",
				UsageText: "rocketpool node deploy",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm deployment",
					},
				},
				Action: func(c *cli.Context) error {
					return deployMegapool(c)
				},
			},
			{
				Name:      "deposit",
				Aliases:   []string{"d"},
				Usage:     "Make a deposit and create a new validator on the megapool",
				UsageText: "rocketpool node deposit [options]",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm deposit",
					},
					cli.BoolFlag{
						Name:  "use-express-ticket, e",
						Usage: "Use an express ticket to create a new validator",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return nodeMegapoolDeposit(c)

				},
			},
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
				Name:      "validators",
				Aliases:   []string{"v"},
				Usage:     "Get a list of the megapool's validators",
				UsageText: "rocketpool megapool validators",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getValidatorStatus(c)

				},
			},
			{
				Name:      "repay-debt",
				Aliases:   []string{"r"},
				Usage:     "Repay megapool debt",
				UsageText: "rocketpool megapool repay-debt",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return repayDebt(c)
				},
			},
			{
				Name:      "reduce-bond",
				Aliases:   []string{"e"},
				Usage:     "Reduce the megapool bond",
				UsageText: "rocketpool megapool reduce-bond",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return reduceBond(c)
				},
			},
			{
				Name:      "claim",
				Aliases:   []string{"r"},
				Usage:     "Claim any megapool rewards that were distributed but not yet claimed",
				UsageText: "rocketpool megapool claim",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return claim(c)
				},
			},
			{
				Name:      "stake",
				Aliases:   []string{"k"},
				Usage:     "Stake a megapool validator",
				UsageText: "rocketpool megapool stake",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
					cli.Uint64Flag{
						Name:  "validator-id",
						Usage: "The validator id to stake",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return stake(c)
				},
			},
			{
				Name:      "exit-queue",
				Aliases:   []string{"x"},
				Usage:     "Exit the megapool queue",
				UsageText: "rocketpool megapool exit-queue",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
					cli.StringFlag{
						Name:  "validator-id",
						Usage: "The validator id to exit",
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
			{
				Name:      "dissolve-validator",
				Usage:     "Dissolve a megapool validator",
				UsageText: "rocketpool megapool dissolve-validator",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
					cli.StringFlag{
						Name:  "validator-id",
						Usage: "The validator id to exit",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return dissolveValidator(c)
				},
			},
			{
				Name:      "exit-validator",
				Usage:     "Request to exit a megapool validator",
				UsageText: "rocketpool megapool exit-validator",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
					cli.StringFlag{
						Name:  "validator-id",
						Usage: "The validator id to exit",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return exitValidator(c)
				},
			},
			{
				Name:      "notify-validator-exit",
				Aliases:   []string{"n"},
				Usage:     "Notify that a validator exit is in progress",
				UsageText: "rocketpool megapool notify-validator-exit",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
					cli.StringFlag{
						Name:  "validator-id",
						Usage: "The validator id for which the exit is being notified",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return notifyValidatorExit(c)
				},
			},
			{
				Name:      "distribute",
				Aliases:   []string{"b"},
				Usage:     "Distribute any accrued execution layer rewards sent to this megapool",
				UsageText: "rocketpool megapool distribute",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return distribute(c)
				},
			},
			{
				Name:      "set-use-latest-delegate",
				Aliases:   []string{"l"},
				Usage:     "Use this to enable or disable the \"use-latest-delegate\" flag on the node's megapool. If enabled, the megapool will ignore its current delegate contract and always use whatever the latest delegate is.",
				UsageText: "rocketpool megapool set-use-latest-delegate [options] true/false",

				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					setting, err := cliutils.ValidateBool("setting", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					return setUseLatestDelegateMegapool(c, setting)

				},
			},
			{
				Name:      "delegate-upgrade",
				Aliases:   []string{"u"},
				Usage:     "Upgrade a megapool's delegate contract to the latest version",
				UsageText: "rocketpool megapool delegate-upgrade",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}
					// Run
					return delegateUpgradeMegapool(c)

				},
			},
		},
	})
}
