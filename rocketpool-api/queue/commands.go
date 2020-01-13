package queue

import (
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register queue commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage deposit queues",
        Subcommands: []cli.Command{

            // Get the deposit queue status
            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the current deposit queue status",
                UsageText: "rocketpool queue status",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return getQueueStatus(c)

                },
            },

            // Process a deposit queue
            cli.Command{
                Name:      "process",
                Aliases:   []string{"p"},
                Usage:     "Process a deposit queue",
                UsageText: "rocketpool queue process durationID",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var durationId string

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 1, func(messages *[]string) {

                        // Get duration ID
                        durationId = c.Args().Get(0)

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return processQueue(c, durationId)

                },
            },

        },
    })
}

