package main

import (
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"

    "github.com/urfave/cli"
)

func main() {

    // Initialise application
    app := cli.NewApp()

    // Configure application
    app.Name     = "Rocket Pool"
    app.Usage    = "Rocket Pool node operator utilities"
    app.Version  = "0.0.1"
    app.Authors  = []cli.Author{
        cli.Author{
            Name:  "Jake Pospischil",
            Email: "jake@rocketpool.net",
        },
    }
    app.Copyright = "(c) 2018 Rocket Pool Pty Ltd"

    // Register commands
    app.Commands = []cli.Command{

        // Deposit RPL
        cli.Command{
            Name:      "deposit",
            Aliases:   []string{"d"},
            Usage:     "Deposit RPL into the node registration contract",
            UsageText: "rocketpool deposit [amount, unit]" + "\n   " +
                       "- amount must be a decimal number" + "\n   " +
                       "- valid units are 'rpl'",
            Category:  "Deposits",
            Action: func(c *cli.Context) error {

                // Check argument count
                if len(c.Args()) != 2 {
                    return cli.NewExitError("USAGE:" + "\n\n" + c.Command.UsageText, 1);
                }

                // Validation messages
                messages := make([]string, 0)

                // Parse amount
                amount, err := strconv.ParseFloat(c.Args().Get(0), 64)
                if err != nil {
                    messages = append(messages, "Invalid amount - must be a decimal number")
                }

                // Parse unit
                unit := c.Args().Get(1)
                switch unit {
                    case "rpl":
                    default:
                        messages = append(messages, "Invalid unit - valid units are 'rpl'")
                }

                // Return validation error
                if len(messages) > 0 {
                    return cli.NewExitError(strings.Join(messages, "\n"), 1)
                }

                // Run command
                fmt.Println("Depositing:", amount, unit)
                return nil

            },
        },

    }

    // Run application
    err := app.Run(os.Args)
    if err != nil {
        log.Fatal(err)
    }

}
