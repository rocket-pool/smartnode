package config

import (
    "github.com/urfave/cli"
)


// Register config command
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Configure Rocket Pool service",
        Action: func(c *cli.Context) error {
            return configureService(c)
        },
    })
}

