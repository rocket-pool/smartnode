package minipool

import (
	"github.com/urfave/cli"

	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage the node's minipools",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get a list of the node's minipools",
                UsageText: "rocketpool minipool status",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return getStatus(c)

                },
            },

            cli.Command{
                Name:      "leader",
                Aliases:   []string{"l"},
                Usage:     "minipool leaderboard",
                UsageText: "rocketpool minipool leader",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return getLeader(c)

                },
            },

            cli.Command{
                Name:      "refund",
                Aliases:   []string{"r"},
                Usage:     "Refund ETH belonging to the node from minipools",
                UsageText: "rocketpool minipool refund [options]",
                Flags: []cli.Flag{
                    cli.StringFlag{
                        Name:  "minipool, m",
                        Usage: "The minipool/s to refund from (address or 'all')",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Validate flags
                    if c.String("minipool") != "" && c.String("minipool") != "all" {
                        if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil { return err }
                    }

                    // Run
                    return refundMinipools(c)

                },
            },

            cli.Command{
                Name:      "dissolve",
                Aliases:   []string{"d"},
                Usage:     "Dissolve initialized or prelaunch minipools",
                UsageText: "rocketpool minipool dissolve [options]",
                Flags: []cli.Flag{
                    cli.BoolFlag{
                        Name:  "yes, y",
                        Usage: "Automatically confirm dissolving minipool/s",
                    },
                    cli.StringFlag{
                        Name:  "minipool, m",
                        Usage: "The minipool/s to dissolve (address or 'all')",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Validate flags
                    if c.String("minipool") != "" && c.String("minipool") != "all" {
                        if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil { return err }
                    }

                    // Run
                    return dissolveMinipools(c)

                },
            },

            cli.Command{
                Name:      "exit",
                Aliases:   []string{"e"},
                Usage:     "Exit staking minipools from the beacon chain",
                UsageText: "rocketpool minipool exit [options]",
                Flags: []cli.Flag{
                    cli.BoolFlag{
                        Name:  "yes, y",
                        Usage: "Automatically confirm exiting minipool/s",
                    },
                    cli.StringFlag{
                        Name:  "minipool, m",
                        Usage: "The minipool/s to exit (address or 'all')",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Validate flags
                    if c.String("minipool") != "" && c.String("minipool") != "all" {
                        if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil { return err }
                    }

                    // Run
                    return exitMinipools(c)

                },
            },

            cli.Command{
                Name:      "close",
                Aliases:   []string{"c"},
                Usage:     "Withdraw balances from dissolved minipools and close them",
                UsageText: "rocketpool minipool close [options]",
                Flags: []cli.Flag{
                    cli.StringFlag{
                        Name:  "minipool, m",
                        Usage: "The minipool/s to close (address or 'all')",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Validate flags
                    if c.String("minipool") != "" && c.String("minipool") != "all" {
                        if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil { return err }
                    }

                    // Run
                    return closeMinipools(c)

                },
            },

        },
    })
}

