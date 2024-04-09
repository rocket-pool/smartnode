package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/shared"
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
		Value:   fmt.Sprintf("v%s", shared.RocketPoolVersion),
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
	rp := client.NewClientFromCtx(c)

	// Install service
	err := rp.InstallService(
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
	fmt.Println(shared.Logo)
	fmt.Printf("%s=== Smart Node v%s ===%s\n\n", terminal.ColorGreen, shared.RocketPoolVersion, terminal.ColorReset)
	fmt.Printf("Changes you should be aware of before starting:\n\n")

	fmt.Printf("%s=== New Notification Module ===%s\n", terminal.ColorGreen, terminal.ColorReset)
	fmt.Println("The Smart Node alert notification functionality allows you to receive notifications about the health and important events of your Rocket Pool Smart Node. Check `https://docs.rocketpool.net/guides/node/maintenance/alerting` for more details.")
	fmt.Println()

	fmt.Printf("%s=== New Geth Mode: PBSS ===%s\n", terminal.ColorGreen, terminal.ColorReset)
	fmt.Println("Geth has been updated to v1.13, which includes the much-anticipated Path-Based State Scheme (PBSS) storage mode. With PBSS, you never have to manually prune Geth again; it prunes automatically behind the scenes during runtime! To enable it, check the \"Enable PBSS\" box in the Execution Client section of the `rocketpool service config` UI. Note you **will have to resync** Geth after enabling this for it to take effect, and will lose attestations if you don't have a fallback client enabled!")
	fmt.Println()
}
