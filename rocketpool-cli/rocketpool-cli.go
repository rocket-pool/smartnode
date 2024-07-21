package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/assets"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/settings"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
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
	app := newCliApp()
	run(app, os.Args)
}

func run(app *cli.App, args []string) {
	// Run application
	_, err := app.Writer.Write([]byte("\n"))
	if err != nil {
		panic(err)
	}
	if err := app.Run(args); err != nil {
		utils.PrettyPrintError(err)
		os.Exit(1)
	}
	_, err = app.Writer.Write([]byte("\n"))
	if err != nil {
		panic(err)
	}
}

func newCliApp() *cli.App {
	// Initialise application
	app := cli.NewApp()

	// Add logo to application help template
	app.CustomAppHelpTemplate = fmt.Sprintf("%s\n%s", assets.Logo(), cli.AppHelpTemplate)

	// Set application info
	app.Name = "rocketpool"
	app.Usage = "Smart Node CLI for Rocket Pool"
	app.Version = assets.RocketPoolVersion()
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
	commands.RegisterCommands(app)

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

	app.After = func(c *cli.Context) error {
		// Close http tracer if any was created
		snSettings = settings.GetSmartNodeSettings(c)
		if snSettings != nil && snSettings.HttpTraceFile != nil {
			snSettings.HttpTraceFile.Close()
		}
		return nil
	}
	return app
}
