package deposits

import (
    "fmt"
    "strconv"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/commands"
)


// Register deposit commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage node deposits",
        Subcommands: []cli.Command{

            // New deposit
            cli.Command{
                Name:      "new",
                Aliases:   []string{"n"},
                Usage:     "Deposit resources into the node",
                UsageText: "rocketpool deposit new [amount, unit]" + "\n   " +
                           "- amount must be a decimal number" + "\n   " +
                           "- valid units are 'rpl'",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var amount float64
                    var unit string

                    // Validate arguments
                    err := commands.ValidateArgs(c, 2, func(messages *[]string) {
                        var err error

                        // Parse amount
                        amount, err = strconv.ParseFloat(c.Args().Get(0), 64)
                        if err != nil {
                            *messages = append(*messages, "Invalid amount - must be a decimal number")
                        }

                        // Parse unit
                        unit = c.Args().Get(1)
                        switch unit {
                            case "rpl":
                            default:
                                *messages = append(*messages, "Invalid unit - valid units are 'rpl'")
                        }

                    })
                    if err != nil {
                        return err
                    }

                    // Run command
                    fmt.Println("Depositing:", amount, unit)
                    return nil

                },
            },

            // List deposits
            cli.Command{
                Name:      "list",
                Aliases:   []string{"l"},
                Usage:     "List all deposits with the node",
                UsageText: "rocketpool deposit list",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    err := commands.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err
                    }

                    // Run command
                    fmt.Println("Deposits:")
                    return nil

                },
            },

            // Withdraw a deposit
            cli.Command{
                Name:      "withdraw",
                Aliases:   []string{"w"},
                Usage:     "Withdraw a specific available deposit",
                UsageText: "rocketpool deposit withdraw [deposit id]" + "\n   " +
                           "- deposit id must match the id of an available listed deposit",
                Action: func(c *cli.Context) error {

                    // Arguments
                    var depositId uint64

                    // Validate arguments
                    err := commands.ValidateArgs(c, 1, func(messages *[]string) {
                        var err error

                        // Parse deposit id
                        depositId, err = strconv.ParseUint(c.Args().Get(0), 10, 64)
                        if err != nil {
                            *messages = append(*messages, "Invalid deposit id - must be an integer")
                        }

                    })
                    if err != nil {
                        return err
                    }

                    // Run command
                    fmt.Println("Withdrawing:", depositId)
                    return nil

                },
            },

        },
    })
}

