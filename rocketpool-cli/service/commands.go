package service

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Creates CLI argument flags from the parameters of the configuration struct
func createFlagsFromConfigParams(sectionName string, params []*cfgtypes.Parameter, configFlags []cli.Flag, network cfgtypes.Network) []cli.Flag {
	for _, param := range params {
		var paramName string
		if sectionName == "" {
			paramName = param.ID
		} else {
			paramName = fmt.Sprintf("%s-%s", sectionName, param.ID)
		}

		defaultVal, err := param.GetDefault(network)
		if err != nil {
			panic(fmt.Sprintf("Error getting default value for [%s]: %s\n", paramName, err.Error()))
		}

		switch param.Type {
		case cfgtypes.ParameterType_Bool:
			configFlags = append(configFlags, cli.BoolFlag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: bool\n", param.Description),
			})
		case cfgtypes.ParameterType_Int:
			configFlags = append(configFlags, cli.IntFlag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: int\n", param.Description),
				Value: int(defaultVal.(int64)),
			})
		case cfgtypes.ParameterType_Float:
			configFlags = append(configFlags, cli.Float64Flag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: float\n", param.Description),
				Value: defaultVal.(float64),
			})
		case cfgtypes.ParameterType_String:
			configFlags = append(configFlags, cli.StringFlag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: string\n", param.Description),
				Value: defaultVal.(string),
			})
		case cfgtypes.ParameterType_Uint:
			configFlags = append(configFlags, cli.UintFlag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: uint\n", param.Description),
				Value: uint(defaultVal.(uint64)),
			})
		case cfgtypes.ParameterType_Uint16:
			configFlags = append(configFlags, cli.UintFlag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: uint16\n", param.Description),
				Value: uint(defaultVal.(uint16)),
			})
		case cfgtypes.ParameterType_Choice:
			optionStrings := []string{}
			for _, option := range param.Options {
				optionStrings = append(optionStrings, fmt.Sprint(option.Value))
			}
			configFlags = append(configFlags, cli.StringFlag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: choice\n\tOptions: %s\n", param.Description, strings.Join(optionStrings, ", ")),
				Value: fmt.Sprint(defaultVal),
			})
		}
	}

	return configFlags
}

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {

	configFlags := []cli.Flag{}
	cfgTemplate := config.NewRocketPoolConfig("", false)
	network := cfgTemplate.Smartnode.Network.Value.(cfgtypes.Network)

	// Root params
	configFlags = createFlagsFromConfigParams("", cfgTemplate.GetParameters(), configFlags, network)

	// Subconfigs
	for sectionName, subconfig := range cfgTemplate.GetSubconfigs() {
		configFlags = createFlagsFromConfigParams(sectionName, subconfig.GetParameters(), configFlags, network)
	}

	app.Commands = append(app.Commands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage Rocket Pool service",
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "compose-file, f",
				Usage: "Optional compose files to override the standard Rocket Pool docker compose YAML files; this flag may be defined multiple times",
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
				Flags:     configFlags,
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
					return startService(c, false)

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
				Name:      "compose",
				Usage:     "View the Rocket Pool service docker compose config",
				UsageText: "rocketpool service compose",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return serviceCompose(c)

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
				Name:      "check-cpu-features",
				Aliases:   []string{"ccf"},
				Usage:     "Checks if your CPU supports all of the features required by the \"modern\" version of certain client images. If not, it prints what features are missing.",
				UsageText: "rocketpool service check-cpu-features",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return checkCpuFeatures()

				},
			},

			{
				Name:      "get-config-yaml",
				Usage:     "Generate YAML that shows the current configuration schema, including all of the parameters and their descriptions",
				UsageText: "rocketpool service get-config-yaml",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return getConfigYaml(c)

				},
			},

			{
				Name:      "export-eth1-data",
				Usage:     "Exports the execution client (eth1) chain data to an external folder. Use this if you want to back up your chain data before switching execution clients.",
				UsageText: "rocketpool service export-eth1-data target-folder",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "force",
						Usage: "Bypass the free space check on the target folder",
					},
					cli.BoolFlag{
						Name:  "dirty",
						Usage: "Exports the execution (eth1) chain data without stopping the client. Requires a second pass (much faster) to sync the remaining files without the client running.",
					},
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					targetDir := c.Args().Get(0)

					// Run command
					return exportEcData(c, targetDir)

				},
			},

			{
				Name:      "import-eth1-data",
				Usage:     "Imports execution client (eth1) chain data from an external folder. Use this if you want to restore the data from an execution client that you previously backed up.",
				UsageText: "rocketpool service import-eth1-data source-folder",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					sourceDir := c.Args().Get(0)

					// Run command
					return importEcData(c, sourceDir)

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
				Name:      "terminate",
				Aliases:   []string{"t"},
				Usage:     fmt.Sprintf("%sDeletes all of the Rocket Pool Docker containers and volumes, including your ETH1 and ETH2 chain data and your Prometheus database (if metrics are enabled). Also removes your entire `.rocketpool` configuration folder, including your wallet, password, and validator keys. Only use this if you are cleaning up the Smartnode and want to start over!%s", colorRed, colorReset),
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
					return terminateService(c)

				},
			},
		},
	})
}
