package tx

import (
	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
	command.Subcommands = append(command.Subcommands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Send transactions created with the node wallet",
		Subcommands: []cli.Command{
			{
				Name:      "sign-tx",
				Aliases:   []string{"s"},
				Usage:     "Signs a transaction and gets its binary representation without sending it to the network",
				UsageText: "rocketpool api tx sign-tx data to value",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 3); err != nil {
						return err
					}
					data, err := cliutils.ValidateByteArray("data", c.Args().Get(0))
					if err != nil {
						return err
					}
					to, err := cliutils.ValidateAddress("to", c.Args().Get(1))
					if err != nil {
						return err
					}
					value, err := cliutils.ValidateBigInt("value", c.Args().Get(2))
					if err != nil {
						return err
					}

					// Regenerate the TX info
					txInfo := &core.TransactionInfo{
						Data:  data,
						To:    to,
						Value: value,
					}

					// Run
					api.PrintResponse(signTx(c, txInfo))
					return nil

				},
			},
		},
	})
}
