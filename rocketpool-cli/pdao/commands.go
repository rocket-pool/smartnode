package pdao

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

const (
	boolUsage             string = "specify 'true', 'false', 'yes', or 'no'"
	floatEthUsage         string = "specify an amount of ETH (e.g., '16.0')"
	floatRplUsage         string = "specify an amount of RPL (e.g., '16.0')"
	blockCountUsage       string = "specify a number, in blocks (e.g., '40000')"
	percentUsage          string = "specify a percentage between 0 and 1 (e.g., '0.51' for 51%)"
	unboundedPercentUsage string = "specify a percentage that can go over 100% (e.g., '1.5' for 150%)"
	uintUsage             string = "specify an integer (e.g., '50')"
	durationUsage         string = "specify a duration using hours, minutes, and seconds (e.g., '20m' or '72h0m0s')"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the Rocket Pool Protocol DAO",
		Subcommands: []cli.Command{

			{
				Name:      "settings",
				Aliases:   []string{"s"},
				Usage:     "Show all of the current Protocol DAO settings and values",
				UsageText: "rocketpool pdao settings",
				Action: func(c *cli.Context) error {

					// Run
					return getSettings(c)

				},
			},

			{
				Name:      "rewards-percentages",
				Aliases:   []string{"rp"},
				Usage:     "View the RPL rewards allocation percentages for node operators, the Oracle DAO, and the Protocol DAO",
				UsageText: "rocketpool pdao rewards-percentages",
				Action: func(c *cli.Context) error {

					// Run
					return getRewardsPercentages(c)

				},
			},
			{
				Name:      "initialize-voting",
				Aliases:   []string{"iv"},
				Usage:     "Unlocks a node operator's voting power (only required for node operators who registered before governance structure was in place)",
				UsageText: "rocketpool network initialize-voting",
				Action: func(c *cli.Context) error {

					// Run
					return initializeVoting(c)

				},
			},

			{
				Name:      "set-voting-delegate",
				Aliases:   []string{"svd"},
				Usage:     "Set the address you want to use when voting on Rocket Pool on-chain governance proposals, or the address you want to delegate your voting power to.",
				UsageText: "rocketpool network set-voting-delegate address",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm delegate setting",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					delegate := c.Args().Get(0)
					// Run
					return networkSetVotingDelegate(c, delegate)

				},
			},

			{
				Name:      "claim-bonds",
				Aliases:   []string{"cb"},
				Usage:     "Unlock any bonded RPL you have for a proposal or set of challenges, and claim any bond rewards for defending or defeating the proposal",
				UsageText: "rocketpool pdao proposals claim-bonds proposal-id indices",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "proposal, p",
						Usage: "The ID of the proposal to claim bonds from (or 'all')",
					},
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
					return claimBonds(c)

				},
			},

			{
				Name:    "propose",
				Aliases: []string{"p"},
				Usage:   "Make a Protocol DAO proposal",
				Subcommands: []cli.Command{

					{
						Name:      "rewards-percentages",
						Aliases:   []string{"rp"},
						Usage:     "Propose updating the RPL rewards allocation percentages for node operators, the Oracle DAO, and the Protocol DAO",
						UsageText: "rocketpool pdao propose rewards-percentages",
						Flags: []cli.Flag{
							cli.BoolFlag{
								Name:  "raw",
								Usage: "Add this flag if you want to use 18-decimal-fixed-point-integer (wei) values instead of floating point percentages",
							},
							cli.BoolFlag{
								Name:  "yes, y",
								Usage: "Automatically confirm all interactive questions",
							},
							cli.StringFlag{
								Name:  "node, n",
								Usage: "The node operator's rewards allocation (a percentage from 0 to 1 if '--raw' is not set)",
							},
							cli.StringFlag{
								Name:  "odao, o",
								Usage: "The Oracle DAO's rewards allocation (a percentage from 0 to 1 if '--raw' is not set)",
							},
							cli.StringFlag{
								Name:  "pdao, p",
								Usage: "The Protocol DAO's rewards allocation (a percentage from 0 to 1 if '--raw' is not set)",
							},
						},
						Action: func(c *cli.Context) error {

							// Validate args
							if err := cliutils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Run
							return proposeRewardsPercentages(c)

						},
					},

					{
						Name:      "one-time-spend",
						Aliases:   []string{"ots"},
						Usage:     "Propose a one-time spend of the Protocol DAO's treasury",
						UsageText: "rocketpool pdao propose one-time-spend",
						Flags: []cli.Flag{
							cli.BoolFlag{
								Name:  "raw",
								Usage: "Add this flag if you want to use 18-decimal-fixed-point-integer (wei) values instead of floating point ETH amounts",
							},
							cli.BoolFlag{
								Name:  "yes, y",
								Usage: "Automatically confirm all interactive questions",
							},
							cli.StringFlag{
								Name:  "invoice-id, i",
								Usage: "The invoice ID / number for this spend",
							},
							cli.StringFlag{
								Name:  "recipient, r",
								Usage: "The recipient of the spend",
							},
							cli.StringFlag{
								Name:  "amount, a",
								Usage: "The amount of RPL to send",
							},
						},
						Action: func(c *cli.Context) error {

							// Validate args
							if err := cliutils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Run
							return proposeOneTimeSpend(c)

						},
					},

					{
						Name:      "recurring-spend",
						Aliases:   []string{"rs"},
						Usage:     "Propose a recurring spend of the Protocol DAO's treasury",
						UsageText: "rocketpool pdao propose recurring-spend",
						Flags: []cli.Flag{
							cli.BoolFlag{
								Name:  "raw",
								Usage: "Add this flag if you want to use 18-decimal-fixed-point-integer (wei) values instead of floating point ETH amounts",
							},
							cli.BoolFlag{
								Name:  "yes, y",
								Usage: "Automatically confirm all interactive questions",
							},
							cli.StringFlag{
								Name:  "contract-name, c",
								Usage: "The name of the recurring spend's contract / invoice (alternatively, the name of the recipient)",
							},
							cli.StringFlag{
								Name:  "recipient, r",
								Usage: "The recipient of the spend",
							},
							cli.StringFlag{
								Name:  "amount-per-period, a",
								Usage: "The amount of RPL to send",
							},
							cli.Uint64Flag{
								Name:  "start-time, s",
								Usage: "The start time of the first payment period (Unix timestamp)",
							},
							cli.StringFlag{
								Name:  "period-length, l",
								Usage: "The length of time between each payment, in hours / minutes / seconds (e.g., 168h0m0s)",
							},
							cli.Uint64Flag{
								Name:  "number-of-periods, n",
								Usage: "The total number of payment periods for the spend",
							},
						},
						Action: func(c *cli.Context) error {

							// Validate args
							if err := cliutils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Run
							return proposeRecurringSpend(c)

						},
					},

					{
						Name:      "recurring-spend-update",
						Aliases:   []string{"rsu"},
						Usage:     "Propose an update to an existing recurring spend plan",
						UsageText: "rocketpool pdao propose recurring-spend-update",
						Flags: []cli.Flag{
							cli.BoolFlag{
								Name:  "raw",
								Usage: "Add this flag if you want to use 18-decimal-fixed-point-integer (wei) values instead of floating point ETH amounts",
							},
							cli.BoolFlag{
								Name:  "yes, y",
								Usage: "Automatically confirm all interactive questions",
							},
							cli.StringFlag{
								Name:  "contract-name, c",
								Usage: "The name of the recurring spend's contract / invoice (alternatively, the name of the recipient)",
							},
							cli.StringFlag{
								Name:  "recipient, r",
								Usage: "The recipient of the spend",
							},
							cli.StringFlag{
								Name:  "amount-per-period, a",
								Usage: "The amount of RPL to send",
							},
							cli.StringFlag{
								Name:  "period-length, l",
								Usage: "The length of time between each payment, in hours / minutes / seconds (e.g., 168h0m0s)",
							},
							cli.Uint64Flag{
								Name:  "number-of-periods, n",
								Usage: "The total number of payment periods for the spend",
							},
						},
						Action: func(c *cli.Context) error {

							// Validate args
							if err := cliutils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Run
							return proposeRecurringSpendUpdate(c)

						},
					},

					{
						Name:    "security-council",
						Aliases: []string{"sc"},
						Usage:   "Modify the security council",
						Subcommands: []cli.Command{

							{
								Name:      "invite",
								Aliases:   []string{"i"},
								Usage:     "Propose an invitation to the security council",
								UsageText: "rocketpool pdao propose security-council invite",
								Flags: []cli.Flag{
									cli.BoolFlag{
										Name:  "yes, y",
										Usage: "Automatically confirm all interactive questions",
									},
									cli.StringFlag{
										Name:  "id, i",
										Usage: "A descriptive ID of the entity being invited",
									},
									cli.StringFlag{
										Name:  "address, a",
										Usage: "The address of the entity being invited",
									},
								},
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 0); err != nil {
										return err
									}

									// Run
									return proposeSecurityCouncilInvite(c)

								},
							},

							{
								Name:      "kick",
								Aliases:   []string{"k"},
								Usage:     "Propose kick a member from the security council",
								UsageText: "rocketpool pdao propose security-council kick",
								Flags: []cli.Flag{
									cli.BoolFlag{
										Name:  "yes, y",
										Usage: "Automatically confirm all interactive questions",
									},
									cli.StringFlag{
										Name:  "addresses, a",
										Usage: "One or more addresses of the entity(s) to kick, separated by commas",
									},
								},
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 0); err != nil {
										return err
									}

									// Run
									return proposeSecurityCouncilKick(c)

								},
							},

							{
								Name:      "replace",
								Aliases:   []string{"r"},
								Usage:     "Propose replacing an existing member of the security council with a new member",
								UsageText: "rocketpool pdao propose security-council replace",
								Flags: []cli.Flag{
									cli.BoolFlag{
										Name:  "yes, y",
										Usage: "Automatically confirm all interactive questions",
									},
									cli.StringFlag{
										Name:  "existing-address, e",
										Usage: "The address of the existing member",
									},
									cli.StringFlag{
										Name:  "new-id, ni",
										Usage: "A descriptive ID of the new entity to invite",
									},
									cli.StringFlag{
										Name:  "new-address, na",
										Usage: "The address of the new entity to invite",
									},
								},
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 0); err != nil {
										return err
									}

									// Run
									return proposeSecurityCouncilReplace(c)

								},
							},
						},
					},

					{
						Name:    "setting",
						Aliases: []string{"s"},
						Usage:   "Make a Protocol DAO setting proposal",
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
										UsageText: "rocketpool pdao propose setting auction is-create-lot-enabled value",
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
										UsageText: "rocketpool pdao propose setting auction is-bid-on-lot-enabled value",
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

									{
										Name:      "lot-minimum-eth-value",
										Aliases:   []string{"lminev"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.LotMinimumEthValueSettingPath, floatEthUsage),
										UsageText: "rocketpool pdao propose setting auction lot-minimum-eth-value value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), false)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingAuctionLotMinimumEthValue(c, value)

										},
									},

									{
										Name:      "lot-maximum-eth-value",
										Aliases:   []string{"lmaxev"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.LotMaximumEthValueSettingPath, floatEthUsage),
										UsageText: "rocketpool pdao propose setting auction lot-maximum-eth-value value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), false)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingAuctionLotMaximumEthValue(c, value)

										},
									},

									{
										Name:      "lot-duration",
										Aliases:   []string{"ld"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.LotDurationSettingPath, durationUsage),
										UsageText: "rocketpool pdao propose setting auction lot-duration value",
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
											value, err := cliutils.ValidateDuration("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingAuctionLotDuration(c, value)

										},
									},

									{
										Name:      "lot-starting-price-ratio",
										Aliases:   []string{"lspr"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.LotStartingPriceRatioSettingPath, percentUsage),
										UsageText: "rocketpool pdao propose setting auction lot-starting-price-ratio value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), true)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingAuctionLotStartingPriceRatio(c, value)

										},
									},

									{
										Name:      "lot-reserve-price-ratio",
										Aliases:   []string{"lrpr"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.LotReservePriceRatioSettingPath, percentUsage),
										UsageText: "rocketpool pdao propose setting auction lot-reserve-price-ratio value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), true)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingAuctionLotReservePriceRatio(c, value)

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
										UsageText: "rocketpool pdao propose setting deposit is-depositing-enabled value",
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
										UsageText: "rocketpool pdao propose setting deposit are-deposit-assignments-enabled value",
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

									{
										Name:      "minimum-deposit",
										Aliases:   []string{"md"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.MinimumDepositSettingPath, floatEthUsage),
										UsageText: "rocketpool pdao propose setting deposit minimum-deposit value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), false)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingDepositMinimumDeposit(c, value)

										},
									},

									{
										Name:      "maximum-deposit-pool-size",
										Aliases:   []string{"mdps"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.MaximumDepositPoolSizeSettingPath, floatEthUsage),
										UsageText: "rocketpool pdao propose setting deposit maximum-deposit-pool-size value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), false)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingDepositMaximumDepositPoolSize(c, value)

										},
									},

									{
										Name:      "maximum-assignments-per-deposit",
										Aliases:   []string{"mapd"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.MaximumDepositAssignmentsSettingPath, uintUsage),
										UsageText: "rocketpool pdao propose setting deposit maximum-assignments-per-deposit value",
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
											value, err := cliutils.ValidatePositiveUint("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingDepositMaximumAssignmentsPerDeposit(c, value)

										},
									},

									{
										Name:      "maximum-socialised-assignments-per-deposit",
										Aliases:   []string{"msapd"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.MaximumSocializedDepositAssignmentsSettingPath, uintUsage),
										UsageText: "rocketpool pdao propose setting deposit maximum-socialised-assignments-per-deposit value",
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
											value, err := cliutils.ValidatePositiveUint("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingDepositMaximumSocialisedAssignmentsPerDeposit(c, value)

										},
									},

									{
										Name:      "deposit-fee",
										Aliases:   []string{"df"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.DepositFeeSettingPath, "specify a percentage between 0 and 0.01 (e.g., '0.001' for 0.10%)"),
										UsageText: "rocketpool pdao propose setting deposit deposit-fee value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), true)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingDepositDepositFee(c, value)

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
										UsageText: "rocketpool pdao propose setting minipool is-submit-withdrawable-enabled value",
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
										Name:      "launch-timeout",
										Aliases:   []string{"lt"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.MinipoolLaunchTimeoutSettingPath, durationUsage),
										UsageText: "rocketpool pdao propose setting minipool launch-timeout value",
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
											value, err := cliutils.ValidateDuration("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingMinipoolLaunchTimeout(c, value)

										},
									},

									{
										Name:      "is-bond-reduction-enabled",
										Aliases:   []string{"ibre"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.BondReductionEnabledSettingPath, boolUsage),
										UsageText: "rocketpool pdao propose setting minipool is-bond-reduction-enabled value",
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

									{
										Name:      "max-count",
										Aliases:   []string{"mc"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.MaximumMinipoolCountSettingPath, uintUsage),
										UsageText: "rocketpool pdao propose setting minipool max-count value",
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
											value, err := cliutils.ValidatePositiveUint("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingMinipoolMaximumCount(c, value)

										},
									},

									{
										Name:      "user-distribute-window-start",
										Aliases:   []string{"udws"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.MinipoolUserDistributeWindowStartSettingPath, durationUsage),
										UsageText: "rocketpool pdao propose setting minipool user-distribute-window-start value",
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
											value, err := cliutils.ValidateDuration("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingMinipoolUserDistributeWindowStart(c, value)

										},
									},

									{
										Name:      "user-distribute-window-length",
										Aliases:   []string{"udwl"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.MinipoolUserDistributeWindowLengthSettingPath, durationUsage),
										UsageText: "rocketpool pdao propose setting minipool user-distribute-window-length value",
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
											value, err := cliutils.ValidateDuration("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingMinipoolUserDistributeWindowLength(c, value)

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
										Name:      "oracle-dao-consensus-threshold",
										Aliases:   []string{"odct"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.NodeConsensusThresholdSettingPath, percentUsage),
										UsageText: "rocketpool pdao propose setting network oracle-dao-consensus-threshold value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), true)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNetworkOracleDaoConsensusThreshold(c, value)

										},
									},

									{
										Name:      "node-penalty-threshold",
										Aliases:   []string{"npt"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.NetworkPenaltyThresholdSettingPath, percentUsage),
										UsageText: "rocketpool pdao propose setting network node-penalty-threshold value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), true)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNetworkNodePenaltyThreshold(c, value)

										},
									},

									{
										Name:      "per-penalty-rate",
										Aliases:   []string{"ppr"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.NetworkPenaltyPerRateSettingPath, percentUsage),
										UsageText: "rocketpool pdao propose setting network per-penalty-rate value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), true)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNetworkPerPenaltyRate(c, value)

										},
									},

									{
										Name:      "is-submit-balances-enabled",
										Aliases:   []string{"isbe"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.SubmitBalancesEnabledSettingPath, boolUsage),
										UsageText: "rocketpool pdao propose setting network is-submit-balances-enabled value",
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
										Name:      "submit-balances-frequency",
										Aliases:   []string{"sbf"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.SubmitBalancesFrequencySettingPath, durationUsage),
										UsageText: "rocketpool pdao propose setting network submit-balances-frequency value",
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
											value, err := cliutils.ValidateDuration("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNetworkSubmitBalancesFrequency(c, value)

										},
									},

									{
										Name:      "is-submit-prices-enabled",
										Aliases:   []string{"ispe"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.SubmitPricesEnabledSettingPath, boolUsage),
										UsageText: "rocketpool pdao propose setting network is-submit-prices-enabled value",
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
										Name:      "submit-prices-frequency",
										Aliases:   []string{"spf"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.SubmitPricesFrequencySettingPath, durationUsage),
										UsageText: "rocketpool pdao propose setting network submit-prices-frequency value",
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
											value, err := cliutils.ValidateDuration("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNetworkSubmitPricesFrequency(c, value)

										},
									},

									{
										Name:      "minimum-node-fee",
										Aliases:   []string{"minnf"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.MinimumNodeFeeSettingPath, percentUsage),
										UsageText: "rocketpool pdao propose setting network minimum-node-fee value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), true)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNetworkMinimumNodeFee(c, value)

										},
									},

									{
										Name:      "target-node-fee",
										Aliases:   []string{"tnf"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.TargetNodeFeeSettingPath, percentUsage),
										UsageText: "rocketpool pdao propose setting network target-node-fee value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), true)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNetworkTargetNodeFee(c, value)

										},
									},

									{
										Name:      "maximum-node-fee",
										Aliases:   []string{"maxnf"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.MaximumNodeFeeSettingPath, percentUsage),
										UsageText: "rocketpool pdao propose setting network maximum-node-fee value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), true)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNetworkMaximumNodeFee(c, value)

										},
									},

									{
										Name:      "node-fee-demand-range",
										Aliases:   []string{"nfdr"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.NodeFeeDemandRangeSettingPath, floatEthUsage),
										UsageText: "rocketpool pdao propose setting network node-fee-demand-range value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), false)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNetworkNodeFeeDemandRange(c, value)

										},
									},

									{
										Name:      "target-reth-collateral-rate",
										Aliases:   []string{"trcr"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.TargetRethCollateralRateSettingPath, percentUsage),
										UsageText: "rocketpool pdao propose setting network target-reth-collateral-rate value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), true)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNetworkTargetRethCollateralRate(c, value)

										},
									},

									{
										Name:      "is-submit-rewards-enabled",
										Aliases:   []string{"isre"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.SubmitRewardsEnabledSettingPath, boolUsage),
										UsageText: "rocketpool pdao propose setting network is-submit-rewards-enabled value",
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
										UsageText: "rocketpool pdao propose setting node is-registration-enabled value",
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
										UsageText: "rocketpool pdao propose setting node is-smoothing-pool-registration-enabled value",
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
										UsageText: "rocketpool pdao propose setting node is-depositing-enabled value",
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
										UsageText: "rocketpool pdao propose setting node are-vacant-minipools-enabled value",
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

									{
										Name:      "minimum-per-minipool-stake",
										Aliases:   []string{"minpms"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.MinimumPerMinipoolStakeSettingPath, unboundedPercentUsage),
										UsageText: "rocketpool pdao propose setting node minimum-per-minipool-stake value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), false)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNodeMinimumPerMinipoolStake(c, value)

										},
									},

									{
										Name:      "maximum-per-minipool-stake",
										Aliases:   []string{"maxpms"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.MaximumPerMinipoolStakeSettingPath, unboundedPercentUsage),
										UsageText: "rocketpool pdao propose setting node maximum-per-minipool-stake value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), false)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingNodeMaximumPerMinipoolStake(c, value)

										},
									},
								},
							},

							{
								Name:    "proposals",
								Aliases: []string{"p"},
								Usage:   "Proposal settings",
								Subcommands: []cli.Command{

									{
										Name:      "vote-phase1-time",
										Aliases:   []string{"vt1"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.VotePhase1TimeSettingPath, durationUsage),
										UsageText: "rocketpool pdao propose setting proposals vote-phase1-time value",
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
											value, err := cliutils.ValidateDuration("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingProposalsVotePhase1Time(c, value)

										},
									},

									{
										Name:      "vote-phase2-time",
										Aliases:   []string{"vt2"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.VotePhase2TimeSettingPath, durationUsage),
										UsageText: "rocketpool pdao propose setting proposals vote-phase2-time value",
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
											value, err := cliutils.ValidateDuration("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingProposalsVotePhase2Time(c, value)

										},
									},

									{
										Name:      "vote-delay-time",
										Aliases:   []string{"vdt"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.VoteDelayTimeSettingPath, durationUsage),
										UsageText: "rocketpool pdao propose setting proposals vote-delay-time value",
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
											value, err := cliutils.ValidateDuration("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingProposalsVoteDelayTime(c, value)

										},
									},

									{
										Name:      "execute-time",
										Aliases:   []string{"et"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.ExecuteTimeSettingPath, durationUsage),
										UsageText: "rocketpool pdao propose setting proposals execute-time value",
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
											value, err := cliutils.ValidateDuration("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingProposalsExecuteTime(c, value)

										},
									},

									{
										Name:      "proposal-bond",
										Aliases:   []string{"pb"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.ProposalBondSettingPath, floatRplUsage),
										UsageText: "rocketpool pdao propose setting proposals proposal-bond value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), false)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingProposalsProposalBond(c, value)

										},
									},

									{
										Name:      "challenge-bond",
										Aliases:   []string{"cb"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.ChallengeBondSettingPath, floatRplUsage),
										UsageText: "rocketpool pdao propose setting proposals challenge-bond value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), false)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingProposalsChallengeBond(c, value)

										},
									},

									{
										Name:      "challenge-period",
										Aliases:   []string{"cp"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.ChallengePeriodSettingPath, durationUsage),
										UsageText: "rocketpool pdao propose setting proposals challenge-period value",
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
											value, err := cliutils.ValidateDuration("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingProposalsChallengePeriod(c, value)

										},
									},

									{
										Name:      "quorum",
										Aliases:   []string{"q"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.ProposalQuorumSettingPath, percentUsage),
										UsageText: "rocketpool pdao propose setting proposals quorum value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), true)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingProposalsQuorum(c, value)

										},
									},

									{
										Name:      "veto-quorum",
										Aliases:   []string{"vq"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.ProposalVetoQuorumSettingPath, percentUsage),
										UsageText: "rocketpool pdao propose setting proposals veto-quorum value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), true)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingProposalsVetoQuorum(c, value)

										},
									},

									{
										Name:      "max-block-age",
										Aliases:   []string{"mba"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.ProposalMaxBlockAgeSettingPath, blockCountUsage),
										UsageText: "rocketpool pdao propose setting proposals max-block-age value",
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
											value, err := cliutils.ValidatePositiveUint("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingProposalsMaxBlockAge(c, value)

										},
									},
								},
							},

							{
								Name:    "rewards",
								Aliases: []string{"r"},
								Usage:   "Rewards settings",
								Subcommands: []cli.Command{

									{
										Name:      "interval-periods",
										Aliases:   []string{"ip"},
										Usage:     fmt.Sprintf("Propose updating the %s setting - the rewards interval will consist of this number of price/balances submission periods; %s", protocol.RewardsClaimIntervalPeriodsSettingPath, uintUsage),
										UsageText: "rocketpool pdao propose setting rewards interval-periods value",
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
											value, err := cliutils.ValidatePositiveUint("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingRewardsIntervalPeriods(c, value)

										},
									},
								},
							},

							{
								Name:    "security",
								Aliases: []string{"s"},
								Usage:   "Security council settings",
								Subcommands: []cli.Command{

									{
										Name:      "members-quorum",
										Aliases:   []string{"mq"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.SecurityMembersQuorumSettingPath, percentUsage),
										UsageText: "rocketpool pdao propose setting security members-quorum value",
										Flags: []cli.Flag{
											cli.BoolFlag{
												Name:  "raw",
												Usage: "Add this flag if your setting is an 18-decimal-fixed-point-integer (wei) value instead of a float",
											},
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
											value, err := parseFloat(c, "value", c.Args().Get(0), true)
											if err != nil {
												return err
											}

											// Run
											return proposeSettingSecurityMembersQuorum(c, value)

										},
									},

									{
										Name:      "members-leave-time",
										Aliases:   []string{"mlt"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.SecurityMembersLeaveTimeSettingPath, durationUsage),
										UsageText: "rocketpool pdao propose setting security members-leave-time value",
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
											value, err := cliutils.ValidateDuration("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingSecurityMembersLeaveTime(c, value)

										},
									},

									{
										Name:      "proposal-vote-time",
										Aliases:   []string{"pvt"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.SecurityProposalVoteTimeSettingPath, durationUsage),
										UsageText: "rocketpool pdao propose setting security proposal-vote-phase1-time value",
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
											value, err := cliutils.ValidateDuration("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingSecurityProposalVoteTime(c, value)

										},
									},

									{
										Name:      "proposal-execute-time",
										Aliases:   []string{"pet"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.SecurityProposalExecuteTimeSettingPath, durationUsage),
										UsageText: "rocketpool pdao propose setting security proposal-execute-time value",
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
											value, err := cliutils.ValidateDuration("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingSecurityProposalExecuteTime(c, value)

										},
									},

									{
										Name:      "proposal-action-time",
										Aliases:   []string{"pat"},
										Usage:     fmt.Sprintf("Propose updating the %s setting; %s", protocol.SecurityProposalActionTimeSettingPath, durationUsage),
										UsageText: "rocketpool pdao propose setting security proposal-action-time value",
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
											value, err := cliutils.ValidateDuration("value", c.Args().Get(0))
											if err != nil {
												return err
											}

											// Run
											return proposeSettingSecurityProposalActionTime(c, value)

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
				Usage:   "Manage Protocol DAO proposals",
				Subcommands: []cli.Command{

					{
						Name:      "list",
						Aliases:   []string{"l"},
						Usage:     "List the Protocol DAO proposals",
						UsageText: "rocketpool pdao proposals list",
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
						UsageText: "rocketpool pdao proposals details proposal-id",
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
						Name:      "vote",
						Aliases:   []string{"v"},
						Usage:     "Vote on a proposal",
						UsageText: "rocketpool pdao proposals vote",
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  "proposal, p",
								Usage: "The ID of the proposal to vote on",
							},
							cli.StringFlag{
								Name:  "vote-direction, v",
								Usage: "How to vote ('abstain', 'for', 'against', 'veto')",
							},
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
							return voteOnProposal(c)

						},
					},
					{
						Name:      "execute",
						Aliases:   []string{"x"},
						Usage:     "Execute a proposal",
						UsageText: "rocketpool pdao proposals execute",
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  "proposal, p",
								Usage: "The ID of the proposal to execute (or 'all')",
							},
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

					{
						Name:      "defeat",
						Aliases:   []string{"t"},
						Usage:     "Defeat a proposal that still has a tree index in the 'Challenged' state after its challenge window has passed",
						UsageText: "rocketpool pdao proposals defeat proposal-id challenged-index",
						Flags: []cli.Flag{
							cli.BoolFlag{
								Name:  "yes, y",
								Usage: "Automatically confirm all interactive questions",
							},
						},
						Action: func(c *cli.Context) error {

							// Validate args
							if err := cliutils.ValidateArgCount(c, 2); err != nil {
								return err
							}
							proposalId, err := cliutils.ValidatePositiveUint("proposal-id", c.Args().Get(0))
							if err != nil {
								return err
							}
							index, err := cliutils.ValidatePositiveUint("challenged-index", c.Args().Get(1))
							if err != nil {
								return err
							}

							// Run
							return defeatProposal(c, proposalId, index)

						},
					},

					{
						Name:      "finalize",
						Aliases:   []string{"f"},
						Usage:     "Finalize a proposal that has been vetoed by burning the proposer's locked RPL bond",
						UsageText: "rocketpool pdao proposals finalize proposal-id",
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
							proposalId, err := cliutils.ValidatePositiveUint("proposal-id", c.Args().Get(0))
							if err != nil {
								return err
							}

							// Run
							return finalizeProposal(c, proposalId)

						},
					},
				},
			},
		},
	})
}
