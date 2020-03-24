package minipool

import (
    "fmt"
    "strings"

    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register minipool subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
    command.Subcommands = append(command.Subcommands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage node minipools and users",
        Subcommands: []cli.Command{

            // Get the node's minipool statuses
            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the node's current minipool statuses",
                UsageText: "rocketpool minipool status [filter]" + "\n   " +
                           "- optionally filter by comma-delimited status codes, from:" + "\n   " +
                           "  'initialized', 'depositassigned', 'prelaunch', 'staking', 'loggedout', 'withdrawn', 'timedout'",
                Action: func(c *cli.Context) error {

                    // Check argument count
                    if len(c.Args()) != 0 && len(c.Args()) != 1 {
                        return cli.NewExitError("USAGE:" + "\n   " + c.Command.UsageText, 1)
                    }

                    // Get & validate status filters
                    var statusFilters []string
                    if len(c.Args()) == 1 {
                        statusFilters = strings.Split(c.Args().Get(0), ",")
                        for _, statusFilter := range statusFilters {
                            filterExists := false
                            for _, filterOption := range []string{"initialized", "depositassigned", "prelaunch", "staking", "loggedout", "withdrawn", "timedout"} {
                                if statusFilter == filterOption {
                                    filterExists = true
                                    break
                                }
                            }
                            if !filterExists {
                                return cli.NewExitError(fmt.Sprintf("Invalid filter '%s' - valid options are 'initialized', 'depositassigned', 'prelaunch', 'staking', 'loggedout', 'withdrawn', 'timedout'", statusFilter), 1)
                            }
                        }
                    }

                    // Run command
                    return getMinipoolStatus(c, statusFilters)

                },
            },

            // Withdraw node deposit from a minipool
            cli.Command{
                Name:      "withdraw",
                Aliases:   []string{"w"},
                Usage:     "Withdraw deposit from an initialized, withdrawn or timed out minipool",
                UsageText: "rocketpool minipool withdraw",
                Action: func(c *cli.Context) error {

                    // Validate arguments
                    if err := cliutils.ValidateArgs(c, 0, nil); err != nil {
                        return err
                    }

                    // Run command
                    return withdrawMinipool(c)

                },
            },

        },
    })
}

