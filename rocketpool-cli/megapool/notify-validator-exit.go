package megapool

import (
	"fmt"
	"sort"

	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

const FarFutureEpoch uint64 = 0xffffffffffffffff

func getExitedValidator() (uint64, bool, error) {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return 0, false, err
	}
	defer rp.Close()
	fmt.Println("Loading megapool validators at the finalized beacon state...")
	status, err := rp.MegapoolStatus(true)
	if err != nil {
		return 0, false, err
	}

	activeValidators := []api.MegapoolValidatorDetails{}

	for _, validator := range status.Megapool.Validators {
		if !validator.Activated || validator.Exiting || validator.Exited {
			continue
		}
		if !validator.BeaconStatus.Exists {
			continue
		}
		if validator.BeaconStatus.ExitEpoch == FarFutureEpoch {
			continue
		}
		if validator.BeaconStatus.WithdrawableEpoch == FarFutureEpoch {
			continue
		}
		activeValidators = append(activeValidators, validator)
	}
	if len(activeValidators) > 0 {
		sort.Sort(ByIndex(activeValidators))
		options := make([]string, len(activeValidators))
		for vi, v := range activeValidators {
			options[vi] = fmt.Sprintf(
				"ID: %d - Index: %d - exit_epoch: %d - withdrawable_epoch: %d - Pubkey: 0x%s",
				v.ValidatorId, v.ValidatorIndex,
				v.BeaconStatus.ExitEpoch, v.BeaconStatus.WithdrawableEpoch,
				v.PubKey.String(),
			)
		}
		selected, _ := prompt.Select("Please select a validator to notify the exit:", options)

		// Get validators
		return uint64(activeValidators[selected].ValidatorId), true, nil
	}
	fmt.Println("No validators are ready to notify exit.")
	fmt.Println("A validator is only listed once its exit_epoch is set on the *finalized* beacon state")
	return 0, false, nil
}

func notifyValidatorExit(validatorId uint64, yes bool) error {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	fmt.Printf("Checking whether validator id %d can be notified...\n", validatorId)

	response, err := rp.CanNotifyValidatorExit(validatorId)
	if err != nil {
		return fmt.Errorf("could not check notify-validator-exit for id %d: %w", validatorId, err)
	}

	if !response.CanExit {
		fmt.Printf("Cannot notify the exit of validator id %d.\n", validatorId)
		if response.InvalidStatus {
			fmt.Println("  The validator is not in a staked state.")
		}
		if response.AlreadyExiting {
			fmt.Println("  Exit has already been notified for this validator.")
		}
		if response.AlreadyExited {
			fmt.Println("  The validator has already been fully exited on the megapool.")
		}
		if response.ExitNotFinalized {
			fmt.Println("  The validator exit is not yet reflected in the finalized beacon state.")
		}
		return fmt.Errorf("cannot notify exit of validator id %d", validatorId)
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(response.GasLimits, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if prompt.Declined(yes, "Are you sure you want to notify about the validator id %d exit?", validatorId) {
		fmt.Println("Cancelled.")
		return nil
	}

	fmt.Printf("Submitting notify-exit for validator id %d ...\n", validatorId)

	// Exit the validator
	resp, err := rp.NotifyValidatorExit(validatorId)
	if err != nil {
		return fmt.Errorf("notify-validator-exit failed for id %d: %w", validatorId, err)
	}

	fmt.Printf("Notifying validator exit...\n")
	cliutils.PrintTransactionHash(rp, resp.TxHash)
	if _, err = rp.WaitForTransaction(resp.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully notified the exit of validator id %d.\n", validatorId)
	return nil

}
