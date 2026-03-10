package main

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool-cli/auction"
	"github.com/rocket-pool/smartnode/rocketpool-cli/claims"
	"github.com/rocket-pool/smartnode/rocketpool-cli/megapool"
	"github.com/rocket-pool/smartnode/rocketpool-cli/minipool"
	"github.com/rocket-pool/smartnode/rocketpool-cli/network"
	"github.com/rocket-pool/smartnode/rocketpool-cli/node"
	"github.com/rocket-pool/smartnode/rocketpool-cli/odao"
	"github.com/rocket-pool/smartnode/rocketpool-cli/pdao"
	"github.com/rocket-pool/smartnode/rocketpool-cli/queue"
	"github.com/rocket-pool/smartnode/rocketpool-cli/security"
	"github.com/rocket-pool/smartnode/rocketpool-cli/service"
	"github.com/rocket-pool/smartnode/rocketpool-cli/update"
	"github.com/rocket-pool/smartnode/rocketpool-cli/wallet"
	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
)

const (
	maxAlertItems int = 3
)

// Run
func main() {

	app := &cli.Command{
		Name:                  "rocketpool",
		Usage:                 "Rocket Pool CLI",
		Version:               shared.RocketPoolVersion(),
		EnableShellCompletion: true,
		Copyright:             "(c) 2026 Rocket Pool Pty Ltd",
		CustomRootCommandHelpTemplate: fmt.Sprintf(`%s
		Authored by the Rocket Pool Core Team
		A special thanks to the Rocket Pool community for all their contributions.
		%s`, shared.Logo(), cli.RootCommandHelpTemplate),
	}

	// Initialize app metadata
	app.Metadata = make(map[string]interface{})

	// Set application flags
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "allow-root",
			Aliases: []string{"r"},
			Usage:   "Allow rocketpool to be run as the root user",
		},
		&cli.StringFlag{
			Name:    "config-path",
			Aliases: []string{"c"},
			Usage:   "Rocket Pool config asset `path`",
			Value:   "~/.rocketpool",
		},
		&cli.StringFlag{
			Name:    "daemon-path",
			Aliases: []string{"d"},
			Usage:   "Interact with a Rocket Pool service daemon at a `path` on the host OS, running outside of docker",
		},
		&cli.Float64Flag{
			Name:    "maxFee",
			Aliases: []string{"f"},
			Usage:   "The max fee (including the priority fee) you want a transaction to cost, in gwei",
		},
		&cli.Float64Flag{
			Name:    "maxPrioFee",
			Aliases: []string{"i"},
			Usage:   "The max priority fee you want a transaction to use, in gwei",
		},
		&cli.Uint64Flag{
			Name:    "gasLimit",
			Aliases: []string{"l"},
			Usage:   "[DEPRECATED] Desired gas limit",
		},
		&cli.StringFlag{
			Name:    "nonce",
			Aliases: []string{"n"},
			Usage:   "Use this flag to explicitly specify the nonce that this transaction should use, so it can override an existing 'stuck' transaction",
		},
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "Enable debug printing of API commands",
		},
		&cli.BoolFlag{
			Name:    "secure-session",
			Aliases: []string{"s"},
			Usage: "Some commands may print sensitive information to your terminal. " +
				"Use this flag when nobody can see your screen to allow sensitive data to be printed without prompting",
		},
	}

	// Register commands
	auction.RegisterCommands(app, "auction", []string{"a"})
	claims.RegisterCommands(app, "claims", []string{"l"})
	minipool.RegisterCommands(app, "minipool", []string{"m"})
	megapool.RegisterCommands(app, "megapool", []string{"g"})
	network.RegisterCommands(app, "network", []string{"e"})
	node.RegisterCommands(app, "node", []string{"n"})
	odao.RegisterCommands(app, "odao", []string{"o"})
	pdao.RegisterCommands(app, "pdao", []string{"p"})
	queue.RegisterCommands(app, "queue", []string{"q"})
	security.RegisterCommands(app, "security", []string{"c"})
	service.RegisterCommands(app, "service", []string{"s"})
	wallet.RegisterCommands(app, "wallet", []string{"w"})

	// Add a command that updates the smart node cli.
	app.Commands = append(app.Commands, &cli.Command{
		Name:  "update",
		Usage: "Update the cli binary",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Automatically confirm the update",
			},
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Force the update even if the current version is the latest",
			},
			&cli.BoolFlag{
				Name:    "skip-signature-verification",
				Aliases: []string{"s"},
				Usage:   "Don't verify the singature of the release",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return update.Update(
				c.Bool("yes"),
				c.Bool("skip-signature-verification"),
				c.Bool("force"),
			)
		},
	})

	app.Before = func(ctx context.Context, c *cli.Command) (context.Context, error) {
		// Check user ID
		if os.Getuid() == 0 && !c.Root().Bool("allow-root") {
			fmt.Fprintln(os.Stderr, "rocketpool should not be run as root. Please try again without 'sudo'.")
			fmt.Fprintln(os.Stderr, "If you want to run rocketpool as root anyway, use the '--allow-root' option to override this warning.")
			os.Exit(1)
		}

		Defaults := rocketpool.Globals{
			ConfigPath: os.ExpandEnv(c.Root().String("config-path")),
			DaemonPath: os.ExpandEnv(c.Root().String("daemon-path")),
			MaxFee:     c.Root().Float64("maxFee"),
			MaxPrioFee: c.Root().Float64("maxPrioFee"),
			GasLimit:   c.Root().Uint64("gasLimit"),
			DebugPrint: c.Root().Bool("debug"),
		}

		// If set, validate custom nonce
		customNonce := c.Root().String("nonce")
		if customNonce != "" {
			nonce, ok := big.NewInt(0).SetString(customNonce, 0)
			if !ok {
				fmt.Fprintf(os.Stderr, "Invalid nonce: %s\n", customNonce)
				os.Exit(1)
			}

			Defaults.CustomNonce = nonce
		}

		rocketpool.SetDefaults(Defaults)

		return ctx, nil
	}

	app.After = func(ctx context.Context, c *cli.Command) error {
		// Skip alert display when no subcommand was actually invoked (e.g. --help, --version).
		if !c.Args().Present() {
			return nil
		}

		rp := rocketpool.NewClient()
		defer rp.Close()

		// Check if the user has enabled the "show alerts after every command" setting.
		// Errors here are intentionally swallowed — config may not exist yet.
		cfg, _, err := rp.LoadConfig()
		if err != nil || cfg.Alertmanager.ShowAlertsOnCLI.Value != true {
			return nil
		}

		// Fetch alerts through the daemon so it works in both Docker and native mode.
		// Errors here are intentionally swallowed — alerts are informational and must
		// never obscure the result of the primary command.
		response, err := rp.NodeAlerts()
		if err != nil {
			return nil
		}

		if len(response.Alerts) > 0 {
			color.YellowPrintln("=== Alerts ===")
			for i, alert := range response.Alerts {
				fmt.Println(alert.ColorString())
				if i == maxAlertItems-1 {
					break
				}
			}
			if len(response.Alerts) > maxAlertItems {
				fmt.Printf("... and %d more.\n", len(response.Alerts)-maxAlertItems)
			}
		}

		return nil
	}

	// Run application
	fmt.Println("")
	if err := app.Run(context.Background(), os.Args); err != nil {
		cliutils.PrettyPrintError(err)
	}

	fmt.Println("")

}
