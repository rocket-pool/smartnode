package node

import (
    "regexp"

    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register node commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage node registration & state",
        Subcommands: []cli.Command{

            // Get the node's status
            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the node's status information",
                UsageText: "rocketpool node status",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return getNodeStatus(c)

                },
            },

            // Initialise the node with an account
            cli.Command{
                Name:      "init",
                Aliases:   []string{"i"},
                Usage:     "Initialize the node with an account",
                UsageText: "rocketpool node init password",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var password string

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 1, func(messages *[]string) {

                        // Check password
                        if password = c.Args().Get(0); len(password) < 8 {
                            *messages = append(*messages, "Password must be at least 8 characters long")
                        }

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return initNode(c, password)

                },
            },

            // Register the node with Rocket Pool
            cli.Command{
                Name:      "register",
                Aliases:   []string{"r"},
                Usage:     "Register the node on the Rocket Pool network",
                UsageText: "rocketpool node register timezone",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var timezone string

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 1, func(messages *[]string) {

                        // Check timezone
                        if timezone = c.Args().Get(0); !regexp.MustCompile("^\\w{2,}\\/\\w{2,}$").MatchString(timezone) {
                            *messages = append(*messages, "Invalid timezone - must be in the format 'Country/City'")
                        }

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return registerNode(c, timezone)

                },
            },

            // Set the node's timezone
            cli.Command{
                Name:      "timezone",
                Aliases:   []string{"t"},
                Usage:     "Set the node's timezone on the Rocket Pool network",
                UsageText: "rocketpool node timezone tz",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var timezone string

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 1, func(messages *[]string) {

                        // Check timezone
                        if timezone = c.Args().Get(0); !regexp.MustCompile("^\\w{2,}\\/\\w{2,}$").MatchString(timezone) {
                            *messages = append(*messages, "Invalid timezone - must be in the format 'Country/City'")
                        }

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return setNodeTimezone(c, timezone)

                },
            },

        },
    })
}

