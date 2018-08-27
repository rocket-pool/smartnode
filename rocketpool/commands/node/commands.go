package node

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/commands"
)


// Register node commands
func RegisterCommands(app *cli.App, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      "node",
        Aliases:   aliases,
        Usage:     "Manage node state",
        Subcommands: []cli.Command{

            // Pause node
            cli.Command{
                Name:      "pause",
                Aliases:   []string{"p"},
                Usage:     "'Pause' the node; stop receiving new minipools",
                UsageText: "rocketpool node pause",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := commands.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err;
                    }

                    // Run command
                    fmt.Println("Pausing...")
                    return nil

                },
            },

            // Resume node
            cli.Command{
                Name:      "resume",
                Aliases:   []string{"r"},
                Usage:     "'Resume' the node; start receiving new minipools again",
                UsageText: "rocketpool node resume",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := commands.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err;
                    }

                    // Run command
                    fmt.Println("Resuming...")
                    return nil

                },
            },

            // Exit node
            cli.Command{
                Name:      "exit",
                Aliases:   []string{"e"},
                Usage:     "'Exit' the node from Rocket Pool permanently; node is paused instead if it has assigned minipools",
                UsageText: "rocketpool node exit",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := commands.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err;
                    }

                    // Run command
                    fmt.Println("Exiting...")
                    return nil

                },
            },

        },
    })
}

