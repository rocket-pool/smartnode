package minipool

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
		Usage:   "Manage the node's minipools",
		Subcommands: []cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get a list of the node's minipools",
				UsageText: "rocketpool api minipool status",
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
				Name:      "can-stake",
				Usage:     "Check whether the minipool is ready to be staked, moving from prelaunch to staking status",
				UsageText: "rocketpool api minipool can-stake minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canStakeMinipool(c, minipoolAddress))
					return nil

				},
			},
			{
				Name:      "stake",
				Aliases:   []string{"t"},
				Usage:     "Stake the minipool, moving it from prelaunch to staking status",
				UsageText: "rocketpool api minipool stake minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(stakeMinipool(c, minipoolAddress))
					return nil

				},
			},

			{
				Name:      "can-promote",
				Usage:     "Check whether a vacant minipool is ready to be promoted",
				UsageText: "rocketpool api minipool can-promote minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canPromoteMinipool(c, minipoolAddress))
					return nil

				},
			},
			{
				Name:      "promote",
				Usage:     "Promote a vacant minipool",
				UsageText: "rocketpool api minipool promote minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(promoteMinipool(c, minipoolAddress))
					return nil

				},
			},

			{
				Name:      "can-refund",
				Usage:     "Check whether the node can refund ETH from the minipool",
				UsageText: "rocketpool api minipool can-refund minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canRefundMinipool(c, minipoolAddress))
					return nil

				},
			},
			{
				Name:      "refund",
				Aliases:   []string{"r"},
				Usage:     "Refund ETH belonging to the node from a minipool",
				UsageText: "rocketpool api minipool refund minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(refundMinipool(c, minipoolAddress))
					return nil

				},
			},

			{
				Name:      "can-dissolve",
				Usage:     "Check whether the minipool can be dissolved",
				UsageText: "rocketpool api minipool can-dissolve minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canDissolveMinipool(c, minipoolAddress))
					return nil

				},
			},
			{
				Name:      "dissolve",
				Aliases:   []string{"d"},
				Usage:     "Dissolve an initialized or prelaunch minipool",
				UsageText: "rocketpool api minipool dissolve minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(dissolveMinipool(c, minipoolAddress))
					return nil

				},
			},

			{
				Name:      "can-exit",
				Usage:     "Check whether the minipool can be exited from the beacon chain",
				UsageText: "rocketpool api minipool can-exit minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canExitMinipool(c, minipoolAddress))
					return nil

				},
			},
			{
				Name:      "exit",
				Aliases:   []string{"e"},
				Usage:     "Exit a staking minipool from the beacon chain",
				UsageText: "rocketpool api minipool exit minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(exitMinipool(c, minipoolAddress))
					return nil

				},
			},

			{
				Name:      "get-minipool-close-details-for-node",
				Usage:     "Check all of the node's minipools for closure eligibility, and return the details of the closeable ones",
				UsageText: "rocketpool api minipool get-minipool-close-details-for-node",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getMinipoolCloseDetailsForNode(c))
					return nil

				},
			},
			{
				Name:      "close",
				Aliases:   []string{"c"},
				Usage:     "Withdraw balance from a dissolved minipool and close it",
				UsageText: "rocketpool api minipool close minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(closeMinipool(c, minipoolAddress))
					return nil

				},
			},

			{
				Name:      "can-delegate-upgrade",
				Usage:     "Check whether the minipool delegate can be upgraded",
				UsageText: "rocketpool api minipool can-delegate-upgrade minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canDelegateUpgrade(c, minipoolAddress))
					return nil

				},
			},
			{
				Name:      "delegate-upgrade",
				Usage:     "Upgrade this minipool to the latest network delegate contract",
				UsageText: "rocketpool api minipool delegate-upgrade minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(delegateUpgrade(c, minipoolAddress))
					return nil

				},
			},

			{
				Name:      "can-delegate-rollback",
				Usage:     "Check whether the minipool delegate can be rolled back",
				UsageText: "rocketpool api minipool can-delegate-rollback minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canDelegateRollback(c, minipoolAddress))
					return nil

				},
			},
			{
				Name:      "delegate-rollback",
				Usage:     "Rollback the minipool to the previous delegate contract",
				UsageText: "rocketpool api minipool delegate-rollback minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(delegateRollback(c, minipoolAddress))
					return nil

				},
			},

			{
				Name:      "can-set-use-latest-delegate",
				Usage:     "Check whether the 'always use latest delegate' toggle can be set",
				UsageText: "rocketpool api minipool can-set-use-latest-delegate minipool-address setting",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}
					setting, err := cliutils.ValidateBool("setting", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canSetUseLatestDelegate(c, minipoolAddress, setting))
					return nil

				},
			},
			{
				Name:      "set-use-latest-delegate",
				Usage:     "Set whether or not to ignore the minipool's current delegate, and always use the latest delegate instead",
				UsageText: "rocketpool api minipool set-use-latest-delegate minipool-address setting",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}
					setting, err := cliutils.ValidateBool("setting", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(setUseLatestDelegate(c, minipoolAddress, setting))
					return nil

				},
			},

			{
				Name:      "get-use-latest-delegate",
				Usage:     "Gets the current setting of the 'always use latest delegate' toggle",
				UsageText: "rocketpool api minipool get-use-latest-delegate minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(getUseLatestDelegate(c, minipoolAddress))
					return nil

				},
			},

			{
				Name:      "get-delegate",
				Usage:     "Gets the address of the current delegate contract used by the minipool",
				UsageText: "rocketpool api minipool get-delegate minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(getDelegate(c, minipoolAddress))
					return nil

				},
			},

			{
				Name:      "get-previous-delegate",
				Usage:     "Gets the address of the previous delegate contract that the minipool will use during a rollback",
				UsageText: "rocketpool api minipool get-previous-delegate minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(getPreviousDelegate(c, minipoolAddress))
					return nil

				},
			},

			{
				Name:      "get-effective-delegate",
				Usage:     "Gets the address of the effective delegate contract used by the minipool, which takes the UseLatestDelegate setting into account",
				UsageText: "rocketpool api minipool get-effective-delegate minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(getEffectiveDelegate(c, minipoolAddress))
					return nil

				},
			},

			{
				Name:      "get-vanity-artifacts",
				Aliases:   []string{"v"},
				Usage:     "Gets the data necessary to search for vanity minipool addresses",
				UsageText: "rocketpool api minipool get-vanity-artifacts deposit node-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					depositAmount, err := cliutils.ValidatePositiveWeiAmount("deposit amount", c.Args().Get(0))
					if err != nil {
						return err
					}
					nodeAddressStr := c.Args().Get(1)

					// Run
					api.PrintResponse(getVanityArtifacts(c, depositAmount, nodeAddressStr))
					return nil

				},
			},

			{
				Name:      "can-begin-reduce-bond-amount",
				Usage:     "Check whether the minipool can begin the bond reduction process",
				UsageText: "rocketpool api minipool can-begin-reduce-bond-amount minipool-address new-bond-amount-wei",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}
					newBondAmountWei, err := cliutils.ValidateWeiAmount("new bond amount", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canBeginReduceBondAmount(c, minipoolAddress, newBondAmountWei))
					return nil

				},
			},
			{
				Name:      "begin-reduce-bond-amount",
				Usage:     "Begin the bond reduction process for a minipool",
				UsageText: "rocketpool api minipool begin-reduce-bond-amount minipool-address new-bond-amount-wei",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}
					newBondAmountWei, err := cliutils.ValidateWeiAmount("new bond amount", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(beginReduceBondAmount(c, minipoolAddress, newBondAmountWei))
					return nil

				},
			},

			{
				Name:      "can-reduce-bond-amount",
				Usage:     "Check if a minipool's bond can be reduced",
				UsageText: "rocketpool api minipool can-reduce-bond-amount minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canReduceBondAmount(c, minipoolAddress))
					return nil

				},
			},
			{
				Name:      "reduce-bond-amount",
				Usage:     "Reduce a minipool's bond",
				UsageText: "rocketpool api minipool reduce-bond-amount minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(reduceBondAmount(c, minipoolAddress))
					return nil

				},
			},

			{
				Name:      "get-distribute-balance-details",
				Usage:     "Get the balance distribution details for all of the node's minipools",
				UsageText: "rocketpool api minipool get-distribute-balance-details",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getDistributeBalanceDetails(c))
					return nil

				},
			},
			{
				Name:      "distribute-balance",
				Usage:     "Distribute a minipool's ETH balance",
				UsageText: "rocketpool api minipool distribute-balance minipool-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(distributeBalance(c, minipoolAddress))
					return nil

				},
			},

			{
				Name:      "import-key",
				Usage:     "Import a validator private key for a vacant minipool",
				UsageText: "rocketpool api minipool import-key minipool-address mnemonic",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}
					mnemonic, err := cliutils.ValidateWalletMnemonic("mnemonic", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(importKey(c, minipoolAddress, mnemonic))
					return nil

				},
			},

			{
				Name:      "can-change-withdrawal-creds",
				Usage:     "Check whether a solo validator's withdrawal credentials can be changed to a minipool address",
				UsageText: "rocketpool api minipool can-change-withdrawal-creds minipool-address mnemonic",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}
					mnemonic, err := cliutils.ValidateWalletMnemonic("mnemonic", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canChangeWithdrawalCreds(c, minipoolAddress, mnemonic))
					return nil

				},
			},
			{
				Name:      "change-withdrawal-creds",
				Usage:     "Change a solo validator's withdrawal credentials to a minipool address",
				UsageText: "rocketpool api minipool change-withdrawal-creds minipool-address mnemonic",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}
					mnemonic, err := cliutils.ValidateWalletMnemonic("mnemonic", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(changeWithdrawalCreds(c, minipoolAddress, mnemonic))
					return nil

				},
			},

			{
				Name:      "get-rescue-dissolved-details-for-node",
				Usage:     "Check all of the node's minipools for rescue eligibility, and return the details of the rescuable ones",
				UsageText: "rocketpool api minipool get-rescue-dissolved-details-for-node",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getMinipoolRescueDissolvedDetailsForNode(c))
					return nil

				},
			},

			{
				Name:      "rescue-dissolved",
				Usage:     "Rescue a dissolved minipool by depositing ETH for it to the Beacon deposit contract",
				UsageText: "rocketpool api minipool rescue-dissolved minipool-address deposit-amount submit",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 3); err != nil {
						return err
					}
					minipoolAddress, err := cliutils.ValidateAddress("minipool address", c.Args().Get(0))
					if err != nil {
						return err
					}
					depositAmount, err := cliutils.ValidateBigInt("deposit amount", c.Args().Get(1))
					if err != nil {
						return err
					}
					submit, err := cliutils.ValidateBool("submit", c.Args().Get(2))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(rescueDissolvedMinipool(c, minipoolAddress, depositAmount, submit))
					return nil

				},
			},
			{
				Name:      "get-bond-reduction-enabled",
				Usage:     "Check whether bond reduction is enabled",
				UsageText: "rocketpool api minipool get-bond-reduction-enabled",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getBondReductionEnabled(c))
					return nil
				},
			},
		},
	})
}
