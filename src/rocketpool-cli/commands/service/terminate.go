package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/terminal"
	"github.com/urfave/cli/v2"
)

// Terminate the Rocket Pool service
func terminateService(c *cli.Context) error {
	// Prompt for confirmation
	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm(fmt.Sprintf("%sWARNING: Are you sure you want to terminate the Rocket Pool service? Any staking minipools will be penalized, your ETH1 and ETH2 chain databases will be deleted, you will lose ALL of your sync progress, and you will lose your Prometheus metrics database!\nAfter doing this, you will have to **reinstall** the Smartnode uses `rocketpool service install -d` in order to use it again.%s", terminal.ColorRed, terminal.ColorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Stop service
	return rp.TerminateService(getComposeFiles(c), rp.Context.ConfigPath)
}
