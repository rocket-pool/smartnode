package exchange

import (
    "strings"

    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register deposit commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage Uniswap token exchanges",
        Subcommands: []cli.Command{

            // Get the current token liquidity
            cli.Command{
                Name:      "liquidity",
                Aliases:   []string{"l"},
                Usage:     "Get the current liquidity available for a token",
                UsageText: "rocketpool exchange liquidity token",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var token string

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 1, func(messages *[]string) {

                        // Parse token type
                        token = strings.ToUpper(c.Args().Get(0))
                        switch token {
                            case "RPL":
                            default:
                                *messages = append(*messages, "Invalid token - valid tokens are 'RPL'")
                        }

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return getTokenLiquidity(c, token)

                },
            },

        },
    })
}

