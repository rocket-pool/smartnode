package run

import (
    "github.com/urfave/cli"
)


// Register config command
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Run Rocket Pool command",
        Action: func(c *cli.Context) error {

            // Run command
            return nil

        },
    })
}

