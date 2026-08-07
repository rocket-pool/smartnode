package megapool

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/color"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func getNotifiableValidator() (uint64, uint64, bool, error) {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return 0, 0, false, err
	}
	defer rp.Close()
	// Get Megapool status (finalized beacon state — required for final balance proofs)
	fmt.Println("Loading megapool validators at the finalized beacon state...")
	status, err := rp.MegapoolStatus(true)
	if err != nil {
		return 0, 0, false, err
	}

	readyValidators := []api.MegapoolValidatorDetails{}
	pendingValidators := []api.MegapoolValidatorDetails{}
	// Beacon exit started, but notify-validator-exit has not been submitted yet.
	needsExitNotify := []api.MegapoolValidatorDetails{}

	for _, validator := range status.Megapool.Validators {
		if validator.Exited {
			continue
		}
		if validator.Exiting {
			if validator.BeaconStatus.Status == beacon.ValidatorState_WithdrawalDone {
				readyValidators = append(readyValidators, validator)
			} else {
				pendingValidators = append(pendingValidators, validator)
			}
			continue
		}
		// Exit is visible on the finalized beacon state but not yet recorded on the megapool.
		if validator.Activated &&
			validator.BeaconStatus.Exists &&
			validator.BeaconStatus.ExitEpoch != 0 &&
			validator.BeaconStatus.ExitEpoch != FarFutureEpoch {
			needsExitNotify = append(needsExitNotify, validator)
		}
	}
	if len(readyValidators) > 0 {
		sort.Sort(ByIndex(readyValidators))
		options := make([]string, len(readyValidators))
		for vi, v := range readyValidators {
			options[vi] = fmt.Sprintf("ID: %d - Index: %d - Pubkey: 0x%s", v.ValidatorId, v.ValidatorIndex, v.PubKey.String())
		}
		selected, _ := prompt.Select("Please select a validator to notify the final balance:", options)

		// Get validators
		return uint64(readyValidators[selected].ValidatorId), uint64(readyValidators[selected].ValidatorIndex), true, nil

	}

	fmt.Println("No validators at the state where the full withdrawal can be proved.")
	printPendingFinalBalanceValidators(pendingValidators, needsExitNotify, status)
	return 0, 0, false, nil
}

func printPendingFinalBalanceValidators(pending, needsExitNotify []api.MegapoolValidatorDetails, status api.MegapoolStatusResponse) {
	currentEpoch := status.BeaconHead.FinalizedEpoch
	if currentEpoch == 0 {
		currentEpoch = status.BeaconHead.Epoch
	}
	secondsPerEpoch := status.SecondsPerEpoch
	if secondsPerEpoch == 0 {
		secondsPerEpoch = 384
	}

	if len(pending) == 0 && len(needsExitNotify) == 0 {
		fmt.Println("There are also no megapool validators currently exiting on the finalized beacon state.")
		fmt.Println("A final balance proof can only be submitted after:")
		fmt.Println("  1. the validator has exited on the beacon chain")
		fmt.Println("  2. notify-validator-exit has been run")
		fmt.Println("  3. the validator reaches beacon status withdrawal_done (full balance withdrawn)")
		return
	}

	if len(needsExitNotify) > 0 {
		sort.Sort(ByIndex(needsExitNotify))
		fmt.Printf("The following %d validator(s) have an exit visible on the finalized beacon state but still need notify-validator-exit first:\n", len(needsExitNotify))
		fmt.Println()
		for _, v := range needsExitNotify {
			printPendingFinalBalanceValidator(v, currentEpoch, secondsPerEpoch)
		}
		fmt.Println("Run: rocketpool megapool notify-validator-exit")
		fmt.Println()
	}

	if len(pending) > 0 {
		sort.Sort(ByIndex(pending))
		fmt.Printf("The following %d validator(s) are exiting on the megapool but not yet fully withdrawn (finalized epoch %d):\n", len(pending), currentEpoch)
		fmt.Println()
		for _, v := range pending {
			printPendingFinalBalanceValidator(v, currentEpoch, secondsPerEpoch)
		}
	}

	fmt.Println("A final balance proof can only be submitted once beacon status is withdrawal_done.")
	fmt.Println("After withdrawable_epoch, the beacon chain still needs to process the full withdrawal (sweep); that can take additional time beyond the estimate above.")
}

