package service

import (
	"fmt"

	"github.com/mitchellh/go-homedir"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/shared"
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
)

// Install the Rocket Pool service
func installService(c *cli.Context) error {
	dataPath := ""

	// Prompt for confirmation
	if !(c.Bool("yes") || utils.Confirm(fmt.Sprintf(
		"The Rocket Pool service will be installed --Version: %s\n\n%sIf you're upgrading, your existing configuration will be backed up and preserved.\nAll of your previous settings will be migrated automatically.%s\nAre you sure you want to continue?",
		c.String("version"), terminal.ColorGreen, terminal.ColorReset,
	))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Attempt to load the config to see if any settings need to be passed along to the install script
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading old configuration: %w", err)
	}
	if !isNew {
		dataPath = cfg.Smartnode.DataPath.Value.(string)
		dataPath, err = homedir.Expand(dataPath)
		if err != nil {
			return fmt.Errorf("error getting data path from old configuration: %w", err)
		}
	}

	// Install service
	err = rp.InstallService(c.Bool("verbose"), c.Bool("no-deps"), c.String("version"), c.String("path"), dataPath)
	if err != nil {
		return err
	}

	// Print success message & return
	fmt.Println("")
	fmt.Println("The Rocket Pool service was successfully installed!")

	printPatchNotes(c)

	// Reload the config after installation
	_, isNew, err = rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading new configuration: %w", err)
	}

	// Report next steps
	fmt.Printf("%s\n=== Next Steps ===\n", terminal.ColorBlue)
	fmt.Printf("Run 'rocketpool service config' to review the settings changes for this update, or to continue setting up your node.%s\n", terminal.ColorReset)

	// Print the docker permissions notice
	if isNew {
		fmt.Printf("\n%sNOTE:\nSince this is your first time installing Rocket Pool, please start a new shell session by logging out and back in or restarting the machine.\n", terminal.ColorYellow)
		fmt.Printf("This is necessary for your user account to have permissions to use Docker.%s", terminal.ColorReset)
	}

	return nil

}

// Print the latest patch notes for this release
// TODO: get this from an external source and don't hardcode it into the CLI
func printPatchNotes(c *cli.Context) {

	fmt.Print(`
______           _        _    ______           _ 
| ___ \         | |      | |   | ___ \         | |
| |_/ /___   ___| | _____| |_  | |_/ /__   ___ | |
|    // _ \ / __| |/ / _ \ __| |  __/ _ \ / _ \| |
| |\ \ (_) | (__|   <  __/ |_  | | | (_) | (_) | |
\_| \_\___/ \___|_|\_\___|\__| \_|  \___/ \___/|_|

`)
	fmt.Printf("%s=== Smartnode v%s ===%s\n\n", terminal.ColorGreen, shared.RocketPoolVersion, terminal.ColorReset)
	fmt.Printf("Changes you should be aware of before starting:\n\n")

	fmt.Printf("%s=== New Testnet: Holesky ===%s\n", terminal.ColorGreen, terminal.ColorReset)
	fmt.Println("A new test network has been deployed named Holesky! This will replace Prater as the new long-term test network for Rocket Pool node operators. To use it, select the \"Holesky Testnet\" option from the Network dialog in the Smartnode section of `rocketpool service config`.\n")

	fmt.Printf("%s=== Prater Deprecation ===%s\n", terminal.ColorGreen, terminal.ColorReset)
	fmt.Println("The Prater test network is now deprecated, as it is being replaced by Holesky. If you are running a Prater node, please exit your minipools to gracefully clean up the network before migration (https://docs.rocketpool.net/guides/node/withdraw.html).\n")

	fmt.Printf("%s=== New Geth Mode: PBSS ===%s\n", terminal.ColorGreen, terminal.ColorReset)
	fmt.Println("Geth has been updated to v1.13, which includes the much-anticipated Path-Based State Scheme (PBSS) storage mode. With PBSS, you never have to manually prune Geth again; it prunes automatically behind the scenes during runtime! To enable it, check the \"Enable PBSS\" box in the Execution Client section of the `rocketpool service config` UI. Note you **will have to resync** Geth after enabling this for it to take effect, and will lose attestations if you don't have a fallback client enabled!\n")

	fmt.Printf("%s=== MEV-Boost Changes ===%s\n", terminal.ColorGreen, terminal.ColorReset)
	fmt.Println("The \"Blocknative\" relay has been shut down, so we have removed it from the MEV-Boost relay options. The other relays are still available.")
}
