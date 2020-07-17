package account

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
        Usage:     "Manage the node account",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the node password and account status",
                UsageText: "rocketpool api account status",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(getStatus(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "init-password",
                Aliases:   []string{"p"},
                Usage:     "Initialize the node password",
                UsageText: "rocketpool api account init-password password",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    password, err := cliutils.ValidateNodePassword("node password", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(initPassword(c, password))
                    return nil

                },
            },

            cli.Command{
                Name:      "init-account",
                Aliases:   []string{"a"},
                Usage:     "Initialize the node account",
                UsageText: "rocketpool api account init-account",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(initAccount(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "export",
                Aliases:   []string{"e"},
                Usage:     "Export the node account in JSON format",
                UsageText: "rocketpool api account export",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(exportAccount(c))
                    return nil

                },
            },

        },
    })
}

