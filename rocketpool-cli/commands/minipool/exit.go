package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

func exitMinipools(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get minipool statuses
	status, err := rp.Api.Minipool.Status()
	if err != nil {
		return err
	}

	// Get active minipools
	activeMinipools := []api.MinipoolDetails{}
	for _, minipool := range status.Data.Minipools {
		if (minipool.Status.Status == types.MinipoolStatus_Staking || (minipool.Status.Status == types.MinipoolStatus_Dissolved && !minipool.Finalised)) && minipool.Validator.Active {
			activeMinipools = append(activeMinipools, minipool)
		}
	}

	// Check for active minipools
	if len(activeMinipools) == 0 {
		fmt.Println("No minipools can be exited.")
		return nil
	}

	// Get selected minipools
	options := make([]utils.SelectionOption[api.MinipoolDetails], len(activeMinipools))
	for i, mp := range activeMinipools {
		option := &options[i]
		option.Element = &activeMinipools[i]
		option.ID = fmt.Sprint(mp.Address)

		if mp.Status.Status == types.MinipoolStatus_Staking {
			option.Display = fmt.Sprintf("%s (staking since %s)", mp.Address.Hex(), mp.Status.StatusTime.Format(TimeFormat))
		} else {
			option.Display = fmt.Sprintf("%s (dissolved since %s)", mp.Address.Hex(), mp.Status.StatusTime.Format(TimeFormat))
		}
	}
	selectedMinipools, err := utils.GetMultiselectIndices(c, minipoolsFlag, options, "Please select a minipool to exit:")
	if err != nil {
		return fmt.Errorf("error determining minipool selection: %w", err)
	}

	// Show a warning message
	fmt.Printf("%sNOTE:\n", terminal.ColorYellow)
	fmt.Println("You are about to exit your minipool. This will tell each one's validator to stop all activities on the Beacon Chain.")
	fmt.Println("Please continue to run your validators until each one you've exited has been processed by the exit queue.\nYou can watch their progress on the https://beaconcha.in explorer.")
	fmt.Println("Your funds will be locked on the Beacon Chain until they've been withdrawn, which will happen automatically (this may take a few days).")
	fmt.Printf("Once your funds have been withdrawn, you can run `rocketpool minipool close` to distribute them to your withdrawal address and close the minipool.\n\n%s", terminal.ColorReset)

	// Prompt for confirmation
	if !(c.Bool("yes") || utils.ConfirmWithIAgree(fmt.Sprintf("Are you sure you want to exit %d minipool(s)? This action cannot be undone!", len(selectedMinipools)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Exit minipools
	addresses := make([]common.Address, len(selectedMinipools))
	for i, mp := range selectedMinipools {
		addresses[i] = mp.Address
	}
	if _, err := rp.Api.Minipool.Exit(addresses); err != nil {
		return fmt.Errorf("error while exiting minipools: %w\n", err)
	} else {
		fmt.Println("Successfully exited all selected minipools.")
		fmt.Println("It may take several hours for your minipools' status to be reflected.")
	}

	// Return
	return nil
}
