package megapool

import (
	"fmt"
	"sort"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

type ByIndex []api.MegapoolValidatorDetails

func (a ByIndex) Len() int           { return len(a) }
func (a ByIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByIndex) Less(i, j int) bool { return a[i].ValidatorIndex < a[j].ValidatorIndex }

func exitValidator(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if Saturn is already deployed
	saturnResp, err := rp.IsSaturnDeployed()
	if err != nil {
		return err
	}
	if !saturnResp.IsSaturnDeployed {
		fmt.Println("This command is only available after the Saturn upgrade.")
		return nil
	}

	// List the validators that can be exited
	var validatorId uint64

	if c.IsSet("validator-id") {
		validatorId = c.Uint64("validator-id")
	} else {
		// Get Megapool status
		status, err := rp.MegapoolStatus()
		if err != nil {
			return err
		}

		activeValidators := []api.MegapoolValidatorDetails{}

		for _, validator := range status.Megapool.Validators {
			if validator.Activated {
				activeValidators = append(activeValidators, validator)
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
			validatorId = uint64(activeValidators[selected].ValidatorId)

		} else {
			fmt.Println("No validators can be exited at the moment")
			return nil
		}
	}

	response, err := rp.CanExitValidator(validatorId)
	if err != nil {
		return err
	}

	if !response.CanExit {
		return nil
	}

	// Show a warning message
	fmt.Printf("%sNOTE:\n", colorYellow)
	fmt.Println("You are about to exit a validator. This will tell each the validator to stop all activities on the Beacon Chain.")
	fmt.Println("Please continue to run your validators until each one you've exited has been processed by the exit queue.\nYou can watch their progress on the https://beaconcha.in explorer.")
	fmt.Println("Your funds will be locked on the Beacon Chain until they've been withdrawn, which will happen automatically (this may take a few days).")
	fmt.Printf("Once your funds have been withdrawn, you can run `rocketpool megapool notify-validator-exit` to distribute them to your withdrawal address.\n\n%s", colorReset)

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to EXIT validator id %d?", validatorId))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Exit the validator
	_, err = rp.ExitValidator(validatorId)
	if err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully requested to exit vaildator id %d.\n", validatorId)
	return nil

}
