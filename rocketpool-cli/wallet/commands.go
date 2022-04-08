package wallet

import (
	"github.com/urfave/cli"

	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the node wallet",
		Subcommands: []cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get the node wallet status",
				UsageText: "rocketpool wallet status",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getStatus(c)

				},
			},

			{
				Name:      "init",
				Aliases:   []string{"i"},
				Usage:     "Initialize the node wallet",
				UsageText: "rocketpool wallet init [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "password, p",
						Usage: "The password to secure the wallet with (if not already set)",
					},
					cli.BoolFlag{
						Name:  "confirm-mnemonic, c",
						Usage: "Automatically confirm the mnemonic phrase",
					},
					cli.StringFlag{
						Name:  "derivation-path, d",
						Usage: "Specify the derivation path for the wallet.\nOmit this flag (or leave it blank) for the default of \"m/44'/60'/0'/0/%d\" (where %d is the index).\nSet this to \"ledgerLive\" to use Ledger Live's path of \"m/44'/60'/%d/0/0\".\nSet this to \"mew\" to use MyEtherWallet's path of \"m/44'/60'/0'/%d\".\nFor custom paths, simply enter them here.",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("password") != "" {
						if _, err := cliutils.ValidateNodePassword("password", c.String("password")); err != nil {
							return err
						}
					}

					// Run
					return initWallet(c)

				},
			},

			{
				Name:      "recover",
				Aliases:   []string{"r"},
				Usage:     "Recover a node wallet from a mnemonic phrase",
				UsageText: "rocketpool wallet recover [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "password, p",
						Usage: "The password to secure the wallet with (if not already set)",
					},
					cli.StringFlag{
						Name:  "mnemonic, m",
						Usage: "The mnemonic phrase to recover the wallet from",
					},
					cli.BoolFlag{
						Name:  "skip-validator-key-recovery, k",
						Usage: "Recover the node wallet, but do not regenerate its validator keys",
					},
					cli.StringFlag{
						Name:  "derivation-path, d",
						Usage: "Specify the derivation path for the wallet.\nOmit this flag (or leave it blank) for the default of \"m/44'/60'/0'/0/%d\" (where %d is the index).\nSet this to \"ledgerLive\" to use Ledger Live's path of \"m/44'/60'/%d/0/0\".\nSet this to \"mew\" to use MyEtherWallet's path of \"m/44'/60'/0'/%d\".\nFor custom paths, simply enter them here.",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("password") != "" {
						if _, err := cliutils.ValidateNodePassword("password", c.String("password")); err != nil {
							return err
						}
					}
					if c.String("mnemonic") != "" {
						if _, err := cliutils.ValidateWalletMnemonic("mnemonic", c.String("mnemonic")); err != nil {
							return err
						}
					}

					// Run
					return recoverWallet(c)

				},
			},

			{
				Name:      "rebuild",
				Aliases:   []string{"b"},
				Usage:     "Rebuild validator keystores from derived keys",
				UsageText: "rocketpool wallet rebuild",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return rebuildWallet(c)

				},
			},

			{
				Name:      "export",
				Aliases:   []string{"e"},
				Usage:     "Export the node wallet in JSON format",
				UsageText: "rocketpool wallet export",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return exportWallet(c)

				},
			},
		},
	})
}
