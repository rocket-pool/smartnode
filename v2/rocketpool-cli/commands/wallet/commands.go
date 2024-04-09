package wallet

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/input"
	cliutils "github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/shared/utils"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, &cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the node wallet",
		Subcommands: []*cli.Command{
			{
				Name:    "status",
				Aliases: []string{"s"},
				Usage:   "Get the node wallet status",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return getStatus(c)
				},
			},

			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Initialize the node wallet",
				Flags: []cli.Flag{
					PasswordFlag,
					initConfirmMnemonicFlag,
					derivationPathFlag,
					walletIndexFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String(PasswordFlag.Name) != "" {
						if _, err := input.ValidateNodePassword("password", c.String(PasswordFlag.Name)); err != nil {
							return err
						}
					}

					// Run
					return InitWallet(c, nil)
				},
			},

			{
				Name:    "recover",
				Aliases: []string{"r"},
				Usage:   "Recover a node wallet from a mnemonic phrase",
				Flags: []cli.Flag{
					PasswordFlag,
					mnemonicFlag,
					skipValidatorRecoveryFlag,
					derivationPathFlag,
					walletIndexFlag,
					addressFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String(PasswordFlag.Name) != "" {
						if _, err := input.ValidateNodePassword("password", c.String(PasswordFlag.Name)); err != nil {
							return err
						}
					}
					if c.String(mnemonicFlag.Name) != "" {
						if _, err := input.ValidateWalletMnemonic("mnemonic", c.String(mnemonicFlag.Name)); err != nil {
							return err
						}
					}

					// Run
					return recoverWallet(c)
				},
			},

			{
				Name:    "rebuild",
				Aliases: []string{"b"},
				Usage:   "Rebuild validator keystores from derived keys",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return rebuildWallet(c)
				},
			},

			{
				Name:    "test-recovery",
				Aliases: []string{"t"},
				Usage:   "Test recovering a node wallet without actually generating any of the node wallet or validator key files to ensure the process works as expected",
				Flags: []cli.Flag{
					mnemonicFlag,
					skipValidatorRecoveryFlag,
					derivationPathFlag,
					walletIndexFlag,
					addressFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String(mnemonicFlag.Name) != "" {
						if _, err := input.ValidateWalletMnemonic("mnemonic", c.String(mnemonicFlag.Name)); err != nil {
							return err
						}
					}

					// Run
					return testRecovery(c)
				},
			},

			{
				Name:    "set-password",
				Aliases: []string{"sp"},
				Usage:   "Upload the node wallet password to the daemon so it can unlock your node wallet keystore, optionally saving it to disk as well",
				Flags: []cli.Flag{
					PasswordFlag,
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return setPassword(c)
				},
			},

			{
				Name:    "delete-password",
				Aliases: []string{"dp"},
				Usage:   "Delete the node wallet password from disk, so it will need to be re-entered manually after each service or node restart.",
				Flags: []cli.Flag{
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return deletePassword(c)
				},
			},

			{
				Name:    "export",
				Aliases: []string{"e"},
				Usage:   "Export the node wallet in JSON format",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return exportWallet(c)
				},
			},

			{
				Name:    "export-as-eth-key",
				Aliases: []string{"ek"},
				Usage:   "Print the node wallet (encrypted with the wallet's password) in the JSON format used by eth-account and other tools for interoperability",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return exportEthKey(c)
				},
			},

			{
				Name:      "set-ens-name",
				Aliases:   []string{"ens"},
				Usage:     "Set a name to the node wallet's ENS reverse record",
				ArgsUsage: "name",
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 1); err != nil {
						return err
					}

					// Run
					return setEnsName(c, c.Args().Get(0))
				},
			},

			{
				Name:  "purge",
				Usage: fmt.Sprintf("%sDeletes your node wallet, your validator keys, and restarts your Validator Client while preserving your chain data. WARNING: Only use this if you want to stop validating with this machine!%s", terminal.ColorRed, terminal.ColorReset),
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return purge(c)
				},
			},

			{
				Name:    "sign-message",
				Aliases: []string{"sm"},
				Usage:   "Sign an arbitrary message with the node's private key",
				Flags: []cli.Flag{
					signMessageFlag,
				},
				Action: func(c *cli.Context) error {
					// Run
					return signMessage(c)
				},
			},

			{
				Name:      "send-message",
				Usage:     "Send a zero-ETH transaction to the target address (or ENS) with the provided hex-encoded message as the data payload",
				ArgsUsage: "to-address hex-message",
				Flags: []cli.Flag{
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 2); err != nil {
						return err
					}
					message, err := input.ValidateByteArray("message", c.Args().Get(1))
					if err != nil {
						return err
					}

					// Run
					return sendMessage(c, c.Args().Get(0), message)
				},
			},

			{
				Name:    "masquerade",
				Aliases: []string{"m"},
				Usage:   "Change your node's effective address to a different one, so your daemon will act as though you were that address. Your node will be in read-only mode while masquerading since you don't have the corresponding wallet's private key.",
				Flags: []cli.Flag{
					cliutils.YesFlag,
					masqueradeAddressFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return masquerade(c)
				},
			},

			{
				Name:      "restore-address",
				Aliases:   []string{"ra"},
				Usage:     "Restore your node's effective address back to your wallet address, ending a masquerade if you have one active. This will take it out of read-only mode.",
				ArgsUsage: "to-address hex-message",
				Flags: []cli.Flag{
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return restoreAddress(c)
				},
			},
		},
	})
}
