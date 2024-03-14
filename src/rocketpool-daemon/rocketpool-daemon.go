package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/api"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/node"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/watchtower"
	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/config"
)

// Run
func main() {
	// Initialise application
	app := cli.NewApp()

	// Set application info
	app.Name = "rocketpool"
	app.Usage = "Rocket Pool service"
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
	app.Copyright = "(C) 2024 Rocket Pool Pty Ltd"

	userDirFlag := &cli.StringFlag{
		Name:     "user-dir",
		Aliases:  []string{"u"},
		Usage:    "The path of the user data directory, which contains the configuration file to load and all of the user's runtime data",
		Required: true,
	}

	// Set application flags
	app.Flags = []cli.Flag{
		userDirFlag,
	}

	// Register primary daemon
	app.Commands = append(app.Commands, &cli.Command{
		Name:    "node",
		Aliases: []string{"n"},
		Usage:   "Run primary Rocket Pool node activity daemon and API server",
		Action: func(c *cli.Context) error {
			// Get the config file
			userDir := c.String(userDirFlag.Name)
			cfgPath := filepath.Join(userDir, config.ConfigFilename)
			_, err := os.Stat(cfgPath)
			if errors.Is(err, fs.ErrNotExist) {
				fmt.Printf("Configuration file not found at [%s].", cfgPath)
				os.Exit(1)
			}

			// Wait group to handle graceful stopping
			stopWg := new(sync.WaitGroup)

			// Create the service provider
			sp, err := services.NewServiceProvider(userDir)
			if err != nil {
				return fmt.Errorf("error creating service provider: %w", err)
			}

			// Create the data dir
			dataDir := sp.GetConfig().UserDataPath.Value
			err = os.MkdirAll(dataDir, 0700)
			if err != nil {
				return fmt.Errorf("error creating user data directory [%s]: %w", dataDir, err)
			}

			// Create the server manager
			serverMgr, err := api.NewServerManager(sp, cfgPath, stopWg)
			if err != nil {
				return fmt.Errorf("error creating server manager: %w", err)
			}

			// Start the task loop
			taskLoop := node.NewTaskLoop(sp, stopWg)
			err = taskLoop.Run()
			if err != nil {
				return fmt.Errorf("error starting task loop: %w", err)
			}

			// Handle process closures
			termListener := make(chan os.Signal, 1)
			signal.Notify(termListener, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-termListener
				fmt.Println("Shutting down node...")
				serverMgr.Stop()
				taskLoop.Stop()
			}()

			// Run the daemon until closed
			fmt.Println("Node online.")
			stopWg.Wait()
			fmt.Println("Node stopped.")
			return nil
		},
	})

	// Register watchtower daemon
	app.Commands = append(app.Commands, &cli.Command{
		Name:    "watchtower",
		Aliases: []string{"w"},
		Usage:   "Run Rocket Pool watchtower activity daemon for Oracle DAO duties",
		Action: func(c *cli.Context) error {
			// Get the config file
			userDir := c.String(userDirFlag.Name)
			cfgPath := filepath.Join(userDir, config.ConfigFilename)
			_, err := os.Stat(cfgPath)
			if errors.Is(err, fs.ErrNotExist) {
				fmt.Printf("Configuration file not found at [%s].", cfgPath)
				os.Exit(1)
			}

			// Wait group to handle graceful stopping
			stopWg := new(sync.WaitGroup)

			// Create the service provider
			sp, err := services.NewServiceProvider(userDir)
			if err != nil {
				return fmt.Errorf("error creating service provider: %w", err)
			}

			// Create the data dir
			dataDir := sp.GetConfig().UserDataPath.Value
			err = os.MkdirAll(dataDir, 0700)
			if err != nil {
				return fmt.Errorf("error creating user data directory [%s]: %w", dataDir, err)
			}

			// Start the task loop
			taskLoop := watchtower.NewTaskLoop(sp, stopWg)
			err = taskLoop.Run()
			if err != nil {
				return fmt.Errorf("error starting task loop: %w", err)
			}

			// Handle process closures
			termListener := make(chan os.Signal, 1)
			signal.Notify(termListener, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-termListener
				fmt.Println("Shutting down watchtower...")
				taskLoop.Stop()
			}()

			// Run the daemon until closed
			fmt.Println("Watchtower online.")
			stopWg.Wait()
			fmt.Println("Watchtower stopped.")
			return nil
		},
	})

	// Run application
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
