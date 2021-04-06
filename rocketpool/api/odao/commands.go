package odao

import (
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/utils/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
    command.Subcommands = append(command.Subcommands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage the Rocket Pool oracle DAO",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get oracle DAO status",
                UsageText: "rocketpool api odao status",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(getStatus(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "members",
                Aliases:   []string{"m"},
                Usage:     "Get the oracle DAO members",
                UsageText: "rocketpool api odao members",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(getMembers(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "proposals",
                Aliases:   []string{"p"},
                Usage:     "Get the oracle DAO proposals",
                UsageText: "rocketpool api odao proposals",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(getProposals(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-propose-invite",
                Usage:     "Check whether the node can propose inviting a new member",
                UsageText: "rocketpool api odao can-propose-invite member-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    memberAddress, err := cliutils.ValidateAddress("member address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canProposeInvite(c, memberAddress))
                    return nil

                },
            },
            cli.Command{
                Name:      "propose-invite",
                Aliases:   []string{"i"},
                Usage:     "Propose inviting a new member",
                UsageText: "rocketpool api odao propose-invite member-address member-id member-email",
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
                    api.PrintResponse(proposeInvite(c, memberAddress, memberId, memberEmail))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-propose-leave",
                Usage:     "Check whether the node can propose leaving the oracle DAO",
                UsageText: "rocketpool api odao can-propose-leave",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(canProposeLeave(c))
                    return nil

                },
            },
            cli.Command{
                Name:      "propose-leave",
                Aliases:   []string{"l"},
                Usage:     "Propose leaving the oracle DAO",
                UsageText: "rocketpool api odao propose-leave",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(proposeLeave(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-propose-replace",
                Usage:     "Check whether the node can propose replacing its position with a new member",
                UsageText: "rocketpool api odao can-propose-replace member-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    memberAddress, err := cliutils.ValidateAddress("member address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canProposeReplace(c, memberAddress))
                    return nil

                },
            },
            cli.Command{
                Name:      "propose-replace",
                Aliases:   []string{"r"},
                Usage:     "Propose replacing the node's position with a new member",
                UsageText: "rocketpool api odao propose-replace member-address member-id member-email",
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
                    api.PrintResponse(proposeReplace(c, memberAddress, memberId, memberEmail))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-propose-kick",
                Usage:     "Check whether the node can propose kicking a member",
                UsageText: "rocketpool api odao can-propose-kick member-address fine-amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    memberAddress, err := cliutils.ValidateAddress("member address", c.Args().Get(0))
                    if err != nil { return err }
                    fineAmountWei, err := cliutils.ValidatePositiveWeiAmount("fine amount", c.Args().Get(1))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canProposeKick(c, memberAddress, fineAmountWei))
                    return nil

                },
            },
            cli.Command{
                Name:      "propose-kick",
                Aliases:   []string{"k"},
                Usage:     "Propose kicking a member",
                UsageText: "rocketpool api odao propose-kick member-address fine-amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    memberAddress, err := cliutils.ValidateAddress("member address", c.Args().Get(0))
                    if err != nil { return err }
                    fineAmountWei, err := cliutils.ValidatePositiveWeiAmount("fine amount", c.Args().Get(1))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(proposeKick(c, memberAddress, fineAmountWei))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-cancel-proposal",
                Usage:     "Check whether the node can cancel a proposal",
                UsageText: "rocketpool api odao can-cancel-proposal proposal-id",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    proposalId, err := cliutils.ValidatePositiveUint("proposal ID", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canCancelProposal(c, proposalId))
                    return nil

                },
            },
            cli.Command{
                Name:      "cancel-proposal",
                Aliases:   []string{"c"},
                Usage:     "Cancel a proposal made by the node",
                UsageText: "rocketpool api odao cancel-proposal proposal-id",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    proposalId, err := cliutils.ValidatePositiveUint("proposal ID", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(cancelProposal(c, proposalId))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-vote-proposal",
                Usage:     "Check whether the node can vote on a proposal",
                UsageText: "rocketpool api odao can-vote-proposal proposal-id",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    proposalId, err := cliutils.ValidatePositiveUint("proposal ID", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canVoteOnProposal(c, proposalId))
                    return nil

                },
            },
            cli.Command{
                Name:      "vote-proposal",
                Aliases:   []string{"v"},
                Usage:     "Vote on a proposal",
                UsageText: "rocketpool api odao vote-proposal proposal-id support",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    proposalId, err := cliutils.ValidatePositiveUint("proposal ID", c.Args().Get(0))
                    if err != nil { return err }
                    support, err := cliutils.ValidateBool("support", c.Args().Get(1))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(voteOnProposal(c, proposalId, support))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-execute-proposal",
                Usage:     "Check whether the node can execute a proposal",
                UsageText: "rocketpool api odao can-execute-proposal proposal-id",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    proposalId, err := cliutils.ValidatePositiveUint("proposal ID", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(canExecuteProposal(c, proposalId))
                    return nil

                },
            },
            cli.Command{
                Name:      "execute-proposal",
                Aliases:   []string{"x"},
                Usage:     "Execute a proposal",
                UsageText: "rocketpool api odao execute-proposal proposal-id",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    proposalId, err := cliutils.ValidatePositiveUint("proposal ID", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(executeProposal(c, proposalId))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-join",
                Usage:     "Check whether the node can join the oracle DAO",
                UsageText: "rocketpool api odao can-join",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(canJoin(c))
                    return nil

                },
            },
            cli.Command{
                Name:      "join",
                Aliases:   []string{"j"},
                Usage:     "Join the oracle DAO (requires an executed invite proposal)",
                UsageText: "rocketpool api odao join",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(join(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-leave",
                Usage:     "Check whether the node can leave the oracle DAO",
                UsageText: "rocketpool api odao can-leave",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(canLeave(c))
                    return nil

                },
            },
            cli.Command{
                Name:      "leave",
                Aliases:   []string{"e"},
                Usage:     "Leave the oracle DAO (requires an executed leave proposal)",
                UsageText: "rocketpool api odao leave bond-refund-address",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    bondRefundAddress, err := cliutils.ValidateAddress("bond refund address", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(leave(c, bondRefundAddress))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-replace",
                Usage:     "Check whether the node can replace its position in the oracle DAO",
                UsageText: "rocketpool api odao can-replace",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(canReplace(c))
                    return nil

                },
            },
            cli.Command{
                Name:      "replace",
                Aliases:   []string{"a"},
                Usage:     "Replace the node's position in the oracle DAO (requires an executed replace proposal)",
                UsageText: "rocketpool api odao replace",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(replace(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-propose-setting",
                Usage:     "Check whether the node can propose an oracle DAO setting update",
                UsageText: "rocketpool api odao can-propose-setting",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(canProposeSetting(c))
                    return nil

                },
            },
            cli.Command{
                Name:      "propose-members-quorum",
                Usage:     "Propose updating the members.quorum setting",
                UsageText: "rocketpool api odao propose-members-quorum value",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    quorum, err := cliutils.ValidateFraction("quorum", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(proposeSettingMembersQuorum(c, quorum))
                    return nil

                },
            },
            cli.Command{
                Name:      "propose-members-rplbond",
                Usage:     "Propose updating the members.rplbond setting",
                UsageText: "rocketpool api odao propose-members-rplbond value",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    bondAmountWei, err := cliutils.ValidateWeiAmount("RPL bond amount", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(proposeSettingMembersRplBond(c, bondAmountWei))
                    return nil

                },
            },
            cli.Command{
                Name:      "propose-members-minipool-unbonded-max",
                Usage:     "Propose updating the members.minipool.unbonded.max setting",
                UsageText: "rocketpool api odao propose-members-minipool-unbonded-max value",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    unbondedMinipoolMax, err := cliutils.ValidateUint("maximum unbonded minipool count", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(proposeSettingMinipoolUnbondedMax(c, unbondedMinipoolMax))
                    return nil

                },
            },
            cli.Command{
                Name:      "propose-proposal-cooldown",
                Usage:     "Propose updating the proposal.cooldown setting",
                UsageText: "rocketpool api odao propose-proposal-cooldown value",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    proposalCooldownBlocks, err := cliutils.ValidateUint("proposal cooldown period", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(proposeSettingProposalCooldown(c, proposalCooldownBlocks))
                    return nil

                },
            },
            cli.Command{
                Name:      "propose-proposal-vote-blocks",
                Usage:     "Propose updating the proposal.vote.blocks setting",
                UsageText: "rocketpool api odao propose-proposal-vote-blocks value",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    proposalVoteBlocks, err := cliutils.ValidateUint("proposal voting period", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(proposeSettingProposalVoteBlocks(c, proposalVoteBlocks))
                    return nil

                },
            },
            cli.Command{
                Name:      "propose-proposal-vote-delay-blocks",
                Usage:     "Propose updating the proposal.vote.delay.blocks setting",
                UsageText: "rocketpool api odao propose-proposal-vote-delay-blocks value",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    proposalDelayBlocks, err := cliutils.ValidateUint("proposal delay period", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(proposeSettingProposalVoteDelayBlocks(c, proposalDelayBlocks))
                    return nil

                },
            },
            cli.Command{
                Name:      "propose-proposal-execute-blocks",
                Usage:     "Propose updating the proposal.execute.blocks setting",
                UsageText: "rocketpool api odao propose-proposal-execute-blocks value",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    proposalExecuteBlocks, err := cliutils.ValidateUint("proposal execution period", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(proposeSettingProposalExecuteBlocks(c, proposalExecuteBlocks))
                    return nil

                },
            },
            cli.Command{
                Name:      "propose-proposal-action-blocks",
                Usage:     "Propose updating the proposal.action.blocks setting",
                UsageText: "rocketpool api odao propose-proposal-action-blocks value",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    proposalActionBlocks, err := cliutils.ValidateUint("proposal action period", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(proposeSettingProposalActionBlocks(c, proposalActionBlocks))
                    return nil

                },
            },
            cli.Command{
                Name:      "get-member-settings",
                Usage:     "Get the ODAO settings related to ODAO members",
                UsageText: "rocketpool api odao get-member-settings",
                Action: func(c *cli.Context) error {

                    // Run
                    api.PrintResponse(getMemberSettings(c))
                    return nil

                },
            },
            cli.Command{
                Name:      "get-proposal-settings",
                Usage:     "Get the ODAO settings related to ODAO proposals",
                UsageText: "rocketpool api odao get-proposal-settings",
                Action: func(c *cli.Context) error {

                    // Run
                    api.PrintResponse(getProposalSettings(c))
                    return nil

                },
            },

        },
    })
}

