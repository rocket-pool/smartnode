package pdao

import (
	"errors"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/utils/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
	command.Subcommands = append(command.Subcommands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the Rocket Pool protocol DAO",
		Subcommands: []cli.Command{

			{
				Name:      "proposals",
				Aliases:   []string{"p"},
				Usage:     "Get the protocol DAO proposals",
				UsageText: "rocketpool api pdao proposals",
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
				UsageText: "rocketpool api pdao proposal-details proposal-id",
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
				Name:      "can-vote-proposal",
				Usage:     "Check whether the node can vote on a proposal",
				UsageText: "rocketpool api pdao can-vote-proposal proposal-id support",
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
					api.PrintResponse(canVoteOnProposal(c, proposalId, support))
					return nil

				},
			},
			{
				Name:      "vote-proposal",
				Aliases:   []string{"v"},
				Usage:     "Vote on a proposal",
				UsageText: "rocketpool api pdao vote-proposal proposal-id support",
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
				UsageText: "rocketpool api pdao can-execute-proposal proposal-id",
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
				UsageText: "rocketpool api pdao execute-proposal proposal-id",
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
				Name:      "get-settings",
				Usage:     "Get the Protocol DAO settings",
				UsageText: "rocketpool api pdao get-member-settings",
				Action: func(c *cli.Context) error {

					// Run
					api.PrintResponse(getSettings(c))
					return nil

				},
			},

			{
				Name:      "can-propose-setting",
				Usage:     "Check whether the node can propose a PDAO setting",
				UsageText: "rocketpool api pdao can-propose-setting setting-name value",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					settingName := c.Args().Get(0)
					value := c.Args().Get(1)

					// Run
					api.PrintResponse(canProposeSetting(c, settingName, value))
					return nil

				},
			},
			{
				Name:      "propose-setting",
				Usage:     "Propose updating a PDAO setting (use can-propose-setting to get the pollard)",
				UsageText: "rocketpool api pdao propose-setting setting-name value pollard",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 4); err != nil {
						return err
					}
					settingName := c.Args().Get(0)
					value := c.Args().Get(1)
					blockNumber, err := cliutils.ValidatePositiveUint32("block-number", c.Args().Get(2))
					if err != nil {
						return err
					}
					pollard := c.Args().Get(3)

					// Run
					api.PrintResponse(proposeSetting(c, settingName, value, blockNumber, pollard))
					return nil

				},
			},

			{
				Name:      "get-rewards-percentages",
				Usage:     "Get the allocation percentages of RPL rewards for the Oracle DAO, the Protocol DAO, and the node operators",
				UsageText: "rocketpool api pdao get-rewards-percentages",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getRewardsPercentages(c))
					return nil

				},
			},
			{
				Name:      "can-propose-rewards-percentages",
				Usage:     "Check whether the node can propose new RPL rewards allocation percentages for the Oracle DAO, the Protocol DAO, and the node operators",
				UsageText: "rocketpool api pdao can-propose-rewards-percentages node odao pdao",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 3); err != nil {
						return err
					}
					node, err1 := cliutils.ValidateBigInt("node", c.Args().Get(0))
					odao, err2 := cliutils.ValidateBigInt("odao", c.Args().Get(1))
					pdao, err3 := cliutils.ValidateBigInt("pdao", c.Args().Get(2))
					err := errors.Join(err1, err2, err3)
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canProposeRewardsPercentages(c, node, odao, pdao))
					return nil

				},
			},
			{
				Name:      "propose-rewards-percentages",
				Usage:     "Propose new RPL rewards allocation percentages for the Oracle DAO, the Protocol DAO, and the node operators",
				UsageText: "rocketpool api pdao propose-rewards-percentages node odao pdao block-number pollard",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 5); err != nil {
						return err
					}
					node, err1 := cliutils.ValidateBigInt("node", c.Args().Get(0))
					odao, err2 := cliutils.ValidateBigInt("odao", c.Args().Get(1))
					pdao, err3 := cliutils.ValidateBigInt("pdao", c.Args().Get(2))
					blockNumber, err4 := cliutils.ValidateUint32("blockNumber", c.Args().Get(3))
					pollard := c.Args().Get(4)
					err := errors.Join(err1, err2, err3, err4)
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(proposeRewardsPercentages(c, node, odao, pdao, blockNumber, pollard))
					return nil

				},
			},
		},
	})
}
