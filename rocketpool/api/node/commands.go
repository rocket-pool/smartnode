package node

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
		Usage:   "Manage the node",
		Subcommands: []cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get the node's status",
				UsageText: "rocketpool api node status",
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
				Name:      "sync",
				Aliases:   []string{"y"},
				Usage:     "Get the sync progress of the eth1 and eth2 clients",
				UsageText: "rocketpool api node sync",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getSyncProgress(c))
					return nil

				},
			},

			{
				Name:      "can-register",
				Usage:     "Check whether the node can be registered with Rocket Pool",
				UsageText: "rocketpool api node can-register timezone-location",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					timezoneLocation, err := cliutils.ValidateTimezoneLocation("timezone location", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canRegisterNode(c, timezoneLocation))
					return nil

				},
			},
			{
				Name:      "register",
				Aliases:   []string{"r"},
				Usage:     "Register the node with Rocket Pool",
				UsageText: "rocketpool api node register timezone-location",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					timezoneLocation, err := cliutils.ValidateTimezoneLocation("timezone location", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(registerNode(c, timezoneLocation))
					return nil

				},
			},

			{
				Name:      "can-set-primary-withdrawal-address",
				Usage:     "Checks if the node can set its primary withdrawal address",
				UsageText: "rocketpool api node can-set-primary-withdrawal-address address confirm",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					withdrawalAddress, err := cliutils.ValidateAddress("withdrawal address", c.Args().Get(0))
					if err != nil {
						return err
					}

					confirm, err := cliutils.ValidateBool("confirm", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canSetPrimaryWithdrawalAddress(c, withdrawalAddress, confirm))
					return nil

				},
			},
			{
				Name:      "set-primary-withdrawal-address",
				Usage:     "Set the node's primary withdrawal address",
				UsageText: "rocketpool api node set-primary-withdrawal-address address confirm",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					withdrawalAddress, err := cliutils.ValidateAddress("withdrawal address", c.Args().Get(0))
					if err != nil {
						return err
					}

					confirm, err := cliutils.ValidateBool("confirm", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(setPrimaryWithdrawalAddress(c, withdrawalAddress, confirm))
					return nil

				},
			},

			{
				Name:      "can-confirm-primary-withdrawal-address",
				Usage:     "Checks if the node can confirm its primary withdrawal address",
				UsageText: "rocketpool api node can-confirm-primary-withdrawal-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(canConfirmPrimaryWithdrawalAddress(c))
					return nil

				},
			},
			{
				Name:      "confirm-primary-withdrawal-address",
				Usage:     "Confirms the node's primary withdrawal address if it was set back to the node address",
				UsageText: "rocketpool api node confirm-primary-withdrawal-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(confirmPrimaryWithdrawalAddress(c))
					return nil

				},
			},

			{
				Name:      "can-set-rpl-withdrawal-address",
				Usage:     "Checks if the node can set its RPL withdrawal address",
				UsageText: "rocketpool api node can-set-rpl-withdrawal-address address confirm",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					withdrawalAddress, err := cliutils.ValidateAddress("withdrawal address", c.Args().Get(0))
					if err != nil {
						return err
					}

					confirm, err := cliutils.ValidateBool("confirm", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canSetRPLWithdrawalAddress(c, withdrawalAddress, confirm))
					return nil

				},
			},
			{
				Name:      "set-rpl-withdrawal-address",
				Usage:     "Set the node's RPL withdrawal address",
				UsageText: "rocketpool api node set-rpl-withdrawal-address address confirm",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					withdrawalAddress, err := cliutils.ValidateAddress("withdrawal address", c.Args().Get(0))
					if err != nil {
						return err
					}

					confirm, err := cliutils.ValidateBool("confirm", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(setRPLWithdrawalAddress(c, withdrawalAddress, confirm))
					return nil

				},
			},

			{
				Name:      "can-confirm-rpl-withdrawal-address",
				Usage:     "Checks if the node can confirm its RPL withdrawal address",
				UsageText: "rocketpool api node can-confirm-rpl-withdrawal-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(canConfirmRPLWithdrawalAddress(c))
					return nil

				},
			},
			{
				Name:      "confirm-rpl-withdrawal-address",
				Usage:     "Confirms the node's RPL withdrawal address if it was set back to the node address",
				UsageText: "rocketpool api node confirm-rpl-withdrawal-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(confirmRPLWithdrawalAddress(c))
					return nil

				},
			},

			{
				Name:      "can-set-timezone",
				Usage:     "Checks if the node can set its timezone location",
				UsageText: "rocketpool api node can-set-timezone timezone-location",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					timezoneLocation, err := cliutils.ValidateTimezoneLocation("timezone location", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canSetTimezoneLocation(c, timezoneLocation))
					return nil

				},
			},
			{
				Name:      "set-timezone",
				Aliases:   []string{"t"},
				Usage:     "Set the node's timezone location",
				UsageText: "rocketpool api node set-timezone timezone-location",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					timezoneLocation, err := cliutils.ValidateTimezoneLocation("timezone location", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(setTimezoneLocation(c, timezoneLocation))
					return nil

				},
			},

			{
				Name:      "can-swap-rpl",
				Usage:     "Check whether the node can swap old RPL for new RPL",
				UsageText: "rocketpool api node can-swap-rpl amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("swap amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canNodeSwapRpl(c, amountWei))
					return nil

				},
			},
			{
				Name:      "swap-rpl-approve-rpl",
				Aliases:   []string{"p1"},
				Usage:     "Approve fixed-supply RPL for swapping to new RPL",
				UsageText: "rocketpool api node swap-rpl-approve-rpl amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("swap amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(approveFsRpl(c, amountWei))
					return nil

				},
			},
			{
				Name:      "wait-and-swap-rpl",
				Aliases:   []string{"p2"},
				Usage:     "Swap old RPL for new RPL, waiting for the approval TX hash to be included in a block first",
				UsageText: "rocketpool api node wait-and-swap-rpl amount tx-hash",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("swap amount", c.Args().Get(0))
					if err != nil {
						return err
					}
					hash, err := cliutils.ValidateTxHash("swap amount", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(waitForApprovalAndSwapFsRpl(c, amountWei, hash))
					return nil

				},
			},
			{
				Name:      "get-swap-rpl-approval-gas",
				Usage:     "Estimate the gas cost of legacy RPL interaction approval",
				UsageText: "rocketpool api node get-swap-rpl-approval-gas",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("approve amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(getSwapApprovalGas(c, amountWei))
					return nil

				},
			},
			{
				Name:      "swap-rpl-allowance",
				Usage:     "Get the node's legacy RPL allowance for new RPL contract",
				UsageText: "rocketpool api node swap-allowance-rpl",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(allowanceFsRpl(c))
					return nil

				},
			},
			{
				Name:      "swap-rpl",
				Aliases:   []string{"p3"},
				Usage:     "Swap old RPL for new RPL",
				UsageText: "rocketpool api node swap-rpl amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("swap amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(swapRpl(c, amountWei))
					return nil

				},
			},

			{
				Name:      "can-stake-rpl",
				Usage:     "Check whether the node can stake RPL",
				UsageText: "rocketpool api node can-stake-rpl amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("stake amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canNodeStakeRpl(c, amountWei))
					return nil

				},
			},
			{
				Name:      "stake-rpl-approve-rpl",
				Aliases:   []string{"k1"},
				Usage:     "Approve RPL for staking against the node",
				UsageText: "rocketpool api node stake-rpl-approve-rpl amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("stake amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(approveRpl(c, amountWei))
					return nil

				},
			},
			{
				Name:      "wait-and-stake-rpl",
				Aliases:   []string{"k2"},
				Usage:     "Stake RPL against the node, waiting for approval tx-hash to be included in a block first",
				UsageText: "rocketpool api node wait-and-stake-rpl amount tx-hash",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("stake amount", c.Args().Get(0))
					if err != nil {
						return err
					}
					hash, err := cliutils.ValidateTxHash("tx-hash", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(waitForApprovalAndStakeRpl(c, amountWei, hash))
					return nil

				},
			},
			{
				Name:      "get-stake-rpl-approval-gas",
				Usage:     "Estimate the gas cost of new RPL interaction approval",
				UsageText: "rocketpool api node get-stake-rpl-approval-gas",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("approve amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(getStakeApprovalGas(c, amountWei))
					return nil

				},
			},
			{
				Name:      "stake-rpl-allowance",
				Usage:     "Get the node's RPL allowance for the staking contract",
				UsageText: "rocketpool api node stake-allowance-rpl",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(allowanceRpl(c))
					return nil

				},
			},
			{
				Name:      "stake-rpl",
				Aliases:   []string{"k3"},
				Usage:     "Stake RPL against the node",
				UsageText: "rocketpool api node stake-rpl amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("stake amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(stakeRpl(c, amountWei))
					return nil

				},
			},
			{
				Name:      "can-set-rpl-locking-allowed",
				Usage:     "Check whether the node can set the RPL lock allowed status",
				UsageText: "rocketpool api node can-set-rpl-locking-allowed caller allowed",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					allowedString := c.Args().Get(0)
					allowed, err := cliutils.ValidateBool("allowed", allowedString)
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canSetRplLockAllowed(c, allowed))
					return nil

				},
			},

			{
				Name:      "set-rpl-locking-allowed",
				Usage:     "Sets the node RPL locking allowed status",
				UsageText: "rocketpool api node set-rpl-locking-allowed caller allowed",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					allowedString := c.Args().Get(0)
					allowed, err := cliutils.ValidateBool("allowed", allowedString)
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(setRplLockAllowed(c, allowed))
					return nil

				},
			},

			{
				Name:      "can-set-stake-rpl-for-allowed",
				Usage:     "Check whether the node can set allowed status for an address to stake RPL on behalf of themself",
				UsageText: "rocketpool api node can-set-stake-rpl-for-allowed caller allowed",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}

					callerAddressString := c.Args().Get(0)
					callerAddress, err := cliutils.ValidateAddress("caller", callerAddressString)
					if err != nil {
						return err
					}

					allowedString := c.Args().Get(1)
					allowed, err := cliutils.ValidateBool("allowed", allowedString)
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canSetStakeRplForAllowed(c, callerAddress, allowed))
					return nil

				},
			},
			{
				Name:      "set-stake-rpl-for-allowed",
				Aliases:   []string{"kf"},
				Usage:     "Sets the allowed status for an address to stake RPL on behalf of your node",
				UsageText: "rocketpool api node set-stake-rpl-for-allowed caller allowed",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}

					callerAddressString := c.Args().Get(0)
					callerAddress, err := cliutils.ValidateAddress("caller", callerAddressString)
					if err != nil {
						return err
					}

					allowedString := c.Args().Get(1)
					allowed, err := cliutils.ValidateBool("allowed", allowedString)
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(setStakeRplForAllowed(c, callerAddress, allowed))

					return nil
				},
			},
			{
				Name:      "can-withdraw-credit",
				Usage:     "Check whether the node can withdraw credit",
				UsageText: "rocketpool api node can-withdraw-credit amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("withdrawal amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canNodeWithdrawCredit(c, amountWei))
					return nil

				},
			},
			{
				Name:      "withdraw-credit",
				Aliases:   []string{"wc"},
				Usage:     "Withdraw credit from the node",
				UsageText: "rocketpool api node withdraw-credit amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("withdrawal amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(nodeWithdrawCredit(c, amountWei))
					return nil

				},
			},
			{
				Name:      "can-withdraw-eth",
				Usage:     "Check whether the node can withdraw ETH staked on its behalf",
				UsageText: "rocketpool api node can-withdraw-eth amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("withdrawal amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canNodeWithdrawEth(c, amountWei))
					return nil

				},
			},
			{
				Name:      "withdraw-eth",
				Aliases:   []string{"i"},
				Usage:     "Withdraw ETH staked on behalf of the node",
				UsageText: "rocketpool api node withdraw-eth amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("withdrawal amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(nodeWithdrawEth(c, amountWei))
					return nil

				},
			},
			{
				Name:      "can-unstake-legacy-rpl",
				Usage:     "Check whether the node can withdraw legacy staked RPL",
				UsageText: "rocketpool api node can-withdraw-legacy-rpl amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("withdrawal amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canNodeUnstakeLegacyRpl(c, amountWei))
					return nil

				},
			},
			{
				Name:      "can-withdraw-rpl",
				Usage:     "Check whether the node can withdraw staked RPL",
				UsageText: "rocketpool api node can-withdraw-rpl",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(canNodeWithdrawRpl(c))
					return nil

				},
			},
			{
				Name:      "withdraw-rpl",
				Aliases:   []string{"w"},
				Usage:     "Withdraw RPL staked against the node",
				UsageText: "rocketpool api node withdraw-rpl",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(nodeWithdrawRpl(c))
					return nil

				},
			},
			{
				Name:      "can-withdraw-rpl-v131",
				Usage:     "Check whether the node can withdraw staked RPL",
				UsageText: "rocketpool api node can-withdraw-rpl-v131 amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("withdrawal amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canNodeWithdrawRplv1_3_1(c, amountWei))
					return nil

				},
			},
			{
				Name:      "withdraw-rpl-v131",
				Aliases:   []string{"w"},
				Usage:     "Withdraw RPL staked against the node",
				UsageText: "rocketpool api node withdraw-rpl-v131 amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("withdrawal amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(nodeWithdrawRplv1_3_1(c, amountWei))
					return nil

				},
			},
			{
				Name:      "can-unstake-rpl",
				Usage:     "Check whether the node can unstake RPL",
				UsageText: "rocketpool api node can-unstake-rpl amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					amountWei, err := cliutils.ValidatePositiveWeiAmount("unstake amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canNodeUnstakeRpl(c, amountWei))
					return nil

				},
			},
			{
				Name:      "unstake-rpl",
				Aliases:   []string{"u"},
				Usage:     "Unstake RPL from the node",
				UsageText: "rocketpool api node withdraw-rpl amount",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					amountWei, err := cliutils.ValidatePositiveWeiAmount("unstake amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(nodeUnstakeRpl(c, amountWei))
					return nil

				},
			},

			{
				Name:      "can-deposit",
				Usage:     "Check whether the node can make a deposit",
				UsageText: "rocketpool api node can-deposit amount min-fee salt use-express-ticket",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 4); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("deposit amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					minNodeFee, err := cliutils.ValidateFraction("minimum node fee", c.Args().Get(1))
					if err != nil {
						return err
					}
					salt, err := cliutils.ValidateBigInt("salt", c.Args().Get(2))
					if err != nil {
						return err
					}

					useExpressTicketString := c.Args().Get(3)
					useExpressTicket, err := cliutils.ValidateBool("use-express-ticket", useExpressTicketString)
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canNodeDeposit(c, amountWei, minNodeFee, salt, useExpressTicket))
					return nil

				},
			},
			{
				Name:      "deposit",
				Aliases:   []string{"d"},
				Usage:     "Make a deposit and create a minipool, or just make and sign the transaction (when submit = false)",
				UsageText: "rocketpool api node deposit amount min-node-fee salt use-credit-balance use-express-ticket submit",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 6); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("deposit amount", c.Args().Get(0))
					if err != nil {
						return err
					}

					minNodeFee, err := cliutils.ValidateFraction("minimum node fee", c.Args().Get(1))
					if err != nil {
						return err
					}

					salt, err := cliutils.ValidateBigInt("salt", c.Args().Get(2))
					if err != nil {
						return err
					}

					useCreditBalanceString := c.Args().Get(3)
					useCreditBalance, err := cliutils.ValidateBool("use-credit-balance", useCreditBalanceString)
					if err != nil {
						return err
					}

					useExpressTicketString := c.Args().Get(4)
					useExpressTicket, err := cliutils.ValidateBool("use-express-ticket", useExpressTicketString)
					if err != nil {
						return err
					}
					submit, err := cliutils.ValidateBool("submit", c.Args().Get(5))
					if err != nil {
						return err
					}

					if err != nil {
						return err
					}

					// Run
					response, err := nodeDeposit(c, amountWei, minNodeFee, salt, useCreditBalance, useExpressTicket, submit)
					if submit {
						api.PrintResponse(response, err)
					} // else nodeDeposit already printed the encoded transaction
					return nil

				},
			},

			{
				Name:      "can-send",
				Usage:     "Check whether the node can send ETH or tokens to an address",
				UsageText: "rocketpool api node can-send amount token to",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 3); err != nil {
						return err
					}
					amountRaw, err := cliutils.ValidatePositiveEthAmount("send amount", c.Args().Get(0))
					if err != nil {
						return err
					}
					token, err := cliutils.ValidateTokenType("token type", c.Args().Get(1))
					if err != nil {
						return err
					}
					toAddress, err := cliutils.ValidateAddress("to address", c.Args().Get(2))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canNodeSend(c, amountRaw, token, toAddress))
					return nil

				},
			},
			{
				Name:      "send",
				Aliases:   []string{"n"},
				Usage:     "Send ETH or tokens from the node account to an address",
				UsageText: "rocketpool api node send amount token to",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 3); err != nil {
						return err
					}
					amountRaw, err := cliutils.ValidatePositiveEthAmount("send amount", c.Args().Get(0))
					if err != nil {
						return err
					}
					token, err := cliutils.ValidateTokenType("token type", c.Args().Get(1))
					if err != nil {
						return err
					}
					toAddress, err := cliutils.ValidateAddress("to address", c.Args().Get(2))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(nodeSend(c, amountRaw, token, toAddress))
					return nil

				},
			},

			{
				Name:      "can-burn",
				Usage:     "Check whether the node can burn tokens for ETH",
				UsageText: "rocketpool api node can-burn amount token",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("burn amount", c.Args().Get(0))
					if err != nil {
						return err
					}
					token, err := cliutils.ValidateBurnableTokenType("token type", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canNodeBurn(c, amountWei, token))
					return nil

				},
			},
			{
				Name:      "burn",
				Aliases:   []string{"b"},
				Usage:     "Burn tokens for ETH",
				UsageText: "rocketpool api node burn amount token",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidatePositiveWeiAmount("burn amount", c.Args().Get(0))
					if err != nil {
						return err
					}
					token, err := cliutils.ValidateBurnableTokenType("token type", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(nodeBurn(c, amountWei, token))
					return nil

				},
			},

			{
				Name:      "can-claim-rpl-rewards",
				Usage:     "Check whether the node has RPL rewards available to claim",
				UsageText: "rocketpool api node can-claim-rpl-rewards",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(canNodeClaimRpl(c))
					return nil

				},
			},
			{
				Name:      "claim-rpl-rewards",
				Usage:     "Claim available RPL rewards",
				UsageText: "rocketpool api node claim-rpl-rewards",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(nodeClaimRpl(c))
					return nil

				},
			},

			{
				Name:      "rewards",
				Usage:     "Get RPL rewards info",
				UsageText: "rocketpool api node rewards",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getRewards(c))
					return nil

				},
			},

			{
				Name:      "deposit-contract-info",
				Usage:     "Get information about the deposit contract specified by Rocket Pool and the Beacon Chain client",
				UsageText: "rocketpool api node deposit-contract-info",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getDepositContractInfo(c))
					return nil

				},
			},

			{
				Name:      "sign",
				Usage:     "Signs a transaction with the node's private key. The TX must be serialized as a hex string.",
				UsageText: "rocketpool api node sign tx",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					data := c.Args().Get(0)

					// Run
					api.PrintResponse(sign(c, data))
					return nil

				},
			},

			{
				Name:      "sign-message",
				Usage:     "Signs an arbitrary message with the node's private key.",
				UsageText: "rocketpool api node sign-message 'message'",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					message := c.Args().Get(0)

					// Run
					api.PrintResponse(signMessage(c, message))
					return nil

				},
			},

			{
				Name:      "is-fee-distributor-initialized",
				Usage:     "Check if the fee distributor contract for this node is initialized and deployed",
				UsageText: "rocketpool api node is-fee-distributor-initialized",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(isFeeDistributorInitialized(c))
					return nil

				},
			},
			{
				Name:      "get-initialize-fee-distributor-gas",
				Usage:     "Estimate the cost of initializing the fee distributor",
				UsageText: "rocketpool api node get-initialize-fee-distributor-gas",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getInitializeFeeDistributorGas(c))
					return nil
				},
			},

			{
				Name:      "initialize-fee-distributor",
				Usage:     "Initialize and deploy the fee distributor contract for this node",
				UsageText: "rocketpool api node initialize-fee-distributor",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(initializeFeeDistributor(c))
					return nil

				},
			},

			{
				Name:      "can-distribute",
				Usage:     "Check if distributing ETH from the node's fee distributor is possible",
				UsageText: "rocketpool api node can-distribute",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(canDistribute(c))
					return nil

				},
			},
			{
				Name:      "distribute",
				Usage:     "Distribute ETH from the node's fee distributor",
				UsageText: "rocketpool api node distribute",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(distribute(c))
					return nil

				},
			},
			{
				Name:      "claim-rpl-rewards",
				Usage:     "Claim available RPL rewards",
				UsageText: "rocketpool api node claim-rpl-rewards",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(nodeClaimRpl(c))
					return nil

				},
			},

			{
				Name:      "get-rewards-info",
				Usage:     "Get info about your eligible rewards periods, including balances and Merkle proofs",
				UsageText: "rocketpool api node get-rewards-info",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getRewardsInfo(c))
					return nil

				},
			},
			{
				Name:      "can-claim-rewards",
				Usage:     "Check if the rewards for the given intervals can be claimed",
				UsageText: "rocketpool api node can-claim-rewards 0,1,2,5,6",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					indicesString := c.Args().Get(0)

					// Run
					api.PrintResponse(canClaimRewards(c, indicesString))
					return nil

				},
			},
			{
				Name:      "claim-rewards",
				Usage:     "Claim rewards for the given reward intervals",
				UsageText: "rocketpool api node claim-rewards 0,1,2,5,6",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					indicesString := c.Args().Get(0)

					// Run
					api.PrintResponse(claimRewards(c, indicesString))
					return nil

				},
			},
			{
				Name:      "can-claim-and-stake-rewards",
				Usage:     "Check if the rewards for the given intervals can be claimed, and RPL restaked automatically",
				UsageText: "rocketpool api node can-claim-and-stake-rewards 0,1,2,5,6 amount-to-restake",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					indicesString := c.Args().Get(0)

					stakeAmount, err := cliutils.ValidateBigInt("stakeAmount", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canClaimAndStakeRewards(c, indicesString, stakeAmount))
					return nil

				},
			},
			{
				Name:      "claim-and-stake-rewards",
				Usage:     "Claim rewards for the given reward intervals and restake RPL automatically",
				UsageText: "rocketpool api node claim-and-stake-rewards 0,1,2,5,6 amount-to-restake",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					indicesString := c.Args().Get(0)

					stakeAmount, err := cliutils.ValidateBigInt("stakeAmount", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(claimAndStakeRewards(c, indicesString, stakeAmount))
					return nil

				},
			},

			{
				Name:      "get-smoothing-pool-registration-status",
				Usage:     "Check whether or not the node is opted into the Smoothing Pool",
				UsageText: "rocketpool api node get-smoothing-pool-registration-status",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getSmoothingPoolRegistrationStatus(c))
					return nil

				},
			},
			{
				Name:      "can-set-smoothing-pool-status",
				Usage:     "Check if the node's Smoothing Pool status can be changed",
				UsageText: "rocketpool api node can-set-smoothing-pool-status status",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					status, err := cliutils.ValidateBool("status", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canSetSmoothingPoolStatus(c, status))
					return nil

				},
			},
			{
				Name:      "set-smoothing-pool-status",
				Usage:     "Sets the node's Smoothing Pool opt-in status",
				UsageText: "rocketpool api node set-smoothing-pool-status status",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					status, err := cliutils.ValidateBool("status", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(setSmoothingPoolStatus(c, status))
					return nil

				},
			},
			{
				Name:      "resolve-ens-name",
				Usage:     "Resolve an ENS name",
				UsageText: "rocketpool api node resolve-ens-name name",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Run
					api.PrintResponse(resolveEnsName(c, c.Args().Get(0)))
					return nil

				},
			},
			{
				Name:      "reverse-resolve-ens-name",
				Usage:     "Reverse resolve an address to an ENS name",
				UsageText: "rocketpool api node reverse-resolve-ens-name address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					address, err := cliutils.ValidateAddress("address", c.Args().Get(0))
					if err != nil {
						return err
					}
					// Run
					api.PrintResponse(reverseResolveEnsName(c, address))
					return nil

				},
			},

			{
				Name:      "check-collateral",
				Usage:     "Check if the node is above the minimum collateralization threshold, including pending bond reductions",
				UsageText: "rocketpool api node check-collateral",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(checkCollateral(c))
					return nil

				},
			},

			{
				Name:      "get-eth-balance",
				Usage:     "Get the ETH balance of the node address",
				UsageText: "rocketpool api node get-eth-balance",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getNodeEthBalance(c))
					return nil

				},
			},

			{
				Name:      "can-send-message",
				Usage:     "Estimates the gas for sending a zero-value message with a payload",
				UsageText: "rocketpool api node can-send-message address message",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					address, err := cliutils.ValidateAddress("address", c.Args().Get(0))
					if err != nil {
						return err
					}
					message, err := cliutils.ValidateByteArray("message", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canSendMessage(c, address, message))
					return nil

				},
			},
			{
				Name:      "send-message",
				Usage:     "Sends a zero-value message with a payload",
				UsageText: "rocketpool api node send-message address message",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					address, err := cliutils.ValidateAddress("address", c.Args().Get(0))
					if err != nil {
						return err
					}
					message, err := cliutils.ValidateByteArray("message", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(sendMessage(c, address, message))
					return nil

				},
			},
			{
				Name:      "get-express-ticket-count",
				Usage:     "Get the number of express tickets available for the node",
				UsageText: "rocketpool api node get-express-ticket-count",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}
					api.PrintResponse(getExpressTicketCount(c))
					return nil
				},
			},
			{
				Name:      "can-claim-unclaimed-rewards",
				Usage:     "Check if any unclaimed rewards can be sent to the node's withdrawal address",
				UsageText: "rocketpool api node can-claim-unclaimed-rewards address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					// Get amount
					nodeAddress, err := cliutils.ValidateAddress("address", c.Args().Get(0))
					if err != nil {
						return err
					}
					// Run
					api.PrintResponse(canClaimUnclaimedRewards(c, nodeAddress))
					return nil

				},
			},
			{
				Name:      "claim-unclaimed-rewards",
				Usage:     "Send unclaimed rewards to the node's withdrawal address",
				UsageText: "rocketpool api node claim-unclaimed-rewards address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					// Get amount
					nodeAddress, err := cliutils.ValidateAddress("address", c.Args().Get(0))
					if err != nil {
						return err
					}
					// Run
					api.PrintResponse(canClaimUnclaimedRewards(c, nodeAddress))
					return nil

				},
			},
		},
	})
}
