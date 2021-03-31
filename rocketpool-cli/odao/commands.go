package odao

import (
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage the Rocket Pool oracle DAO",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get oracle DAO status",
                UsageText: "rocketpool odao status",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return getStatus(c)

                },
            },

            cli.Command{
                Name:      "members",
                Aliases:   []string{"m"},
                Usage:     "Get the oracle DAO members",
                UsageText: "rocketpool odao members",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return getMembers(c)

                },
            },

            cli.Command{
                Name:      "propose",
                Aliases:   []string{"p"},
                Usage:     "Make an oracle DAO proposal",
                Subcommands: []cli.Command{

                    cli.Command{
                        Name:      "member",
                        Aliases:   []string{"m"},
                        Usage:     "Make an oracle DAO member proposal",
                        Subcommands: []cli.Command{

                            cli.Command{
                                Name:      "invite",
                                Aliases:   []string{"i"},
                                Usage:     "Propose inviting a new member",
                                UsageText: "rocketpool odao propose member invite member-address member-id member-email",
                                Action: func(c *cli.Context) error {

                                    // Validate args
                                    if err := cliutils.ValidateArgCount(c, 3); err != nil { return err }
                                    memberAddress, err := cliutils.ValidateAddress("member address", c.Args().Get(0))
                                    if err != nil { return err }
                                    memberId, err := cliutils.ValidateDAOMemberID("member ID", c.Args().Get(1))
                                    if err != nil { return err }
                                    memberEmail, err := cliutils.ValidateDAOMemberEmail("member email address", c.Args().Get(2))
                                    if err != nil { return err }

                                    // Run
                                    return proposeInvite(c, memberAddress, memberId, memberEmail)

                                },
                            },

                            cli.Command{
                                Name:      "leave",
                                Aliases:   []string{"l"},
                                Usage:     "Propose leaving the oracle DAO",
                                UsageText: "rocketpool odao propose member leave",
                                Action: func(c *cli.Context) error {

                                    // Validate args
                                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                                    // Run
                                    return proposeLeave(c)

                                },
                            },

                            cli.Command{
                                Name:      "replace",
                                Aliases:   []string{"r"},
                                Usage:     "Propose replacing the node's position with a new member",
                                UsageText: "rocketpool odao propose member replace member-address member-id member-email",
                                Action: func(c *cli.Context) error {

                                    // Validate args
                                    if err := cliutils.ValidateArgCount(c, 3); err != nil { return err }
                                    memberAddress, err := cliutils.ValidateAddress("member address", c.Args().Get(0))
                                    if err != nil { return err }
                                    memberId, err := cliutils.ValidateDAOMemberID("member ID", c.Args().Get(1))
                                    if err != nil { return err }
                                    memberEmail, err := cliutils.ValidateDAOMemberEmail("member email address", c.Args().Get(2))
                                    if err != nil { return err }

                                    // Run
                                    return proposeReplace(c, memberAddress, memberId, memberEmail)

                                },
                            },

                            cli.Command{
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
                                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                                    // Validate flags
                                    if c.String("member") != "" {
                                        if _, err := cliutils.ValidateAddress("member address", c.String("member")); err != nil { return err }
                                    }
                                    if c.String("fine") != "" && c.String("fine") != "max" {
                                        if _, err := cliutils.ValidatePositiveEthAmount("fine amount", c.String("fine")); err != nil { return err }
                                    }

                                    // Run
                                    return proposeKick(c)

                                },
                            },

                        },
                    },

                    cli.Command{
                        Name:      "setting",
                        Aliases:   []string{"s"},
                        Usage:     "Make an oracle DAO setting proposal",
                        Subcommands: []cli.Command{

                            cli.Command{
                                Name:       "members-quorum",
                                Aliases:    []string{"q"},
                                Usage:      "Propose updating the members.quorum setting",
                                UsageText:  "rocketpool odao propose setting members-quorum value",
                                Action: func(c *cli.Context) error {

                                    // Validate args
                                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                                    quorumPercent, err := cliutils.ValidatePercentage("quorum percentage", c.Args().Get(0))
                                    if err != nil { return err }

                                    // Run
                                    return proposeSettingMembersQuorum(c, quorumPercent)

                                },
                            },
                            cli.Command{
                                Name:       "members-rplbond",
                                Aliases:    []string{"b"},
                                Usage:      "Propose updating the members.rplbond setting",
                                UsageText:  "rocketpool odao propose setting members-rplbond value",
                                Action: func(c *cli.Context) error {

                                    // Validate args
                                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                                    bondAmountEth, err := cliutils.ValidateEthAmount("RPL bond amount", c.Args().Get(0))
                                    if err != nil { return err }

                                    // Run
                                    return proposeSettingMembersRplBond(c, bondAmountEth)

                                },
                            },
                            cli.Command{
                                Name:       "members-minipool-unbonded-max",
                                Aliases:    []string{"u"},
                                Usage:      "Propose updating the members.minipool.unbonded.max setting",
                                UsageText:  "rocketpool odao propose setting members-minipool-unbonded-max value",
                                Action: func(c *cli.Context) error {

                                    // Validate args
                                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                                    unbondedMinipoolMax, err := cliutils.ValidateUint("maximum unbonded minipool count", c.Args().Get(0))
                                    if err != nil { return err }

                                    // Run
                                    return proposeSettingMinipoolUnbondedMax(c, unbondedMinipoolMax)

                                },
                            },
                            cli.Command{
                                Name:       "proposal-cooldown",
                                Aliases:    []string{"c"},
                                Usage:      "Propose updating the proposal.cooldown setting",
                                UsageText:  "rocketpool odao propose setting proposal-cooldown value",
                                Action: func(c *cli.Context) error {

                                    // Validate args
                                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                                    proposalCooldownBlocks, err := cliutils.ValidateUint("proposal cooldown period", c.Args().Get(0))
                                    if err != nil { return err }

                                    // Run
                                    return proposeSettingProposalCooldown(c, proposalCooldownBlocks)

                                },
                            },
                            cli.Command{
                                Name:       "proposal-vote-blocks",
                                Aliases:    []string{"v"},
                                Usage:      "Propose updating the proposal.vote.blocks setting",
                                UsageText:  "rocketpool odao propose setting proposal-vote-blocks value",
                                Action: func(c *cli.Context) error {

                                    // Validate args
                                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                                    proposalVoteBlocks, err := cliutils.ValidateUint("proposal voting period", c.Args().Get(0))
                                    if err != nil { return err }

                                    // Run
                                    return proposeSettingProposalVoteBlocks(c, proposalVoteBlocks)

                                },
                            },
                            cli.Command{
                                Name:       "proposal-vote-delay-blocks",
                                Aliases:    []string{"d"},
                                Usage:      "Propose updating the proposal.vote.delay.blocks setting",
                                UsageText:  "rocketpool odao propose setting proposal-vote-delay-blocks value",
                                Action: func(c *cli.Context) error {

                                    // Validate args
                                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                                    proposalDelayBlocks, err := cliutils.ValidateUint("proposal delay period", c.Args().Get(0))
                                    if err != nil { return err }

                                    // Run
                                    return proposeSettingProposalVoteDelayBlocks(c, proposalDelayBlocks)

                                },
                            },
                            cli.Command{
                                Name:       "proposal-execute-blocks",
                                Aliases:    []string{"x"},
                                Usage:      "Propose updating the proposal.execute.blocks setting",
                                UsageText:  "rocketpool odao propose setting proposal-execute-blocks value",
                                Action: func(c *cli.Context) error {

                                    // Validate args
                                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                                    proposalExecuteBlocks, err := cliutils.ValidateUint("proposal execution period", c.Args().Get(0))
                                    if err != nil { return err }

                                    // Run
                                    return proposeSettingProposalExecuteBlocks(c, proposalExecuteBlocks)

                                },
                            },
                            cli.Command{
                                Name:       "proposal-action-blocks",
                                Aliases:    []string{"a"},
                                Usage:      "Propose updating the proposal.action.blocks setting",
                                UsageText:  "rocketpool odao propose setting proposal-action-blocks value",
                                Action: func(c *cli.Context) error {

                                    // Validate args
                                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                                    proposalActionBlocks, err := cliutils.ValidateUint("proposal action period", c.Args().Get(0))
                                    if err != nil { return err }

                                    // Run
                                    return proposeSettingProposalActionBlocks(c, proposalActionBlocks)

                                },
                            },

                        },
                    },

                },
            },

            cli.Command{
                Name:      "proposals",
                Aliases:   []string{"o"},
                Usage:     "Manage oracle DAO proposals",
                Subcommands: []cli.Command{

                    cli.Command{
                        Name:      "list",
                        Aliases:   []string{"l"},
                        Usage:     "List the oracle DAO proposals",
                        UsageText: "rocketpool odao proposals list",
                        Action: func(c *cli.Context) error {

                            // Validate args
                            if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                            // Run
                            return getProposals(c)

                        },
                    },

                    cli.Command{
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
                            if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                            // Validate flags
                            if c.String("proposal") != "" {
                                if _, err := cliutils.ValidatePositiveUint("proposal ID", c.String("proposal")); err != nil { return err }
                            }

                            // Run
                            return cancelProposal(c)

                        },
                    },

                    cli.Command{
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
                            if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                            // Validate flags
                            if c.String("proposal") != "" {
                                if _, err := cliutils.ValidatePositiveUint("proposal ID", c.String("proposal")); err != nil { return err }
                            }
                            if c.String("support") != "" {
                                if _, err := cliutils.ValidateBool("support", c.String("support")); err != nil { return err }
                            }

                            // Run
                            return voteOnProposal(c)

                        },
                    },

                    cli.Command{
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
                            if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                            // Validate flags
                            if c.String("proposal") != "" && c.String("proposal") != "all" {
                                if _, err := cliutils.ValidatePositiveUint("proposal ID", c.String("proposal")); err != nil { return err }
                            }

                            // Run
                            return executeProposal(c)

                        },
                    },

                },
            },

            cli.Command{
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
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return join(c)

                },
            },

            cli.Command{
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
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Validate flags
                    if c.String("refund-address") != "" && c.String("refund-address") != "node" {
                        if _, err := cliutils.ValidateAddress("bond refund address", c.String("refund-address")); err != nil { return err }
                    }

                    // Run
                    return leave(c)

                },
            },

            cli.Command{
                Name:      "replace",
                Aliases:   []string{"r"},
                Usage:     "Replace the node's position in the oracle DAO (requires an executed replace proposal)",
                UsageText: "rocketpool odao replace [options]",
                Flags: []cli.Flag{
                    cli.BoolFlag{
                        Name:  "yes, y",
                        Usage: "Automatically confirm replacement",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return replace(c)

                },
            },

        },
    })
}

