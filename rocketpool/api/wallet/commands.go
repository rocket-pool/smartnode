package wallet

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
        Usage:     "Manage the node wallet",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the node wallet status",
                UsageText: "rocketpool api wallet status",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(getStatus(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "set-password",
                Aliases:   []string{"p"},
                Usage:     "Set the node wallet password",
                UsageText: "rocketpool api wallet set-password password",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    password, err := cliutils.ValidateNodePassword("wallet password", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(setPassword(c, password))
                    return nil

                },
            },

            cli.Command{
                Name:      "init",
                Aliases:   []string{"i"},
                Usage:     "Initialize the node wallet",
                UsageText: "rocketpool api wallet init",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(initWallet(c))
                    return nil

                },
            },

            cli.Command{
                Name:      "recover",
                Aliases:   []string{"r"},
                Usage:     "Recover a node wallet from a mnemonic phrase",
                UsageText: "rocketpool api wallet recover mnemonic",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 1); err != nil { return err }
                    mnemonic, err := cliutils.ValidateWalletMnemonic("mnemonic", c.Args().Get(0))
                    if err != nil { return err }

                    // Run
                    api.PrintResponse(recoverWallet(c, mnemonic))
                    return nil

                },
            },

            cli.Command{
                Name:      "export",
                Aliases:   []string{"e"},
                Usage:     "Export the node wallet in JSON format",
                UsageText: "rocketpool api wallet export",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    api.PrintResponse(exportWallet(c))
                    return nil

                },
            },

        },
    })
}

