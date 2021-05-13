package metrics

import (
    "github.com/urfave/cli"
)


// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Rocket Pool Metrics",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "print",
                Aliases:   []string{"p"},
                Usage:     "Print the ouptput of metrics",
                UsageText: "rocketpool metrics print",
                Action: func(c *cli.Context) error {

                    // Run
                    return print(c)

                },
            },

        },
    })
}

