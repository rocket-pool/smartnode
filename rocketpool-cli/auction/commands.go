package auction

import (
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage Rocket Pool RPL auctions",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get RPL auction status",
                UsageText: "rocketpool auction status",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return getStatus(c)

                },
            },

            cli.Command{
                Name:      "lots",
                Aliases:   []string{"l"},
                Usage:     "Get RPL lots for auction",
                UsageText: "rocketpool auction lots",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return getLots(c)

                },
            },

            cli.Command{
                Name:      "create-lot",
                Aliases:   []string{"t"},
                Usage:     "Create a new lot",
                UsageText: "rocketpool auction create-lot",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return createLot(c)

                },
            },

            cli.Command{
                Name:      "bid-lot",
                Aliases:   []string{"b"},
                Usage:     "Bid on a lot",
                UsageText: "rocketpool auction bid-lot amount",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return bidOnLot(c)

                },
            },

            cli.Command{
                Name:      "claim-lot",
                Aliases:   []string{"c"},
                Usage:     "Claim RPL from a lot",
                UsageText: "rocketpool auction claim-lot",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return claimFromLot(c)

                },
            },

            cli.Command{
                Name:      "recover-lot",
                Aliases:   []string{"r"},
                Usage:     "Recover unclaimed RPL from a lot (returning it to the auction contract)",
                UsageText: "rocketpool auction recover-lot",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return recoverRplFromLot(c)

                },
            },

        },
    })
}

