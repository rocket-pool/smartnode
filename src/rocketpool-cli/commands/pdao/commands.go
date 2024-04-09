package pdao

import (
	"fmt"

	"github.com/urfave/cli/v2"

	input "github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/shared/utils"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	// Create the auction settings commands
	auctionContract := rocketpool.ContractName_RocketDAOProtocolSettingsAuction
	auctionSettingsCmd := cliutils.CreateSetterCategory("auction", "Auction", "a", auctionContract)
	auctionSettingsCmd.Subcommands = []*cli.Command{
		cliutils.CreateBoolSetter("is-create-lot-enabled", "icle", auctionContract, protocol.SettingName_Auction_IsCreateLotEnabled, proposeSetting),
		cliutils.CreateBoolSetter("is-bid-on-lot-enabled", "ibole", auctionContract, protocol.SettingName_Auction_IsBidOnLotEnabled, proposeSetting),
		cliutils.CreateEthSetter("lot-minimum-eth-value", "lminev", auctionContract, protocol.SettingName_Auction_LotMinimumEthValue, proposeSetting),
		cliutils.CreateEthSetter("lot-maximum-eth-value", "lmaxev", auctionContract, protocol.SettingName_Auction_LotMaximumEthValue, proposeSetting),
		cliutils.CreateDurationSetter("lot-duration", "ld", auctionContract, protocol.SettingName_Auction_LotDuration, proposeSetting),
		cliutils.CreatePercentSetter("lot-starting-price-ratio", "lspr", auctionContract, protocol.SettingName_Auction_LotStartingPriceRatio, proposeSetting),
		cliutils.CreatePercentSetter("lot-reserve-price-ratio", "lrpr", auctionContract, protocol.SettingName_Auction_LotReservePriceRatio, proposeSetting),
	}

	// Create the deposit settings commands
	depositContract := rocketpool.ContractName_RocketDAOProtocolSettingsDeposit
	depositSettingsCmd := cliutils.CreateSetterCategory("deposit", "Deposit pool", "d", depositContract)
	depositSettingsCmd.Subcommands = []*cli.Command{
		cliutils.CreateBoolSetter("is-depositing-enabled", "ide", depositContract, protocol.SettingName_Deposit_IsDepositingEnabled, proposeSetting),
		cliutils.CreateBoolSetter("are-deposit-assignments-enabled", "adae", depositContract, protocol.SettingName_Deposit_AreDepositAssignmentsEnabled, proposeSetting),
		cliutils.CreateEthSetter("minimum-deposit", "md", depositContract, protocol.SettingName_Deposit_MinimumDeposit, proposeSetting),
		cliutils.CreateEthSetter("maximum-deposit-pool-size", "mdps", depositContract, protocol.SettingName_Deposit_MaximumDepositPoolSize, proposeSetting),
		cliutils.CreateUintSetter("maximum-assignments-per-deposit", "mapd", depositContract, protocol.SettingName_Deposit_MaximumAssignmentsPerDeposit, proposeSetting),
		cliutils.CreateUintSetter("maximum-socialised-assignments-per-deposit", "msapd", depositContract, protocol.SettingName_Deposit_MaximumSocialisedAssignmentsPerDeposit, proposeSetting),
	}
	df := cliutils.CreateUnboundedPercentSetter("deposit-fee", "df", depositContract, protocol.SettingName_Deposit_DepositFee, proposeSetting)
	df.Usage = fmt.Sprintf("Propose updating the %s setting; specify a percentage between 0 and 0.01 (e.g., '0.001' for 0.10%%)", string(protocol.SettingName_Deposit_DepositFee))
	depositSettingsCmd.Subcommands = append(depositSettingsCmd.Subcommands, df)

	// Create the minipool settings commands
	minipoolContract := rocketpool.ContractName_RocketDAOProtocolSettingsMinipool
	minipoolSettingsCmd := cliutils.CreateSetterCategory("minipool", "Minipool", "m", minipoolContract)
	minipoolSettingsCmd.Subcommands = []*cli.Command{
		cliutils.CreateBoolSetter("is-submit-withdrawable-enabled", "iswe", minipoolContract, protocol.SettingName_Minipool_IsSubmitWithdrawableEnabled, proposeSetting),
		cliutils.CreateDurationSetter("launch-timeout", "lt", minipoolContract, protocol.SettingName_Minipool_LaunchTimeout, proposeSetting),
		cliutils.CreateBoolSetter("is-bond-reduction-enabled", "ibre", minipoolContract, protocol.SettingName_Minipool_IsBondReductionEnabled, proposeSetting),
		cliutils.CreateUintSetter("max-count", "mc", minipoolContract, protocol.SettingName_Minipool_MaximumCount, proposeSetting),
		cliutils.CreateDurationSetter("user-distribute-window-start", "udws", minipoolContract, protocol.SettingName_Minipool_UserDistributeWindowStart, proposeSetting),
		cliutils.CreateDurationSetter("user-distribute-window-length", "udwl", minipoolContract, protocol.SettingName_Minipool_UserDistributeWindowLength, proposeSetting),
	}

	// Create the network settings commands
	networkContract := rocketpool.ContractName_RocketDAOProtocolSettingsNetwork
	networkSettingsCmd := cliutils.CreateSetterCategory("network", "Network", "ne", networkContract)
	networkSettingsCmd.Subcommands = []*cli.Command{
		cliutils.CreatePercentSetter("oracle-dao-consensus-threshold", "odct", networkContract, protocol.SettingName_Network_OracleDaoConsensusThreshold, proposeSetting),
		cliutils.CreatePercentSetter("node-penalty-threshold", "npt", networkContract, protocol.SettingName_Network_NodePenaltyThreshold, proposeSetting),
		cliutils.CreatePercentSetter("per-penalty-rate", "ppr", networkContract, protocol.SettingName_Network_PerPenaltyRate, proposeSetting),
		cliutils.CreateBoolSetter("is-submit-balances-enabled", "isbe", networkContract, protocol.SettingName_Network_IsSubmitBalancesEnabled, proposeSetting),
		cliutils.CreateDurationSetter("submit-balances-frequency", "sbf", networkContract, protocol.SettingName_Network_SubmitBalancesFrequency, proposeSetting),
		cliutils.CreateBoolSetter("is-submit-prices-enabled", "ispe", networkContract, protocol.SettingName_Network_IsSubmitPricesEnabled, proposeSetting),
		cliutils.CreateDurationSetter("submit-prices-frequency", "spf", networkContract, protocol.SettingName_Network_SubmitPricesFrequency, proposeSetting),
		cliutils.CreatePercentSetter("minimum-node-fee", "minnf", networkContract, protocol.SettingName_Network_MinimumNodeFee, proposeSetting),
		cliutils.CreatePercentSetter("target-node-fee", "tnf", networkContract, protocol.SettingName_Network_TargetNodeFee, proposeSetting),
		cliutils.CreatePercentSetter("maximum-node-fee", "maxnf", networkContract, protocol.SettingName_Network_MaximumNodeFee, proposeSetting),
		cliutils.CreateEthSetter("node-fee-demand-range", "nfdr", networkContract, protocol.SettingName_Network_NodeFeeDemandRange, proposeSetting),
		cliutils.CreatePercentSetter("target-reth-collateral-rate", "trcr", networkContract, protocol.SettingName_Network_TargetRethCollateralRate, proposeSetting),
		cliutils.CreateBoolSetter("is-submit-rewards-enabled", "isre", networkContract, protocol.SettingName_Network_IsSubmitRewardsEnabled, proposeSetting),
	}

	// Create the node settings commands
	nodeContract := rocketpool.ContractName_RocketDAOProtocolSettingsNode
	nodeSettingsCmd := cliutils.CreateSetterCategory("node", "Node", "no", nodeContract)
	nodeSettingsCmd.Subcommands = []*cli.Command{
		cliutils.CreateBoolSetter("is-registration-enabled", "ire", nodeContract, protocol.SettingName_Node_IsRegistrationEnabled, proposeSetting),
		cliutils.CreateBoolSetter("is-smoothing-pool-registration-enabled", "ispre", nodeContract, protocol.SettingName_Node_IsSmoothingPoolRegistrationEnabled, proposeSetting),
		cliutils.CreateBoolSetter("is-depositing-enabled", "ide", nodeContract, protocol.SettingName_Node_IsDepositingEnabled, proposeSetting),
		cliutils.CreateBoolSetter("are-vacant-minipools-enabled", "avme", nodeContract, protocol.SettingName_Node_AreVacantMinipoolsEnabled, proposeSetting),
		cliutils.CreateUnboundedPercentSetter("minimum-per-minipool-stake", "minpms", nodeContract, protocol.SettingName_Node_MinimumPerMinipoolStake, proposeSetting),
		cliutils.CreateUnboundedPercentSetter("maximum-per-minipool-stake", "maxpms", nodeContract, protocol.SettingName_Node_MaximumPerMinipoolStake, proposeSetting),
	}

	// Create the proposal setting commands
	proposalsContract := rocketpool.ContractName_RocketDAOProtocolSettingsProposals
	proposalsSettingsCmd := cliutils.CreateSetterCategory("proposals", "Proposal", "p", proposalsContract)
	proposalsSettingsCmd.Subcommands = []*cli.Command{
		cliutils.CreateDurationSetter("vote-phase1-time", "vt1", proposalsContract, protocol.SettingName_Proposals_VotePhase1Time, proposeSetting),
		cliutils.CreateDurationSetter("vote-phase2-time", "vt2", proposalsContract, protocol.SettingName_Proposals_VotePhase2Time, proposeSetting),
		cliutils.CreateDurationSetter("vote-delay-time", "vdt", proposalsContract, protocol.SettingName_Proposals_VoteDelayTime, proposeSetting),
		cliutils.CreateDurationSetter("execute-time", "et", proposalsContract, protocol.SettingName_Proposals_ExecuteTime, proposeSetting),
		cliutils.CreateRplSetter("proposal-bond", "pb", proposalsContract, protocol.SettingName_Proposals_ProposalBond, proposeSetting),
		cliutils.CreateRplSetter("challenge-bond", "cb", proposalsContract, protocol.SettingName_Proposals_ChallengeBond, proposeSetting),
		cliutils.CreateDurationSetter("challenge-period", "cp", proposalsContract, protocol.SettingName_Proposals_ChallengePeriod, proposeSetting),
		cliutils.CreatePercentSetter("quorum", "q", proposalsContract, protocol.SettingName_Proposals_ProposalQuorum, proposeSetting),
		cliutils.CreatePercentSetter("veto-quorum", "vq", proposalsContract, protocol.SettingName_Proposals_ProposalVetoQuorum, proposeSetting),
		cliutils.CreateBlockCountSetter("max-block-age", "mba", proposalsContract, protocol.SettingName_Proposals_ProposalMaxBlockAge, proposeSetting),
	}

	// Create the rewards setting commands
	rewardsContract := rocketpool.ContractName_RocketDAOProtocolSettingsRewards
	rewardsSettingsCmd := cliutils.CreateSetterCategory("rewards", "Rewards", "r", rewardsContract)
	rewardsSettingsCmd.Subcommands = []*cli.Command{
		cliutils.CreateUintSetter("interval-periods", "ip", rewardsContract, protocol.SettingName_Rewards_IntervalPeriods, proposeSetting),
	}

	// Create the security council setting commands
	securityContract := rocketpool.ContractName_RocketDAOProtocolSettingsSecurity
	securitySettingsCmd := cliutils.CreateSetterCategory("security", "Security council", "s", securityContract)
	securitySettingsCmd.Subcommands = []*cli.Command{
		cliutils.CreatePercentSetter("members-quorum", "mq", securityContract, protocol.SettingName_Security_MembersQuorum, proposeSetting),
		cliutils.CreateDurationSetter("members-leave-time", "mlt", securityContract, protocol.SettingName_Security_MembersLeaveTime, proposeSetting),
		cliutils.CreateDurationSetter("proposal-vote-time", "pvt", securityContract, protocol.SettingName_Security_ProposalVoteTime, proposeSetting),
		cliutils.CreateDurationSetter("proposal-execute-time", "pet", securityContract, protocol.SettingName_Security_ProposalExecuteTime, proposeSetting),
		cliutils.CreateDurationSetter("proposal-action-time", "pat", securityContract, protocol.SettingName_Security_ProposalActionTime, proposeSetting),
	}

	app.Commands = append(app.Commands, &cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the Rocket Pool Protocol DAO",
		Subcommands: []*cli.Command{
			{
				Name:    "settings",
				Aliases: []string{"s"},
				Usage:   "Show all of the current Protocol DAO settings and values",
				Action: func(c *cli.Context) error {
					// Run
					return getSettings(c)
				},
			},

			{
				Name:    "rewards-percentages",
				Aliases: []string{"rp"},
				Usage:   "View the RPL rewards allocation percentages for node operators, the Oracle DAO, and the Protocol DAO",
				Action: func(c *cli.Context) error {
					// Run
					return getRewardsPercentages(c)
				},
			},

			{
				Name:    "claim-bonds",
				Aliases: []string{"cb"},
				Usage:   "Unlock any bonded RPL you have for a proposal or set of challenges, and claim any bond rewards for defending or defeating the proposal",
				Flags: []cli.Flag{
					cliutils.InstantiateFlag(proposalFlag, "The ID of the proposal to claim bonds from (or 'all')"),
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
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
				Subcommands: []*cli.Command{
					{
						Name:    "rewards-percentages",
						Aliases: []string{"rp"},
						Usage:   "Propose updating the RPL rewards allocation percentages for node operators, the Oracle DAO, and the Protocol DAO",
						Flags: []cli.Flag{
							cliutils.RawFlag,
							cliutils.YesFlag,
							proposeRewardsPercentagesNodeFlag,
							proposeRewardsPercentagesOdaoFlag,
							proposeRewardsPercentagesPdaoFlag,
						},
						Action: func(c *cli.Context) error {
							// Validate args
							if err := utils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Run
							return proposeRewardsPercentages(c)
						},
					},

					{
						Name:    "one-time-spend",
						Aliases: []string{"ots"},
						Usage:   "Propose a one-time spend of the Protocol DAO's treasury",
						Flags: []cli.Flag{
							cliutils.RawFlag,
							cliutils.YesFlag,
							oneTimeSpendInvoiceFlag,
							recipientFlag,
							amountFlag,
						},
						Action: func(c *cli.Context) error {
							// Validate args
							if err := utils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Run
							return proposeOneTimeSpend(c)
						},
					},

					{
						Name:    "recurring-spend",
						Aliases: []string{"rs"},
						Usage:   "Propose a recurring spend of the Protocol DAO's treasury",
						Flags: []cli.Flag{
							cliutils.RawFlag,
							cliutils.YesFlag,
							contractNameFlag,
							recipientFlag,
							amountPerPeriodFlag,
							recurringSpendStartTimeFlag,
							periodLengthFlag,
							numberOfPeriodsFlag,
						},
						Action: func(c *cli.Context) error {
							// Validate args
							if err := utils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Run
							return proposeRecurringSpend(c)
						},
					},

					{
						Name:    "recurring-spend-update",
						Aliases: []string{"rsu"},
						Usage:   "Propose an update to an existing recurring spend plan",
						Flags: []cli.Flag{
							cliutils.RawFlag,
							cliutils.YesFlag,
							contractNameFlag,
							recipientFlag,
							amountPerPeriodFlag,
							periodLengthFlag,
							numberOfPeriodsFlag,
						},
						Action: func(c *cli.Context) error {
							// Validate args
							if err := utils.ValidateArgCount(c, 0); err != nil {
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
						Subcommands: []*cli.Command{
							{
								Name:    "invite",
								Aliases: []string{"i"},
								Usage:   "Propose an invitation to the security council",
								Flags: []cli.Flag{
									cliutils.YesFlag,
									scInviteIdFlag,
									scInviteAddressFlag,
								},
								Action: func(c *cli.Context) error {
									// Validate args
									if err := utils.ValidateArgCount(c, 0); err != nil {
										return err
									}

									// Run
									return proposeSecurityCouncilInvite(c)
								},
							},

							{
								Name:    "kick",
								Aliases: []string{"k"},
								Usage:   "Propose kicking one or more members from the security council",
								Flags: []cli.Flag{
									cliutils.YesFlag,
									scKickAddressesFlag,
								},
								Action: func(c *cli.Context) error {
									// Validate args
									if err := utils.ValidateArgCount(c, 0); err != nil {
										return err
									}

									// Run
									return proposeSecurityCouncilKick(c)
								},
							},

							{
								Name:    "replace",
								Aliases: []string{"r"},
								Usage:   "Propose replacing an existing member of the security council with a new member",
								Flags: []cli.Flag{
									cliutils.YesFlag,
									scReplaceExistingAddressFlag,
									scReplaceNewIdFlag,
									scReplaceNewAddressFlag,
								},
								Action: func(c *cli.Context) error {
									// Validate args
									if err := utils.ValidateArgCount(c, 0); err != nil {
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
						Subcommands: []*cli.Command{
							auctionSettingsCmd,
							depositSettingsCmd,
							minipoolSettingsCmd,
							networkSettingsCmd,
							nodeSettingsCmd,
							proposalsSettingsCmd,
							rewardsSettingsCmd,
							securitySettingsCmd,
						},
					},
				},
			},

			{
				Name:    "proposals",
				Aliases: []string{"o"},
				Usage:   "Manage Protocol DAO proposals",
				Subcommands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"l"},
						Usage:   "List the Protocol DAO proposals",
						Flags: []cli.Flag{
							proposalsListStatesFlag,
						},
						Action: func(c *cli.Context) error {
							// Validate args
							if err := utils.ValidateArgCount(c, 0); err != nil {
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
						ArgsUsage: "proposal-id",
						Action: func(c *cli.Context) error {
							// Validate args
							var err error
							if err = utils.ValidateArgCount(c, 1); err != nil {
								return err
							}
							id, err := input.ValidateUint("proposal-id", c.Args().Get(0))
							if err != nil {
								return err
							}

							// Run
							return getProposal(c, id)
						},
					},

					{
						Name:    "vote",
						Aliases: []string{"v"},
						Usage:   "Vote on a proposal",
						Flags: []cli.Flag{
							cliutils.InstantiateFlag(proposalFlag, "The ID of the proposal to vote on"),
							voteDirectionFlag,
							cliutils.YesFlag,
						},
						Action: func(c *cli.Context) error {
							// Validate args
							if err := utils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Run
							return voteOnProposal(c)
						},
					},

					{
						Name:    "execute",
						Aliases: []string{"x"},
						Usage:   "Execute a proposal",
						Flags: []cli.Flag{
							executeProposalFlag,
							cliutils.YesFlag,
						},
						Action: func(c *cli.Context) error {
							// Validate args
							if err := utils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Validate flags
							if c.String(executeProposalFlag.Name) != "" && c.String(executeProposalFlag.Name) != "all" {
								if _, err := input.ValidatePositiveUint("proposal ID", c.String(executeProposalFlag.Name)); err != nil {
									return err
								}
							}

							// Run
							return executeProposals(c)
						},
					},

					{
						Name:      "defeat",
						Aliases:   []string{"t"},
						Usage:     "Defeat a proposal that still has a tree index in the 'Challenged' state after its challenge window has passed",
						ArgsUsage: "proposal-id challenged-index",
						Flags: []cli.Flag{
							cliutils.YesFlag,
						},
						Action: func(c *cli.Context) error {
							// Validate args
							if err := utils.ValidateArgCount(c, 2); err != nil {
								return err
							}
							proposalId, err := input.ValidatePositiveUint("proposal-id", c.Args().Get(0))
							if err != nil {
								return err
							}
							index, err := input.ValidatePositiveUint("challenged-index", c.Args().Get(1))
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
						ArgsUsage: "proposal-id",
						Flags: []cli.Flag{
							cliutils.YesFlag,
						},
						Action: func(c *cli.Context) error {
							// Validate args
							if err := utils.ValidateArgCount(c, 1); err != nil {
								return err
							}
							proposalId, err := input.ValidatePositiveUint("proposal-id", c.Args().Get(0))
							if err != nil {
								return err
							}

							// Run
							return finalizeProposal(c, proposalId)
						},
					},
				},
			},

			{
				Name:    "initialize-voting",
				Aliases: []string{"iv"},
				Usage:   "Unlocks a node operator's voting power (only required for node operators who registered before governance structure was in place)",
				Action: func(c *cli.Context) error {
					// Run
					return initializeVoting(c)
				},
			},

			{
				Name:      "set-voting-delegate",
				Aliases:   []string{"svd"},
				Usage:     "Set the address you want to use when voting on Rocket Pool on-chain governance proposals, or the address you want to delegate your voting power to.",
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
					return setVotingDelegate(c, delegate)
				},
			},
		},
	})
}
