package megapool

import (
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/shared/utils"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, &cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the node's minipools",
		Subcommands: []*cli.Command{
			{
				Name:    "status",
				Aliases: []string{"s"},
				Usage:   "Get a list of the node's minipools",
				Action: func(c *cli.Context) error {
					// Validate args
					utils.ValidateArgCount(c, 0)

					// Run
					return getStatus(c)
				},
			},

			// {
			// 	Name:    "stake",
			// 	Aliases: []string{"t"},
			// 	Usage:   "Stake a minipool after the scrub check, moving it from prelaunch to staking.",
			// 	Flags: []cli.Flag{
			// 		&cli.StringFlag{
			// 			Name:    minipoolsFlag,
			// 			Aliases: []string{"m"},
			// 			Usage:   "A comma-separated list of addresses for minipools to stake (or 'all' to stake all available minipools)",
			// 		},
			// 	},
			// 	Action: func(c *cli.Context) error {
			// 		// Validate args
			// 		utils.ValidateArgCount(c, 0)

			// 		// Run
			// 		return stakeMinipools(c)
			// 	},
			// },

			// {
			// 	Name:    "delegate-upgrade",
			// 	Aliases: []string{"u"},
			// 	Usage:   "Upgrade a minipool's delegate contract to the latest version",
			// 	Flags: []cli.Flag{
			// 		&cli.StringFlag{
			// 			Name:    minipoolsFlag,
			// 			Aliases: []string{"m"},
			// 			Usage:   "The comma-separated addresses of the minipools to upgrade (or 'all' to upgrade every available minipool)",
			// 		},
			// 	},
			// 	Action: func(c *cli.Context) error {
			// 		// Validate args
			// 		utils.ValidateArgCount(c, 0)

			// 		// Run
			// 		return upgradeDelegates(c)
			// 	},
			// },

			// {
			// 	Name:      "set-use-latest-delegate",
			// 	Aliases:   []string{"l"},
			// 	Usage:     "Use this to enable or disable the \"use-latest-delegate\" flag on one or more minipools. If enabled, the minipool will ignore its current delegate contract and always use whatever the latest delegate is.",
			// 	ArgsUsage: "true/false",
			// 	Flags: []cli.Flag{
			// 		&cli.StringFlag{
			// 			Name:    minipoolsFlag,
			// 			Aliases: []string{"m"},
			// 			Usage:   "The comma-separated addresses of the minipools to set the use-latest setting for (or 'all' to set it on every available minipool)",
			// 		},
			// 	},
			// 	Action: func(c *cli.Context) error {
			// 		// Validate args
			// 		utils.ValidateArgCount(c, 1)
			// 		setting, err := input.ValidateBool("setting", c.Args().Get(0))
			// 		if err != nil {
			// 			return err
			// 		}

			// 		// Run
			// 		return setUseLatestDelegates(c, setting)
			// 	},
			// },
		},
	})
}
