package service

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
)

// Get the compose file paths for a CLI context
func getComposeFiles(c *cli.Command) []string {
	return cliutils.Parent(c).StringSlice("compose-file")
}

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
			configFlags = append(configFlags, &cli.BoolFlag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: bool\n", param.Description),
			})
		case cfgtypes.ParameterType_Int:
			configFlags = append(configFlags, &cli.IntFlag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: int\n", param.Description),
				Value: int(defaultVal.(int64)),
			})
		case cfgtypes.ParameterType_Float:
			configFlags = append(configFlags, &cli.Float64Flag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: float\n", param.Description),
				Value: defaultVal.(float64),
			})
		case cfgtypes.ParameterType_String:
			configFlags = append(configFlags, &cli.StringFlag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: string\n", param.Description),
				Value: defaultVal.(string),
			})
		case cfgtypes.ParameterType_Uint:
			configFlags = append(configFlags, &cli.Uint64Flag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: uint\n", param.Description),
				Value: defaultVal.(uint64),
			})
		case cfgtypes.ParameterType_Uint16:
			configFlags = append(configFlags, &cli.UintFlag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: uint16\n", param.Description),
				Value: uint(defaultVal.(uint16)),
			})
		case cfgtypes.ParameterType_Choice:
			optionStrings := []string{}
			for _, option := range param.Options {
				optionStrings = append(optionStrings, fmt.Sprint(option.Value))
			}
			configFlags = append(configFlags, &cli.StringFlag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: choice\n\tOptions: %s\n", param.Description, strings.Join(optionStrings, ", ")),
				Value: fmt.Sprint(defaultVal),
			})
		}
	}

	return configFlags
}

