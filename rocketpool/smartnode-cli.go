package main

import (
    "fmt"
    "log"
    "os"
    "regexp"
    "strconv"
    "strings"

    "github.com/urfave/cli"
)


func main() {

    // Add logo to help template
    cli.AppHelpTemplate = fmt.Sprintf(
        "______           _        _    ______           _ " + "\n" +
        "| ___ \\         | |      | |   | ___ \\         | |" + "\n" +
        "| |_/ /___   ___| | _____| |_  | |_/ /__   ___ | |" + "\n" +
        "|    // _ \\ / __| |/ / _ \\ __| |  __/ _ \\ / _ \\| |" + "\n" +
        "| |\\ \\ (_) | (__|   <  __/ |_  | | | (_) | (_) | |" + "\n" +
        "\\_| \\_\\___/ \\___|_|\\_\\___|\\__| \\_|  \\___/ \\___/|_|" + "\n\n" +
    "%s", cli.AppHelpTemplate)

    // Initialise application
    app := cli.NewApp()

    // Configure application
    app.Name     = "rocketpool"
    app.Usage    = "Rocket Pool node operator utilities"
    app.Version  = "0.0.1"
    app.Authors  = []cli.Author{
        cli.Author{
            Name:  "Darren Langley",
            Email: "darren@rocketpool.net",
        },
        cli.Author{
            Name:  "David Rugendyke",
            Email: "david@rocketpool.net",
        },
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

        // Fee management commands
        cli.Command{
            Name:      "fee",
            Aliases:   []string{"f"},
            Usage:     "Manage user fees",
            Subcommands: []cli.Command{

                // Display fee
                cli.Command{
                    Name:      "display",
                    Aliases:   []string{"d"},
                    Usage:     "Display the current user fee percentage",
                    UsageText: "rocketpool fee display",
                    Action: func(c *cli.Context) error {

                        // Validate arguments
                        err := validateArgs(c, 0, nil)
                        if err != nil {
                            return err;
                        }

                        // Run command
                        fmt.Println("User fee:")
                        return nil

                    },
                },

                // Vote on fee
                cli.Command{
                    Name:      "vote",
                    Aliases:   []string{"v"},
                    Usage:     "Vote on the user fee percentage to be charged",
                    UsageText: "rocketpool fee vote [fee percentage]" + "\n   " +
                               "- fee percentage must be a decimal number between 0 and 100",
                    Action: func(c *cli.Context) error {

                        // Arguments
                        var feePercent float64

                        // Validate arguments
                        err := validateArgs(c, 1, func(messages *[]string) {
                            var err error

                            // Parse fee percentage
                            feePercent, err = strconv.ParseFloat(c.Args().Get(0), 64)
                            if err != nil {
                                *messages = append(*messages, "Invalid fee percentage - must be a decimal number")
                            }
                            if feePercent < 0 || feePercent > 100 {
                                *messages = append(*messages, "Invalid fee percentage - must be between 0 and 100")
                            }

                        });
                        if err != nil {
                            return err;
                        }

                        // Run command
                        fmt.Println("Voting:", feePercent)
                        return nil

                    },
                },

            },
        },

        // User commands
        cli.Command{
            Name:      "user",
            Aliases:   []string{"u"},
            Usage:     "Manage users",
            Subcommands: []cli.Command{

                // List users
                cli.Command{
                    Name:      "list",
                    Aliases:   []string{"l"},
                    Usage:     "List minipools and users assigned to the node",
                    UsageText: "rocketpool user list",
                    Action: func(c *cli.Context) error {

                        // Validate arguments
                        err := validateArgs(c, 0, nil)
                        if err != nil {
                            return err;
                        }

                        // Run command
                        fmt.Println("Users:")
                        return nil

                    },
                },

            },
        },

        // Node commands
        cli.Command{
            Name:      "node",
            Aliases:   []string{"n"},
            Usage:     "Manage node state",
            Subcommands: []cli.Command{

                // Pause node
                cli.Command{
                    Name:      "pause",
                    Aliases:   []string{"p"},
                    Usage:     "'Pause' the node; stop receiving new minipools",
                    UsageText: "rocketpool node pause",
                    Action: func(c *cli.Context) error {

                        // Validate arguments
                        err := validateArgs(c, 0, nil)
                        if err != nil {
                            return err;
                        }

                        // Run command
                        fmt.Println("Pausing...")
                        return nil

                    },
                },

                // Resume node
                cli.Command{
                    Name:      "resume",
                    Aliases:   []string{"r"},
                    Usage:     "'Resume' the node; start receiving new minipools again",
                    UsageText: "rocketpool node resume",
                    Action: func(c *cli.Context) error {

                        // Validate arguments
                        err := validateArgs(c, 0, nil)
                        if err != nil {
                            return err;
                        }

                        // Run command
                        fmt.Println("Resuming...")
                        return nil

                    },
                },

                // Exit node
                cli.Command{
                    Name:      "exit",
                    Aliases:   []string{"e"},
                    Usage:     "'Exit' the node from Rocket Pool permanently; node is paused instead if it has assigned minipools",
                    UsageText: "rocketpool node exit",
                    Action: func(c *cli.Context) error {

                        // Validate arguments
                        err := validateArgs(c, 0, nil)
                        if err != nil {
                            return err;
                        }

                        // Run command
                        fmt.Println("Exiting...")
                        return nil

                    },
                },

            },
        },

        // RPIP commands
        cli.Command{
            Name:      "rpip",
            Usage:     "Manage Rocket Pool Improvement Proposals",
            Subcommands: []cli.Command{

                // List proposals
                cli.Command{
                    Name:      "list",
                    Aliases:   []string{"l"},
                    Usage:     "List current Rocket Pool Improvement Proposals",
                    UsageText: "rocketpool rpip list",
                    Action: func(c *cli.Context) error {

                        // Validate arguments
                        err := validateArgs(c, 0, nil)
                        if err != nil {
                            return err;
                        }

                        // Run command
                        fmt.Println("Proposals:")
                        return nil

                    },
                },

                // Subscribe to alerts
                cli.Command{
                    Name:      "alert",
                    Aliases:   []string{"a"},
                    Usage:     "Subscribe an email address to new RPIP alerts",
                    UsageText: "rocketpool rpip alert [email address]",
                    Action: func(c *cli.Context) error {

                        // Arguments
                        var email string

                        // Validate arguments
                        err := validateArgs(c, 1, func(messages *[]string) {

                            // Parse email address
                            email = c.Args().Get(0)
                            if !regexp.MustCompile("^[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$").MatchString(email) {
                                *messages = append(*messages, "Invalid email address")
                            }

                        });
                        if err != nil {
                            return err;
                        }

                        // Run command
                        fmt.Printf("Subscribing %v to alerts...\n", email)
                        return nil

                    },
                },

                // Vote on a proposal
                cli.Command{
                    Name:      "vote",
                    Aliases:   []string{"v"},
                    Usage:     "Vote on a Rocket Pool Improvement Proposal",
                    UsageText: "rocketpool rpip vote [proposal id, vote]" + "\n   " +
                               "- proposal id must match the id of a current proposal" + "\n   " +
                               "- valid vote values are 'yes', 'y', 'no', and 'n'",
                    Action: func(c *cli.Context) error {

                        // Arguments
                        var proposalId uint64
                        var vote bool

                        // Validate arguments
                        err := validateArgs(c, 2, func(messages *[]string) {
                            var err error

                            // Parse proposal id
                            proposalId, err = strconv.ParseUint(c.Args().Get(0), 10, 64)
                            if err != nil {
                                *messages = append(*messages, "Invalid proposal id - must be an integer")
                            }

                            // Parse vote
                            switch c.Args().Get(1) {
                                case "yes":
                                    vote = true
                                case "y":
                                    vote = true
                                case "no":
                                    vote = false
                                case "n":
                                    vote = false
                                default:
                                    *messages = append(*messages, "Invalid vote - valid values are 'yes', 'y', 'no', and 'n'")
                            }

                        });
                        if err != nil {
                            return err;
                        }

                        // Run command
                        fmt.Println("Voting:", proposalId, vote)
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

