package minipool

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func exitMinipools(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get minipool statuses
	status, err := rp.MinipoolStatus()
	if err != nil {
		return err
	}

	// Get active minipools
	activeMinipools := []api.MinipoolDetails{}
	for _, minipool := range status.Minipools {
		if (minipool.Status.Status == types.Staking || (minipool.Status.Status == types.Dissolved && !minipool.Finalised)) && minipool.Validator.Active {
			activeMinipools = append(activeMinipools, minipool)
		}
	}

	// Check for active minipools
	if len(activeMinipools) == 0 {
		fmt.Println("No minipools can be exited.")
		return nil
	}

	// Get selected minipools
	var selectedMinipools []api.MinipoolDetails
	if c.String("minipool") == "" {

		// Prompt for minipool selection
		options := make([]string, len(activeMinipools)+1)
		options[0] = "All available minipools"
		for mi, minipool := range activeMinipools {
			if minipool.Status.Status == types.Staking {
				options[mi+1] = fmt.Sprintf("%s (staking since %s)", minipool.Address.Hex(), minipool.Status.StatusTime.Format(TimeFormat))
			} else {
				options[mi+1] = fmt.Sprintf("%s (dissolved since %s)", minipool.Address.Hex(), minipool.Status.StatusTime.Format(TimeFormat))
			}
		}
		selected, _ := prompt.Select("Please select a minipool to exit:", options)

		// Get minipools
		if selected == 0 {
			selectedMinipools = activeMinipools
		} else {
			selectedMinipools = []api.MinipoolDetails{activeMinipools[selected-1]}
		}

	} else {

		// Get matching minipools
		if c.String("minipool") == "all" {
			selectedMinipools = activeMinipools
		} else {
			selectedAddress := common.HexToAddress(c.String("minipool"))
			for _, minipool := range activeMinipools {
				if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
					selectedMinipools = []api.MinipoolDetails{minipool}
					break
				}
			}
			if selectedMinipools == nil {
				return fmt.Errorf("The minipool %s is not available for exiting.", selectedAddress.Hex())
			}
		}

	}

	// Show a warning message
	fmt.Printf("%sNOTE:\n", colorYellow)
	fmt.Println("You are about to exit your minipool. This will tell each one's validator to stop all activities on the Beacon Chain.")
	fmt.Println("Please continue to run your validators until each one you've exited has been processed by the exit queue.\nYou can watch their progress on the https://beaconcha.in explorer.")
	fmt.Println("Your funds will be locked on the Beacon Chain until they've been withdrawn, which will happen automatically (this may take a few days).")
	fmt.Printf("Once your funds have been withdrawn, you can run `rocketpool minipool close` to distribute them to your withdrawal address and close the minipool.\n\n%s", colorReset)

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.ConfirmWithIAgree(fmt.Sprintf("Are you sure you want to exit %d minipool(s)? This action cannot be undone!", len(selectedMinipools)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Exit minipools
	for _, minipool := range selectedMinipools {
		if _, err := rp.ExitMinipool(minipool.Address); err != nil {
			fmt.Printf("Could not exit minipool %s: %s.\n", minipool.Address.Hex(), err)
		} else {
			fmt.Printf("Successfully exited minipool %s.\n", minipool.Address.Hex())
			fmt.Println("It may take several hours for your minipool's status to be reflected.")
		}
	}

	// Return
	return nil

}
