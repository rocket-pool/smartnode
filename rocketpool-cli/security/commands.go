package security

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

const (
	boolUsage string = "specify 'true', 'false', 'yes', or 'no'"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the Rocket Pool security council",
		Subcommands: []cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get security council status",
				UsageText: "rocketpool security status",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getStatus(c)

				},
			},

			{
				Name:      "members",
				Aliases:   []string{"m"},
				Usage:     "Get the security council members",
				UsageText: "rocketpool security members",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getMembers(c)

				},
			},

			{
				Name:    "propose",
				Aliases: []string{"p"},
				Usage:   "Make a security council proposal",
				Subcommands: []cli.Command{

					{
						Name:    "member",
						Aliases: []string{"m"},
						Usage:   "Make a security council member proposal",
						Subcommands: []cli.Command{
							{
								Name:      "leave",
								Aliases:   []string{"l"},
								Usage:     "Propose leaving the security council",
								UsageText: "rocketpool security propose member leave",
								Flags: []cli.Flag{
									cli.BoolFlag{
										Name:  "yes, y",
										Usage: "Automatically confirm all interactive questions",
									},
								},
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 0); err != nil {
										return err
									}

									// Run
									return proposeLeave(c)

								},
							},
						},
					},
					{
						Name:    "setting",
						Aliases: []string{"s"},
						Usage:   "Make a proposal to update a Protocol DAO setting",
						Subcommands: []cli.Command{

							{
								Name:    "auction",
								Aliases: []string{"a"},
								Usage:   "Auction settings",
								Subcommands: []cli.Command{

									{
										Name:      "is-create-lot-enabled",
										Aliases:   []string{"icle"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.CreateLotEnabledSettingPath, boolUsage),
										UsageText: "rocketpool security propose setting auction is-create-lot-enabled value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "yes, y",
												Usage: "Automatically confirm all interactive questions",
											},
										},
										Action: func(c *cli.Context) error {

											// Validate args
											if err := cliutils.ValidateArgCount(c, 1); err != nil {
												return err
											}
											value, err := cliutils.ValidateBool("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingAuctionIsCreateLotEnabled(c, value)

										},
									},

									{
										Name:      "is-bid-on-lot-enabled",
										Aliases:   []string{"ibole"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.BidOnLotEnabledSettingPath, boolUsage),
										UsageText: "rocketpool security propose setting auction is-bid-on-lot-enabled value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "yes, y",
												Usage: "Automatically confirm all interactive questions",
											},
										},
										Action: func(c *cli.Context) error {

											// Validate args
											if err := cliutils.ValidateArgCount(c, 1); err != nil {
												return err
											}
											value, err := cliutils.ValidateBool("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingAuctionIsBidOnLotEnabled(c, value)

										},
									},
								},
							},

							{
								Name:    "deposit",
								Aliases: []string{"d"},
								Usage:   "Deposit pool settings",
								Subcommands: []cli.Command{

									{
										Name:      "is-depositing-enabled",
										Aliases:   []string{"ide"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.DepositEnabledSettingPath, boolUsage),
										UsageText: "rocketpool security propose setting deposit is-depositing-enabled value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "yes, y",
												Usage: "Automatically confirm all interactive questions",
											},
										},
										Action: func(c *cli.Context) error {

											// Validate args
											if err := cliutils.ValidateArgCount(c, 1); err != nil {
												return err
											}
											value, err := cliutils.ValidateBool("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingDepositIsDepositingEnabled(c, value)

										},
									},

									{
										Name:      "are-deposit-assignments-enabled",
										Aliases:   []string{"adae"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.AssignDepositsEnabledSettingPath, boolUsage),
										UsageText: "rocketpool security propose setting deposit are-deposit-assignments-enabled value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "yes, y",
												Usage: "Automatically confirm all interactive questions",
											},
										},
										Action: func(c *cli.Context) error {

											// Validate args
											if err := cliutils.ValidateArgCount(c, 1); err != nil {
												return err
											}
											value, err := cliutils.ValidateBool("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingDepositAreDepositAssignmentsEnabled(c, value)

										},
									},
								},
							},

							{
								Name:    "minipool",
								Aliases: []string{"m"},
								Usage:   "Minipool settings",
								Subcommands: []cli.Command{

									{
										Name:      "is-submit-withdrawable-enabled",
										Aliases:   []string{"iswe"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.MinipoolSubmitWithdrawableEnabledSettingPath, boolUsage),
										UsageText: "rocketpool security propose setting minipool is-submit-withdrawable-enabled value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "yes, y",
												Usage: "Automatically confirm all interactive questions",
											},
										},
										Action: func(c *cli.Context) error {

											// Validate args
											if err := cliutils.ValidateArgCount(c, 1); err != nil {
												return err
											}
											value, err := cliutils.ValidateBool("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingMinipoolIsSubmitWithdrawableEnabled(c, value)

										},
									},

									{
										Name:      "is-bond-reduction-enabled",
										Aliases:   []string{"ibre"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.BondReductionEnabledSettingPath, boolUsage),
										UsageText: "rocketpool security propose setting minipool is-bond-reduction-enabled value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "yes, y",
												Usage: "Automatically confirm all interactive questions",
											},
										},
										Action: func(c *cli.Context) error {

											// Validate args
											if err := cliutils.ValidateArgCount(c, 1); err != nil {
												return err
											}
											value, err := cliutils.ValidateBool("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingMinipoolIsBondReductionEnabled(c, value)

										},
									},
								},
							},

							{
								Name:    "network",
								Aliases: []string{"ne"},
								Usage:   "Network settings",
								Subcommands: []cli.Command{

									{
										Name:      "is-submit-balances-enabled",
										Aliases:   []string{"isbe"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.SubmitBalancesEnabledSettingPath, boolUsage),
										UsageText: "rocketpool security propose setting network is-submit-balances-enabled value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "yes, y",
												Usage: "Automatically confirm all interactive questions",
											},
										},
										Action: func(c *cli.Context) error {

											// Validate args
											if err := cliutils.ValidateArgCount(c, 1); err != nil {
												return err
											}
											value, err := cliutils.ValidateBool("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNetworkIsSubmitBalancesEnabled(c, value)

										},
									},

									{
										Name:      "is-submit-prices-enabled",
										Aliases:   []string{"ispe"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.SubmitPricesEnabledSettingPath, boolUsage),
										UsageText: "rocketpool security propose setting network is-submit-prices-enabled value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "yes, y",
												Usage: "Automatically confirm all interactive questions",
											},
										},
										Action: func(c *cli.Context) error {

											// Validate args
											if err := cliutils.ValidateArgCount(c, 1); err != nil {
												return err
											}
											value, err := cliutils.ValidateBool("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNetworkIsSubmitPricesEnabled(c, value)

										},
									},

									{
										Name:      "is-submit-rewards-enabled",
										Aliases:   []string{"isre"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.SubmitRewardsEnabledSettingPath, boolUsage),
										UsageText: "rocketpool security propose setting network is-submit-rewards-enabled value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "yes, y",
												Usage: "Automatically confirm all interactive questions",
											},
										},
										Action: func(c *cli.Context) error {

											// Validate args
											if err := cliutils.ValidateArgCount(c, 1); err != nil {
												return err
											}
											value, err := cliutils.ValidateBool("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNetworkIsSubmitRewardsEnabled(c, value)

										},
									},

									{
										Name:      "node-commission-share-council-adder",
										Aliases:   []string{"ncsca"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.NodeComissionShareSecurityCouncilAdder, boolUsage),
										UsageText: "rocketpool security propose setting network node-commission-share-council-adder",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "yes, y",
												Usage: "Automatically confirm all interactive questions",
											},
										},
										Action: func(c *cli.Context) error {

											// Validate args
											if err := cliutils.ValidateArgCount(c, 1); err != nil {
												return err
											}
											value, err := cliutils.ValidateBigInt("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNodeComissionShareSecurityCouncilAdder(c, value)

										},
									},
								},
							},

							{
								Name:    "node",
								Aliases: []string{"no"},
								Usage:   "Node settings",
								Subcommands: []cli.Command{

									{
										Name:      "is-registration-enabled",
										Aliases:   []string{"ire"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.NodeRegistrationEnabledSettingPath, boolUsage),
										UsageText: "rocketpool security propose setting node is-registration-enabled value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "yes, y",
												Usage: "Automatically confirm all interactive questions",
											},
										},
										Action: func(c *cli.Context) error {

											// Validate args
											if err := cliutils.ValidateArgCount(c, 1); err != nil {
												return err
											}
											value, err := cliutils.ValidateBool("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNodeIsRegistrationEnabled(c, value)

										},
									},

									{
										Name:      "is-smoothing-pool-registration-enabled",
										Aliases:   []string{"ispre"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.SmoothingPoolRegistrationEnabledSettingPath, boolUsage),
										UsageText: "rocketpool security propose setting node is-smoothing-pool-registration-enabled value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "yes, y",
												Usage: "Automatically confirm all interactive questions",
											},
										},
										Action: func(c *cli.Context) error {

											// Validate args
											if err := cliutils.ValidateArgCount(c, 1); err != nil {
												return err
											}
											value, err := cliutils.ValidateBool("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNodeIsSmoothingPoolRegistrationEnabled(c, value)

										},
									},

									{
										Name:      "is-depositing-enabled",
										Aliases:   []string{"ide"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.NodeDepositEnabledSettingPath, boolUsage),
										UsageText: "rocketpool security propose setting node is-depositing-enabled value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "yes, y",
												Usage: "Automatically confirm all interactive questions",
											},
										},
										Action: func(c *cli.Context) error {

											// Validate args
											if err := cliutils.ValidateArgCount(c, 1); err != nil {
												return err
											}
											value, err := cliutils.ValidateBool("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNodeIsDepositingEnabled(c, value)

										},
									},

									{
										Name:      "are-vacant-minipools-enabled",
										Aliases:   []string{"avme"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.VacantMinipoolsEnabledSettingPath, boolUsage),
										UsageText: "rocketpool security propose setting node are-vacant-minipools-enabled value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "yes, y",
												Usage: "Automatically confirm all interactive questions",
											},
										},
										Action: func(c *cli.Context) error {

											// Validate args
											if err := cliutils.ValidateArgCount(c, 1); err != nil {
												return err
											}
											value, err := cliutils.ValidateBool("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNodeAreVacantMinipoolsEnabled(c, value)

										},
									},
								},
							},
						},
					},
				},
			},

			{
				Name:    "proposals",
				Aliases: []string{"o"},
				Usage:   "Manage security council proposals",
				Subcommands: []cli.Command{

					{
						Name:      "list",
						Aliases:   []string{"l"},
						Usage:     "List the security council proposals",
						UsageText: "rocketpool security proposals list",
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  "states, s",
								Usage: "Comma separated list of states to filter ('pending', 'active', 'succeeded', 'executed', 'cancelled', 'defeated', or 'expired')",
								Value: "",
							},
						},
						Action: func(c *cli.Context) error {

							// Validate args
							if err := cliutils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Run
							return getProposals(c, c.String("states"))

						},
					},

					{
						Name:      "details",
						Aliases:   []string{"d"},
						Usage:     "View proposal details",
						UsageText: "rocketpool security proposals details proposal-id",
						Action: func(c *cli.Context) error {

							// Validate args
							var err error
							if err = cliutils.ValidateArgCount(c, 1); err != nil {
								return err
							}
							id, err := cliutils.ValidateUint("proposal-id", c.Args().Get(0))
							if err != nil {
								return err
							}

							// Run
							return getProposal(c, id)

						},
					},

					{
						Name:      "cancel",
						Aliases:   []string{"c"},
						Usage:     "Cancel a proposal made by the node",
						UsageText: "rocketpool security proposals cancel",
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  "proposal, p",
								Usage: "The ID of the proposal to cancel",
							},
						},
						Action: func(c *cli.Context) error {

							// Validate args
							if err := cliutils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Validate flags
							if c.String("proposal") != "" {
								if _, err := cliutils.ValidatePositiveUint("proposal ID", c.String("proposal")); err != nil {
									return err
								}
							}

							// Run
							return cancelProposal(c)

						},
					},

					{
						Name:      "vote",
						Aliases:   []string{"v"},
						Usage:     "Vote on a proposal",
						UsageText: "rocketpool security proposals vote",
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  "proposal, p",
								Usage: "The ID of the proposal to vote on",
							},
							cli.StringFlag{
								Name:  "support, s",
								Usage: "Whether to support the proposal ('yes' or 'no')",
							},
							cli.BoolFlag{
								Name:  "yes, y",
								Usage: "Automatically confirm vote",
							},
						},
						Action: func(c *cli.Context) error {

							// Validate args
							if err := cliutils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Validate flags
							if c.String("proposal") != "" {
								if _, err := cliutils.ValidatePositiveUint("proposal ID", c.String("proposal")); err != nil {
									return err
								}
							}
							if c.String("support") != "" {
								if _, err := cliutils.ValidateBool("support", c.String("support")); err != nil {
									return err
								}
							}

							// Run
							return voteOnProposal(c)

						},
					},

					{
						Name:      "execute",
						Aliases:   []string{"x"},
						Usage:     "Execute a proposal",
						UsageText: "rocketpool security proposals execute",
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  "proposal, p",
								Usage: "The ID of the proposal to execute (or 'all')",
							},
						},
						Action: func(c *cli.Context) error {

							// Validate args
							if err := cliutils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Validate flags
							if c.String("proposal") != "" && c.String("proposal") != "all" {
								if _, err := cliutils.ValidatePositiveUint("proposal ID", c.String("proposal")); err != nil {
									return err
								}
							}

							// Run
							return executeProposal(c)

						},
					},
				},
			},

			{
				Name:      "join",
				Aliases:   []string{"j"},
				Usage:     "Join the security council (requires an executed invite proposal)",
				UsageText: "rocketpool security join",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm joining",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return join(c)

				},
			},

			{
				Name:      "leave",
				Aliases:   []string{"l"},
				Usage:     "Leave the security council (requires an executed leave proposal)",
				UsageText: "rocketpool security leave",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm leaving",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return leave(c)

				},
			},
		},
	})
}
