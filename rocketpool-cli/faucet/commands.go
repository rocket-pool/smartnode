package faucet

import (
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Withdraw from the Rocket Pool faucet",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "withdraw",
                Aliases:   []string{"w"},
                Usage:     "Withdraw ETH or tokens from the Rocket Pool faucet",
                UsageText: "rocketpool faucet withdraw token",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    token, err := cliutils.ValidateWithdrawableTokenType("token type", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    return faucetWithdraw(c, token)

                },
            },

        },
    })
}

