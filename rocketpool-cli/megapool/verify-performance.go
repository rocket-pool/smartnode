package megapool

import (
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	verifyperf "github.com/rocket-pool/smartnode/shared/utils/cli/verify-performance"
)

// validateMegapoolTargets checks that the verify-performance targets argument
// is either "all" or a comma-separated list of valid validator ids.
func validateMegapoolTargets(targets string) error {
	if strings.EqualFold(strings.TrimSpace(targets), "all") {
		return nil
	}
	found := false
	for _, raw := range strings.Split(targets, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if _, err := cliutils.ValidateUint32("validator-id", raw); err != nil {
			return err
		}
		found = true
	}
	if !found {
		return fmt.Errorf("no validator id provided; supply an id, a comma-separated list, or 'all'")
	}
	return nil
}

func verifyMegapoolPerformance(megapoolAddress common.Address, targetValidators string, startEpoch uint64, epochs uint64, yes bool) error {
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
	resp, err := rp.VerifyMegapoolValidatorPerformance(megapoolAddress, targetValidators, startEpoch, endEpoch)
	if err != nil {
		return err
	}
	elapsed := time.Since(start)

	verifyperf.PrintBatchResults(resp, func(r api.VerifyPerformanceResult) string {
		if (megapoolAddress != common.Address{}) {
			return fmt.Sprintf("megapool %s validator %d", megapoolAddress.Hex(), r.ValidatorId)
		}
		return fmt.Sprintf("megapool validator %d", r.ValidatorId)
	})
	verifyperf.PrintElapsed(elapsed)
	return nil
}
