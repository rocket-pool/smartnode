package node

import (
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/input"
	cliutils "github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/shared/utils"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, &cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the node",
		Subcommands: []*cli.Command{

			{
				Name:    "status",
				Aliases: []string{"s"},
				Usage:   "Get the node's status",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getStatus(c)
				},
			},

			{
				Name:    "register",
				Aliases: []string{"r"},
				Usage:   "Register the node with Rocket Pool",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    registerTimezoneFlag,
						Aliases: []string{"t"},
						Usage:   "The timezone location to register the node with (in the format 'Country/City')",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String(registerTimezoneFlag) != "" {
						if _, err := input.ValidateTimezoneLocation("timezone location", c.String(registerTimezoneFlag)); err != nil {
							return err
						}
					}

					// Run
					return registerNode(c)
				},
			},

			{
				Name:    "rewards",
				Aliases: []string{"e"},
				Usage:   "Get the time and your expected RPL rewards of the next checkpoint",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getRewards(c)
				},
			},

			{
				Name:      "set-primary-withdrawal-address",
				Aliases:   []string{"w"},
				Usage:     "Set the node's primary withdrawal address, which will receive all ETH rewards (and RPL if the RPL withdrawal address is not set)",
				ArgsUsage: "address",
				Flags: []cli.Flag{
					cliutils.YesFlag,
					&cli.BoolFlag{
						Name:  setPrimaryWithdrawalAddressForceFlag,
						Usage: "Force update the primary withdrawal address, bypassing the 'pending' state that requires a confirmation transaction from the new address",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					withdrawalAddress := c.Args().Get(0)

					// Run
					return setPrimaryWithdrawalAddress(c, withdrawalAddress)
				},
			},

			{
				Name:    "confirm-primary-withdrawal-address",
				Aliases: []string{"f"},
				Usage:   "Confirm the node's pending primary withdrawal address if it has been set back to the node's address itself",
				Flags: []cli.Flag{
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return confirmPrimaryWithdrawalAddress(c)
				},
			},

			{
				Name:      "set-rpl-withdrawal-address",
				Aliases:   []string{"srwa"},
				Usage:     "Set the node's RPL withdrawal address, which will receive all RPL rewards and staked RPL withdrawals",
				ArgsUsage: "address",
				Flags: []cli.Flag{
					cliutils.YesFlag,
					&cli.BoolFlag{
						Name:  setRplWithdrawalAddressForceFlag,
						Usage: "Force update the RPL withdrawal address, bypassing the 'pending' state that requires a confirmation transaction from the new address",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					withdrawalAddress := c.Args().Get(0)

					// Run
					return setRplWithdrawalAddress(c, withdrawalAddress)
				},
			},

			{
				Name:    "confirm-rpl-withdrawal-address",
				Aliases: []string{"crwa"},
				Usage:   "Confirm the node's pending RPL withdrawal address if it has been set back to the node's address itself",
				Flags: []cli.Flag{
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return confirmRplWithdrawalAddress(c)
				},
			},

			{
				Name:    "allow-rpl-locking",
				Aliases: []string{"arl"},
				Usage:   "Allow the node to lock RPL when creating governance proposals/challenges",
				Flags: []cli.Flag{
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return setRplLockingAllowed(c, true)
				},
			},

			{
				Name:    "deny-rpl-locking",
				Aliases: []string{"drl"},
				Usage:   "Do not allow the node to lock RPL when creating governance proposals/challenges",
				Flags: []cli.Flag{
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}
					// Run
					return setRplLockingAllowed(c, false)
				},
			},

			{
				Name:    "set-timezone",
				Aliases: []string{"t"},
				Usage:   "Set the node's timezone location",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    timezoneFlag,
						Aliases: []string{"t"},
						Usage:   "The timezone location to set for the node (in the format 'Country/City')",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String(timezoneFlag) != "" {
						if _, err := input.ValidateTimezoneLocation("timezone location", c.String("timezone")); err != nil {
							return err
						}
					}

					// Run
					return setTimezoneLocation(c)
				},
			},

			{
				Name:    "swap-rpl",
				Aliases: []string{"p"},
				Usage:   "Swap old RPL for new RPL",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    amountFlag,
						Aliases: []string{"a"},
						Usage:   "The amount of old RPL to swap (or 'all')",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String(amountFlag) != "" && c.String(amountFlag) != "all" {
						if _, err := input.ValidatePositiveEthAmount("swap amount", c.String(amountFlag)); err != nil {
							return err
						}
					}

					// Run
					return nodeSwapRpl(c)
				},
			},

			{
				Name:    "stake-rpl",
				Aliases: []string{"k"},
				Usage:   "Stake RPL against the node",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    amountFlag,
						Aliases: []string{"a"},
						Usage:   "The amount of RPL to stake (also accepts 'min8' / 'max8' for 8-ETH minipools, 'min16' / 'max16' for 16-ETH minipools, or 'all' for all of your RPL)",
					},
					cliutils.YesFlag,
					&cli.BoolFlag{
						Name:    swapFlag,
						Aliases: []string{"s"},
						Usage:   "Automatically confirm swapping legacy RPL before staking",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String(amountFlag) != "" &&
						c.String(amountFlag) != "min8" &&
						c.String(amountFlag) != "max8" &&
						c.String(amountFlag) != "min16" &&
						c.String(amountFlag) != "max16" &&
						c.String(amountFlag) != "all" {
						if _, err := input.ValidatePositiveEthAmount("stake amount", c.String(amountFlag)); err != nil {
							return err
						}
					}

					// Run
					return nodeStakeRpl(c)
				},
			},

			{
				Name:      "add-address-to-stake-rpl-whitelist",
				Aliases:   []string{"asw"},
				Usage:     "Adds an address to your node's RPL staking whitelist, so it can stake RPL on behalf of your node.",
				ArgsUsage: "address",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					addressOrEns := c.Args().Get(0)

					// Run
					return setStakeRplForAllowed(c, addressOrEns, true)
				},
			},

			{
				Name:      "remove-address-from-stake-rpl-whitelist",
				Aliases:   []string{"rsw"},
				Usage:     "Removes an address from your node's RPL staking whitelist, so it can no longer stake RPL on behalf of your node.",
				ArgsUsage: "address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := utils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					addressOrEns := c.Args().Get(0)

					// Run
					return setStakeRplForAllowed(c, addressOrEns, false)

				},
			},

			{
				Name:    "claim-rewards",
				Aliases: []string{"c"},
				Usage:   "Claim available RPL and ETH rewards for any checkpoint you haven't claimed yet",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    claimRestakeFlag,
						Aliases: []string{"a"},
						Usage:   "The amount of RPL to automatically restake during claiming (or 'all' for all available RPL)",
					},
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return nodeClaimRewards(c)
				},
			},

			{
				Name:    "withdraw-rpl",
				Aliases: []string{"wr"},
				Usage:   "Withdraw RPL staked against the node",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    amountFlag,
						Aliases: []string{"a"},
						Usage:   "The amount of RPL to withdraw (or 'max')",
					},
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String(amountFlag) != "" && c.String(amountFlag) != "max" {
						if _, err := input.ValidatePositiveEthAmount("withdrawal amount", c.String(amountFlag)); err != nil {
							return err
						}
					}

					// Run
					return nodeWithdrawRpl(c)
				},
			},
			{
				Name:    "withdraw-eth",
				Aliases: []string{"we"},
				Usage:   "Withdraw ETH staked on behalf of the node",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    amountFlag,
						Aliases: []string{"a"},
						Usage:   "The amount of ETH to withdraw (or 'max')",
					},
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String(amountFlag) != "" && c.String(amountFlag) != "max" {
						if _, err := input.ValidatePositiveEthAmount("withdrawal amount", c.String(amountFlag)); err != nil {
							return err
						}
					}

					// Run
					return nodeWithdrawEth(c)
				},
			},

			{
				Name:    "deposit",
				Aliases: []string{"d"},
				Usage:   "Make a deposit and create a minipool",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    amountFlag,
						Aliases: []string{"a"},
						Usage:   "The amount of ETH to deposit (8 or 16)",
					},
					&cli.StringFlag{
						Name:    maxSlippageFlag,
						Aliases: []string{"s"},
						Usage:   "The maximum acceptable slippage in node commission rate for the deposit (or 'auto'). Only relevant when the commission rate is not fixed.",
					},
					cliutils.YesFlag,
					&cli.StringFlag{
						Name:    saltFlag,
						Aliases: []string{"l"},
						Usage:   "An optional seed to use when generating the new minipool's address. Use this if you want it to have a custom vanity address.",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String(amountFlag) != "" {
						if _, err := input.ValidatePositiveEthAmount("deposit amount", c.String(amountFlag)); err != nil {
							return err
						}
					}
					if c.String(maxSlippageFlag) != "" && c.String(maxSlippageFlag) != "auto" {
						if _, err := input.ValidatePercentage("maximum commission rate slippage", c.String(maxSlippageFlag)); err != nil {
							return err
						}
					}
					if c.String(saltFlag) != "" {
						if _, err := input.ValidateBigInt("salt", c.String(saltFlag)); err != nil {
							return err
						}
					}

					// Run
					return nodeDeposit(c)
				},
			},

			{
				Name:    "create-vacant-minipool",
				Aliases: []string{"cvm"},
				Usage:   "Create an empty minipool, which can be used to migrate an existing solo staking validator as part of the 0x00 to 0x01 withdrawal credentials upgrade",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    amountFlag,
						Aliases: []string{"a"},
						Usage:   "The amount of ETH to deposit (8 or 16)",
					},
					&cli.StringFlag{
						Name:    maxSlippageFlag,
						Aliases: []string{"s"},
						Usage:   "The maximum acceptable slippage in node commission rate for the deposit (or 'auto'). Only relevant when the commission rate is not fixed.",
					},
					cliutils.YesFlag,
					&cli.StringFlag{
						Name:    saltFlag,
						Aliases: []string{"l"},
						Usage:   "An optional seed to use when generating the new minipool's address. Use this if you want it to have a custom vanity address.",
					},
					&cli.StringFlag{
						Name:    cvmMnemonicFlag,
						Aliases: []string{"m"},
						Usage:   "Use this flag if you want to recreate your validator's private key within the Smartnode's VC instead of running it via your own VC, and have the Smartnode reassign your validator's withdrawal credentials to the new minipool address automatically.",
					},
					&cli.BoolFlag{
						Name:  cliutils.NoRestartFlag,
						Usage: "Don't restart the Validator Client after importing the key. Note that the key won't be loaded (and won't attest) until you restart the VC to load it.",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					pubkey, err := input.ValidatePubkey("pubkey", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Validate flags
					if c.String("amount") != "" {
						if _, err := input.ValidatePositiveEthAmount("deposit amount", c.String("amount")); err != nil {
							return err
						}
					}
					if c.String("max-slippage") != "" && c.String("max-slippage") != "auto" {
						if _, err := input.ValidatePercentage("maximum commission rate slippage", c.String("max-slippage")); err != nil {
							return err
						}
					}
					if c.String("salt") != "" {
						if _, err := input.ValidateBigInt("salt", c.String("salt")); err != nil {
							return err
						}
					}

					// Run
					return createVacantMinipool(c, pubkey)
				},
			},

			{
				Name:      "send",
				Aliases:   []string{"n"},
				Usage:     "Send ETH or tokens from the node account to an address. ENS names supported. <token> can be 'rpl', 'eth', 'fsrpl' (for the old RPL v1 token), 'reth', or the address of an arbitrary token you want to send (including the 0x prefix).",
				ArgsUsage: "amount token to",
				Flags: []cli.Flag{
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 3); err != nil {
						return err
					}
					amount, err := input.ValidatePositiveEthAmount("send amount", c.Args().Get(0))
					if err != nil {
						return err
					}
					token, err := utils.ValidateTokenType("token type", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					return nodeSend(c, amount, token, c.Args().Get(2))
				},
			},

			{
				Name:      "set-voting-delegate",
				Aliases:   []string{"sv"},
				Usage:     "Set the address you want to use when voting on Rocket Pool Snapshot governance proposals, or the address you want to delegate your voting power to.",
				ArgsUsage: "address",
				Flags: []cli.Flag{
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					delegate := c.Args().Get(0)

					// Run
					return nodeSetVotingDelegate(c, delegate)
				},
			},
			{
				Name:    "clear-voting-delegate",
				Aliases: []string{"cv"},
				Usage:   "Remove the address you've set for voting on Rocket Pool governance proposals.",
				Flags: []cli.Flag{
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return nodeClearVotingDelegate(c)
				},
			},

			{
				Name:    "initialize-fee-distributor",
				Aliases: []string{"z"},
				Usage:   "Create the fee distributor contract for your node, so you can withdraw priority fees and MEV rewards after the merge",
				Flags: []cli.Flag{
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return initializeFeeDistributor(c)
				},
			},

			{
				Name:    "distribute-fees",
				Aliases: []string{"b"},
				Usage:   "Distribute the priority fee and MEV rewards from your fee distributor to your withdrawal address and the rETH contract (based on your node's average commission)",
				Flags: []cli.Flag{
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return distribute(c)
				},
			},

			{
				Name:    "join-smoothing-pool",
				Aliases: []string{"js"},
				Usage:   "Opt your node into the Smoothing Pool",
				Flags: []cli.Flag{
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return setSmoothingPoolState(c, true)
				},
			},

			{
				Name:    "leave-smoothing-pool",
				Aliases: []string{"ls"},
				Usage:   "Leave the Smoothing Pool",
				Flags: []cli.Flag{
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return setSmoothingPoolState(c, false)
				},
			},
		},
	})
}
