package exchange

import (
    "strconv"
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

            // Get the current token price
            cli.Command{
                Name:      "price",
                Aliases:   []string{"p"},
                Usage:     "Get the current price for a token purchase",
                UsageText: "rocketpool exchange price amount token",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var amount float64
                    var token string

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 2, func(messages *[]string) {
                        var err error

                        // Parse amount
                        if amount, err = strconv.ParseFloat(c.Args().Get(0), 64); err != nil || amount <= 0 {
                            *messages = append(*messages, "Invalid amount - must be a positive decimal number")
                        }

                        // Parse token type
                        token = strings.ToUpper(c.Args().Get(1))
                        switch token {
                            case "RPL":
                            default:
                                *messages = append(*messages, "Invalid token - valid tokens are 'RPL'")
                        }

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return getTokenPrice(c, amount, token)

                },
            },

            // Buy tokens with ether
            cli.Command{
                Name:      "buy",
                Aliases:   []string{"b"},
                Usage:     "Buy tokens with ether",
                UsageText: "rocketpool exchange buy etherAmount tokenAmount token",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var etherAmount float64
                    var tokenAmount float64
                    var token string

                    // Validate arguments
                    if err := cliutils.ValidateAPIArgs(c, 3, func(messages *[]string) {
                        var err error

                        // Parse amounts
                        if etherAmount, err = strconv.ParseFloat(c.Args().Get(0), 64); err != nil || etherAmount <= 0 {
                            *messages = append(*messages, "Invalid ether amount - must be a positive decimal number")
                        }
                        if tokenAmount, err = strconv.ParseFloat(c.Args().Get(1), 64); err != nil || tokenAmount <= 0 {
                            *messages = append(*messages, "Invalid token amount - must be a positive decimal number")
                        }

                        // Parse token type
                        token = strings.ToUpper(c.Args().Get(2))
                        switch token {
                            case "RPL":
                            default:
                                *messages = append(*messages, "Invalid token - valid tokens are 'RPL'")
                        }

                    }); err != nil {
                        return err
                    }

                    // Run command
                    return buyTokens(c, etherAmount, tokenAmount, token)

                },
            },

        },
    })
}

