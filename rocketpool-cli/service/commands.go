package service

import (
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage Rocket Pool service",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "install",
                Aliases:   []string{"i"},
                Usage:     "Install the Rocket Pool service",
                UsageText: "rocketpool service install [options]",
                Flags: []cli.Flag{
                    cli.BoolFlag{
                        Name:  "verbose, r",
                        Usage: "Print installation script command output",
                    },
                    cli.BoolFlag{
                        Name:  "use-wget, w",
                        Usage: "Attempt to download the installation script using wget instead of cURL",
                    },
                    cli.BoolFlag{
                        Name:  "ignore-deps, i",
                        Usage: "Do not install Operating System dependencies",
                    },
                    cli.StringFlag{
                        Name:  "network, n",
                        Usage: "The Eth 2.0 network to run Rocket Pool on",
                        Value: "medalla",
                    },
                    cli.StringFlag{
                        Name:  "version, v",
                        Usage: "The smart node package version to install",
                        Value: "latest",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run command
                    return installService(c)

                },
            },

            cli.Command{
                Name:      "status",
                Aliases:   []string{"u"},
                Usage:     "View the Rocket Pool service status",
                UsageText: "rocketpool service status",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run command
                    return serviceStatus(c)

                },
            },

            cli.Command{
                Name:      "start",
                Aliases:   []string{"s"},
                Usage:     "Start the Rocket Pool service",
                UsageText: "rocketpool service start",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run command
                    return startService(c)

                },
            },

            cli.Command{
                Name:      "pause",
                Aliases:   []string{"p"},
                Usage:     "Pause the Rocket Pool service",
                UsageText: "rocketpool service pause",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run command
                    return pauseService(c)

                },
            },

            cli.Command{
                Name:      "stop",
                Aliases:   []string{"o"},
                Usage:     "Stop the Rocket Pool service",
                UsageText: "rocketpool service stop",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run command
                    return stopService(c)

                },
            },

            cli.Command{
                Name:      "logs",
                Aliases:   []string{"l"},
                Usage:     "View the Rocket Pool service logs",
                UsageText: "rocketpool service logs [services...]",
                Action: func(c *cli.Context) error {

                    // Run command
                    return serviceLogs(c, c.Args()...)

                },
            },

            cli.Command{
                Name:      "stats",
                Aliases:   []string{"a"},
                Usage:     "View the Rocket Pool service stats",
                UsageText: "rocketpool service stats",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run command
                    return serviceStats(c)

                },
            },

        },
    })
}

