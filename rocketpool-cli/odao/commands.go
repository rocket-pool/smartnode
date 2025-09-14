package odao

import (
	"github.com/urfave/cli"

	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the Rocket Pool oracle DAO",
		Subcommands: []cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get oracle DAO status",
				UsageText: "rocketpool odao status",
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
				Usage:     "Get the oracle DAO members",
				UsageText: "rocketpool odao members",
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
				Name:      "member-settings",
				Aliases:   []string{"b"},
				Usage:     "Get the oracle DAO settings related to oracle DAO members",
				UsageText: "rocketpool odao member-settings",
				Action: func(c *cli.Context) error {

					// Run
					return getMemberSettings(c)

				},
			},

			{
				Name:      "proposal-settings",
				Aliases:   []string{"a"},
				Usage:     "Get the oracle DAO settings related to oracle DAO proposals",
				UsageText: "rocketpool odao proposal-settings",
				Action: func(c *cli.Context) error {

					// Run
					return getProposalSettings(c)

				},
			},

			{
				Name:      "minipool-settings",
				Aliases:   []string{"i"},
				Usage:     "Get the oracle DAO settings related to minipools",
				UsageText: "rocketpool odao minipool-settings",
				Action: func(c *cli.Context) error {

					// Run
					return getMinipoolSettings(c)

				},
			},

			{
				Name:      "penalise-megapool",
				Aliases:   []string{"pm"},
				Usage:     "(Saturn) Penalise a megapool",
				UsageText: "rocketpool odao penalise-megapool megapool-adress block",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes",
						Usage: "Automatically confirm the action",
					},
				},
				Action: func(c *cli.Context) error {

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
					return penaliseMegapool(c, megapoolAddress, block)

				},
			},

			{
				Name:    "propose",
				Aliases: []string{"p"},
				Usage:   "Make an oracle DAO proposal",
				Subcommands: []cli.Command{

					{
						Name:    "member",
						Aliases: []string{"m"},
						Usage:   "Make an oracle DAO member proposal",
						Subcommands: []cli.Command{

							{
								Name:      "invite",
								Aliases:   []string{"i"},
								Usage:     "Propose inviting a new member",
								UsageText: "rocketpool odao propose member invite member-address member-id member-url",
								Action: func(c *cli.Context) error {

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
									return proposeInvite(c, memberAddress, memberId, c.Args().Get(2))

								},
							},

							{
								Name:      "leave",
								Aliases:   []string{"l"},
								Usage:     "Propose leaving the oracle DAO",
								UsageText: "rocketpool odao propose member leave",
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 0); err != nil {
										return err
									}

									// Run
									return proposeLeave(c)

								},
							},

							{
								Name:      "kick",
								Aliases:   []string{"k"},
								Usage:     "Propose kicking a member",
								UsageText: "rocketpool odao propose member kick [options]",
								Flags: []cli.Flag{
									cli.StringFlag{
										Name:  "member, m",
										Usage: "The address of the member to propose kicking",
									},
									cli.StringFlag{
										Name:  "fine, f",
										Usage: "The amount of RPL to fine the member (or 'max')",
									},
								},
								Action: func(c *cli.Context) error {

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
									return proposeKick(c)

								},
							},
						},
					},

					{
						Name:    "setting",
						Aliases: []string{"s"},
						Usage:   "Make an oracle DAO setting proposal",
						Subcommands: []cli.Command{

							{
								Name:      "members-quorum",
								Aliases:   []string{"q"},
								Usage:     "Propose updating the members.quorum setting - takes a percent, from 0 to 100",
								UsageText: "rocketpool odao propose setting members-quorum value",
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}
									quorumPercent, err := cliutils.ValidatePercentage("quorum percentage", c.Args().Get(0))
									if err != nil {
										return err
									}

									// Run
									return proposeSettingMembersQuorum(c, quorumPercent)

								},
							},
							{
								Name:      "members-rplbond",
								Aliases:   []string{"b"},
								Usage:     "Propose updating the members.rplbond setting - takes an RPL amount (e.g. 5000)",
								UsageText: "rocketpool odao propose setting members-rplbond value",
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}
									bondAmountEth, err := cliutils.ValidateEthAmount("RPL bond amount", c.Args().Get(0))
									if err != nil {
										return err
									}

									// Run
									return proposeSettingMembersRplBond(c, bondAmountEth)

								},
							},
							{
								Name:      "members-minipool-unbonded-max",
								Aliases:   []string{"u"},
								Usage:     "Propose updating the members.minipool.unbonded.max setting - takes a number (e.g 100)",
								UsageText: "rocketpool odao propose setting members-minipool-unbonded-max value",
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}
									unbondedMinipoolMax, err := cliutils.ValidateUint("maximum unbonded minipool count", c.Args().Get(0))
									if err != nil {
										return err
									}

									// Run
									return proposeSettingMinipoolUnbondedMax(c, unbondedMinipoolMax)

								},
							},
							{
								Name:      "proposal-cooldown",
								Aliases:   []string{"c"},
								Usage:     "Propose updating the proposal.cooldown.time setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting proposal-cooldown value",
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingProposalCooldown(c, c.Args().Get(0))

								},
							},
							{
								Name:      "proposal-vote-timespan",
								Aliases:   []string{"v"},
								Usage:     "Propose updating the proposal.vote.time setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting proposal-vote-timespan value",
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingProposalVoteTimespan(c, c.Args().Get(0))

								},
							},
							{
								Name:      "proposal-vote-delay-timespan",
								Aliases:   []string{"d"},
								Usage:     "Propose updating the proposal.vote.delay.time setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting proposal-vote-delay-timespan value",
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingProposalVoteDelayTimespan(c, c.Args().Get(0))

								},
							},
							{
								Name:      "proposal-execute-timespan",
								Aliases:   []string{"x"},
								Usage:     "Propose updating the proposal.execute.time setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting proposal-execute-timespan value",
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingProposalExecuteTimespan(c, c.Args().Get(0))

								},
							},
							{
								Name:      "proposal-action-timespan",
								Aliases:   []string{"a"},
								Usage:     "Propose updating the proposal.action.time setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting proposal-action-timespan value",
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingProposalActionTimespan(c, c.Args().Get(0))

								},
							},
							{
								Name:      "scrub-period",
								Aliases:   []string{"s"},
								Usage:     "Propose updating the minipool.scrub.period setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting scrub-period value",
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingScrubPeriod(c, c.Args().Get(0))

								},
							},
							{
								Name:      "promotion-scrub-period",
								Aliases:   []string{"p"},
								Usage:     "Propose updating the minipool.promotion.scrub.period setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting promotion-scrub-period value",
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingPromotionScrubPeriod(c, c.Args().Get(0))

								},
							},
							{
								Name:      "scrub-penalty-enabled",
								Aliases:   []string{"spe"},
								Usage:     "Propose updating the minipool.scrub.penalty.enabled setting - format is true / false",
								UsageText: "rocketpool odao propose setting scrub-penalty-enabled value",
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}
									enabled, err := cliutils.ValidateBool("enabled", c.Args().Get(0))
									if err != nil {
										return err
									}

									// Run
									return proposeSettingScrubPenaltyEnabled(c, enabled)

								},
							},
							{
								Name:      "bond-reduction-window-start",
								Aliases:   []string{"brws"},
								Usage:     "Propose updating the minipool.bond.reduction.window.start setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting bond-reduction-window-start value",
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingBondReductionWindowStart(c, c.Args().Get(0))

								},
							},
							{
								Name:      "bond-reduction-window-length",
								Aliases:   []string{"brwl"},
								Usage:     "Propose updating the minipool.bond.reduction.window.length setting - format is e.g. 1h30m45s",
								UsageText: "rocketpool odao propose setting bond-reduction-window-length value",
								Action: func(c *cli.Context) error {

									// Validate args
									if err := cliutils.ValidateArgCount(c, 1); err != nil {
										return err
									}

									// Run
									return proposeSettingBondReductionWindowLength(c, c.Args().Get(0))

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
				Subcommands: []cli.Command{

					{
						Name:      "list",
						Aliases:   []string{"l"},
						Usage:     "List the oracle DAO proposals",
						UsageText: "rocketpool odao proposals list",
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
						UsageText: "rocketpool odao proposals details proposal-id",
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
						UsageText: "rocketpool odao proposals cancel [options]",
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
						UsageText: "rocketpool odao proposals vote [options]",
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
						UsageText: "rocketpool odao proposals execute [options]",
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
				Usage:     "Join the oracle DAO (requires an executed invite proposal)",
				UsageText: "rocketpool odao join [options]",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm joining",
					},
					cli.BoolFlag{
						Name:  "swap, s",
						Usage: "Automatically confirm swapping old RPL before joining",
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
				Usage:     "Leave the oracle DAO (requires an executed leave proposal)",
				UsageText: "rocketpool odao leave [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "refund-address, r",
						Usage: "The address to refund the node's RPL bond to (or 'node')",
					},
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

					// Validate flags
					if c.String("refund-address") != "" && c.String("refund-address") != "node" {
						if _, err := cliutils.ValidateAddress("bond refund address", c.String("refund-address")); err != nil {
							return err
						}
					}

					// Run
					return leave(c)

				},
			},
		},
	})
}
