// Package verifyperformance contains shared CLI helpers for the
// `rocketpool minipool verify-performance` and
// `rocketpool megapool verify-performance` commands.
package verifyperformance

import (
	"fmt"
	"math/big"
	"time"

	"github.com/rocket-pool/smartnode/shared/services/performance"
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
			epochs = performance.DefaultPerformancePeriodEpochs
		}
	}
	endEpoch := startEpoch + epochs - 1
	return startEpoch, endEpoch, nil
}

// ConfirmLargeRange prompts the user when the requested epoch range exceeds
// LargeEpochRangeWarning and returns true if the user accepts the warning.
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

// PrintElapsed prints the wall-clock time the verification took, rounded to the
// millisecond.
func PrintElapsed(elapsed time.Duration) {
	fmt.Printf("\nCompleted in %s\n", elapsed.Round(time.Millisecond))
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
		if resp.Challengeable {
			fmt.Printf("\nMissed target epochs (challengeable):\n")
		} else {
			fmt.Printf("\nMissed target epochs (not challengeable):\n")
		}
		printEpochList(resp.MissedEpochList)
	}
}

// PrintBatchResults writes a batch of verify-performance results to stdout. For
// each result it prints the per-validator detail (or the per-validator error),
// then a closing summary. labelFor produces the human-facing label for a
// result, e.g. "minipool 0x..." or "megapool validator 3".
func PrintBatchResults(resp api.VerifyPerformanceBatchResponse, labelFor func(api.VerifyPerformanceResult) string) {
	var passed int
	var failedLabels []string
	var erroredLabels []string
	for i, result := range resp.Results {
		if i > 0 {
			fmt.Println()
		}
		label := labelFor(result)
		if result.Error != "" {
			fmt.Printf("RPIP-73 target-vote performance for %s\n", label)
			fmt.Printf("  Error:            %s\n", result.Error)
			erroredLabels = append(erroredLabels, label)
			continue
		}
		if result.Performance == nil {
			fmt.Printf("RPIP-73 target-vote performance for %s\n", label)
			fmt.Printf("  Error:            no result returned\n")
			erroredLabels = append(erroredLabels, label)
			continue
		}
		PrintResult(*result.Performance, label)
		if result.Performance.PassesThreshold {
			passed++
		} else {
			failedLabels = append(failedLabels, label)
		}
	}

	if len(resp.Results) > 1 {
		fmt.Printf("\nSummary: %d validator(s) checked - %d pass, %d fail, %d errored\n",
			len(resp.Results), passed, len(failedLabels), len(erroredLabels))
	}

	// List the validators that failed (and any that errored) at the very end so
	// they are easy to spot without scrolling back through every result.
	if len(failedLabels) > 0 {
		fmt.Printf("\nFailed validators (exit-eligible under RPIP-73):\n")
		for _, label := range failedLabels {
			fmt.Printf("  - %s\n", label)
		}
	}
	if len(erroredLabels) > 0 {
		fmt.Printf("\nErrored validators (could not be verified):\n")
		for _, label := range erroredLabels {
			fmt.Printf("  - %s\n", label)
		}
	}
}

// ChallengeGroup is a set of validators sharing an identical missed-epoch
// set, challengeable together in a single challengeMegapool call.
type ChallengeGroup struct {
	ValidatorIds  []uint32
	StartEpoch    uint64
	Participation []*big.Int
	MissedEpochs  []uint64
}

// GroupChallengeable groups the challengeable validators of a batch result by
// identical missed-epoch sets, preserving the order in which they appear in
// the results. Errored, passing, and non-challengeable validators are
// excluded.
func GroupChallengeable(results []api.VerifyPerformanceResult) []ChallengeGroup {
	groups := []ChallengeGroup{}
	groupIndexByKey := map[string]int{}
	for _, result := range results {
		perf := result.Performance
		if result.Error != "" || perf == nil || !perf.Challengeable || len(perf.MissedEpochList) == 0 {
			continue
		}
		key := fmt.Sprint(perf.StartEpoch, perf.MissedEpochList)
		if i, ok := groupIndexByKey[key]; ok {
			groups[i].ValidatorIds = append(groups[i].ValidatorIds, result.ValidatorId)
			continue
		}
		groupIndexByKey[key] = len(groups)
		groups = append(groups, ChallengeGroup{
			ValidatorIds:  []uint32{result.ValidatorId},
			StartEpoch:    perf.StartEpoch,
			Participation: perf.Participation,
			MissedEpochs:  perf.MissedEpochList,
		})
	}
	return groups
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
