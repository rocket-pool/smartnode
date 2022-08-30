package supernode

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
		Usage:   "Run commands related to supernodes for SaaS providers and consumers",
		Subcommands: []cli.Command{
			{
				Name:      "can-deposit",
				Usage:     "Check whether the node can make a deposit for the provided supernode",
				UsageText: "rocketpool api node can-deposit amount supernode-address salt",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 3); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidateDepositWeiAmount("deposit amount", c.Args().Get(0))
					if err != nil {
						return err
					}
					supernodeAddress, err := cliutils.ValidateAddress("supernode-address", c.Args().Get(1))
					if err != nil {
						return err
					}
					salt, err := cliutils.ValidateBigInt("salt", c.Args().Get(2))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(canNodeDeposit(c, amountWei, supernodeAddress, salt))
					return nil

				},
			},
			{
				Name:      "deposit",
				Aliases:   []string{"d"},
				Usage:     "Make a deposit and create a minipool for the provided supernode",
				UsageText: "rocketpool api node deposit amount supernode-address salt",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 3); err != nil {
						return err
					}
					amountWei, err := cliutils.ValidateDepositWeiAmount("deposit amount", c.Args().Get(0))
					if err != nil {
						return err
					}
					supernodeAddress, err := cliutils.ValidateAddress("supernode-address", c.Args().Get(1))
					if err != nil {
						return err
					}
					salt, err := cliutils.ValidateBigInt("salt", c.Args().Get(2))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(nodeDeposit(c, amountWei, supernodeAddress, salt))
					return nil

				},
			},
		},
	})
}
