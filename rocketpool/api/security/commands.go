package security

import (
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/utils/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
	command.Subcommands = append(command.Subcommands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the Rocket Pool security council",
		Subcommands: []cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get security council status",
				UsageText: "rocketpool api security status",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getStatus(c))
					return nil

				},
			},
			{
				Name:      "members",
				Aliases:   []string{"m"},
				Usage:     "Get the security council members",
				UsageText: "rocketpool api security members",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getMembers(c))
					return nil

				},
			},

			{
				Name:      "proposals",
				Aliases:   []string{"p"},
				Usage:     "Get the security council proposals",
				UsageText: "rocketpool api security proposals",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getProposals(c))
					return nil

				},
			},

			{
				Name:      "proposal-details",
				Aliases:   []string{"d"},
				Usage:     "Get details of a proposal",
				UsageText: "rocketpool api security proposal-details proposal-id",
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
					api.PrintResponse(getProposal(c, id))
					return nil

				},
			},

			{
				Name:      "can-propose-invite",
				Usage:     "Check whether the node can propose inviting a new member",
				UsageText: "rocketpool api security can-propose-invite member-id member-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					memberId, err := cliutils.ValidateDAOMemberID("member ID", c.Args().Get(0))
					if err != nil {
						return err
					}
					memberAddress, err := cliutils.ValidateAddress("member address", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canProposeInvite(c, memberId, memberAddress))
					return nil

				},
			},
			{
				Name:      "propose-invite",
				Aliases:   []string{"i"},
				Usage:     "Propose inviting a new member",
				UsageText: "rocketpool api security propose-invite member-id member-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					memberId, err := cliutils.ValidateDAOMemberID("member ID", c.Args().Get(0))
					if err != nil {
						return err
					}
					memberAddress, err := cliutils.ValidateAddress("member address", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(proposeInvite(c, memberId, memberAddress))
					return nil

				},
			},

			{
				Name:      "can-propose-leave",
				Usage:     "Check whether the node can propose leaving the security council",
				UsageText: "rocketpool api security can-propose-leave",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(canProposeLeave(c))
					return nil

				},
			},
			{
				Name:      "propose-leave",
				Aliases:   []string{"l"},
				Usage:     "Propose leaving the security council",
				UsageText: "rocketpool api security propose-leave",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(proposeLeave(c))
					return nil

				},
			},

			{
				Name:      "can-propose-kick",
				Usage:     "Check whether the node can propose kicking a member",
				UsageText: "rocketpool api security can-propose-kick member-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					memberAddress, err := cliutils.ValidateAddress("member address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canProposeKick(c, memberAddress))
					return nil

				},
			},
			{
				Name:      "propose-kick",
				Aliases:   []string{"k"},
				Usage:     "Propose kicking a member",
				UsageText: "rocketpool api security propose-kick member-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					memberAddress, err := cliutils.ValidateAddress("member address", c.Args().Get(0))
					if err != nil {
						return err
					}
					// Run
					api.PrintResponse(proposeKick(c, memberAddress))
					return nil

				},
			},

			{
				Name:      "can-propose-kick-multi",
				Usage:     "Check whether the node can propose kicking multiple members",
				UsageText: "rocketpool api security can-propose-kick-multi member-addresses",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					memberAddresses, err := cliutils.ValidateAddresses("member addresses", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canProposeKickMulti(c, memberAddresses))
					return nil

				},
			},
			{
				Name:      "propose-kick-multi",
				Usage:     "Propose kicking multiple member",
				UsageText: "rocketpool api security propose-kick-multi member-addresses",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					memberAddresses, err := cliutils.ValidateAddresses("member addresses", c.Args().Get(0))
					if err != nil {
						return err
					}
					// Run
					api.PrintResponse(proposeKickMulti(c, memberAddresses))
					return nil

				},
			},

			{
				Name:      "can-propose-replace-member",
				Usage:     "Check whether the node can propose replacing someone on the security council with another member",
				UsageText: "rocketpool api security can-propose-replace-member existing-address new-id new-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 3); err != nil {
						return err
					}
					existingAddress, err := cliutils.ValidateAddress("existingAddress", c.Args().Get(0))
					if err != nil {
						return err
					}
					newID := c.Args().Get(1)
					newAddress, err := cliutils.ValidateAddress("newAddress", c.Args().Get(2))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canProposeReplaceMember(c, existingAddress, newID, newAddress))
					return nil

				},
			},
			{
				Name:      "propose-replace-member",
				Usage:     "Propose replacing someone on the security council with another member",
				UsageText: "rocketpool api pdao propose-replace-member existing-address new-id new-address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 3); err != nil {
						return err
					}
					existingAddress, err := cliutils.ValidateAddress("existingAddress", c.Args().Get(0))
					if err != nil {
						return err
					}
					newID := c.Args().Get(1)
					newAddress, err := cliutils.ValidateAddress("newAddress", c.Args().Get(2))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(proposeReplaceMember(c, existingAddress, newID, newAddress))
					return nil

				},
			},

			{
				Name:      "can-cancel-proposal",
				Usage:     "Check whether the node can cancel a proposal",
				UsageText: "rocketpool api security can-cancel-proposal proposal-id",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					proposalId, err := cliutils.ValidatePositiveUint("proposal ID", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canCancelProposal(c, proposalId))
					return nil

				},
			},
			{
				Name:      "cancel-proposal",
				Aliases:   []string{"c"},
				Usage:     "Cancel a proposal made by the node",
				UsageText: "rocketpool api security cancel-proposal proposal-id",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					proposalId, err := cliutils.ValidatePositiveUint("proposal ID", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(cancelProposal(c, proposalId))
					return nil

				},
			},

			{
				Name:      "can-vote-proposal",
				Usage:     "Check whether the node can vote on a proposal",
				UsageText: "rocketpool api security can-vote-proposal proposal-id",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					proposalId, err := cliutils.ValidatePositiveUint("proposal ID", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canVoteOnProposal(c, proposalId))
					return nil

				},
			},
			{
				Name:      "vote-proposal",
				Aliases:   []string{"v"},
				Usage:     "Vote on a proposal",
				UsageText: "rocketpool api security vote-proposal proposal-id support",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					proposalId, err := cliutils.ValidatePositiveUint("proposal ID", c.Args().Get(0))
					if err != nil {
						return err
					}
					support, err := cliutils.ValidateBool("support", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(voteOnProposal(c, proposalId, support))
					return nil

				},
			},

			{
				Name:      "can-execute-proposal",
				Usage:     "Check whether the node can execute a proposal",
				UsageText: "rocketpool api security can-execute-proposal proposal-id",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					proposalId, err := cliutils.ValidatePositiveUint("proposal ID", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canExecuteProposal(c, proposalId))
					return nil

				},
			},
			{
				Name:      "execute-proposal",
				Aliases:   []string{"x"},
				Usage:     "Execute a proposal",
				UsageText: "rocketpool api security execute-proposal proposal-id",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					proposalId, err := cliutils.ValidatePositiveUint("proposal ID", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(executeProposal(c, proposalId))
					return nil

				},
			},

			{
				Name:      "can-join",
				Usage:     "Check whether the node can join the security council",
				UsageText: "rocketpool api security can-join",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(canJoin(c))
					return nil

				},
			},
			{
				Name:      "join",
				Aliases:   []string{"j"},
				Usage:     "Join the security council (requires an executed invite proposal)",
				UsageText: "rocketpool api security join",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(join(c))
					return nil

				},
			},

			{
				Name:      "can-leave",
				Usage:     "Check whether the node can leave the security council",
				UsageText: "rocketpool api security can-leave",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(canLeave(c))
					return nil

				},
			},
			{
				Name:      "leave",
				Aliases:   []string{"e"},
				Usage:     "Leave the security council (requires an executed leave proposal)",
				UsageText: "rocketpool api security leave",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(leave(c))
					return nil

				},
			},

			{
				Name:      "can-propose-setting",
				Usage:     "Check whether the node can propose a PDAO setting",
				UsageText: "rocketpool api security can-propose-setting contract-name setting-name value",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 3); err != nil {
						return err
					}
					contractName := c.Args().Get(0)
					settingName := c.Args().Get(1)
					value := c.Args().Get(2)

					// Run
					api.PrintResponse(canProposeSetting(c, contractName, settingName, value))
					return nil

				},
			},
			{
				Name:      "propose-setting",
				Usage:     "Propose updating a PDAO setting",
				UsageText: "rocketpool api security propose-setting contract-name setting-name value",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 3); err != nil {
						return err
					}
					contractName := c.Args().Get(0)
					settingName := c.Args().Get(1)
					value := c.Args().Get(2)

					// Run
					api.PrintResponse(proposeSetting(c, contractName, settingName, value))
					return nil

				},
			},
		},
	})
}
