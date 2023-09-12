package faucet

import (
	"github.com/urfave/cli"

	types "github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
	command.Subcommands = append(command.Subcommands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Access the legacy RPL faucet",
		Subcommands: []cli.Command{
			// Status
			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get the faucet's status",
				UsageText: "rocketpool api faucet status",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					response, err := runFaucetCall[types.FaucetStatusResponse](c, &faucetStatusHandler{})
					api.PrintResponse(response, err)
					return nil

				},
			},

			// Withdraw RPL
			{
				Name:      "withdraw-rpl",
				Aliases:   []string{"w"},
				Usage:     "Withdraw legacy RPL from the faucet",
				UsageText: "rocketpool api faucet withdraw-rpl",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					response, err := runFaucetCall[types.FaucetWithdrawRplResponse](c, &faucetWithdrawHandler{})
					api.PrintResponse(response, err)
					return nil

				},
			},
		},
	})
}
