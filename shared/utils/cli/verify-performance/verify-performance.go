// Package verifyperformance contains shared CLI helpers for the
// `rocketpool minipool verify-performance` and
// `rocketpool megapool verify-performance` commands.
package verifyperformance

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

// LargeEpochRangeWarning is the number of epochs above which the CLI prompts
// the user to confirm, because each epoch issues a committees fetch plus a
// few-dozen block fetches against the beacon node.
const LargeEpochRangeWarning uint64 = 256

// ResolveEpochRange fills in defaults for the start/length range. If epochs
// is 0, it defaults to the on-chain performance_period setting. Returns the
// inclusive [startEpoch, endEpoch] range.
func ResolveEpochRange(rp *rocketpool.Client, startEpoch, epochs uint64) (uint64, uint64, error) {
	if epochs == 0 {
		settings, err := rp.PDAOGetSettings()
		if err != nil {
			return 0, 0, fmt.Errorf("error fetching pDAO settings for default performance_period: %w", err)
		}
		epochs = settings.Performance.Period
		if epochs == 0 {
			return 0, 0, fmt.Errorf("on-chain performance_period is 0 and no --epochs was provided")
		}
	}
	endEpoch := startEpoch + epochs - 1
	return startEpoch, endEpoch, nil
}

// ConfirmLargeRange prompts the user when the requested epoch range exceeds
// LargeEpochRangeWarning and returns true if the user accepts the warning.
// Small ranges return true immediately.
func ConfirmLargeRange(total uint64) bool {
	if total <= LargeEpochRangeWarning {
		return true
	}
	return prompt.Confirm(
		"This will fetch attestation data for %d epochs from your beacon node. "+
			"For historical epochs this requires an archival beacon node and may take a while. "+
			"Continue?",
		total,
	)
}

// PrintCancelled prints a generic cancellation message and returns nil so it
// can be returned directly from a CLI Action.
func PrintCancelled() error {
	fmt.Println("Cancelled.")
	return nil
}

// PrintResult writes a VerifyPerformanceResponse to stdout in a human-readable
// format. `label` is the human-facing string identifying the validator being
// verified, e.g. "minipool 0x...".
func PrintResult(resp api.VerifyPerformanceResponse, label string) {
	fmt.Printf("RPIP-73 target-vote performance for %s\n", label)
	fmt.Printf("  Validator:        %s (index %d)\n", resp.ValidatorPubkey.Hex(), resp.ValidatorIndex)
	fmt.Printf("  Epoch range:      %d - %d (inclusive, %d epochs)\n", resp.StartEpoch, resp.EndEpoch, resp.TotalEpochs)
	fmt.Printf("  Timely target:    %d epochs\n", resp.TimelyEpochs)
	fmt.Printf("  Missed target:    %d epochs\n", resp.MissedEpochs)
	if resp.InactiveEpochs > 0 {
		fmt.Printf("  Inactive:         %d epochs (no committee assignment, excluded from %%)\n", resp.InactiveEpochs)
	}
	fmt.Printf("  Performance:      %.2f%%\n", resp.PerformancePct)
	fmt.Printf("  Threshold:        %.2f%%\n", resp.PerformanceThresholdPct)
	if resp.PassesThreshold {
		fmt.Println("  Result:           PASS (not exit-eligible under RPIP-73 with these parameters)")
	} else {
		fmt.Println("  Result:           FAIL (exit-eligible under RPIP-73 with these parameters)")
	}

	if len(resp.MissedEpochList) > 0 {
		fmt.Printf("\nMissed target epochs (challengeable):\n")
		printEpochList(resp.MissedEpochList)
	}
}

func printEpochList(epochs []uint64) {
	const perLine = 8
	for i, e := range epochs {
		if i%perLine == 0 {
			if i > 0 {
				fmt.Println()
			}
			fmt.Print("  ")
		}
		fmt.Printf("%d ", e)
	}
	fmt.Println()
}
