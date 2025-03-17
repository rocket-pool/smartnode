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
				Name:      "can-distribute-megapool",
				Usage:     "Check if can distribute megapool rewards",
				UsageText: "rocketpool api node can-distribute-megapool",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(canDistributeMegapool(c))
					return nil

				},
			},

			{
				Name:      "distribute-megapool",
				Usage:     "Distribute megapool rewards",
				UsageText: "rocketpool api node distribute-megapool",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(distributeMegapool(c))
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
				Name:      "can-stake",
				Usage:     "Check if we can stake a megapool validator",
				UsageText: "rocketpool api megapool can-stake validator-id",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Get the validatorId
					validatorId, err := cliutils.ValidateUint("validatorId", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canStake(c, validatorId))
					return nil

				},
			},
			{
				Name:      "stake",
				Aliases:   []string{"st"},
				Usage:     "Stake a megapool validator",
				UsageText: "rocketpool api megapool stake validator-id",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Get validatorId
					validatorId, err := cliutils.ValidateUint("validatorId", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(stake(c, validatorId))
					return nil

				},
			},
			{
				Name:      "can-exit-queue",
				Usage:     "Check whether the node can exit the megapool queue",
				UsageText: "rocketpool api megapool can-exit-queue validator-id",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Check the validator-id
					validatorId, err := cliutils.ValidateUint32("validatorId", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canExitQueue(c, validatorId))
					return nil

				},
			},
			{
				Name:      "exit-queue",
				Usage:     "Exit the megapool queue",
				UsageText: "rocketpool api megapool exit-queue validator-id",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Check the validatorId
					validatorId, err := cliutils.ValidateUint32("validatorId", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(exitQueue(c, validatorId))
					return nil

				},
			},
			{
				Name:      "can-dissolve-validator",
				Usage:     "Check if we can dissolve a megapool validator",
				UsageText: "rocketpool api megapool can-dissolve-validator validator-id",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Get the validatorId
					validatorId, err := cliutils.ValidateUint32("validatorId", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canDissolveValidator(c, validatorId))
					return nil

				},
			},
			{
				Name:      "dissolve-validator",
				Aliases:   []string{"dv"},
				Usage:     "Dissolve a megapool validator",
				UsageText: "rocketpool api megapool dissolve-validator validator-id",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Get validatorId
					validatorId, err := cliutils.ValidateUint32("validatorId", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(dissolveValidator(c, validatorId))
					return nil

				},
			},
			{
				Name:      "can-exit-validator",
				Usage:     "Check if we can exit a megapool validator",
				UsageText: "rocketpool api megapool can-exit-validator validator-id",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Get the validatorId
					validatorId, err := cliutils.ValidateUint32("validatorId", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canExitValidator(c, validatorId))
					return nil

				},
			},
			{
				Name:      "exit-validator",
				Aliases:   []string{"ev"},
				Usage:     "Exit a megapool validator",
				UsageText: "rocketpool api megapool exit-validator validator-id",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Get validatorId
					validatorId, err := cliutils.ValidateUint32("validatorId", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(exitValidator(c, validatorId))
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
			{
				Name:      "get-effective-delegate",
				Usage:     "Gets the address of the effective delegate contract used by the megapool, which takes the UseLatestDelegate setting into account",
				UsageText: "rocketpool api megapool get-effective-delegate megapool-address",
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
					api.PrintResponse(getEffectiveDelegate(c, megapoolAddress))
					return nil

				},
			},
			{
				Name:      "can-delegate-upgrade",
				Usage:     "Check whether the megapool delegate can be upgraded",
				UsageText: "rocketpool api megapool can-delegate-upgrade megapool-address",
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
					api.PrintResponse(canDelegateUpgrade(c, megapoolAddress))
					return nil

				},
			},
			{
				Name:      "delegate-upgrade",
				Usage:     "Upgrade this megapool to the latest network delegate contract",
				UsageText: "rocketpool api megapool delegate-upgrade megapool-address",
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
					api.PrintResponse(delegateUpgrade(c, megapoolAddress))
					return nil

				},
			},
			{
				Name:      "calculate-rewards",
				Usage:     "Calculate the rewards split given an eth amount",
				UsageText: "rocketpool api megapool calculate-rewards amount",
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
					api.PrintResponse(calculateRewards(c, amount))
					return nil

				},
			},
			{
				Name:      "pending-rewards",
				Usage:     "Calculate the pending rewards split",
				UsageText: "rocketpool api megapool pending-rewards",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}
					// Run
					api.PrintResponse(calculatePendingRewards(c))
					return nil

				},
			},
		},
	})
}