// Register commands
func RegisterCommands(app *cli.Command, name string, aliases []string) {

	configFlags := []cli.Flag{}
	cfgTemplate := config.NewRocketPoolConfig("", false)
	network := cfgTemplate.Smartnode.Network.Value.(cfgtypes.Network)

	// Root params
	configFlags = createFlagsFromConfigParams("", cfgTemplate.GetParameters(), configFlags, network)

	// Subconfigs
	for sectionName, subconfig := range cfgTemplate.GetSubconfigs() {
		configFlags = createFlagsFromConfigParams(sectionName, subconfig.GetParameters(), configFlags, network)
	}

	app.Commands = append(app.Commands, &cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage Rocket Pool service",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "compose-file",
				Aliases: []string{"f"},
				Usage:   "Optional compose files to override the standard Rocket Pool docker compose YAML files; this flag may be defined multiple times",
			},
		},
		Commands: []*cli.Command{

			{
				Name:      "install",
				Aliases:   []string{"i"},
				Usage:     "Install the Rocket Pool service",
				UsageText: "rocketpool service install [options]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm service installation",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"r"},
						Usage:   "Print installation script command output",
					},
					&cli.BoolFlag{
						Name:    "no-deps",
						Aliases: []string{"d"},
						Usage:   "Do not install Operating System dependencies",
					},
					&cli.StringFlag{
						Name:    "path",
						Aliases: []string{"p"},
						Usage:   "A custom path to install Rocket Pool to",
					},
					&cli.StringFlag{
						Name:    "version",
						Aliases: []string{"v"},
						Usage:   "The smart node package version to install",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					if c.String("version") != "" {
						fmt.Fprintf(os.Stderr, "--version/-v is no longer supported. Instead, download the correct version of the `rocketpool` binary and install that. Current version: %s\n", shared.RocketPoolVersion())
						os.Exit(1)
					}

					// Run command
					return installService(c.Bool("yes"), c.Bool("verbose"), c.Bool("no-deps"), c.String("path"))

				},
			},

			{
				Name:      "config",
				Aliases:   []string{"c"},
				Usage:     "Configure the Rocket Pool service",
				UsageText: "rocketpool service config",
				Flags:     configFlags,
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}
					configPath := c.Root().String("config-path")
					path, err := homedir.Expand(configPath)
					if err != nil {
						return fmt.Errorf("error expanding config path [%s]: %w", configPath, err)
					}

					_, err = os.Stat(path)
					if os.IsNotExist(err) {
						color.YellowPrintf("Your configured Rocket Pool directory of [%s] does not exist.\n", path)
						color.YellowPrintln("Please follow the instructions at https://docs.rocketpool.net/node-staking/docker to install the Smart Node.")
					}
					if err != nil {
						return fmt.Errorf("error checking if config path exists: %w", err)
					}

					isHeadless := c.NumFlags() > c.Root().NumFlags()

					if isHeadless {
						return configureServiceHeadless(c)
					}

					// Run command
					return configureService(
						c.Root().String("config-path"),
						/*isNative=*/ c.Root().IsSet("daemon-path"),
						c.Bool("yes"),
						getComposeFiles(c),
					)

				},
			},

			{
				Name:      "status",
				Aliases:   []string{"u"},
				Usage:     "View the Rocket Pool service status",
				UsageText: "rocketpool service status",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return serviceStatus(getComposeFiles(c))

				},
			},

			{
				Name:      "start",
				Aliases:   []string{"s"},
				Usage:     "Start the Rocket Pool service",
				UsageText: "rocketpool service start",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "ignore-slash-timer",
						Usage: "Bypass the safety timer that forces a delay when switching to a new ETH2 client",
					},
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Ignore service config prompt after upgrading",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return startService(startServiceParams{
						yes:                    c.Bool("yes"),
						ignoreSlashTimer:       c.Bool("ignore-slash-timer"),
						ignoreConfigSuggestion: false,
						composeFiles:           getComposeFiles(c),
					})

				},
			},

			{
				Name:      "stop",
				Aliases:   []string{"pause", "p", "o"},
				Usage:     "Pause the Rocket Pool service",
				UsageText: "rocketpool service stop [options]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm service suspension",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					_, err := pauseService(c.Bool("yes"), getComposeFiles(c))
					return err

				},
			},

			{
				Name:      "reset-docker",
				Aliases:   []string{"rd"},
				Usage:     "Cleanup Docker resources, including stopped containers, unused images and networks. Stops and restarts Smart Node.",
				UsageText: "rocketpool service reset [options]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm service suspension",
					},
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a"},
						Usage:   "Removes all Docker images, including those currently used by the Smart Node stack. This will force a full re-download of all images when the Smart Node is restarted.",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return resetDocker(c.Bool("yes"), c.Bool("all"), getComposeFiles(c))
				},
			},

			{
				Name:      "prune-docker",
				Aliases:   []string{"pd"},
				Usage:     "Cleanup unused Docker resources, including stopped containers, unused images, networks and volumes. Does not restart smartnode, so the running containers and the images and networks they reference will not be pruned.",
				UsageText: "rocketpool service prune",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a"},
						Usage:   "Removes all Docker images, including those currently used by the Smart Node stack. This will force a full re-download of all images when the Smart Node is restarted.",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return pruneDocker(c.Bool("all"), getComposeFiles(c))
				},
			},

			{
				Name:      "logs",
				Aliases:   []string{"l"},
				Usage:     "View the Rocket Pool service logs",
				UsageText: "rocketpool service logs [options] [services...]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "tail",
						Aliases: []string{"t"},
						Usage:   "The number of lines to show from the end of the logs (number or \"all\")",
						Value:   "100",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Run command
					return serviceLogs(c.String("tail"), getComposeFiles(c), c.Args().Slice()...)

				},
			},

			{
				Name:      "compose",
				Usage:     "View the Rocket Pool service docker compose config",
				UsageText: "rocketpool service compose",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return serviceCompose(getComposeFiles(c))

				},
			},

			{
				Name:      "version",
				Aliases:   []string{"v"},
				Usage:     "View the Rocket Pool service version information",
				UsageText: "rocketpool service version",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return serviceVersion()

				},
			},

			{
				Name:      "prune-eth1",
				Aliases:   []string{"n"},
				Usage:     "Shuts down the main ETH1 client and prunes its database, freeing up disk space, then restarts it when it's done.",
				UsageText: "rocketpool service prune-eth1",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return pruneExecutionClient(c.Bool("yes"))

				},
			},

			{
				Name:      "install-update-tracker",
				Aliases:   []string{"d"},
				Usage:     "Install the update tracker that provides the available system update count to the metrics dashboard",
				UsageText: "rocketpool service install-update-tracker [options]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm service installation",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"r"},
						Usage:   "Print installation script command output",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					if c.String("version") != "" {
						fmt.Fprintf(os.Stderr, "--version/-v is no longer supported. Instead, download the correct version of the `rocketpool` binary and install the update tracker from there. Current version: %s\n", shared.RocketPoolVersion())
						os.Exit(1)
					}

					// Run command
					return installUpdateTracker(c.Bool("yes"), c.Bool("verbose"))

				},
			},

			{
				Name:      "get-config-yaml",
				Usage:     "Generate YAML that shows the current configuration schema, including all of the parameters and their descriptions",
				UsageText: "rocketpool service get-config-yaml",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return getConfigYaml()

				},
			},

			{
				Name:      "resync-eth1",
				Usage:     color.Red("Deletes the main ETH1 client's chain data and resyncs it from scratch. Only use this as a last resort!"),
				UsageText: "rocketpool service resync-eth1",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return resyncEth1(c.Bool("yes"), getComposeFiles(c))

				},
			},

			{
				Name:      "resync-eth2",
				Usage:     color.Red("Deletes the ETH2 client's chain data and resyncs it from scratch. Only use this as a last resort!"),
				UsageText: "rocketpool service resync-eth2",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return resyncEth2(c.Bool("yes"), getComposeFiles(c))

				},
			},

			{
				Name:      "terminate",
				Aliases:   []string{"t"},
				Usage:     color.Red("Deletes all of the Rocket Pool Docker containers and volumes, including your ETH1 and ETH2 chain data and your Prometheus database (if metrics are enabled). Also removes your entire `.rocketpool` configuration folder, including your wallet, password, and validator keys. Only use this if you are cleaning up the Smart Node and want to start over!"),
				UsageText: "rocketpool service terminate [options]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm service termination",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return terminateService(c.Bool("yes"), getComposeFiles(c), c.Root().String("config-path"))

				},
			},
		},
	})
}
