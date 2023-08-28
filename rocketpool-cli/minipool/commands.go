package minipool

import (
	"github.com/urfave/cli"

	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Manage the node's minipools",
		Subcommands: []cli.Command{

			{
				Name:      "status",
				Aliases:   []string{"s"},
				Usage:     "Get a list of the node's minipools",
				UsageText: "rocketpool minipool status",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "include-finalized, f",
						Usage: "Include finalized minipools in the list (default is to hide them).",
					},
				},
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
				Name:      "stake",
				Aliases:   []string{"t"},
				Usage:     "Stake a minipool after the scrub check, moving it from prelaunch to staking.",
				UsageText: "rocketpool minipool stake [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "minipool, m",
						Usage: "The minipool/s to stake (address or 'all')",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("minipool") != "" && c.String("minipool") != "all" {
						if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil {
							return err
						}
					}

					// Run
					return stakeMinipools(c)

				},
			},

			{
				Name:      "set-withdrawal-creds",
				Aliases:   []string{"swc"},
				Usage:     "Convert the withdrawal credentials for a migrated solo validator from the old 0x00 value to the minipool address. Required to complete the migration process.",
				UsageText: "rocketpool minipool set-withdrawal-creds minipool-address [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "mnemonic, m",
						Usage: "Use this flag to provide the mnemonic for your validator key instead of typing it interactively.",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					address, err := cliutils.ValidateAddress("minipool-address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					return setWithdrawalCreds(c, address)

				},
			},
			{
				Name:      "import-key",
				Aliases:   []string{"ik"},
				Usage:     "Import the externally-derived key for a minipool that was previously a solo validator, so the Smartnode's VC manages it instead of your externally-managed VC.",
				UsageText: "rocketpool minipool import-key minipool-address [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "mnemonic, m",
						Usage: "Use this flag to provide the mnemonic for your validator key instead of typing it interactively.",
					},
					cli.BoolFlag{
						Name:  "no-restart",
						Usage: "Don't restart the Validator Client after importing the key. Note that the key won't be loaded (and won't attest) until you restart the VC to load it.",
					},
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm all interactive questions",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					address, err := cliutils.ValidateAddress("minipool-address", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Run
					return importKey(c, address)

				},
			},
			{
				Name:      "promote",
				Aliases:   []string{"p"},
				Usage:     "Promote a vacant minipool after the scrub check, completing a solo validator migration.",
				UsageText: "rocketpool minipool promote [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "minipool, m",
						Usage: "The minipool/s to promote (address or 'all')",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("minipool") != "" && c.String("minipool") != "all" {
						if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil {
							return err
						}
					}

					// Run
					return promoteMinipools(c)

				},
			},

			{
				Name:      "refund",
				Aliases:   []string{"r"},
				Usage:     "Refund ETH belonging to the node from minipools",
				UsageText: "rocketpool minipool refund [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "minipool, m",
						Usage: "The minipool/s to refund from (address or 'all')",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("minipool") != "" && c.String("minipool") != "all" {
						if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil {
							return err
						}
					}

					// Run
					return refundMinipools(c)

				},
			},

			{
				Name:      "begin-bond-reduction",
				Aliases:   []string{"bbr"},
				Usage:     "Begins the ETH bond reduction process for a minipool, taking it from 16 ETH down to 8 ETH (begins conversion of a 16 ETH minipool to an LEB8)",
				UsageText: "rocketpool minipool begin-bond-reduction [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "minipool, m",
						Usage: "The minipool/s to begin the bond reduction for (address or 'all')",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("minipool") != "" && c.String("minipool") != "all" {
						if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil {
							return err
						}
					}

					// Run
					return beginReduceBondAmount(c)

				},
			},

			{
				Name:      "reduce-bond",
				Aliases:   []string{"rb"},
				Usage:     "Manually completes the ETH bond reduction process for a minipool from 16 ETH down to 8 ETH once it is eligible. Please run `begin-bond-reduction` first to start this process.",
				UsageText: "rocketpool minipool reduce-bond [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "minipool, m",
						Usage: "The minipool/s to reduce the bond for (address or 'all')",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("minipool") != "" && c.String("minipool") != "all" {
						if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil {
							return err
						}
					}

					// Run
					return reduceBondAmount(c)

				},
			},

			{
				Name:      "distribute-balance",
				Aliases:   []string{"d"},
				Usage:     "Distribute a minipool's ETH balance between your withdrawal address and the rETH holders.",
				UsageText: "rocketpool minipool distribute-balance [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "minipool, m",
						Usage: "The minipool/s to distribute the balance of (address or 'all')",
					},
					cli.Float64Flag{
						Name:  "threshold, t",
						Usage: "Filter on a minimum amount of ETH that can be distributed - minipools below this amount won't be shown",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("minipool") != "" && c.String("minipool") != "all" {
						if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil {
							return err
						}
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
			       UsageText: "rocketpool minipool dissolve [options]",
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
				Name:      "exit",
				Aliases:   []string{"e"},
				Usage:     "Exit staking minipools from the beacon chain",
				UsageText: "rocketpool minipool exit [options]",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm exiting minipool/s",
					},
					cli.StringFlag{
						Name:  "minipool, m",
						Usage: "The minipool/s to exit (address or 'all')",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("minipool") != "" && c.String("minipool") != "all" {
						if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil {
							return err
						}
					}

					// Run
					return exitMinipools(c)

				},
			},

			{
				Name:      "close",
				Aliases:   []string{"c"},
				Usage:     "Withdraw any remaining balance from a minipool and close it",
				UsageText: "rocketpool minipool close [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "minipool, m",
						Usage: "The minipool/s to close (address or 'all')",
					},
					cli.BoolFlag{
						Name:  "confirm-slashing",
						Usage: "Reserved for acknowledging situations where you've been slashed by the Beacon Chain, and closing a minipool will result in the complete loss of the ETH bond and your RPL collateral. DO NOT use this flag unless you have been explicitly instructed to do so.",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("minipool") != "" && c.String("minipool") != "all" {
						if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil {
							return err
						}
					}

					// Run
					return closeMinipools(c)

				},
			},

			{
				Name:      "delegate-upgrade",
				Aliases:   []string{"u"},
				Usage:     "Upgrade a minipool's delegate contract to the latest version",
				UsageText: "rocketpool minipool delegate-upgrade [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "minipool, m",
						Usage: "The minipool/s to upgrade (address or 'all')",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("minipool") != "" && c.String("minipool") != "all" {
						if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil {
							return err
						}
					}

					// Run
					return delegateUpgradeMinipools(c)

				},
			},

			{
				Name:      "delegate-rollback",
				Aliases:   []string{"b"},
				Usage:     "Roll a minipool's delegate contract back to its previous version",
				UsageText: "rocketpool minipool delegate-rollback [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "minipool, m",
						Usage: "The minipool/s to rollback (address or 'all')",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("minipool") != "" && c.String("minipool") != "all" {
						if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil {
							return err
						}
					}

					// Run
					return delegateRollbackMinipools(c)

				},
			},

			{
				Name:      "set-use-latest-delegate",
				Aliases:   []string{"l"},
				Usage:     "Use this to enable or disable the \"use-latest-delegate\" flag on one or more minipools. If enabled, the minipool will ignore its current delegate contract and always use whatever the latest delegate is.",
				UsageText: "rocketpool minipool set-use-latest-delegate [options] true/false",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "minipool, m",
						Usage: "The minipool/s to configure the use-latest setting on (address or 'all')",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 1); err != nil {
						return err
					}
					setting, err := cliutils.ValidateBool("setting", c.Args().Get(0))
					if err != nil {
						return err
					}

					// Validate flags
					if c.String("minipool") != "" && c.String("minipool") != "all" {
						if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil {
							return err
						}
					}

					// Run
					return setUseLatestDelegateMinipools(c, setting)

				},
			},

			{
				Name:      "find-vanity-address",
				Aliases:   []string{"v"},
				Usage:     "Search for a custom vanity minipool address",
				UsageText: "rocketpool minipool find-vanity-address [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "prefix, p",
						Usage: "The prefix of the address to search for (must start with 0x)",
					},
					cli.StringFlag{
						Name:  "salt, s",
						Usage: "The salt to start searching from (must start with 0x)",
					},
					cli.IntFlag{
						Name:  "threads, t",
						Usage: "The number of threads to use for searching (defaults to your CPU thread count)",
					},
					cli.StringFlag{
						Name:  "node-address, n",
						Usage: "The node address to search for (leave blank to use the local node)",
					},
					cli.StringFlag{
						Name:  "amount, a",
						Usage: "The bond amount to be used for the minipool, in ETH (impacts vanity address generation)",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags

					// Run
					return findVanitySalt(c)

				},
			},

			{
				Name:      "rescue-dissolved",
				Aliases:   []string{"rd"},
				Usage:     "Manually deposit ETH into the Beacon deposit contract for a dissolved minipool, activating it on the Beacon Chain so it can be exited.",
				UsageText: "rocketpool minipool rescue-dissolved [options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "minipool, m",
						Usage: "The minipool/s to rescue (address, starting with 0x)",
					},
					cli.StringFlag{
						Name:  "amount, a",
						Usage: "The amount of ETH to deposit into the minipool",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Validate flags
					if c.String("minipool") != "" {
						if _, err := cliutils.ValidateAddress("minipool address", c.String("minipool")); err != nil {
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
