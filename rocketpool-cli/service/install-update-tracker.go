package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/shared"
	"github.com/urfave/cli/v2"
)

var (
	installUpdateTrackerVerboseFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:    "verbose",
		Aliases: []string{"r"},
		Usage:   "Print installation script command output",
	}
	installUpdateTrackerVersionFlag *cli.StringFlag = &cli.StringFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "The update tracker package version to install",
		Value:   fmt.Sprintf("v%s", shared.RocketPoolVersion),
	}
)

// Install the Rocket Pool update tracker for the metrics dashboard
func installUpdateTracker(c *cli.Context) error {
	// Prompt for confirmation
	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm(
		"This will add the ability to display any available Operating System updates or new Rocket Pool versions on the metrics dashboard. "+
			"Are you sure you want to install the update tracker?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Install service
	err := rp.InstallUpdateTracker(c.Bool(installUpdateTrackerVerboseFlag.Name), c.String(installUpdateTrackerVersionFlag.Name))
	if err != nil {
		return err
	}

	// Print success message & return
	fmt.Println("")
	fmt.Println("The Rocket Pool update tracker service was successfully installed!")
	fmt.Println("")
	fmt.Printf("%sNOTE:\nPlease restart the Smartnode stack to enable update tracking on the metrics dashboard.%s\n", terminal.ColorYellow, terminal.ColorReset)
	fmt.Println("")
	return nil
}
