package upgrade

import (
	"context"

	"github.com/rocket-pool/smartnode/shared/utils/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli/v3"
)

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
	command.Commands = append(command.Commands, &cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the Rocket Pool trusted node DAO upgrades",
		Commands: []*cli.Command{
			{
				Name:      "get-upgrade-proposals",
				Usage:     "List the available upgrades",
				UsageText: "rocketpool api upgrades get-upgrade-proposals",
				Action: func(ctx context.Context, c *cli.Command) error {
					// Run
					api.PrintResponse(getUpgradeProposals(c))
					return nil
				},
			},
			{

				Name:      "can-execute-upgrade",
				Usage:     "Check whether the node can execute a proposal",
				UsageText: "rocketpool api upgrades can-execute-upgrade upgrade-proposal-id",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					upgradeProposalId, err := cliutils.ValidatePositiveUint("upgrade proposal ID", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canExecuteUpgrade(c, upgradeProposalId))
					return nil

				},
			},
			{
				Name:      "execute-upgrade",
				Aliases:   []string{"x"},
				Usage:     "Execute an upgrade",
				UsageText: "rocketpool api upgrades execute-upgrade upgrade-proposal-id",
				Action: func(ctx context.Context, c *cli.Command) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					upgradeProposalId, err := cliutils.ValidatePositiveUint("upgrade proposal ID", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(executeUpgrade(c, upgradeProposalId))
					return nil

				},
			},
		},
	})
}
