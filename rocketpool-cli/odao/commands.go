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
                Name:      "proposals",
                Aliases:   []string{"p"},
                Usage:     "Get the oracle DAO proposals",
                UsageText: "rocketpool odao proposals",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return getProposals(c)

                },
            },

            cli.Command{
                Name:      "propose-invite",
                Aliases:   []string{"i"},
                Usage:     "Propose inviting a new member",
                UsageText: "rocketpool odao propose-invite member-address member-id member-email",
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
                Name:      "propose-leave",
                Aliases:   []string{"l"},
                Usage:     "Propose leaving the oracle DAO",
                UsageText: "rocketpool odao propose-leave",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return proposeLeave(c)

                },
            },

            cli.Command{
                Name:      "propose-replace",
                Aliases:   []string{"r"},
                Usage:     "Propose replacing the node's position with a new member",
                UsageText: "rocketpool odao propose-replace member-address member-id member-email",
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
                Name:      "propose-kick",
                Aliases:   []string{"k"},
                Usage:     "Propose kicking a member",
                UsageText: "rocketpool odao propose-kick [options]",
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

            cli.Command{
                Name:      "cancel-proposal",
                Aliases:   []string{"c"},
                Usage:     "Cancel a proposal made by the node",
                UsageText: "rocketpool odao cancel-proposal [options]",
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
                Name:      "vote-proposal",
                Aliases:   []string{"v"},
                Usage:     "Vote on a proposal",
                UsageText: "rocketpool odao vote-proposal [options]",
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
                Name:      "execute-proposal",
                Aliases:   []string{"x"},
                Usage:     "Execute a proposal",
                UsageText: "rocketpool api odao execute-proposal [options]",
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
                Aliases:   []string{"e"},
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
                Aliases:   []string{"a"},
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

