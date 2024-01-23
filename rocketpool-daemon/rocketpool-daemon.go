package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/api"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/node"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/watchtower"
	"github.com/rocket-pool/smartnode/shared"
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
	app.Copyright = "(C) 2023 Rocket Pool Pty Ltd"

	// Set application flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "settings, s",
			Usage: "Rocket Pool service user config absolute `path`",
			Value: "/.rocketpool/user-settings.yml",
		},
		cli.StringFlag{
			Name:  "metricsAddress, m",
			Usage: "Address to serve metrics on if enabled",
			Value: "0.0.0.0",
		},
		cli.BoolFlag{
			Name:  "use-protected-api",
			Usage: "Set this to true to use the Flashbots Protect RPC instead of your local Execution Client. Useful to ensure your transactions aren't front-run.",
		},
	}

	// Register primary daemon
	app.Commands = append(app.Commands, cli.Command{
		Name:    "node",
		Aliases: []string{"n"},
		Usage:   "Run primary Rocket Pool node activity daemon and API server",
		Action: func(c *cli.Context) error {
			// Create env vars
			metricsAddress := c.String("metricsAddress")
			if metricsAddress == "" {
				metricsAddress = "0.0.0.0"
			}
			os.Setenv("NODE_METRICS_ADDRESS", metricsAddress)

			// Create the service provider
			sp, err := services.NewServiceProvider(c)
			if err != nil {
				return fmt.Errorf("error creating service provider: %w", err)
			}

			// Create the API server
			apiMgr := api.NewApiManager(sp)
			err = apiMgr.Start()
			if err != nil {
				return fmt.Errorf("error starting API server: %w", err)
			}

			return node.Run(sp)
		},
	})

	// Register watchtower daemon
	app.Commands = append(app.Commands, cli.Command{
		Name:    "watchtower",
		Aliases: []string{"w"},
		Usage:   "Run Rocket Pool watchtower activity daemon for Oracle DAO duties",
		Action: func(c *cli.Context) error {
			// Create env vars
			metricsAddress := c.String("metricsAddress")
			if metricsAddress == "" {
				metricsAddress = "0.0.0.0"
			}
			os.Setenv("WATCHTOWER_METRICS_ADDRESS", metricsAddress)

			// Create the service provider
			sp, err := services.NewServiceProvider(c)
			if err != nil {
				return fmt.Errorf("error creating service provider: %w", err)
			}
			return watchtower.Run(sp)
		},
	})

	// Run application
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
