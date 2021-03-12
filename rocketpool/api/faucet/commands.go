package faucet

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/utils/api"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
    command.Subcommands = append(command.Subcommands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Access the RPL faucet",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the faucet's status",
                UsageText: "rocketpool api faucet status",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(getStatus(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "can-get-rpl",
                Usage:     "Check whether the node can get RPL from the faucet",
                UsageText: "rocketpool api faucet can-get-rpl",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(canGetRpl(c))
                    return nil

                },
            },
            cli.Command{
                Name:      "get-rpl",
                Aliases:   []string{"g"},
                Usage:     "Get RPL from the faucet",
                UsageText: "rocketpool api faucet get-rpl",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(getRpl(c))
                    return nil

                },
            },

        },
    })
}

