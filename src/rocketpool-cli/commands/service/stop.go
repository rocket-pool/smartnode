package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/terminal"
	"github.com/urfave/cli/v2"
)

// Pause the Rocket Pool service
func stopService(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get the config
	cfg, _, err := rp.LoadConfig()
	if err != nil {
		return err
	}

	// Write a note on doppelganger protection
	if cfg.ValidatorClient.VcCommon.DoppelgangerDetection.Value {
		fmt.Printf("%sNOTE: You currently have Doppelganger Protection enabled.\nIf you stop your validator, it will miss up to 3 attestations when it next starts.\nThis is *intentional* and does not indicate a problem with your node.%s\n\n", terminal.ColorYellow, terminal.ColorReset)
	}

	// Prompt for confirmation
	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm("Are you sure you want to pause the Rocket Pool service? Any staking minipools will be penalized!")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Pause service
	return rp.PauseService(getComposeFiles(c))
}
