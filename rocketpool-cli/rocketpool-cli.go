package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/auction"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/minipool"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/network"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/node"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/odao"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/pdao"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/queue"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/security"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/service"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/wallet"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/settings"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/shared"
)

// allowRootFlag is the only one this file deals with- simply so it can exit early.
var (
	allowRootFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:    "allow-root",
		Aliases: []string{"r"},
		Usage:   "Allow rocketpool to be run as the root user",
	}
)

// Run
func main() {
	// Add logo to application help template
	cli.AppHelpTemplate = fmt.Sprintf(`
______           _        _    ______           _ 
| ___ \         | |      | |   | ___ \         | |
| |_/ /___   ___| | _____| |_  | |_/ /__   ___ | |
|    // _ \ / __| |/ / _ \ __| |  __/ _ \ / _ \| |
| |\ \ (_) | (__|   <  __/ |_  | | | (_) | (_) | |
\_| \_\___/ \___|_|\_\___|\__| \_|  \___/ \___/|_|

%s`, cli.AppHelpTemplate)

	// Initialise application
	app := cli.NewApp()

	// Set application info
	app.Name = "rocketpool"
	app.Usage = "Smart Node CLI for Rocket Pool"
	app.Version = shared.RocketPoolVersion
	app.Authors = []*cli.Author{
		{
			Name:  "David Rugendyke",
			Email: "david@rocketpool.net",
		},
		{
			Name:  "Jake Pospischil",
			Email: "jake@rocketpool.net",
		},
		{
			Name:  "Joe Clapis",
			Email: "joe@rocketpool.net",
		},
		{
			Name:  "Kane Wallmann",
			Email: "kane@rocketpool.net",
		},
	}
	app.Copyright = "(c) 2024 Rocket Pool Pty Ltd"

	// Initialize app metadata
	app.Metadata = make(map[string]interface{})

	// Set allowedRootFlag
	app.Flags = []cli.Flag{
		allowRootFlag,
	}

	// Set global smart node flags
	app.Flags = settings.AppendSmartNodeSettingsFlags(app.Flags)

	// Set utility flags
	app.Flags = utils.AppendFlags(app.Flags)

	// Register commands
	auction.RegisterCommands(app, "auction", []string{"a"})
	minipool.RegisterCommands(app, "minipool", []string{"m"})
	network.RegisterCommands(app, "network", []string{"e"})
	node.RegisterCommands(app, "node", []string{"n"})
	odao.RegisterCommands(app, "odao", []string{"o"})
	pdao.RegisterCommands(app, "pdao", []string{"p"})
	queue.RegisterCommands(app, "queue", []string{"q"})
	security.RegisterCommands(app, "security", []string{"c"})
	service.RegisterCommands(app, "service", []string{"s"})
	wallet.RegisterCommands(app, "wallet", []string{"w"})

	var snSettings *settings.SmartNodeSettings
	app.Before = func(c *cli.Context) error {
		// Check user ID
		if os.Getuid() == 0 && !c.Bool(allowRootFlag.Name) {
			fmt.Fprintln(os.Stderr, "rocketpool should not be run as root. Please try again without 'sudo'.")
			fmt.Fprintf(os.Stderr, "If you want to run rocketpool as root anyway, use the '--%s' option to override this warning.\n", allowRootFlag.Name)
			os.Exit(1)
		}

		var err error
		snSettings, err = settings.NewSmartNodeSettings(c)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			os.Exit(1)
		}
		return nil
	}

	// Run application
	if snSettings != nil && snSettings.HttpTraceFile != nil {
		defer snSettings.HttpTraceFile.Close()
	}
	fmt.Println()
	if err := app.Run(os.Args); err != nil {
		utils.PrettyPrintError(err)
		os.Exit(1)
	}
	fmt.Println()
}
