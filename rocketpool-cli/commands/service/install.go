package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/v2/assets"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/urfave/cli/v2"
)

var (
	installVerboseFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:    "verbose",
		Aliases: []string{"r"},
		Usage:   "Print installation script command output",
	}
	installNoDepsFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:    "no-deps",
		Aliases: []string{"d"},
		Usage:   "Do not install Operating System dependencies",
	}
	installPathFlag *cli.StringFlag = &cli.StringFlag{
		Name:    "path",
		Aliases: []string{"p"},
		Usage:   "A custom path to install Rocket Pool to",
	}
	installVersionFlag *cli.StringFlag = &cli.StringFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "The smart node package version to install",
		Value:   fmt.Sprintf("v%s", assets.RocketPoolVersion()),
	}
	installUpdateDefaultsFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:    "update-defaults",
		Aliases: []string{"u"},
		Usage:   "Certain configuration values are reset when the Smart Node is updated, such as Docker container tags; use this flag to force that reset, even if the Smart Node hasn't been updated",
	}
	installLocalFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:    "local-script",
		Aliases: []string{"l"},
		Usage:   fmt.Sprintf("Use a local installer script instead of pulling it down from the source repository. The script and the installer package must be in your current working directory.%sMake sure you absolutely trust the script before using this flag.%s", terminal.ColorRed, terminal.ColorReset),
	}
)

// Install the Rocket Pool service
func installService(c *cli.Context) error {
	// Prompt for confirmation
	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm(fmt.Sprintf(
		"The Rocket Pool Smart Node service will be installed --Version: %s\n\n%sIf you're upgrading, your existing configuration will be backed up and preserved.\nAll of your previous settings will be migrated automatically.%s\nAre you sure you want to continue?",
		c.String(installVersionFlag.Name), terminal.ColorGreen, terminal.ColorReset,
	))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Get RP client
	rp, err := client.NewClientFromCtx(c)
	if err != nil {
		return err
	}

	// Install service
	err = rp.InstallService(
		c.Bool(installVerboseFlag.Name),
		c.Bool(installNoDepsFlag.Name),
		c.String(installVersionFlag.Name),
		c.String(installPathFlag.Name),
		c.Bool(installLocalFlag.Name),
	)
	if err != nil {
		return err
	}

	// Print success message & return
	fmt.Println("")
	fmt.Println("The Rocket Pool Smart Node service was successfully installed!")

	printPatchNotes()

	// Reload the config after installation
	_, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading new configuration: %w", err)
	}

	// Report next steps
	fmt.Printf("%s\n=== Next Steps ===\n", terminal.ColorBlue)
	fmt.Printf("Run 'rocketpool service config' to review the settings changes for this update, or to continue setting up your node.%s\n", terminal.ColorReset)

	// Print the docker permissions notice
	if isNew {
		fmt.Printf("\n%sNOTE:\nSince this is your first time installing the Smart Node, please start a new shell session by logging out and back in or restarting the machine.\n", terminal.ColorYellow)
		fmt.Printf("This is necessary for your user account to have permissions to use Docker.%s", terminal.ColorReset)
	}

	return nil
}

// Print the latest patch notes for this release
// TODO: get this from an external source and don't hardcode it into the CLI
func printPatchNotes() {
	fmt.Println()
	fmt.Println(assets.Logo())
	fmt.Printf("%s=== Smart Node v%s ===%s\n\n", terminal.ColorGreen, assets.RocketPoolVersion(), terminal.ColorReset)
	fmt.Printf("Changes you should be aware of before starting:\n\n")

	fmt.Printf("%s=== Welcome to the v2.0 Beta! ===%s\n", terminal.ColorGreen, terminal.ColorReset)
	fmt.Println("Welcome to Smart Node v2! This is a completely redesigned Smart Node from the ground up, taking advantage of years of lessons learned and user feedback. The list of features and changes is far too long to list here, but here are some highlights:")
	fmt.Println("- Support for installing and updating via Debian's `apt` package manager (other package managers coming soon!)")
	fmt.Println("- Support for printing transaction data or signed transactions without submitting them to the network")
	fmt.Println("- Passwordless mode: an opt-in feature that will no longer save your node wallet's password to disk")
	fmt.Println("- Overhauled Smart Node service with an HTTP API, support for batching Ethereum queries and transactions together, a new logging system, and consolidation of the api / node / watchtower containers into one")
	fmt.Println("- And much more!")
	fmt.Println()
	fmt.Println("To learn all about what's changed in Smart Node v2 and how to use it, take a look at our guide: https://github.com/rocket-pool/smartnode/blob/v2/v2.md")
}
