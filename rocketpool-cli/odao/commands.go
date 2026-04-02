package odao

import (
	"context"

	"github.com/urfave/cli/v3"

	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register commands
func RegisterCommands(app *cli.Command, name string, aliases []string) {
	app.Commands = append(app.Commands, &cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the Rocket Pool oracle DAO",
		Commands: []*cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get oracle DAO status",
				UsageText: "rocketpool odao status",
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
				Name:      "members",
				Aliases:   []string{"m"},
				Usage:     "Get the oracle DAO members",
				UsageText: "rocketpool odao members",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getMembers()

				},
			},

			{
				Name:      "member-settings",
				Aliases:   []string{"b"},
				Usage:     "Get the oracle DAO settings related to oracle DAO members",
				UsageText: "rocketpool odao member-settings",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Run
					return getMemberSettings()

				},
			},

			{
				Name:      "proposal-settings",
				Aliases:   []string{"a"},
				Usage:     "Get the oracle DAO settings related to oracle DAO proposals",
				UsageText: "rocketpool odao proposal-settings",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Run
					return getProposalSettings()

				},
			},

			{
				Name:      "minipool-settings",
				Aliases:   []string{"i"},
				Usage:     "Get the oracle DAO settings related to minipools",
				UsageText: "rocketpool odao minipool-settings",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Run
					return getMinipoolSettings()

				},
			},

			{
				Name:      "penalise-megapool",
				Aliases:   []string{"pm"},
				Usage:     "(Saturn) Penalise a megapool",
				UsageText: "rocketpool odao penalise-megapool megapool-address block",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm the action",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}

					megapoolAddress, err := cliutils.ValidateAddress("megapool address", c.Args().Get(0))
					if err != nil {
						return err
					}

					block, err := cliutils.ValidateBigInt("block number", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					return penaliseMegapool(megapoolAddress, block, c.Bool("yes"))

				},
			},

			{
				Name:    "propose",
				Aliases: []string{"p"},
				Usage:   "Make an oracle DAO proposal",
				Commands: []*cli.Command{

					{
						Name:    "member",
						Aliases: []string{"m"},
						Usage:   "Make an oracle DAO member proposal",
						Commands: []*cli.Command{

							{
								Name:      "invite",
								Aliases:   []string{"i"},
								Usage:     "Propose inviting a new member",
								UsageText: "rocketpool odao propose member invite member-address member-id member-url",
								Action: func(ctx context.Context, c *cli.Command) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 3); err != nil {
										return err
									}
									memberAddress, err := cliutils.ValidateAddress("member address", c.Args().Get(0))
									if err != nil {
										return err
									}
									memberId, err := cliutils.ValidateDAOMemberID("member ID", c.Args().Get(1))
									if err != nil {
										return err
									}

									// Run
									return proposeInvite(memberAddress, memberId, c.Args().Get(2), c.Bool("yes"))

								},
							},

							{
								Name:      "leave",
								Aliases:   []string{"l"},
								Usage:     "Propose leaving the oracle DAO",
								UsageText: "rocketpool odao propose member leave",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name:    "yes",
										Aliases: []string{"y"},
										Usage:   "Automatically confirm the action",
									},
								},
								Action: func(ctx context.Context, c *cli.Command) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 0); err != nil {
										return err
									}

									// Run
									return proposeLeave(c.Bool("yes"))

								},
							},

							{
								Name:      "kick",
								Aliases:   []string{"k"},
								Usage:     "Propose kicking a member",
								UsageText: "rocketpool odao propose member kick [options]",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name:    "yes",
										Aliases: []string{"y"},
										Usage:   "Automatically confirm the action",
									},
									&cli.StringFlag{
										Name:    "member",
										Aliases: []string{"m"},
										Usage:   "The address of the member to propose kicking",
									},
									&cli.StringFlag{
										Name:    "fine",
										Aliases: []string{"f"},
										Usage:   "The amount of RPL to fine the member (or 'max')",
									},
								},
								Action: func(ctx context.Context, c *cli.Command) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 0); err != nil {
										return err
									}

									// Validate flags
									if c.String("member") != "" {
										if _, err := cliutils.ValidateAddress("member address", c.String("member")); err != nil {
											return err
										}
									}
									if c.String("fine") != "" && c.String("fine") != "max" {
										if _, err := cliutils.ValidatePositiveEthAmount("fine amount", c.String("fine")); err != nil {
											return err
										}
									}

									// Run
									return proposeKick(c.String("member"), c.String("fine"), c.Bool("yes"))

								},
							},
						},
					},

					{
						Name:    "setting",
						Aliases: []string{"s"},
						Usage:   "Make an oracle DAO setting proposal",
						Commands: []*cli.Command{

							{
								Name:      "members-quorum",
								Aliases:   []string{"q"},
								Usage:     "Propose updating the members.quorum setting - takes a percent, from 0 to 100",
								UsageText: "rocketpool odao propose setting members-quorum value",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name:    "yes",
										Aliases: []string{"y"},
										Usage:   "Automatically confirm the action",
									},
								},
								Action: func(ctx context.Context, c *cli.Command) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}
									quorumPercent, err := cliutils.ValidatePercentage("quorum percentage", c.Args().Get(0))
									if err != nil {
										return err
									}

									// Run
									return proposeSettingMembersQuorum(quorumPercent, c.Bool("yes"))

								},
							},
							{
								Name:      "members-rplbond",
								Aliases:   []string{"b"},
								Usage:     "Propose updating the members.rplbond setting - takes an RPL amount (e.g. 5000)",
								UsageText: "rocketpool odao propose setting members-rplbond value",
								Action: func(ctx context.Context, c *cli.Command) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}
									bondAmountEth, err := cliutils.ValidateEthAmount("RPL bond amount", c.Args().Get(0))
									if err != nil {
										return err
									}

									// Run
									return proposeSettingMembersRplBond(bondAmountEth, c.Bool("yes"))

								},
							},
							{
								Name:      "members-minipool-unbonded-max",
								Aliases:   []string{"u"},
								Usage:     "Propose updating the members.minipool.unbonded.max setting - takes a number (e.g 100)",
								UsageText: "rocketpool odao propose setting members-minipool-unbonded-max value",
								Action: func(ctx context.Context, c *cli.Command) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}
									unbondedMinipoolMax, err := cliutils.ValidateUint("maximum unbonded minipool count", c.Args().Get(0))
									if err != nil {
										return err
									}

									// Run
									return proposeSettingMinipoolUnbondedMax(unbondedMinipoolMax, c.Bool("yes"))

								},
							},
							{
								Name:      "proposal-cooldown",
								Aliases:   []string{"c"},
								Usage:     "Propose updating the proposal.cooldown.time setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting proposal-cooldown value",
								Action: func(ctx context.Context, c *cli.Command) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingProposalCooldown(c.Args().Get(0), c.Bool("yes"))

								},
							},
							{
								Name:      "proposal-vote-timespan",
								Aliases:   []string{"v"},
								Usage:     "Propose updating the proposal.vote.time setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting proposal-vote-timespan value",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name:    "yes",
										Aliases: []string{"y"},
										Usage:   "Automatically confirm the action",
									},
								},
								Action: func(ctx context.Context, c *cli.Command) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingProposalVoteTimespan(c.Args().Get(0), c.Bool("yes"))

								},
							},
							{
								Name:      "proposal-vote-delay-timespan",
								Aliases:   []string{"d"},
								Usage:     "Propose updating the proposal.vote.delay.time setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting proposal-vote-delay-timespan value",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name:    "yes",
										Aliases: []string{"y"},
										Usage:   "Automatically confirm the action",
									},
								},
								Action: func(ctx context.Context, c *cli.Command) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingProposalVoteDelayTimespan(c.Args().Get(0), c.Bool("yes"))

								},
							},
							{
								Name:      "proposal-execute-timespan",
								Aliases:   []string{"x"},
								Usage:     "Propose updating the proposal.execute.time setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting proposal-execute-timespan value",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name:    "yes",
										Aliases: []string{"y"},
										Usage:   "Automatically confirm the action",
									},
								},
								Action: func(ctx context.Context, c *cli.Command) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingProposalExecuteTimespan(c.Args().Get(0), c.Bool("yes"))

								},
							},
							{
								Name:      "proposal-action-timespan",
								Aliases:   []string{"a"},
								Usage:     "Propose updating the proposal.action.time setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting proposal-action-timespan value",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name:    "yes",
										Aliases: []string{"y"},
										Usage:   "Automatically confirm the action",
									},
								},
								Action: func(ctx context.Context, c *cli.Command) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingProposalActionTimespan(c.Args().Get(0), c.Bool("yes"))

								},
							},
							{
								Name:      "scrub-period",
								Aliases:   []string{"s"},
								Usage:     "Propose updating the minipool.scrub.period setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting scrub-period value",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name:    "yes",
										Aliases: []string{"y"},
										Usage:   "Automatically confirm the action",
									},
								},
								Action: func(ctx context.Context, c *cli.Command) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingScrubPeriod(c.Args().Get(0), c.Bool("yes"))

								},
							},
							{
								Name:      "promotion-scrub-period",
								Aliases:   []string{"p"},
								Usage:     "Propose updating the minipool.promotion.scrub.period setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting promotion-scrub-period value",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name:    "yes",
										Aliases: []string{"y"},
										Usage:   "Automatically confirm the action",
									},
								},
								Action: func(ctx context.Context, c *cli.Command) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingPromotionScrubPeriod(c.Args().Get(0), c.Bool("yes"))

								},
							},
							{
								Name:      "scrub-penalty-enabled",
								Aliases:   []string{"spe"},
								Usage:     "Propose updating the minipool.scrub.penalty.enabled setting - format is true / false",
								UsageText: "rocketpool odao propose setting scrub-penalty-enabled value",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name:    "yes",
										Aliases: []string{"y"},
										Usage:   "Automatically confirm the action",
									},
								},
								Action: func(ctx context.Context, c *cli.Command) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}
									enabled, err := cliutils.ValidateBool("enabled", c.Args().Get(0))
									if err != nil {
										return err
									}

									// Run
									return proposeSettingScrubPenaltyEnabled(enabled, c.Bool("yes"))

								},
							},
							{
								Name:      "bond-reduction-window-start",
								Aliases:   []string{"brws"},
								Usage:     "Propose updating the minipool.bond.reduction.window.start setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting bond-reduction-window-start value",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name:    "yes",
										Aliases: []string{"y"},
										Usage:   "Automatically confirm the action",
									},
								},
								Action: func(ctx context.Context, c *cli.Command) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingBondReductionWindowStart(c.Args().Get(0), c.Bool("yes"))

								},
							},
							{
								Name:      "bond-reduction-window-length",
								Aliases:   []string{"brwl"},
								Usage:     "Propose updating the minipool.bond.reduction.window.length setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting bond-reduction-window-length value",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name:    "yes",
										Aliases: []string{"y"},
										Usage:   "Automatically confirm the action",
									},
								},
								Action: func(ctx context.Context, c *cli.Command) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingBondReductionWindowLength(c.Args().Get(0), c.Bool("yes"))

								},
							},
						},
					},
				},
			},

			{
				Name:    "proposals",
				Aliases: []string{"o"},
				Usage:   "Manage oracle DAO proposals",
				Commands: []*cli.Command{

					{
						Name:      "list",
						Aliases:   []string{"l"},
						Usage:     "List the oracle DAO proposals",
						UsageText: "rocketpool odao proposals list",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "states",
								Aliases: []string{"s"},
								Usage:   "Comma separated list of states to filter ('pending', 'active', 'succeeded', 'executed', 'cancelled', 'defeated', or 'expired')",
								Value:   "",
							},
						},
						Action: func(ctx context.Context, c *cli.Command) error {

							// Validate args
							if err := cliutils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Run
							return getProposals(c.String("states"))

						},
					},

					{
						Name:      "details",
						Aliases:   []string{"d"},
						Usage:     "View proposal details",
						UsageText: "rocketpool odao proposals details proposal-id",
						Action: func(ctx context.Context, c *cli.Command) error {

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
							return getProposal(id)

						},
					},

					{
						Name:      "cancel",
						Aliases:   []string{"c"},
						Usage:     "Cancel a proposal made by the node",
						UsageText: "rocketpool odao proposals cancel [options]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "proposal",
								Aliases: []string{"p"},
								Usage:   "The ID of the proposal to cancel",
							},
						},
						Action: func(ctx context.Context, c *cli.Command) error {

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
							return cancelProposal(c.String("proposal"), c.Bool("yes"))

						},
					},

					{
						Name:      "vote",
						Aliases:   []string{"v"},
						Usage:     "Vote on a proposal",
						UsageText: "rocketpool odao proposals vote [options]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "proposal",
								Aliases: []string{"p"},
								Usage:   "The ID of the proposal to vote on",
							},
							&cli.StringFlag{
								Name:    "support",
								Aliases: []string{"s"},
								Usage:   "Whether to support the proposal ('yes' or 'no')",
							},
							&cli.BoolFlag{
								Name:    "yes",
								Aliases: []string{"y"},
								Usage:   "Automatically confirm vote",
							},
						},
						Action: func(ctx context.Context, c *cli.Command) error {

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
							return voteOnProposal(c.String("proposal"), c.String("support"), c.Bool("yes"))

						},
					},

					{
						Name:      "execute",
						Aliases:   []string{"x"},
						Usage:     "Execute a proposal",
						UsageText: "rocketpool odao proposals execute [options]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "proposal",
								Aliases: []string{"p"},
								Usage:   "The ID of the proposal to execute (or 'all')",
							},
						},
						Action: func(ctx context.Context, c *cli.Command) error {

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
							return executeProposal(c.String("proposal"), c.Bool("yes"))

						},
					},
				},
			},

			{
				Name:      "join",
				Aliases:   []string{"j"},
				Usage:     "Join the oracle DAO (requires an executed invite proposal)",
				UsageText: "rocketpool odao join [options]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm joining",
					},
					&cli.BoolFlag{
						Name:    "swap",
						Aliases: []string{"s"},
						Usage:   "Automatically confirm swapping old RPL before joining",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return join(c.Bool("yes"), c.Bool("swap"))

				},
			},

			{
				Name:      "leave",
				Aliases:   []string{"l"},
				Usage:     "Leave the oracle DAO (requires an executed leave proposal)",
				UsageText: "rocketpool odao leave [options]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "refund-address",
						Aliases: []string{"r"},
						Usage:   "The address to refund the node's RPL bond to (or 'node')",
					},
					&cli.BoolFlag{
						Name:    "yes",
						Aliases: []string{"y"},
						Usage:   "Automatically confirm leaving",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("refund-address") != "" && c.String("refund-address") != "node" {
						if _, err := cliutils.ValidateAddress("bond refund address", c.String("refund-address")); err != nil {
							return err
						}
					}

					// Run
					return leave(c.String("refund-address"), c.Bool("yes"))

				},
			},
			{
				Name:    "upgrade",
				Aliases: []string{"u"},
				Usage:   "Upgrade Proposals",
				Commands: []*cli.Command{

					{
						Name:      "get-upgrade-proposals",
						Aliases:   []string{"g"},
						Usage:     "Get the upgrade proposals",
						UsageText: "rocketpool odao upgrade get-upgrade-proposals",
						Action: func(ctx context.Context, c *cli.Command) error {
							return getUpgradeProposals()
						},
					},
					{
						Name:      "execute-upgrade",
						Aliases:   []string{"eu"},
						Usage:     "Execute an upgrade",
						UsageText: "rocketpool odao upgrade execute-upgrade upgrade-proposal-id",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:    "yes",
								Aliases: []string{"y"},
								Usage:   "Automatically confirm the action",
							},
						},
						Action: func(ctx context.Context, c *cli.Command) error {
							return executeUpgrade(c.String("proposal"), c.Bool("yes"))
						},
					},
				},
			},
		},
	})
}
