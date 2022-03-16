package service

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage Rocket Pool service",
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "compose-file, f",
				Usage: "Optional compose files to override the standard Rocket Pool docker-compose.yml; this flag may be defined multiple times",
			},
		},
		Subcommands: []cli.Command{

			{
				Name:      "install",
				Aliases:   []string{"i"},
				Usage:     "Install the Rocket Pool service",
				UsageText: "rocketpool service install [options]",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm service installation",
					},
					cli.BoolFlag{
						Name:  "verbose, r",
						Usage: "Print installation script command output",
					},
					cli.BoolFlag{
						Name:  "no-deps, d",
						Usage: "Do not install Operating System dependencies",
					},
					cli.StringFlag{
						Name:  "network, n",
						Usage: "[DEPRECATED] The Eth 2.0 network to run Rocket Pool on - use 'prater' for Rocket Pool's test network",
					},
					cli.StringFlag{
						Name:  "path, p",
						Usage: "A custom path to install Rocket Pool to",
					},
					cli.StringFlag{
						Name:  "version, v",
						Usage: "The smart node package version to install",
						Value: fmt.Sprintf("v%s", shared.RocketPoolVersion),
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return installService(c)

				},
			},

			{
				Name:      "config",
				Aliases:   []string{"c"},
				Usage:     "Configure the Rocket Pool service",
				UsageText: "rocketpool service config",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "headless",
						Usage: "Create a config file without going through the TUI for 3rd-party post-processing",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return configureService(c)

				},
			},

			{
				Name:      "status",
				Aliases:   []string{"u"},
				Usage:     "View the Rocket Pool service status",
				UsageText: "rocketpool service status",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return serviceStatus(c)

				},
			},

			{
				Name:      "start",
				Aliases:   []string{"s"},
				Usage:     "Start the Rocket Pool service",
				UsageText: "rocketpool service start",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "ignore-slash-timer",
						Usage: "Bypass the safety timer that forces a delay when switching to a new ETH2 client",
					},
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Ignore service config prompt after upgrading",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return startService(c)

				},
			},

			{
				Name:      "pause",
				Aliases:   []string{"p"},
				Usage:     "Pause the Rocket Pool service",
				UsageText: "rocketpool service pause [options]",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm service suspension",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return pauseService(c)

				},
			},
			{
				Name:      "stop",
				Aliases:   []string{"o"},
				Usage:     "Pause the Rocket Pool service (alias of 'rocketpool service pause')",
				UsageText: "rocketpool service stop [options]",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm service suspension",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return pauseService(c)

				},
			},

			{
				Name:      "terminate",
				Aliases:   []string{"t"},
				Usage:     "Stop the Rocket Pool service and tear down the service stack",
				UsageText: "rocketpool service terminate [options]",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm service termination",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return stopService(c)

				},
			},

			{
				Name:      "logs",
				Aliases:   []string{"l"},
				Usage:     "View the Rocket Pool service logs",
				UsageText: "rocketpool service logs [options] [services...]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "tail, t",
						Usage: "The number of lines to show from the end of the logs (number or \"all\")",
						Value: "100",
					},
				},
				Action: func(c *cli.Context) error {

					// Run command
					return serviceLogs(c, c.Args()...)

				},
			},

			{
				Name:      "stats",
				Aliases:   []string{"a"},
				Usage:     "View the Rocket Pool service stats",
				UsageText: "rocketpool service stats",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return serviceStats(c)

				},
			},

			{
				Name:      "version",
				Aliases:   []string{"v"},
				Usage:     "View the Rocket Pool service version information",
				UsageText: "rocketpool service version",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return serviceVersion(c)

				},
			},

			{
				Name:      "prune-eth1",
				Aliases:   []string{"n"},
				Usage:     "Shuts down the main ETH1 client and prunes its database, freeing up disk space, then restarts it when it's done.",
				UsageText: "rocketpool service prune-eth1",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return pruneExecutionClient(c)

				},
			},

			{
				Name:      "install-update-tracker",
				Aliases:   []string{"d"},
				Usage:     "Install the update tracker that provides the available system update count to the metrics dashboard",
				UsageText: "rocketpool service install-update-tracker [options]",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm service installation",
					},
					cli.BoolFlag{
						Name:  "verbose, r",
						Usage: "Print installation script command output",
					},
					cli.StringFlag{
						Name:  "version, v",
						Usage: "The update tracker package version to install",
						Value: fmt.Sprintf("v%s", shared.RocketPoolVersion),
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return installUpdateTracker(c)

				},
			},

			{
				Name:      "resync-eth1",
				Usage:     fmt.Sprintf("%sDeletes the main ETH1 client's chain data and resyncs it from scratch. Only use this as a last resort!%s", colorRed, colorReset),
				UsageText: "rocketpool service resync-eth1",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return resyncEth1(c)

				},
			},

			{
				Name:      "resync-eth2",
				Usage:     fmt.Sprintf("%sDeletes the ETH2 client's chain data and resyncs it from scratch. Only use this as a last resort!%s", colorRed, colorReset),
				UsageText: "rocketpool service resync-eth2",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return resyncEth2(c)

				},
			},

			{
				Name:      "migrate-config",
				Usage:     "<DEBUG FUNCTION> Migrate a legacy RP config to a new config.",
				UsageText: "rocketpool service migrate-config",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 3); err != nil {
						return err
					}
					oldConfig := c.Args().Get(0)
					oldSettings := c.Args().Get(1)
					newConfig := c.Args().Get(2)

					// Run command
					return migrateConfig(c, oldConfig, oldSettings, newConfig)

				},
			},
		},
	})
}
