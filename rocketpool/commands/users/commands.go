package users

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/commands"
)


// Register user commands
func RegisterCommands(app *cli.App, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      "user",
        Aliases:   aliases,
        Usage:     "Manage users",
        Subcommands: []cli.Command{

            // List users
            cli.Command{
                Name:      "list",
                Aliases:   []string{"l"},
                Usage:     "List minipools and users assigned to the node",
                UsageText: "rocketpool user list",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := commands.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err;
                    }

                    // Run command
                    fmt.Println("Users:")
                    return nil

                },
            },

        },
    })
}

