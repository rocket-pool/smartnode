package faucet

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
        Usage:     "Withdraw from the Rocket Pool faucet",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "withdraw",
                Aliases:   []string{"w"},
                Usage:     "Withdraw ETH or tokens from the Rocket Pool faucet",
                UsageText: "rocketpool api faucet withdraw token",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    token, err := cliutils.ValidateWithdrawableTokenType("token type", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(faucetWithdraw(c, token))
                    return nil

                },
            },

        },
    })
}

