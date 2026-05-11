package megapool

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

type ByIndex []api.MegapoolValidatorDetails

func (a ByIndex) Len() int           { return len(a) }
func (a ByIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByIndex) Less(i, j int) bool { return a[i].ValidatorIndex < a[j].ValidatorIndex }

func formatGwei(gwei uint64) string {
	eth := float64(gwei) / 1e9
	if eth == float64(uint64(eth)) {
		return fmt.Sprintf("%d", uint64(eth))
	}
	return fmt.Sprintf("%.2f", eth)
}

// formatDaysHours formats a duration as a "Xd Yh" string. Sub-hour values are
// rendered in minutes; sub-minute values fall back to "<1m".
func formatDaysHours(d time.Duration) string {
	totalSeconds := int64(d.Seconds())
	if totalSeconds < 60 {
		return "<1m"
	}
	const secondsPerHour = 3600
	const secondsPerDay = 24 * secondsPerHour
	days := totalSeconds / secondsPerDay
	hours := (totalSeconds % secondsPerDay) / secondsPerHour
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	minutes := totalSeconds / 60
	return fmt.Sprintf("%dm", minutes)
}

// sortExitingValidatorsBySweep orders validators in withdrawal-sweep order relative to a given last validator index
func sortExitingValidatorsBySweep(validators []api.MegapoolValidatorDetails, lastWithdrawnIndex uint64, hasLastWithdrawnIndex bool) {
	if !hasLastWithdrawnIndex {
		sort.Sort(ByIndex(validators))
		return
	}
	sort.SliceStable(validators, func(i, j int) bool {
		ai := validators[i].ValidatorIndex
		aj := validators[j].ValidatorIndex
		iAfter := ai > lastWithdrawnIndex
		jAfter := aj > lastWithdrawnIndex
		if iAfter != jAfter {
			return iAfter
		}
		return ai < aj
	})
}

func getExitableValidator(fetchExitQueueEstimate bool) (uint64, bool, error) {
	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return 0, false, err
	}
	defer rp.Close()

	// Get the latest block and identify the withdrawals present in it
	var lastWithdrawnIndex uint64
	var hasLastWithdrawnIndex bool
	withdrawalsResp, err := rp.GetLatestBlockWithdrawals()
	if err != nil {
		fmt.Printf("Warning: could not fetch latest beacon block withdrawals: %s\n\n", err.Error())
	} else if len(withdrawalsResp.Withdrawals) == 0 {
		fmt.Printf("Latest beacon block (slot %d, exec block %d) has no validator withdrawals.\n\n",
			withdrawalsResp.Slot, withdrawalsResp.BlockNumber)
	} else {
		indexes := make([]string, 0, len(withdrawalsResp.Withdrawals))
		seen := make(map[string]struct{}, len(withdrawalsResp.Withdrawals))
		for _, wd := range withdrawalsResp.Withdrawals {
			if idx, perr := strconv.ParseUint(wd.ValidatorIndex, 10, 64); perr == nil {
				if !hasLastWithdrawnIndex || idx > lastWithdrawnIndex {
					lastWithdrawnIndex = idx
					hasLastWithdrawnIndex = true
				}
			}
			if _, ok := seen[wd.ValidatorIndex]; ok {
				continue
			}
			seen[wd.ValidatorIndex] = struct{}{}
			indexes = append(indexes, wd.ValidatorIndex)
		}
		fmt.Printf("Latest beacon block (slot %d, exec block %d) processed withdrawals for %d validator(s):\n",
			withdrawalsResp.Slot, withdrawalsResp.BlockNumber, len(indexes))
		fmt.Printf("  %s\n\n", strings.Join(indexes, ", "))
	}
	var estimate api.BeaconWithdrawalQueueEstimateResponse
	if fetchExitQueueEstimate {
		// Print an estimate of the beacon chain withdrawal queue time
		fmt.Println("Fetching beacon chain exit queue estimate... This may take a while...")
		if estimate, err = rp.GetBeaconWithdrawalQueueEstimate(); err != nil {
			fmt.Printf("Warning: could not fetch beacon chain exit queue estimate: %s\n\n", err.Error())
		} else if estimate.ExitQueueGwei == 0 {
			fmt.Println("The beacon chain exit queue is currently empty.")
			fmt.Printf("At the current churn limit of %s ETH/epoch, a new exit request would be processed in the next epoch (~%s).\n\n",
				formatGwei(estimate.ChurnPerEpochGwei), (time.Duration(estimate.SecondsPerEpoch) * time.Second).Round(time.Second))
		} else {
			wait := formatDaysHours(time.Duration(estimate.EstimatedQueueSeconds) * time.Second)
			fmt.Printf("Beacon chain exit queue: %s ETH waiting to exit.\n",
				formatGwei(estimate.ExitQueueGwei))
			fmt.Printf("Churn limit: %s ETH/epoch -> estimated %d epochs (~%s) to process the queue.\n\n",
				formatGwei(estimate.ChurnPerEpochGwei), estimate.EstimatedQueueEpochs, wait)
		}
	} else {
		fmt.Println("Skipping the beacon chain exit queue estimate. Use the --fetch-estimate flag to fetch it.")
		fmt.Println()
	}

	// Get Megapool status
	status, err := rp.MegapoolStatus(false)
	if err != nil {
		return 0, false, err
	}

	activeValidators := []api.MegapoolValidatorDetails{}
	exitingValidators := []api.MegapoolValidatorDetails{}

	for _, validator := range status.Megapool.Validators {
		if validator.Activated && !validator.Exiting && !validator.Exited && validator.BeaconStatus.Status != beacon.ValidatorState_ActiveExiting {
			// Check if validator is old enough to exit
			earliestExitEpoch := validator.BeaconStatus.ActivationEpoch + 256
			if status.BeaconHead.Epoch >= earliestExitEpoch {
				activeValidators = append(activeValidators, validator)
			}
		}
		if validator.BeaconStatus.Status == beacon.ValidatorState_ActiveExiting {
			exitingValidators = append(exitingValidators, validator)
		}
	}

	// Print exiting validators
	if len(exitingValidators) > 0 {
		sortExitingValidatorsBySweep(exitingValidators, lastWithdrawnIndex, hasLastWithdrawnIndex)
		fmt.Println("The following validators are still active and have already received their exit request on the Beacon Chain:")
		for _, v := range exitingValidators {
			fmt.Printf("ID %d: - Index %d Pubkey: 0x%s\n", v.ValidatorId, v.ValidatorIndex, v.PubKey.String())
		}
		fmt.Println()
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

	}
	fmt.Println("No validators can be exited at the moment")
	return 0, false, nil
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
