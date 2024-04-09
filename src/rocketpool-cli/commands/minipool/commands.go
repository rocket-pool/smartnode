package minipool

import (
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/input"
	cliutils "github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/shared/utils"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, &cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the node's minipools",
		Subcommands: []*cli.Command{
			{
				Name:    "status",
				Aliases: []string{"s"},
				Usage:   "Get a list of the node's minipools",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    statusIncludeFinalizedFlag,
						Aliases: []string{"f"},
						Usage:   "Include finalized minipools in the list (default is to hide them).",
					},
				},
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
				Name:    "stake",
				Aliases: []string{"t"},
				Usage:   "Stake a minipool after the scrub check, moving it from prelaunch to staking.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    minipoolsFlag,
						Aliases: []string{"m"},
						Usage:   "A comma-separated list of addresses for minipools to stake (or 'all' to stake all available minipools)",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return stakeMinipools(c)
				},
			},

			{
				Name:    "set-withdrawal-creds",
				Aliases: []string{"swc"},
				Usage:   "Convert the withdrawal credentials for a migrated solo validator from the old 0x00 value to the minipool address. Required to complete the migration process.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    cliutils.MnemonicFlag,
						Aliases: []string{"m"},
						Usage:   "Use this flag to provide the mnemonic for your validator key instead of typing it interactively.",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					address, err := input.ValidateAddress("minipool-address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					return setWithdrawalCreds(c, address)
				},
			},
			{
				Name:    "import-key",
				Aliases: []string{"ik"},
				Usage:   "Import the externally-derived key for a minipool that was previously a solo validator, so the Smartnode's VC manages it instead of your externally-managed VC.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    cliutils.MnemonicFlag,
						Aliases: []string{"m"},
						Usage:   "Use this flag to provide the mnemonic for your validator key instead of typing it interactively.",
					},
					&cli.BoolFlag{
						Name:  cliutils.NoRestartFlag,
						Usage: "Don't restart the Validator Client after importing the key. Note that the key won't be loaded (and won't attest) until you restart the VC to load it.",
					},
					cliutils.YesFlag,
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					address, err := input.ValidateAddress("minipool-address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					return importKey(c, address)
				},
			},
			{
				Name:    "promote",
				Aliases: []string{"p"},
				Usage:   "Promote a vacant minipool after the scrub check, completing a solo validator migration.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    minipoolsFlag,
						Aliases: []string{"m"},
						Usage:   "The comma-separated addresses of the minipools to promote (or 'all' for every available minipool)",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return promoteMinipools(c)
				},
			},

			{
				Name:    "refund",
				Aliases: []string{"r"},
				Usage:   "Refund ETH belonging to the node from minipools",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    minipoolsFlag,
						Aliases: []string{"m"},
						Usage:   "The comma-separated addresses of the minipools to refund (or 'all' for every available minipool)",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return refundMinipools(c)
				},
			},

			{
				Name:    "begin-bond-reduction",
				Aliases: []string{"bbr"},
				Usage:   "Begins the ETH bond reduction process for a minipool, taking it from 16 ETH down to 8 ETH (begins conversion of a 16 ETH minipool to an LEB8)",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    minipoolsFlag,
						Aliases: []string{"m"},
						Usage:   "The comma-separated addresses of the minipools to begin the bond reduction for (or 'all' for every available minipool)",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return beginReduceBondAmount(c)
				},
			},

			{
				Name:    "reduce-bond",
				Aliases: []string{"rb"},
				Usage:   "Manually completes the ETH bond reduction process for a minipool from 16 ETH down to 8 ETH once it is eligible. Please run `begin-bond-reduction` first to start this process.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    minipoolsFlag,
						Aliases: []string{"m"},
						Usage:   "The comma-separated addresses of the minipools to reduce the bond for (or 'all' for every available minipool)",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return reduceBondAmount(c)
				},
			},

			{
				Name:    "distribute-balance",
				Aliases: []string{"d"},
				Usage:   "Distribute a minipool's ETH balance between your withdrawal address and the rETH holders.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    minipoolsFlag,
						Aliases: []string{"m"},
						Usage:   "The comma-separated addresses of the minipools to distribute the balance of (or 'all' for every available minipool)",
					},
					&cli.Float64Flag{
						Name:    distributeThresholdFlag,
						Aliases: []string{"t"},
						Usage:   "Filter on a minimum amount of ETH that can be distributed - minipools below this amount won't be shown",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return distributeBalance(c)
				},
			},

			/*
			   REMOVED UNTIL BEACON WITHDRAWALS
			   cli.Command{
			       Name:      "dissolve",
			       Aliases:   []string{"d"},
			       Usage:     "Dissolve initialized or prelaunch minipools",
			       Flags: []cli.Flag{
			           cli.BoolFlag{
			               Name:  "yes, y",
			               Usage: "Automatically confirm dissolving minipool/s",
			           },
			           cli.StringFlag{
			               Name:  "minipool, m",
			               Usage: "The minipool/s to dissolve (address or 'all')",
			           },
			       },
			       Action: func(c *cli.Context) error {

			           // Validate args
			           if err := cliutils.ValidateArgCount(c, 0); err != nil { return err }

			           // Validate flags
			           if c.String("minipool") != "" && c.String("minipool") != "all" {
			               if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil { return err }
			           }

			           // Run
			           return dissolveMinipools(c)

			       },
			   },
			*/
			{
				Name:    "exit",
				Aliases: []string{"e"},
				Usage:   "Exit staking minipools from the beacon chain",
				Flags: []cli.Flag{
					cliutils.YesFlag,
					&cli.StringFlag{
						Name:    minipoolsFlag,
						Aliases: []string{"m"},
						Usage:   "The comma-separated addresses of the minipools to exit (or 'all' to exit every available minipool)",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return exitMinipools(c)
				},
			},

			{
				Name:    "close",
				Aliases: []string{"c"},
				Usage:   "Withdraw any remaining balance from a minipool and close it",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    minipoolsFlag,
						Aliases: []string{"m"},
						Usage:   "The comma-separated addresses of the minipools to close (or 'all' to close every available minipool)",
					},
					&cli.BoolFlag{
						Name:  closeConfirmFlag,
						Usage: "Reserved for acknowledging situations where you've been slashed by the Beacon Chain, and closing a minipool will result in the complete loss of the ETH bond and your RPL collateral. DO NOT use this flag unless you have been explicitly instructed to do so.",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return closeMinipools(c)
				},
			},

			{
				Name:    "delegate-upgrade",
				Aliases: []string{"u"},
				Usage:   "Upgrade a minipool's delegate contract to the latest version",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    minipoolsFlag,
						Aliases: []string{"m"},
						Usage:   "The comma-separated addresses of the minipools to upgrade (or 'all' to upgrade every available minipool)",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return upgradeDelegates(c)
				},
			},

			{
				Name:    "delegate-rollback",
				Aliases: []string{"b"},
				Usage:   "Roll a minipool's delegate contract back to its previous version",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    minipoolsFlag,
						Aliases: []string{"m"},
						Usage:   "The comma-separated addresses of the minipools to rollback (or 'all' to rollback every available minipool)",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return rollbackDelegates(c)
				},
			},

			{
				Name:      "set-use-latest-delegate",
				Aliases:   []string{"l"},
				Usage:     "Use this to enable or disable the \"use-latest-delegate\" flag on one or more minipools. If enabled, the minipool will ignore its current delegate contract and always use whatever the latest delegate is.",
				ArgsUsage: "true/false",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    minipoolsFlag,
						Aliases: []string{"m"},
						Usage:   "The comma-separated addresses of the minipools to set the use-latest setting for (or 'all' to set it on every available minipool)",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					setting, err := input.ValidateBool("setting", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					return setUseLatestDelegates(c, setting)
				},
			},

			{
				Name:    "find-vanity-address",
				Aliases: []string{"v"},
				Usage:   "Search for a custom vanity minipool address",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    vanityPrefixFlag,
						Aliases: []string{"p"},
						Usage:   "The prefix of the address to search for (must start with 0x)",
					},
					&cli.StringFlag{
						Name:    vanitySaltFlag,
						Aliases: []string{"s"},
						Usage:   "The salt to start searching from (must start with 0x)",
					},
					&cli.IntFlag{
						Name:    vanityThreadsFlag,
						Aliases: []string{"t"},
						Usage:   "The number of threads to use for searching (defaults to your CPU thread count)",
					},
					&cli.StringFlag{
						Name:    vanityAddressFlag,
						Aliases: []string{"n"},
						Usage:   "The node address to search for (leave blank to use the local node)",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run
					return findVanitySalt(c)
				},
			},

			{
				Name:    "rescue-dissolved",
				Aliases: []string{"rd"},
				Usage:   "Manually deposit ETH into the Beacon deposit contract for a dissolved minipool, activating it on the Beacon Chain so it can be exited.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    rescueMinipoolFlag,
						Aliases: []string{"m"},
						Usage:   "The minipool/s to rescue (address, starting with 0x)",
					},
					&cli.StringFlag{
						Name:    rescueAmountFlag,
						Aliases: []string{"a"},
						Usage:   "The amount of ETH to deposit into the minipool",
					},
				},
				Action: func(c *cli.Context) error {
					// Validate args
					if err := utils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String(rescueMinipoolFlag) != "" {
						if _, err := input.ValidateAddress("minipool address", c.String("minipool")); err != nil {
							return err
						}
					}

					// Run
					return rescueDissolved(c)
				},
			},
		},
	})
}
