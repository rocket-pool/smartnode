package megapool

import (
	"fmt"
	"sort"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

type ByIndex []api.MegapoolValidatorDetails

func (a ByIndex) Len() int           { return len(a) }
func (a ByIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByIndex) Less(i, j int) bool { return a[i].ValidatorIndex < a[j].ValidatorIndex }

func getExitableValidator() (uint64, bool, error) {
	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return 0, false, err
	}
	defer rp.Close()

	// Get Megapool status
	status, err := rp.MegapoolStatus(false)
	if err != nil {
		return 0, false, err
	}

	activeValidators := []api.MegapoolValidatorDetails{}

	for _, validator := range status.Megapool.Validators {
		if validator.Activated && !validator.Exiting && !validator.Exited {
			// Check if validator is old enough to exit
			earliestExitEpoch := validator.BeaconStatus.ActivationEpoch + 256
			if status.BeaconHead.Epoch >= earliestExitEpoch {
				activeValidators = append(activeValidators, validator)
			}
		}
	}
	if len(activeValidators) > 0 {
		sort.Sort(ByIndex(activeValidators))

		options := make([]string, len(activeValidators))
		for vi, v := range activeValidators {
			options[vi] = fmt.Sprintf("ID: %d - Index: %d Pubkey: 0x%s", v.ValidatorId, v.ValidatorIndex, v.PubKey.String())
		}
		selected, _ := prompt.Select("Please select a validator to EXIT:", options)

		// Get validators
		return uint64(activeValidators[selected].ValidatorId), true, nil

	} else {
		fmt.Println("No validators can be exited at the moment")
		return 0, false, nil
	}
}

func exitValidator(validatorId uint64, yes bool) error {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	response, err := rp.CanExitValidator(validatorId)
	if err != nil {
		return err
	}

	if !response.CanExit {
		return nil
	}

	// Show a warning message
	color.YellowPrintln("NOTE:")
	color.YellowPrintln("You are about to exit a validator. This will tell each the validator to stop all activities on the Beacon Chain.")
	color.YellowPrintln("Please continue to run your validators until each one you've exited has been processed by the exit queue.")
	color.YellowPrintln("You can watch their progress on the https://beaconcha.in explorer.")
	color.YellowPrintln("Your funds will be locked on the Beacon Chain until they've been withdrawn, which will happen automatically (this may take a few days).")
	color.YellowPrintln("Once your funds have been withdrawn, you can run `rocketpool megapool notify-validator-exit` to distribute them to your withdrawal address.")
	fmt.Println()

	// Prompt for confirmation
	if prompt.Declined(yes, "Are you sure you want to EXIT validator id %d?", validatorId) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Exit the validator
	_, err = rp.ExitValidator(validatorId)
	if err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully requested to exit validator id %d.\n", validatorId)
	return nil

}
