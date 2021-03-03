package tndao

import (
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage the Rocket Pool trusted node DAO",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get trusted node DAO status",
                UsageText: "rocketpool tndao status",
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
                Usage:     "Get the trusted node DAO members",
                UsageText: "rocketpool tndao members",
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
                Usage:     "Get the trusted node DAO proposals",
                UsageText: "rocketpool tndao proposals",
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
                UsageText: "rocketpool tndao propose-invite member-address member-id member-email",
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
                Usage:     "Propose leaving the trusted node DAO",
                UsageText: "rocketpool tndao propose-leave",
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
                UsageText: "rocketpool tndao propose-replace member-address member-id member-email",
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
                UsageText: "rocketpool tndao propose-kick [options]",
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
    })
}

