package service

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/commands"
    "github.com/rocket-pool/smartnode-cli/rocketpool/daemon"
)


// Register user commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage smartnode daemon service",
        Subcommands: []cli.Command{

            // Run daemon
            cli.Command{
                Name:      "run",
                Aliases:   []string{"r"},
                Usage:     "Run smartnode daemon service; for manual / advanced use only",
                UsageText: "rocketpool service run",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := commands.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err
                    }

                    // Run command
                    daemon.Run()
                    
                    // Return
                    return nil

                },
            },

        },
    })
}

