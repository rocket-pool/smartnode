package node

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/commands"
)


// Register node commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
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
                        return err
                    }

                    // Run command
                    fmt.Println("Node pausing...")
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
                        return err
                    }

                    // Run command
                    fmt.Println("Node resuming...")
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
                        return err
                    }

                    // Run command
                    exited, err := exitNode()
                    if err != nil {
                        return cli.NewExitError("The node could not be exited from the network", 1)
                    }

                    // Return
                    if exited {
                        fmt.Println("The node successfully exited from the network")
                    } else {
                        fmt.Println("The node paused receiving new minipools and will exit when able")
                    }
                    return nil

                },
            },

        },
    })
}

