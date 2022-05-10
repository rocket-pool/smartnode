package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/rocketpool/api"
	"github.com/rocket-pool/smartnode/rocketpool/node"
	"github.com/rocket-pool/smartnode/rocketpool/watchtower"
	"github.com/rocket-pool/smartnode/shared"
	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
)

// Run
func main() {

	// Initialise application
	app := cli.NewApp()

	// Set application info
	app.Name = "rocketpool"
	app.Usage = "Rocket Pool service"
	app.Version = shared.RocketPoolVersion
	app.Authors = []cli.Author{
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
	app.Copyright = "(c) 2021 Rocket Pool Pty Ltd"

	// Set application flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "settings, s",
			Usage: "Rocket Pool service user config absolute `path`",
			Value: "/.rocketpool/user-settings.yml",
		},
		cli.StringFlag{
			Name:  "watchtowerFolder, w",
			Usage: "Absolute path to the directory the watchtower stores persistent state",
			Value: "/.rocketpool/watchtower",
		},
		cli.StringFlag{
			Name:  "storageAddress, a",
			Usage: "Rocket Pool storage contract `address`",
		},
		cli.StringFlag{
			Name:  "oneInchOracleAddress, o",
			Usage: "1inch exchange oracle contract `address`",
		},
		cli.StringFlag{
			Name:  "rplTokenAddress, t",
			Usage: "RPL token contract `address`",
		},
		cli.StringFlag{
			Name:  "rplFaucetAddress, f",
			Usage: "Rocket Pool RPL token faucet `address`",
		},
		cli.StringFlag{
			Name:  "password, p",
			Usage: "Rocket Pool wallet password file absolute `path`",
		},
		cli.StringFlag{
			Name:  "wallet, w",
			Usage: "Rocket Pool wallet file absolute `path`",
		},
		cli.StringFlag{
			Name:  "validatorKeychain, k",
			Usage: "Rocket Pool validator keychain absolute `path`",
		},
		cli.StringFlag{
			Name:  "eth1Provider, e",
			Usage: "Eth 1.0 provider `address`",
		},
		cli.StringFlag{
			Name:  "eth2Provider, b",
			Usage: "Eth 2.0 provider `address`",
		},
		cli.Float64Flag{
			Name:  "maxFee",
			Usage: "Desired max fee in gwei",
		},
		cli.Float64Flag{
			Name:  "maxPrioFee",
			Usage: "Desired max priority fee in gwei",
		},
		cli.Uint64Flag{
			Name:  "gasLimit, l",
			Usage: "Desired gas limit",
		},
		cli.StringFlag{
			Name:  "nonce",
			Usage: "Use this flag to explicitly specify the nonce that this transaction should use, so it can override an existing 'stuck' transaction",
		},
		cli.StringFlag{
			Name:  "metricsAddress, m",
			Usage: "Address to serve metrics on if enabled",
			Value: "0.0.0.0",
		},
		cli.UintFlag{
			Name:  "metricsPort, r",
			Usage: "Port to serve metrics on if enabled",
			Value: 9102,
		},
		cli.BoolFlag{
			Name:  "ignore-sync-check",
			Usage: "Set this to true if you already checked the sync status of the execution client(s) and don't need to re-check it for this command",
		},
		cli.BoolFlag{
			Name:  "force-fallback-ec",
			Usage: "Set this to true if you know the primary EC is offline and want to bypass its health checks, and just use the fallback EC instead",
		},
	}

	// Register commands
	api.RegisterCommands(app, "api", []string{"a"})
	node.RegisterCommands(app, "node", []string{"n"})
	watchtower.RegisterCommands(app, "watchtower", []string{"w"})

	// Get command being run
	var commandName string
	app.Before = func(c *cli.Context) error {
		commandName = c.Args().First()
		return nil
	}

	// Run application
	if err := app.Run(os.Args); err != nil {
		if commandName == "api" {
			apiutils.PrintErrorResponse(err)
		} else {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

}
