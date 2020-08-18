package config

import (
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register command
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Configure Rocket Pool service",
        Action: func(c *cli.Context) error {

            // Validate args
            if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

            // Run command
            return configureService(c)

        },
    })
}

