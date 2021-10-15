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
            /*
            REMOVED UNTIL BEACON WITHDRAWALS
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
            */
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
            /*
            REMOVED UNTIL BEACON WITHDRAWALS
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
            */
            cli.Command{
                Name:      "delegate-upgrade",
                Aliases:   []string{"u"},
                Usage:     "Upgrade a minipool's delegate contract to the latest version",
                UsageText: "rocketpool minipool delegate-upgrade [options]",
                Flags: []cli.Flag{
                    cli.StringFlag{
                        Name:  "minipool, m",
                        Usage: "The minipool/s to upgrade (address or 'all')",
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
                    return delegateUpgradeMinipools(c)

                },
            },

            cli.Command{
                Name:      "delegate-rollback",
                Aliases:   []string{"b"},
                Usage:     "Roll a minipool's delegate contract back to its previous version",
                UsageText: "rocketpool minipool delegate-rollback [options]",
                Flags: []cli.Flag{
                    cli.StringFlag{
                        Name:  "minipool, m",
                        Usage: "The minipool/s to rollback (address or 'all')",
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
                    return delegateRollbackMinipools(c)

                },
            },

            cli.Command{
                Name:      "set-use-latest-delegate",
                Aliases:   []string{"l"},
                Usage:     "If enabled, the minipool will ignore its current delegate contract and always use whatever the latest delegate is",
                UsageText: "rocketpool minipool set-use-latest-delegate [options] setting",
                Flags: []cli.Flag{
                    cli.StringFlag{
                        Name:  "minipool, m",
                        Usage: "The minipool/s to configure the use-latest setting on (address or 'all')",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    setting, err := cliutils.ValidateBool("setting", c.Args().Get(0))
                    if err != nil { return err }

                    // Validate flags
                    if c.String("minipool") != "" && c.String("minipool") != "all" {
                        if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil { return err }
                    }

                    // Run
                    return setUseLatestDelegateMinipools(c, setting)

                },
            },

            cli.Command{
                Name:      "find-vanity-address",
                Aliases:   []string{"v"},
                Usage:     "Search for a custom vanity minipool address",
                UsageText: "rocketpool minipool find-vanity-address [options]",
                Flags: []cli.Flag{
                    cli.StringFlag{
                        Name:  "prefix, p",
                        Usage: "The prefix of the address to search for (must start with 0x)",
                    },
                    cli.StringFlag{
                        Name:  "salt, s",
                        Usage: "The salt to start searching from (must start with 0x)",
                    },
                    cli.IntFlag{
                        Name:  "threads, t",
                        Usage: "The number of threads to use for searching (defaults to your CPU thread count)",
                    },
                    cli.StringFlag{
                        Name:  "node-address, n",
                        Usage: "The node address to search for (leave blank to use the local node)",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Validate flags

                    // Run
                    return findVanitySalt(c)

                },
            },

        },
    })
}

