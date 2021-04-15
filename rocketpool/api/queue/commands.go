package queue

import (
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/utils/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
    command.Subcommands = append(command.Subcommands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage the Rocket Pool deposit queue",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the deposit pool and minipool queue status",
                UsageText: "rocketpool api queue status",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(getStatus(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-process",
                Usage:     "Check whether the deposit pool can be processed",
                UsageText: "rocketpool api queue can-process",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(canProcessQueue(c))
                    return nil

                },
            },
            cli.Command{
                Name:      "process",
                Aliases:   []string{"p"},
                Usage:     "Process the deposit pool",
                UsageText: "rocketpool api queue process",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(processQueue(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "wait",
                Usage:     "Wait for a queue transaction to be mined",
                UsageText: "rocketpool api queue wait tx-hash",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    hash, err := cliutils.ValidateTxHash("tx-hash", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(waitForTransaction(c, hash))
                    return nil

                },
            },

        },
    })
}