func printPendingFinalBalanceValidator(v api.MegapoolValidatorDetails, currentEpoch, secondsPerEpoch uint64) {
	bs := v.BeaconStatus
	fmt.Printf("  ID %d - Index %d - status: %s\n", v.ValidatorId, v.ValidatorIndex, bs.Status)

	exitEpoch := bs.ExitEpoch
	withdrawableEpoch := bs.WithdrawableEpoch
	if withdrawableEpoch == 0 || withdrawableEpoch == FarFutureEpoch {
		// Prefer the megapool-recorded withdrawable epoch if beacon still reports FAR_FUTURE.
		if v.WithdrawableEpoch != 0 && v.WithdrawableEpoch != FarFutureEpoch {
			withdrawableEpoch = v.WithdrawableEpoch
		}
	}

	if exitEpoch != 0 && exitEpoch != FarFutureEpoch {
		fmt.Printf("    exit_epoch:          %d%s\n", exitEpoch, epochTimingSuffix(exitEpoch, currentEpoch, secondsPerEpoch))
	} else {
		fmt.Printf("    exit_epoch:          not yet set on finalized state\n")
	}

	if withdrawableEpoch != 0 && withdrawableEpoch != FarFutureEpoch {
		fmt.Printf("    withdrawable_epoch:  %d%s\n", withdrawableEpoch, epochTimingSuffix(withdrawableEpoch, currentEpoch, secondsPerEpoch))
		if currentEpoch >= withdrawableEpoch {
			switch bs.Status {
			case beacon.ValidatorState_WithdrawalPossible:
				fmt.Printf("    note:                withdrawable; waiting for the beacon withdrawal sweep (full balance)\n")
			case beacon.ValidatorState_ExitedUnslashed, beacon.ValidatorState_ExitedSlashed:
				fmt.Printf("    note:                exited; waiting to become withdrawal_possible, then for the sweep\n")
			default:
				fmt.Printf("    note:                withdrawable epoch reached; waiting for full withdrawal (status %s)\n", bs.Status)
			}
		}
	} else {
		fmt.Printf("    withdrawable_epoch:  not yet set on finalized state\n")
	}
	fmt.Println()
}

func epochTimingSuffix(targetEpoch, currentEpoch, secondsPerEpoch uint64) string {
	if targetEpoch <= currentEpoch {
		return " (reached)"
	}
	remaining := targetEpoch - currentEpoch
	wait := formatDaysHours(time.Duration(remaining*secondsPerEpoch) * time.Second)
	return fmt.Sprintf(" (in %d epochs, ~%s)", remaining, wait)
}

func notifyFinalBalance(validatorId, validatorIndex, slot uint64, yes bool) error {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get the config
	cfg, _, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading configuration: %w", err)
	}

	if slot == 0 {
		fmt.Println("The Smart Node needs to find the slot containing the validator withdrawal. This may take a while. You can speed up the final balance proof generation by submitting the withdrawal slot for your validator.")
		fmt.Println()

		if validatorIndex != 0 {
			beaconChainUrl := getBeaconChainURL(validatorIndex, cfg)
			fmt.Printf("The withdrawal slot for validator ID: %d can be found under the 'Consensus Layer' tab on this page: %s\n", validatorId, beaconChainUrl)
			fmt.Println()
		}

		if prompt.Confirm("Would you like to manually input the withdrawal slot?") {
			slotString := prompt.Prompt("Please enter the withdrawal slot:", "^\\d+$", "Invalid slot. Please provide a slot number.")
			slot, err = strconv.ParseUint(slotString, 0, 64)
			if err != nil {
				return fmt.Errorf("'%s' is not a valid slot: %w.\n", slotString, err)
			}
		}
	}

	color.YellowPrintln("Fetching the beacon state to craft a final balance proof. This process can take several minutes and is CPU and memory intensive.")
	fmt.Println()

	response, err := rp.CanNotifyFinalBalance(validatorId, slot)
	if err != nil {
		return err
	}

	if !response.CanExit {
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(response.GasLimits, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if prompt.Declined(yes, "Are you sure you want to notify the final balance for validator id %d exit?", validatorId) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Exit the validator
	resp, err := rp.NotifyFinalBalance(validatorId, slot)
	if err != nil {
		return err
	}

	fmt.Printf("Notifying validator final balance...\n")
	cliutils.PrintTransactionHash(rp, resp.TxHash)
	if _, err = rp.WaitForTransaction(resp.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully notified final balance for validator id %d.\n", validatorId)
	return nil

}

// returns the Beaconcha.in withdrawals URL for a validator index.
func getBeaconChainURL(index uint64, cfg *config.RocketPoolConfig) string {
	network := cfg.GetNetwork()

	var baseURL string
	switch network {
	case cfgtypes.Network_Mainnet:
		baseURL = "https://beaconcha.in"
	case cfgtypes.Network_Devnet, cfgtypes.Network_Testnet:
		baseURL = "https://hoodi.beaconcha.in"
	default:
		return ""
	}

	return fmt.Sprintf("%s/validator/%d#withdrawals", baseURL, index)
}
