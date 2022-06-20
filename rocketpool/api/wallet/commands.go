package wallet

import (
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/utils/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register subcommands
func RegisterSubcommands(command *cli.Command, name string, aliases []string) {
	command.Subcommands = append(command.Subcommands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the node wallet",
		Subcommands: []cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get the node wallet status",
				UsageText: "rocketpool api wallet status",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(getStatus(c))
					return nil

				},
			},

			{
				Name:      "set-password",
				Aliases:   []string{"p"},
				Usage:     "Set the node wallet password",
				UsageText: "rocketpool api wallet set-password password",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					password, err := cliutils.ValidateNodePassword("wallet password", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(setPassword(c, password))
					return nil

				},
			},

			{
				Name:      "init",
				Aliases:   []string{"i"},
				Usage:     "Initialize the node wallet",
				UsageText: "rocketpool api wallet init",
				Flags: []cli.Flag{
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

					// Run
					api.PrintResponse(initWallet(c))
					return nil

				},
			},

			{
				Name:      "recover",
				Aliases:   []string{"r"},
				Usage:     "Recover a node wallet from a mnemonic phrase",
				UsageText: "rocketpool api wallet recover mnemonic",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "skip-validator-key-recovery, k",
						Usage: "Recover the node wallet, but do not regenerate its validator keys",
					},
					cli.StringFlag{
						Name:  "derivation-path, d",
						Usage: "Specify the derivation path for the wallet.\nOmit this flag (or leave it blank) for the default of \"m/44'/60'/0'/0/%d\" (where %d is the index).\nSet this to \"ledgerLive\" to use Ledger Live's path of \"m/44'/60'/%d/0/0\".\nSet this to \"mew\" to use MyEtherWallet's path of \"m/44'/60'/0'/%d\".\nFor custom paths, simply enter them here.",
					},
					cli.UintFlag{
						Name:  "wallet-index, i",
						Usage: "Specify the index to use with the derivation path when recovering your wallet",
						Value: 0,
					},
					cli.StringSliceFlag{
						Name:  "password",
						Usage: "Specify the password for a minipool's predefined validator private key file; format is `--password 0xabcd...=\"password goes here\"",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					mnemonic, err := cliutils.ValidateWalletMnemonic("mnemonic", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(recoverWallet(c, mnemonic))
					return nil

				},
			},

			{
				Name:      "search-and-recover",
				Aliases:   []string{"r"},
				Usage:     "Search for and recover a node wallet's derivation key and index using a mnemonic phrase and a well-known address.",
				UsageText: "rocketpool api wallet search-and-recover mnemonic address",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "skip-validator-key-recovery, k",
						Usage: "Recover the node wallet, but do not regenerate its validator keys",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					mnemonic, err := cliutils.ValidateWalletMnemonic("mnemonic", c.Args().Get(0))
					if err != nil {
						return err
					}
					address, err := cliutils.ValidateAddress("address", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(searchAndRecoverWallet(c, mnemonic, address))
					return nil

				},
			},

			{
				Name:      "rebuild",
				Aliases:   []string{"b"},
				Usage:     "Rebuild validator keystores from derived keys",
				UsageText: "rocketpool api wallet rebuild",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(rebuildWallet(c))
					return nil

				},
			},

			{
				Name:      "test-mnemonic",
				Aliases:   []string{"t"},
				Usage:     "Test recovering a node wallet from a mnemonic phrase to ensure the phrase is correct",
				UsageText: "rocketpool api wallet test-mnemonic mnemonic",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "derivation-path, d",
						Usage: "Specify the derivation path for the wallet.\nOmit this flag (or leave it blank) for the default of \"m/44'/60'/0'/0/%d\" (where %d is the index).\nSet this to \"ledgerLive\" to use Ledger Live's path of \"m/44'/60'/%d/0/0\".\nSet this to \"mew\" to use MyEtherWallet's path of \"m/44'/60'/0'/%d\".\nFor custom paths, simply enter them here.",
					},
					cli.UintFlag{
						Name:  "wallet-index, i",
						Usage: "Specify the index to use with the derivation path when recovering your wallet",
						Value: 0,
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					mnemonic, err := cliutils.ValidateWalletMnemonic("mnemonic", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(testMnemonic(c, mnemonic))
					return nil

				},
			},

			{
				Name:      "export",
				Aliases:   []string{"e"},
				Usage:     "Export the node wallet in JSON format",
				UsageText: "rocketpool api wallet export",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(exportWallet(c))
					return nil

				},
			},
		},
	})
}
