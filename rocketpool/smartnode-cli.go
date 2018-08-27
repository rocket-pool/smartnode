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
    app.Name     = "rocketpool"
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

        // Deposit commands
        cli.Command{
            Name:      "deposit",
            Aliases:   []string{"d"},
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
                        err := validateArgs(c, 2, func(messages *[]string) {
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

                        });
                        if err != nil {
                            return err;
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
                        err := validateArgs(c, 0, nil)
                        if err != nil {
                            return err;
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
                        err := validateArgs(c, 1, func(messages *[]string) {
                            var err error

                            // Parse deposit id
                            depositId, err = strconv.ParseUint(c.Args().Get(0), 10, 64)
                            if err != nil {
                                *messages = append(*messages, "Invalid deposit id - must be an integer")
                            }

                        });
                        if err != nil {
                            return err;
                        }

                        // Run command
                        fmt.Println("Withdrawing:", depositId)
                        return nil

                    },
                },

            },
        },

        // Resource management commands
        cli.Command{
            Name:      "resource",
            Aliases:   []string{"r"},
            Usage:     "Manage resources",
            Subcommands: []cli.Command{

                // Check free resources
                cli.Command{
                    Name:      "free",
                    Aliases:   []string{"f"},
                    Usage:     "Check free resources assigned to the node",
                    UsageText: "rocketpool resource free [type]" + "\n   " +
                               "- valid types are 'eth' and 'rpl'",
                    Action: func(c *cli.Context) error {

                        // Arguments
                        var resourceType string

                        // Validate arguments
                        err := validateArgs(c, 1, func(messages *[]string) {

                            // Parse type
                            resourceType = c.Args().Get(0)
                            switch resourceType {
                                case "eth":
                                case "rpl":
                                default:
                                    *messages = append(*messages, "Invalid type - valid types are 'eth' and 'rpl'")
                            }

                        });
                        if err != nil {
                            return err;
                        }

                        // Run command
                        fmt.Printf("Free %v: 0\n", resourceType)
                        return nil

                    },
                },

                // Check used resources
                cli.Command{
                    Name:      "used",
                    Aliases:   []string{"u"},
                    Usage:     "Check used resources assigned to the node",
                    UsageText: "rocketpool resource used [type]" + "\n   " +
                               "- valid types are 'eth' and 'rpl'",
                    Action: func(c *cli.Context) error {

                        // Arguments
                        var resourceType string

                        // Validate arguments
                        err := validateArgs(c, 1, func(messages *[]string) {

                            // Parse type
                            resourceType = c.Args().Get(0)
                            switch resourceType {
                                case "eth":
                                case "rpl":
                                default:
                                    *messages = append(*messages, "Invalid type - valid types are 'eth' and 'rpl'")
                            }

                        });
                        if err != nil {
                            return err;
                        }

                        // Run command
                        fmt.Printf("Used %v: 0\n", resourceType)
                        return nil

                    },
                },

                // Check resources required
                cli.Command{
                    Name:      "required",
                    Aliases:   []string{"r"},
                    Usage:     "Check resources required based on current network utilisation",
                    UsageText: "rocketpool resource required [type, ether amount]" + "\n   " +
                               "- valid types are 'rpl'" + "\n   " +
                               "- ether amount must be a decimal number",
                    Action: func(c *cli.Context) error {

                        // Arguments
                        var resourceType string
                        var etherAmount float64

                        // Validate arguments
                        err := validateArgs(c, 2, func(messages *[]string) {
                            var err error

                            // Parse type
                            resourceType = c.Args().Get(0)
                            switch resourceType {
                                case "rpl":
                                default:
                                    *messages = append(*messages, "Invalid type - valid types are 'rpl'")
                            }

                            // Parse ether amount
                            etherAmount, err = strconv.ParseFloat(c.Args().Get(1), 64)
                            if err != nil {
                                *messages = append(*messages, "Invalid ether amount - must be a decimal number")
                            }

                        });
                        if err != nil {
                            return err;
                        }

                        // Run command
                        fmt.Printf("Required %v for %v eth: 0\n", resourceType, etherAmount)
                        return nil

                    },
                },

            },
        },

    }

    // Run application
    err := app.Run(os.Args)
    if err != nil {
        log.Fatal(err)
    }

}


// Validate a command's arguments
func validateArgs(c *cli.Context, count int, validate func(*[]string)) error {

    // Check argument count
    if len(c.Args()) != count {
        return cli.NewExitError("USAGE:" + "\n   " + c.Command.UsageText, 1);
    }

    // Validate
    messages := make([]string, 0)
    if validate != nil {
        validate(&messages)
    }

    // Return validation error or nil
    if len(messages) > 0 {
        return cli.NewExitError(strings.Join(messages, "\n"), 1)
    }
    return nil;

}

