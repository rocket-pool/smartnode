package queue

import (
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/utils/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
	command.Subcommands = append(command.Subcommands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the Rocket Pool deposit queue",
		Subcommands: []cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get the deposit pool and minipool queue status",
				UsageText: "rocketpool api queue status",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getStatus(c))
					return nil

				},
			},

			{
				Name:      "can-process",
				Usage:     "Check whether the deposit pool can be processed",
				UsageText: "rocketpool api queue can-process max-validators",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					max, err := cliutils.ValidatePositiveUint32("max-validators", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canProcessQueue(c, int64(max)))
					return nil

				},
			},
			{
				Name:      "process",
				Aliases:   []string{"p"},
				Usage:     "Process the deposit pool",
				UsageText: "rocketpool api queue process max-validators",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					max, err := cliutils.ValidatePositiveUint32("max-validators", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(processQueue(c, int64(max)))
					return nil

				},
			},
			{
				Name:      "get-total-length",
				Usage:     "Gets the total queue length. It considers legacy plus megapool (express and standard)",
				UsageText: "rocketpool api queue get-total-length",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getTotalQueueLength(c))
					return nil

				},
			},
			{
				Name:      "get-express-length",
				Usage:     "Gets the express queue length.",
				UsageText: "rocketpool api queue get-express-length",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getExpressQueueLength(c))
					return nil

				},
			},
			{
				Name:      "get-standard-length",
				Usage:     "Gets the standard queue length.",
				UsageText: "rocketpool api queue get-standard-length",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getStandardQueueLength(c))
					return nil

				},
			},
		},
	})
}
