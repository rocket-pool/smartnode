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
				UsageText: "rocketpool api megapool status finalized-state",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Get finalized state
					finalizedState, err := cliutils.ValidateBool("finalized-state", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(getStatus(c, finalizedState))
					return nil

				},
			},
			{
				Name:      "validator-map-and-balances",
				Aliases:   []string{"gvm"},
				Usage:     "Get a map of the node's validators and beacon balances",
				UsageText: "rocketpool api megapool validator-map-and-balances",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getValidatorMapAndBalances(c))
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
				Name:      "can-reduce-bond",
				Usage:     "Check if we can reduce the megapool bond",
				UsageText: "rocketpool api megapool can-reduce-bond amount",
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
					api.PrintResponse(canReduceBond(c, amount))
					return nil

				},
			},
			{
				Name:      "reduce-bond",
				Aliases:   []string{"rb"},
				Usage:     "Reduce the megapool bond",
				UsageText: "rocketpool api megapool reduce-bond amount",
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
					api.PrintResponse(reduceBond(c, amount))
					return nil

				},
			},
			{
				Name:      "can-claim-refund",
				Usage:     "Check if we can claim a megapool refund",
				UsageText: "rocketpool api megapool can-claim-refund",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(canClaimRefund(c))
					return nil

				},
			},
			{
				Name:      "claim-refund",
				Aliases:   []string{"cr"},
				Usage:     "Claim a megapool refund",
				UsageText: "rocketpool api megapool claim-refund",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(claimRefund(c))
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
				Name:      "can-notify-validator-exit",
				Usage:     "Check if we can notify the exit of a megapool validator",
				UsageText: "rocketpool api megapool can-notify-validator-exit validator-id",
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
					api.PrintResponse(canNotifyValidatorExit(c, validatorId))
					return nil

				},
			},
			{
				Name:      "notify-validator-exit",
				Aliases:   []string{"ev"},
				Usage:     "Notify a megapool validator exit",
				UsageText: "rocketpool api megapool notify-validator-exit validator-id",
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
					api.PrintResponse(notifyValidatorExit(c, validatorId))
					return nil

				},
			},
			{
				Name:      "can-notify-final-balance",
				Usage:     "Check if we can notify the final balance of a megapool validator",
				UsageText: "rocketpool api megapool can-notify-final-balance validator-id slot",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}

					// Get the validatorId
					validatorId, err := cliutils.ValidateUint32("validatorId", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Get slot
					slot, err := cliutils.ValidateUint("slot", c.Args().Get(1))
					if err != nil {
						return err
					}
					// Run
					api.PrintResponse(canNotifyFinalBalance(c, validatorId, slot))
					return nil

				},
			},
			{
				Name:      "notify-final-balance",
				Aliases:   []string{"ev"},
				Usage:     "Notify a megapool validator final balance",
				UsageText: "rocketpool api megapool notify-final-balance validator-id slot",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}

					// Get validatorId
					validatorId, err := cliutils.ValidateUint32("validatorId", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Get slot
					slot, err := cliutils.ValidateUint("slot", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(notifyFinalBalance(c, validatorId, slot))
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
				UsageText: "rocketpool api megapool can-set-use-latest-delegate megapool-address",
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
					api.PrintResponse(canSetUseLatestDelegate(c, megapoolAddress))
					return nil

				},
			},
			{
				Name:      "set-use-latest-delegate",
				Usage:     "Set to ignore the megapool's current delegate, and always use the latest delegate instead",
				UsageText: "rocketpool api megapool set-use-latest-delegate",
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
					api.PrintResponse(setUseLatestDelegate(c, megapoolAddress))
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
			{
				Name:      "get-new-validator-bond-requirement",
				Usage:     "Get the bond amount required for the megapool's next validator",
				UsageText: "rocketpool api megapool get-new-validator-bond-requirement",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}
					// Run
					api.PrintResponse(getNewValidatorBondRequirement(c))
					return nil

				},
			},
		},
	})
}
