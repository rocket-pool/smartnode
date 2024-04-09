package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/urfave/cli/v2"
)

// Pause the Rocket Pool service. Returns whether the action proceeded (was confirmed by user and no error occurred before starting it)
func stopService(c *cli.Context) (bool, error) {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get the config
	cfg, _, err := rp.LoadConfig()
	if err != nil {
		return false, err
	}

	// Write a note on doppelganger protection
	if cfg.ValidatorClient.VcCommon.DoppelgangerDetection.Value {
		fmt.Printf("%sNOTE: You currently have Doppelganger Protection enabled.\nIf you stop your validator, it will miss up to 3 attestations when it next starts.\nThis is *intentional* and does not indicate a problem with your node.%s\n\n", terminal.ColorYellow, terminal.ColorReset)
	}

	// Prompt for confirmation
	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm("Are you sure you want to pause the Smart Node service? Any staking minipools will be penalized!")) {
		fmt.Println("Cancelled.")
		return false, nil
	}

	// Pause service
	err = rp.PauseService(getComposeFiles(c))
	return true, err
}
