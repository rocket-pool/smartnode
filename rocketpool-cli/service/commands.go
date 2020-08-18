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

        },
    })
}

