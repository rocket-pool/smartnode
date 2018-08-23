package main

import (
    "fmt"
    "log"
    "os"
    "strconv"

    "github.com/urfave/cli"
)

func main() {

    // Initialise application
    app := cli.NewApp()

    // Configure application
    app.Name = "Rocket Pool"
    app.Version = "0.0.1"
    app.Authors = []cli.Author{
        cli.Author{
            Name:  "Jake Pospischil",
            Email: "jake@rocketpool.net",
        },
    }
    app.Copyright = "(c) 2018 Rocket Pool Pty Ltd"
    app.Usage = "Rocket Pool node operator utilities"

    // Register commands
    app.Commands = []cli.Command{

        // Deposit RPL
        cli.Command{
            Name:        "deposit",
            Aliases:     []string{"d"},
            Category:    "Deposits",
            Usage:       "Deposit RPL into the node registration contract",
            Action: func(c *cli.Context) error {

                // Parse amount
                amount, err := strconv.ParseUint(c.Args().Get(0), 10, 64)
                if err != nil {
                    return cli.NewExitError("Invalid amount", 1)
                }

                // Parse unit
                unit := c.Args().Get(1)
                switch unit {
                    case "rpl":
                    default:
                        return cli.NewExitError("Invalid unit", 1)
                }

                // Run command
                fmt.Println("Deposit!", amount, unit)
                return nil

            },
            OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
                fmt.Fprintf(c.App.Writer, "Incorrect usage\n")
                return err
            },
        },

    }

    // Run application
    err := app.Run(os.Args)
    if err != nil {
        log.Fatal(err)
    }

}
