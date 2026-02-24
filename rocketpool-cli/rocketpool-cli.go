package main

import (
	"fmt"
	"math/big"
	"os"

	"github.com/urfave/cli"

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
	"github.com/rocket-pool/smartnode/rocketpool-cli/wallet"
	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

const (
	colorReset    string = "\033[0m"
	colorYellow   string = "\033[33m"
	maxAlertItems int    = 3
)

// Run
func main() {

	// Add logo to application help template
	cli.AppHelpTemplate = fmt.Sprintf(`
%s

Authored by the Rocket Pool Core Team
A special thanks to the Rocket Pool community for all their contributions.

%s`, shared.Logo(), cli.AppHelpTemplate)

	// Initialise application
	app := cli.NewApp()

	// Set application info
	app.Name = "rocketpool"
	app.Usage = "Rocket Pool CLI"
	app.Version = shared.RocketPoolVersion()
	app.Copyright = "(c) 2026 Rocket Pool Pty Ltd"

	// Initialize app metadata
	app.Metadata = make(map[string]interface{})

	// Set application flags
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "allow-root, r",
			Usage: "Allow rocketpool to be run as the root user",
		},
		cli.StringFlag{
			Name:  "config-path, c",
			Usage: "Rocket Pool config asset `path`",
			Value: "~/.rocketpool",
		},
		cli.StringFlag{
			Name:  "daemon-path, d",
			Usage: "Interact with a Rocket Pool service daemon at a `path` on the host OS, running outside of docker",
		},
		cli.Float64Flag{
			Name:  "maxFee, f",
			Usage: "The max fee (including the priority fee) you want a transaction to cost, in gwei",
		},
		cli.Float64Flag{
			Name:  "maxPrioFee, i",
			Usage: "The max priority fee you want a transaction to use, in gwei",
		},
		cli.Uint64Flag{
			Name:  "gasLimit, l",
			Usage: "[DEPRECATED] Desired gas limit",
		},
		cli.StringFlag{
			Name:  "nonce",
			Usage: "Use this flag to explicitly specify the nonce that this transaction should use, so it can override an existing 'stuck' transaction",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Enable debug printing of API commands",
		},
		cli.BoolFlag{
			Name: "secure-session, s",
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

	app.Before = func(c *cli.Context) error {
		// Check user ID
		if os.Getuid() == 0 && !c.GlobalBool("allow-root") {
			fmt.Fprintln(os.Stderr, "rocketpool should not be run as root. Please try again without 'sudo'.")
			fmt.Fprintln(os.Stderr, "If you want to run rocketpool as root anyway, use the '--allow-root' option to override this warning.")
			os.Exit(1)
		}

		// If set, validate custom nonce
		customNonce := c.GlobalString("nonce")
		if customNonce != "" {
			nonce, ok := big.NewInt(0).SetString(customNonce, 0)
			if !ok {
				fmt.Fprintf(os.Stderr, "Invalid nonce: %s\n", customNonce)
				os.Exit(1)
			}

			// Save the parsed value on Metadata so we don't need to reparse it later
			c.App.Metadata["nonce"] = nonce
		}

		return nil
	}

	app.After = func(c *cli.Context) error {
		// Skip alert display when no subcommand was actually invoked (e.g. --help, --version).
		if !c.Args().Present() {
			return nil
		}

		rp := rocketpool.NewClientFromCtx(c)
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
			fmt.Printf("\n%s=== Alerts ===%s\n", colorYellow, colorReset)
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
	if err := app.Run(os.Args); err != nil {
		cliutils.PrettyPrintError(err)
	}

	fmt.Println("")

}
