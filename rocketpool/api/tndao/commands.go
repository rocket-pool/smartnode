package tndao

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
        Usage:     "Manage Rocket Pool trusted node DAO",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "members",
                Aliases:   []string{"m"},
                Usage:     "View the trusted node DAO members",
                UsageText: "rocketpool api tndao members",
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
                Usage:     "View the trusted node DAO proposals",
                UsageText: "rocketpool api tndao proposals",
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
                UsageText: "rocketpool api tndao can-propose-invite member-address",
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
                UsageText: "rocketpool api tndao propose-invite member-address member-id member-email",
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
                Usage:     "Check whether the node can propose leaving the trusted node DAO",
                UsageText: "rocketpool api tndao can-propose-leave",
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
                Usage:     "Propose leaving the trusted node DAO",
                UsageText: "rocketpool api tndao propose-leave",
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
                UsageText: "rocketpool api tndao can-propose-replace member-address",
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
                UsageText: "rocketpool api tndao propose-replace member-address member-id member-email",
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
                UsageText: "rocketpool api tndao can-propose-kick member-address fine-amount",
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
                UsageText: "rocketpool api tndao propose-kick member-address fine-amount",
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
                UsageText: "rocketpool api tndao can-cancel-proposal proposal-id",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    proposalId, err := cliutils.ValidateInteger("proposal id", c.Args().Get(0))
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
                UsageText: "rocketpool api tndao cancel-proposal proposal-id",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    proposalId, err := cliutils.ValidateInteger("proposal id", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(cancelProposal(c, proposalId))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-vote-proposal",
                Usage:     "Check whether the node can vote on a proposal",
                UsageText: "rocketpool api tndao can-vote-proposal proposal-id",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    proposalId, err := cliutils.ValidateInteger("proposal id", c.Args().Get(0))
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
                UsageText: "rocketpool api tndao vote-proposal proposal-id support",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 2); err != nil { return err }
                    proposalId, err := cliutils.ValidateInteger("proposal id", c.Args().Get(0))
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
                UsageText: "rocketpool api tndao can-execute-proposal proposal-id",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    proposalId, err := cliutils.ValidateInteger("proposal id", c.Args().Get(0))
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
                UsageText: "rocketpool api tndao execute-proposal proposal-id",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    proposalId, err := cliutils.ValidateInteger("proposal id", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(executeProposal(c, proposalId))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-join",
                Usage:     "Check whether the node can join the trusted node DAO",
                UsageText: "rocketpool api tndao can-join",
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
                Usage:     "Join the trusted node DAO (requires an executed invite proposal)",
                UsageText: "rocketpool api tndao join",
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
                Usage:     "Check whether the node can leave the trusted node DAO",
                UsageText: "rocketpool api tndao can-leave",
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
                Usage:     "Leave the trusted node DAO (requires an executed leave proposal)",
                UsageText: "rocketpool api tndao leave bond-refund-address",
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
                Usage:     "Check whether the node can replace its position in the trusted node DAO",
                UsageText: "rocketpool api tndao can-replace",
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
                Usage:     "Replace the node's position in the trusted node DAO (requires an executed replace proposal)",
                UsageText: "rocketpool api tndao replace",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(replace(c))
                    return nil

                },
            },

        },
    })
}

