package wallet

import (
	"fmt"
	"os"
	"strings"

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
				UsageText: "rocketpool api wallet recover [options] [mnemonic]",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "skip-validator-key-recovery, k",
						Usage: "Recover the node wallet, but do not regenerate its validator keys",
					},
					cli.StringFlag{
						Name:  "derivation-path, d",
						Usage: "Specify the derivation path for the wallet.\nOmit this flag (or leave it blank) for the default of \"m/44'/60'/0'/0/%d\" (where %d is the index).\nSet this to \"ledgerLive\" to use Ledger Live's path of \"m/44'/60'/%d/0/0\".\nSet this to \"mew\" to use MyEtherWallet's path of \"m/44'/60'/0'/%d\".\nFor custom paths, simply enter them here.",
					},
					cli.StringFlag{
						Name:  "mnemonic-file, f",
						Usage: "Specify the path to the mnemonic.\nOmit this flag to enter the mnemonic via plain text.",
					},
					cli.UintFlag{
						Name:  "wallet-index, i",
						Usage: "Specify the index to use with the derivation path when recovering your wallet",
						Value: 0,
					},
				},
				Action: func(c *cli.Context) error {

					// Validate input
					// Must supply either --mnemonic-file or via stdin, but not both
					if (c.String("mnemonic-file") == "") == (c.Args().Get(0) == "") {
						return fmt.Errorf("Please specify a mnemonic file or mnemonic via stdin, but not both")
					}

					// Read mnemonic from file
					var providedMnemonic string
					if c.String("mnemonic-file") != "" {
						bytes, err := os.ReadFile(c.String("mnemonic-file"))
						if err != nil {
							return err
						}
						providedMnemonic = strings.TrimSpace(string(bytes))
					}

					// Read mnemonic from stdin
					if c.Args().Get(0) != "" {
						providedMnemonic = c.Args().Get(0)
					}

					// Validate mnemonic
					mnemonic, err := cliutils.ValidateWalletMnemonic("mnemonic", providedMnemonic)
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
				Name:      "test-recovery",
				Aliases:   []string{"r"},
				Usage:     "Test recovery of a node wallet and its validator keys without actually saving the recovered files",
				UsageText: "rocketpool api wallet test-recovery mnemonic",
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
					api.PrintResponse(testRecoverWallet(c, mnemonic))
					return nil

				},
			},

			{
				Name:      "test-search-and-recover",
				Aliases:   []string{"r"},
				Usage:     "Test searching for and recovery of a node wallet's derivation key, index, and validator keys using a mnemonic phrase and a well-known address.",
				UsageText: "rocketpool api wallet test-search-and-recover mnemonic address",
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
					api.PrintResponse(testSearchAndRecoverWallet(c, mnemonic, address))
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

			{
				Name:      "estimate-gas-set-ens-name",
				Usage:     "Estimate the gas required to set the name for the node wallet's ENS reverse record",
				UsageText: "rocketpool api node estimate-gas-set-ens-name name",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Run
					api.PrintResponse(setEnsName(c, c.Args().Get(0), true))
					return nil

				},
			},

			{
				Name:      "set-ens-name",
				Usage:     "Set a name to the node wallet's ENS reverse record",
				UsageText: "rocketpool api node set-ens-name name",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Run
					api.PrintResponse(setEnsName(c, c.Args().Get(0), false))
					return nil

				},
			},

			{
				Name:      "masquerade",
				Usage:     "Change your node's effective address to a different one. Your node will not be able to submit transactions or sign messages since you don't have the corresponding wallet's private key.",
				UsageText: "rocketpool api wallet masquerade address",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					address, err := cliutils.ValidateAddress("address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					api.PrintResponse(masquerade(c, address))
					return nil

				},
			},

			{
				Name:      "end-masquerade",
				Usage:     "End a masquerade, restoring your node's effective address back to your wallet address if one is loaded.",
				UsageText: "rocketpool api wallet end-masquerade",
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					api.PrintResponse(endMasquerade(c))
					return nil

				},
			},
		},
	})
}
