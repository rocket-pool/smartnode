package megapool

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
		Usage:   "Manage the node's megapool",
		Subcommands: []cli.Command{
			{
				Name:      "can-deploy-megapool",
				Usage:     "Check if the node can deploy a megapool",
				UsageText: "rocketpool api node can-deploy-megapool",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(canDeployMegapool(c))
					return nil

				},
			},

			{
				Name:      "deploy-megapool",
				Usage:     "Deploy the node's megapool",
				UsageText: "rocketpool api node deploy-megapool",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(deployMegapool(c))
					return nil

				},
			},
			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get the node's megapool status",
				UsageText: "rocketpool api megapool status",
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
				Name:      "can-repay-debt",
				Usage:     "Check if we can repay the megapool debt",
				UsageText: "rocketpool api megapool can-repay-debt amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Get amount
					amount, err := cliutils.ValidatePositiveWeiAmount("amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canRepayDebt(c, amount))
					return nil

				},
			},
			{
				Name:      "repay-debt",
				Aliases:   []string{"rd"},
				Usage:     "Repay the megapool debt",
				UsageText: "rocketpool api megapool repay-debt amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Get amount
					amount, err := cliutils.ValidatePositiveWeiAmount("amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(repayDebt(c, amount))
					return nil

				},
			},
			{
				Name:      "can-exit-queue",
				Usage:     "Check whether the node can exit the megapool queue",
				UsageText: "rocketpool api megapool can-exit-queue validator-index express-queue",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Check the validator-index
					validatorIndex, err := cliutils.ValidatePositiveUint(c.Args().Get(0), "validator-index")
					if err != nil {
						return err
					}

					// Check the express-queue value
					expressQueue, err := cliutils.ValidateBool(c.Args().Get(1), "express-queue")
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canExitQueue(c, validatorIndex, expressQueue))
					return nil

				},
			},
			{
				Name:      "exit-queue",
				Usage:     "Exit the megapool queue",
				UsageText: "rocketpool api megapool exit-queue validator-index express-queue",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}

					// Check the validator-index
					validatorIndex, err := cliutils.ValidatePositiveUint(c.Args().Get(0), "validator-index")
					if err != nil {
						return err
					}

					// Check the express-queue value
					expressQueue, err := cliutils.ValidateBool(c.Args().Get(1), "express-queue")
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(exitQueue(c, validatorIndex, expressQueue))
					return nil

				},
			},
			{
				Name:      "get-use-latest-delegate",
				Usage:     "Gets the current setting of the 'always use latest delegate' toggle",
				UsageText: "rocketpool api megapool get-use-latest-delegate megapool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					megapoolAddress, err := cliutils.ValidateAddress("megapool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(getUseLatestDelegate(c, megapoolAddress))
					return nil

				},
			},
			{
				Name:      "can-set-use-latest-delegate",
				Usage:     "Check whether the 'always use latest delegate' toggle can be set",
				UsageText: "rocketpool api megapool can-set-use-latest-delegate megapool-address setting",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					megapoolAddress, err := cliutils.ValidateAddress("megapool address", c.Args().Get(0))
					if err != nil {
						return err
					}
					setting, err := cliutils.ValidateBool("setting", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canSetUseLatestDelegate(c, megapoolAddress, setting))
					return nil

				},
			},
			{
				Name:      "set-use-latest-delegate",
				Usage:     "Set whether or not to ignore the megapool's current delegate, and always use the latest delegate instead",
				UsageText: "rocketpool api megapool set-use-latest-delegate setting",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					megapoolAddress, err := cliutils.ValidateAddress("megapool address", c.Args().Get(0))
					if err != nil {
						return err
					}
					setting, err := cliutils.ValidateBool("setting", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(setUseLatestDelegate(c, megapoolAddress, setting))
					return nil

				},
			},
			{
				Name:      "get-delegate",
				Usage:     "Gets the address of the current delegate contract used by the megapool",
				UsageText: "rocketpool api megapool get-delegate megapool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					megapoolAddress, err := cliutils.ValidateAddress("megapool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(getDelegate(c, megapoolAddress))
					return nil

				},
			},
		},
	})
}
