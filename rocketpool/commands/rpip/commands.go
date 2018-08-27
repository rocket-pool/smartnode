package rpip

import (
    "fmt"
    "regexp"
    "strconv"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/commands"
)


// Register rpip commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
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
                    err := commands.ValidateArgs(c, 0, nil)
                    if err != nil {
                        return err;
                    }

                    // Run command
                    fmt.Println("Proposals:")
                    return nil

                },
            },

            // RPIP alert commands
            cli.Command{
                Name:      "alert",
                Aliases:   []string{"a"},
                Usage:     "Manage RPIP email alerts",
                Subcommands: []cli.Command{

                    // Subscribe to alerts
                    cli.Command{
                        Name:      "subscribe",
                        Aliases:   []string{"s"},
                        Usage:     "Subscribe an email address to new RPIP alerts",
                        UsageText: "rocketpool rpip alert subscribe [email address]",
                        Action: func(c *cli.Context) error {

                            // Arguments
                            var email string

                            // Validate arguments
                            err := commands.ValidateArgs(c, 1, func(messages *[]string) {

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
                            fmt.Printf("Subscribing %v to RPIP alerts...\n", email)
                            return nil

                        },
                    },

                    // Check alert subscription
                    cli.Command{
                        Name:      "check",
                        Aliases:   []string{"c"},
                        Usage:     "Check for new RPIP alert subscriptions",
                        UsageText: "rocketpool rpip alert check",
                        Action: func(c *cli.Context) error {

                            // Validate arguments
                            err := commands.ValidateArgs(c, 0, nil)
                            if err != nil {
                                return err;
                            }

                            // Run command
                            fmt.Println("Checking RPIP alerts...")
                            return nil

                        },
                    },

                    // Unsubscribe from alerts
                    cli.Command{
                        Name:      "unsubscribe",
                        Aliases:   []string{"u"},
                        Usage:     "Unsubscribe from new RPIP alerts",
                        UsageText: "rocketpool rpip alert unsubscribe",
                        Action: func(c *cli.Context) error {

                            // Validate arguments
                            err := commands.ValidateArgs(c, 0, nil)
                            if err != nil {
                                return err;
                            }

                            // Run command
                            fmt.Println("Unsubscribing from RPIP alerts...")
                            return nil

                        },
                    },

                },
            },

            // RPIP vote commands
            cli.Command{
                Name:      "vote",
                Aliases:   []string{"v"},
                Usage:     "Manage RPIP votes",
                Subcommands: []cli.Command{

                    // Commit a vote on a proposal
                    cli.Command{
                        Name:      "commit",
                        Aliases:   []string{"c"},
                        Usage:     "Commit a vote on a Rocket Pool Improvement Proposal",
                        UsageText: "rocketpool rpip vote commit [proposal id, vote]" + "\n   " +
                                   "- proposal id must match the id of a current proposal" + "\n   " +
                                   "- valid vote values are 'yes', 'y', 'no', and 'n'",
                        Action: func(c *cli.Context) error {

                            // Arguments
                            var proposalId uint64
                            var vote bool

                            // Validate arguments
                            err := commands.ValidateArgs(c, 2, func(messages *[]string) {
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
                            fmt.Println("Committing vote:", proposalId, vote)
                            return nil

                        },
                    },

                    // Check vote on a proposal
                    cli.Command{
                        Name:      "check",
                        Usage:     "Check your vote on a Rocket Pool Improvement Proposal",
                        UsageText: "rocketpool rpip vote commit [proposal id]" + "\n   " +
                                   "- proposal id must match the id of a current proposal",
                        Action: func(c *cli.Context) error {

                            // Arguments
                            var proposalId uint64

                            // Validate arguments
                            err := commands.ValidateArgs(c, 1, func(messages *[]string) {
                                var err error

                                // Parse proposal id
                                proposalId, err = strconv.ParseUint(c.Args().Get(0), 10, 64)
                                if err != nil {
                                    *messages = append(*messages, "Invalid proposal id - must be an integer")
                                }

                            });
                            if err != nil {
                                return err;
                            }

                            // Run command
                            fmt.Println("Checking vote:", proposalId)
                            return nil

                        },
                    },

                    // Reveal a vote on a proposal
                    cli.Command{
                        Name:      "reveal",
                        Aliases:   []string{"r"},
                        Usage:     "Reveal a vote on a Rocket Pool Improvement Proposal",
                        UsageText: "rocketpool rpip vote reveal [proposal id, vote]" + "\n   " +
                                   "- proposal id must match the id of a current proposal" + "\n   " +
                                   "- valid vote values are 'yes', 'y', 'no', and 'n'",
                        Action: func(c *cli.Context) error {

                            // Arguments
                            var proposalId uint64
                            var vote bool

                            // Validate arguments
                            err := commands.ValidateArgs(c, 2, func(messages *[]string) {
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
                            fmt.Println("Revealing vote:", proposalId, vote)
                            return nil

                        },
                    },

                },
            },

        },
    })
}

