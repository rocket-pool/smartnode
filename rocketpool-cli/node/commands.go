package node

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"

	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register commands
func RegisterCommands(app *cli.Command, name string, aliases []string) {
	app.Commands = append(app.Commands, &cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the node",
		Commands: []*cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get the node's status",
				UsageText: "rocketpool node status",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getStatus()

				},
			},

			{
				Name:      "sync",
				Aliases:   []string{"y"},
				Usage:     "Get the sync progress of the eth1 and eth2 clients",
				UsageText: "rocketpool node sync",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getSyncProgress()

				},
			},

			{
				Name:      "register",
				Aliases:   []string{"r"},
				Usage:     "Register the node with Rocket Pool",
				UsageText: "rocketpool node register [options]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "timezone",
						Aliases: []string{"t"},
						Usage:   "The timezone location to register the node with (in the format 'Country/City')",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("timezone") != "" {
						if _, err := cliutils.ValidateTimezoneLocation("timezone location", c.String("timezone")); err != nil {
							return err
						}
					}

					// Run
					return registerNode(c.String("timezone"), c.Bool("yes"))

				},
			},

			{
				Name:      "rewards",
				Aliases:   []string{"e"},
				Usage:     "Get the time and your expected RPL rewards of the next checkpoint",
				UsageText: "rocketpool node rewards",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm actions",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getRewards(c.Bool("yes"))

				},
			},

			{
				Name:      "set-primary-withdrawal-address",
				Aliases:   []string{"w"},
				Usage:     "Set the node's primary withdrawal address, which will receive all ETH rewards (and RPL if the RPL withdrawal address is not set)",
				UsageText: "rocketpool node set-primary-withdrawal-address [options] address",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm setting primary withdrawal address",
					},
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "Force update the primary withdrawal address, bypassing the 'pending' state that requires a confirmation transaction from the new address",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					withdrawalAddress := c.Args().Get(0)

					// Run
					return setPrimaryWithdrawalAddress(withdrawalAddress, c.Bool("yes"), c.Bool("force"))

				},
			},

			{
				Name:      "confirm-primary-withdrawal-address",
				Aliases:   []string{"f"},
				Usage:     "Confirm the node's pending primary withdrawal address if it has been set back to the node's address itself",
				UsageText: "rocketpool node confirm-primary-withdrawal-address [options]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm withdrawal address",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return confirmPrimaryWithdrawalAddress(c.Bool("yes"))

				},
			},

			{
				Name:      "set-rpl-withdrawal-address",
				Aliases:   []string{"srwa"},
				Usage:     "Set the node's RPL withdrawal address, which will receive all RPL rewards and staked RPL withdrawals",
				UsageText: "rocketpool node set-rpl-withdrawal-address [options] address",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm setting rpl withdrawal address",
					},
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "Force update the rpl withdrawal address, bypassing the 'pending' state that requires a confirmation transaction from the new address",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					withdrawalAddress := c.Args().Get(0)

					// Run
					return setRPLWithdrawalAddress(withdrawalAddress, c.Bool("yes"), c.Bool("force"))

				},
			},

			{
				Name:      "confirm-rpl-withdrawal-address",
				Aliases:   []string{"crwa"},
				Usage:     "Confirm the node's pending rpl withdrawal address if it has been set back to the node's address itself",
				UsageText: "rocketpool node confirm-rpl-withdrawal-address [options]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm withdrawal address",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return confirmRPLWithdrawalAddress(c.Bool("yes"))

				},
			},
			{
				Name:      "allow-rpl-locking",
				Aliases:   []string{"arl"},
				Usage:     "Allow the node to lock RPL when creating governance proposals/challenges",
				UsageText: "rocketpool node allow-rpl-locking [options]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm allowing RPL locking",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}
					// Run
					return setRPLLockingAllowed(c.Bool("yes"), true)

				},
			},
			{
				Name:      "deny-rpl-locking",
				Aliases:   []string{"drl"},
				Usage:     "Do not allow the node to lock RPL when creating governance proposals/challenges",
				UsageText: "rocketpool node deny-rpl-locking [options]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm not allowing RPL locking",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}
					// Run
					return setRPLLockingAllowed(c.Bool("yes"), false)

				},
			},
			{
				Name:      "set-timezone",
				Aliases:   []string{"t"},
				Usage:     "Set the node's timezone location",
				UsageText: "rocketpool node set-timezone [options]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "timezone",
						Aliases: []string{"t"},
						Usage:   "The timezone location to set for the node (in the format 'Country/City')",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("timezone") != "" {
						if _, err := cliutils.ValidateTimezoneLocation("timezone location", c.String("timezone")); err != nil {
							return err
						}
					}

					// Run
					return setTimezoneLocation(c.String("timezone"), c.Bool("yes"))

				},
			},

			{
				Name:      "swap-rpl",
				Aliases:   []string{"p"},
				Usage:     "Swap old RPL for new RPL",
				UsageText: "rocketpool node swap-rpl [options]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "amount",
						Aliases: []string{"a"},
						Usage:   "The amount of old RPL to swap (or 'all')",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("amount") != "" && c.String("amount") != "all" {
						if _, err := cliutils.ValidatePositiveEthAmount("swap amount", c.String("amount")); err != nil {
							return err
						}
					}

					// Run
					return nodeSwapRpl(c.String("amount"), c.Bool("yes"))

				},
			},

			{
				Name:      "stake-rpl",
				Aliases:   []string{"k"},
				Usage:     "Stake RPL against the node",
				UsageText: "rocketpool node stake-rpl [options]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "amount",
						Aliases: []string{"a"},
						Usage:   "The amount of RPL to stake (also accepts custom percentages for a validator (eg. 3% of borrowed ETH as RPL), or 'all' for all of your RPL)",
					},
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm RPL stake",
					},
					&cli.BoolFlag{
						Name:    "swap",
						Aliases: []string{"s"},
						Usage:   "Automatically confirm swapping old RPL before staking",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					amount := c.String("amount")
					if amount != "" {
						if strings.HasSuffix(amount, "%") {
							trimmedAmount := strings.TrimSuffix(amount, "%")
							stakePercent, err := strconv.ParseFloat(trimmedAmount, 64)
							if err != nil || stakePercent <= 0 {
								return fmt.Errorf("invalid percentage value: %s", amount)
							}

						} else if amount != "all" {
							// Validate it as a positive ETH amount if it's not a percentage or "all"
							if _, err := cliutils.ValidatePositiveEthAmount("stake amount", amount); err != nil {
								return err
							}
						}
					}

					// Run
					return nodeStakeRpl(c.String("amount"), c.Bool("swap"), c.Bool("yes"))

				},
			},

			{
				Name:      "add-address-to-stake-rpl-whitelist",
				Aliases:   []string{"asw"},
				Usage:     "Adds an address to your node's RPL staking whitelist, so it can stake RPL on behalf of your node.",
				UsageText: "rocketpool node add-address-to-stake-rpl-whitelist address",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					addressOrENS := c.Args().Get(0)

					// Run
					return addAddressToStakeRplWhitelist(addressOrENS, c.Bool("yes"))

				},
			},

			{
				Name:      "remove-address-from-stake-rpl-whitelist",
				Aliases:   []string{"rsw"},
				Usage:     "Removes an address from your node's RPL staking whitelist, so it can no longer stake RPL on behalf of your node.",
				UsageText: "rocketpool node remove-address-from-stake-rpl-whitelist address",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					addressOrENS := c.Args().Get(0)

					// Run
					return removeAddressFromStakeRplWhitelist(addressOrENS, c.Bool("yes"))

				},
			},

			{
				Name:      "claim-rewards",
				Aliases:   []string{"c"},
				Usage:     "Claim available RPL and ETH rewards for any checkpoint you haven't claimed yet",
				UsageText: "rocketpool node claim-rpl [options]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "restake-amount",
						Aliases: []string{"a"},
						Usage:   "The amount of RPL to automatically restake during claiming (or 'all' for all available RPL)",
					},
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm rewards claim",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return nodeClaimRewards(c.String("restake-amount"), c.Bool("yes"))

				},
			},

			{
				Name:      "withdraw-rpl",
				Aliases:   []string{"i"},
				Usage:     "Withdraw RPL staked against the node",
				UsageText: "rocketpool node withdraw-rpl [options]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "amount",
						Aliases: []string{"a"},
						Usage:   "The amount of RPL to withdraw (or 'max')",
					},
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm RPL withdrawal",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("amount") != "" && c.String("amount") != "max" {
						if _, err := cliutils.ValidatePositiveEthAmount("withdrawal amount", c.String("amount")); err != nil {
							return err
						}
					}

					// Run
					return nodeWithdrawRpl(c.String("amount"), c.Bool("yes"))

				},
			},
			{
				Name:      "withdraw-eth",
				Aliases:   []string{"h"},
				Usage:     "Withdraw ETH staked on behalf of the node",
				UsageText: "rocketpool node withdraw-eth [options]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "amount",
						Aliases: []string{"a"},
						Usage:   "The amount of ETH to withdraw (or 'max')",
					},
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm ETH withdrawal",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("amount") != "" && c.String("amount") != "max" {
						if _, err := cliutils.ValidatePositiveEthAmount("withdrawal amount", c.String("amount")); err != nil {
							return err
						}
					}

					// Run
					return nodeWithdrawEth(c.String("amount"), c.Bool("yes"))

				},
			},

			{
				Name:      "withdraw-credit",
				Aliases:   []string{"wc"},
				Usage:     "(Saturn) Withdraw ETH credit from the node as rETH",
				UsageText: "rocketpool node withdraw-credit [options]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "amount",
						Aliases: []string{"a"},
						Usage:   "The amount of ETH to withdraw (or 'max')",
					},
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm ETH withdrawal",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("amount") != "" && c.String("amount") != "max" {
						if _, err := cliutils.ValidatePositiveEthAmount("withdrawal amount", c.String("amount")); err != nil {
							return err
						}
					}

					// Run
					return nodeWithdrawCredit(c.String("amount"), c.Bool("yes"))

				},
			},

			{
				Name:      "send",
				Aliases:   []string{"n"},
				Usage:     "Send ETH or tokens from the node account to an address. ENS names supported. Use 'all' as the amount to send the entire balance. <token> can be 'rpl', 'eth', 'fsrpl' (for the old RPL v1 token), 'reth', or the address of an arbitrary token you want to send (including the 0x prefix).",
				UsageText: "rocketpool node send [options] amount token to",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm token send",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 3); err != nil {
						return err
					}

					amountStr := c.Args().Get(0)
					sendAll := strings.EqualFold(amountStr, "all")
					var amount float64
					if !sendAll {
						var err error
						amount, err = cliutils.ValidatePositiveEthAmount("send amount", amountStr)
						if err != nil {
							return err
						}
					}

					token, err := cliutils.ValidateTokenType("token type", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					return nodeSend(amount, sendAll, token, c.Args().Get(2), c.Bool("yes"))

				},
			},

			{
				Name:      "initialize-fee-distributor",
				Aliases:   []string{"z"},
				Usage:     "Create the fee distributor contract for your node, so you can withdraw priority fees and MEV rewards after the merge",
				UsageText: "rocketpool node initialize-fee-distributor",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm initialization gas costs",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return initializeFeeDistributor(c.Bool("yes"))

				},
			},

			{
				Name:      "distribute-fees",
				Aliases:   []string{"b"},
				Usage:     "Distribute the priority fee and MEV rewards from your fee distributor to your withdrawal address and the rETH contract (based on your node's average commission)",
				UsageText: "rocketpool node distribute-fees",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm distribution",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return distribute(c.Bool("yes"))

				},
			},

			{
				Name:      "join-smoothing-pool",
				Aliases:   []string{"js"},
				Usage:     "Opt your node into the Smoothing Pool",
				UsageText: "rocketpool node join-smoothing-pool",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm opt-in",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return joinSmoothingPool(c.Bool("yes"))

				},
			},

			{
				Name:      "leave-smoothing-pool",
				Aliases:   []string{"ls"},
				Usage:     "Leave the Smoothing Pool",
				UsageText: "rocketpool node leave-smoothing-pool",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm opt-out",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return leaveSmoothingPool(c.Bool("yes"))

				},
			},

			{
				Name:      "sign-message",
				Aliases:   []string{"sm"},
				Usage:     "Sign an arbitrary message with the node's private key",
				UsageText: "rocketpool node sign-message [-m message]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "message",
						Aliases: []string{"m"},
						Usage:   "The 'quoted message' to be signed",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					// Run
					return signMessage(c.String("message"))
				},
			},

			{
				Name:      "send-message",
				Usage:     "Send a zero-ETH transaction to the target address (or ENS) with the provided hex-encoded message as the data payload",
				UsageText: "rocketpool node send-message [-y] to-address hex-message",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm message send",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					message, err := cliutils.ValidateByteArray("message", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					return sendMessage(c.Args().Get(0), message, c.Bool("yes"))

				},
			},

			{
				Name:      "claim-unclaimed-rewards",
				Aliases:   []string{"cur"},
				Usage:     "Sends any unclaimed rewards to the node's withdrawal address",
				UsageText: "rocketpool node claim-unclaimed-rewards",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}
					// Run
					return claimUnclaimedRewards(c.Bool("yes"))

				},
			},

			{
				Name:      "provision-express-tickets",
				Aliases:   []string{"pet"},
				Usage:     "Provision the node's express tickets",
				UsageText: "rocketpool node provision-express-tickets",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return provisionExpressTickets()
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm provision",
					},
				},
			},
		},
	})
}
