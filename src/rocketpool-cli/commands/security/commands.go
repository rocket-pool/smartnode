package security

import (
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/input"
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
	}

	// Create the deposit settings commands
	depositContract := rocketpool.ContractName_RocketDAOProtocolSettingsDeposit
	depositSettingsCmd := cliutils.CreateSetterCategory("deposit", "Deposit pool", "d", depositContract)
	depositSettingsCmd.Subcommands = []*cli.Command{
		cliutils.CreateBoolSetter("is-depositing-enabled", "ide", depositContract, protocol.SettingName_Deposit_IsDepositingEnabled, proposeSetting),
		cliutils.CreateBoolSetter("are-deposit-assignments-enabled", "adae", depositContract, protocol.SettingName_Deposit_AreDepositAssignmentsEnabled, proposeSetting),
	}

	// Create the minipool settings commands
	minipoolContract := rocketpool.ContractName_RocketDAOProtocolSettingsMinipool
	minipoolSettingsCmd := cliutils.CreateSetterCategory("minipool", "Minipool", "m", minipoolContract)
	minipoolSettingsCmd.Subcommands = []*cli.Command{
		cliutils.CreateBoolSetter("is-submit-withdrawable-enabled", "iswe", minipoolContract, protocol.SettingName_Minipool_IsSubmitWithdrawableEnabled, proposeSetting),
		cliutils.CreateBoolSetter("is-bond-reduction-enabled", "ibre", minipoolContract, protocol.SettingName_Minipool_IsBondReductionEnabled, proposeSetting),
	}

	// Create the network settings commands
	networkContract := rocketpool.ContractName_RocketDAOProtocolSettingsNetwork
	networkSettingsCmd := cliutils.CreateSetterCategory("network", "Network", "ne", networkContract)
	networkSettingsCmd.Subcommands = []*cli.Command{
		cliutils.CreateBoolSetter("is-submit-balances-enabled", "isbe", networkContract, protocol.SettingName_Network_IsSubmitBalancesEnabled, proposeSetting),
		cliutils.CreateBoolSetter("is-submit-prices-enabled", "ispe", networkContract, protocol.SettingName_Network_IsSubmitPricesEnabled, proposeSetting),
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
	}

	app.Commands = append(app.Commands, &cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the Rocket Pool security council",
		Subcommands: []*cli.Command{
			{
				Name:    "status",
				Aliases: []string{"s"},
				Usage:   "Get security council status",
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
				Name:    "members",
				Aliases: []string{"m"},
				Usage:   "Get the security council members",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
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
				Subcommands: []*cli.Command{
					{
						Name:    "member",
						Aliases: []string{"m"},
						Usage:   "Make a security council member proposal",
						Subcommands: []*cli.Command{
							{
								Name:    "leave",
								Aliases: []string{"l"},
								Usage:   "Propose leaving the security council",
								Flags: []cli.Flag{
									cliutils.YesFlag,
								},
								Action: func(c *cli.Context) error {
									// Validate args
									if err := utils.ValidateArgCount(c, 0); err != nil {
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
						Subcommands: []*cli.Command{
							auctionSettingsCmd,
							depositSettingsCmd,
							minipoolSettingsCmd,
							networkSettingsCmd,
							nodeSettingsCmd,
						},
					},
				},
			},

			{
				Name:    "proposals",
				Aliases: []string{"o"},
				Usage:   "Manage security council proposals",
				Subcommands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"l"},
						Usage:   "List the security council proposals",
						Flags: []cli.Flag{
							proposalsListStatesFlag,
						},
						Action: func(c *cli.Context) error {
							// Validate args
							if err := utils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Run
							return getProposals(c, c.String(proposalsListStatesFlag.Name))
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
						Name:    "cancel",
						Aliases: []string{"c"},
						Usage:   "Cancel a proposal made by the node",
						Flags: []cli.Flag{
							cliutils.InstantiateFlag(proposalFlag, "The ID of the proposal to cancel"),
						},
						Action: func(c *cli.Context) error {
							// Validate args
							if err := utils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Run
							return cancelProposal(c)
						},
					},

					{
						Name:    "vote",
						Aliases: []string{"v"},
						Usage:   "Vote on a proposal",
						Flags: []cli.Flag{
							cliutils.InstantiateFlag(proposalFlag, "The ID of the proposal to vote on"),
							voteSupportFlag,
							cliutils.YesFlag,
						},
						Action: func(c *cli.Context) error {
							// Validate args
							if err := utils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Validate flags
							if c.String(voteSupportFlag.Name) != "" {
								if _, err := input.ValidateBool("support", c.String(voteSupportFlag.Name)); err != nil {
									return err
								}
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
						},
						Action: func(c *cli.Context) error {
							// Validate args
							if err := utils.ValidateArgCount(c, 0); err != nil {
								return err
							}

							// Run
							return executeProposal(c)
						},
					},
				},
			},

			{
				Name:    "join",
				Aliases: []string{"j"},
				Usage:   "Join the security council (requires an executed invite proposal)",
				Flags: []cli.Flag{
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return join(c)
				},
			},

			{
				Name:    "leave",
				Aliases: []string{"l"},
				Usage:   "Leave the security council (requires an executed leave proposal)",
				Flags: []cli.Flag{
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return leave(c)
				},
			},
		},
	})
}
