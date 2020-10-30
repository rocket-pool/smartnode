package wallet

import (
    "github.com/urfave/cli"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
    app.Commands = append(app.Commands, cli.Command{
        Name:      name,
        Aliases:   aliases,
        Usage:     "Manage the node wallet",
        Subcommands: []cli.Command{

            cli.Command{
                Name:      "status",
                Aliases:   []string{"s"},
                Usage:     "Get the node wallet status",
                UsageText: "rocketpool wallet status",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return getStatus(c)

                },
            },

            cli.Command{
                Name:      "init",
                Aliases:   []string{"i"},
                Usage:     "Initialize the node wallet",
                UsageText: "rocketpool wallet init",
                Flags: []cli.Flag{
                    cli.StringFlag{
                        Name:  "password, p",
                        Usage: "The password to secure the wallet with (if not already set)",
                    },
                    cli.BoolFlag{
                        Name:  "confirm-mnemonic, c",
                        Usage: "Automatically confirm the mnemonic phrase",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return initWallet(c)

                },
            },

            cli.Command{
                Name:      "recover",
                Aliases:   []string{"r"},
                Usage:     "Recover a node wallet from a mnemonic phrase",
                UsageText: "rocketpool wallet recover",
                Flags: []cli.Flag{
                    cli.StringFlag{
                        Name:  "password, p",
                        Usage: "The password to secure the wallet with (if not already set)",
                    },
                    cli.StringFlag{
                        Name:  "mnemonic, m",
                        Usage: "The mnemonic phrase to recover the wallet from",
                    },
                },
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return recoverWallet(c)

                },
            },

            cli.Command{
                Name:      "rebuild",
                Aliases:   []string{"b"},
                Usage:     "Rebuild validator keystores from derived keys",
                UsageText: "rocketpool wallet rebuild",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return rebuildWallet(c)

                },
            },

            cli.Command{
                Name:      "export",
                Aliases:   []string{"e"},
                Usage:     "Export the node wallet in JSON format",
                UsageText: "rocketpool wallet export",
                Action: func(c *cli.Context) error {

                    // Validate args
                    if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

                    // Run
                    return exportWallet(c)

                },
            },

        },
    })
}

