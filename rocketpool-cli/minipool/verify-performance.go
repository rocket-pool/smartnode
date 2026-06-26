package minipool

import (
	"fmt"
	"strings"
	"time"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	verifyperf "github.com/rocket-pool/smartnode/shared/utils/cli/verify-performance"
)

// validateMinipoolTargets checks that the verify-performance targets argument
// is either "all" or a comma-separated list of valid minipool addresses.
func validateMinipoolTargets(targets string) error {
	if strings.EqualFold(strings.TrimSpace(targets), "all") {
		return nil
	}
	found := false
	for _, raw := range strings.Split(targets, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if _, err := cliutils.ValidateAddress("minipool address", raw); err != nil {
			return err
		}
		found = true
	}
	if !found {
		return fmt.Errorf("no minipool address provided; supply an address, a comma-separated list, or 'all'")
	}
	return nil
}

func verifyMinipoolPerformance(targets string, startEpoch uint64, epochs uint64, yes bool) error {
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	startEpoch, endEpoch, err := verifyperf.ResolveEpochRange(rp, startEpoch, epochs)
	if err != nil {
		return err
	}
	if !yes && !verifyperf.ConfirmLargeRange(endEpoch-startEpoch+1) {
		return verifyperf.PrintCancelled()
	}

	start := time.Now()
	resp, err := rp.VerifyMinipoolPerformance(targets, startEpoch, endEpoch)
	if err != nil {
		return err
	}
	elapsed := time.Since(start)

	verifyperf.PrintBatchResults(resp, func(r api.VerifyPerformanceResult) string {
		return "minipool " + r.MinipoolAddress.Hex()
	})
	verifyperf.PrintElapsed(elapsed)
	return nil
}
