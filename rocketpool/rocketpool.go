package main

import (
	"context"
	"fmt"
	"os"

	cli "github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/node"
	"github.com/rocket-pool/smartnode/rocketpool/watchtower"
	"github.com/rocket-pool/smartnode/shared"

	blsversionpin "github.com/herumi/bls-eth-go-binary/bls"
)

// This variable simple ensures we have a direct dependency on bls-eth-go-binary, so it doesn't get tidied out of go.mod
// we should probably re-test latest versions of the herumi packages on ARM to see if the issue Joe encountered has been
// resolved yet.
var _ blsversionpin.ID

// Run
func main() {

	// Initialise application
	app := &cli.Command{
		Name:      "rocketpool",
		Usage:     "Rocket Pool service",
		Version:   shared.RocketPoolVersion(),
		Copyright: "(c) 2026 Rocket Pool Pty Ltd",
	}

	// Set application flags
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "settings",
			Aliases: []string{"s"},
			Usage:   "Rocket Pool service user config absolute `path`",
			Value:   "/.rocketpool/user-settings.yml",
		},
		&cli.Float64Flag{
			Name:  "maxFee",
			Usage: "Desired max fee in gwei",
		},
		&cli.Float64Flag{
			Name:  "maxPrioFee",
			Usage: "Desired max priority fee in gwei",
		},
		&cli.Uint64Flag{
			Name:    "gasLimit",
			Aliases: []string{"l"},
			Usage:   "Desired gas limit",
		},
		&cli.StringFlag{
			Name:    "metricsAddress",
			Aliases: []string{"m"},
			Usage:   "Address to serve metrics on if enabled",
			Value:   "0.0.0.0",
		},
		&cli.Uint64Flag{
			Name:    "metricsPort",
			Aliases: []string{"r"},
			Usage:   "Port to serve metrics on if enabled",
		},
		&cli.BoolFlag{
			Name:  "ignore-sync-check",
			Usage: "Set this to true if you already checked the sync status of the execution client(s) and don't need to re-check it for this command",
		},
		&cli.BoolFlag{
			Name:  "force-fallbacks",
			Usage: "Set this to true if you know the primary EC or CC is offline and want to bypass its health checks, and just use the fallback EC and CC instead",
		},
		&cli.BoolFlag{
			Name:  "use-protected-api",
			Usage: "Set this to true to use the Flashbots Protect RPC instead of your local Execution Client. Useful to ensure your transactions aren't front-run.",
		},
	}

	// Register commands
	node.RegisterCommands(app, "node", []string{"n"})
	watchtower.RegisterCommands(app, "watchtower", []string{"w"})

	// Run application
	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}
