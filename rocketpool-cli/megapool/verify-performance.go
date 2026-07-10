package megapool

import (
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	verifyperf "github.com/rocket-pool/smartnode/shared/utils/cli/verify-performance"
	"github.com/rocket-pool/smartnode/shared/utils/math"
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

	return challengePerformance(rp, megapoolAddress, resp, yes)
}

// challengePerformance drives the on-chain challenge flow for the
// challengeable validators of a verify-performance run: it groups validators
// sharing the same missed epochs so they can be confirmed together, then
// submits one challengeMegapool call per validator (challenges are
// per-validator on-chain, each requiring its own RPL bond) after the gas
// confirmation.
func challengePerformance(rp *rocketpool.Client, megapoolAddress common.Address, resp api.VerifyPerformanceBatchResponse, yes bool) error {
	groups := verifyperf.GroupChallengeable(resp.Results)
	if len(groups) == 0 {
		return nil
	}

	settings, err := rp.PDAOGetSettings()
	if err != nil {
		return fmt.Errorf("error fetching pDAO settings for the challenge bond: %w", err)
	}
	if !settings.Saturn2Deployed {
		fmt.Println("\nPerformance challenges are not available until Saturn 2 is deployed.")
		return nil
	}
	bondRpl := math.RoundDown(eth.WeiToEth(settings.Performance.ChallengeBond), 6)

	for _, group := range groups {
		ids := make([]string, len(group.ValidatorIds))
		for i, id := range group.ValidatorIds {
			ids[i] = fmt.Sprint(id)
		}
		fmt.Printf("\nValidator id(s) %s missed the same %d target epoch(s).\n", strings.Join(ids, ", "), len(group.MissedEpochs))
		fmt.Printf("Each validator is challenged individually and requires a bond of %.6f RPL.\n", bondRpl)

		if prompt.Declined(yes, "Do you want to challenge validator id(s) %s with a bond of %.6f RPL each?", strings.Join(ids, ", "), bondRpl) {
			fmt.Println("Skipped.")
			continue
		}

		for _, validatorId := range group.ValidatorIds {
			can, err := rp.CanChallengeMegapoolPerformance(megapoolAddress, validatorId, group.StartEpoch, group.Participation)
			if err != nil {
				return err
			}
			if can.InsufficientRplBalance {
				fmt.Printf("The node wallet holds %.6f RPL but the challenge bond requires %.6f RPL. Skipping validator %d.\n",
					math.RoundDown(eth.WeiToEth(can.RplBalance), 6), math.RoundDown(eth.WeiToEth(can.ChallengeBond), 6), validatorId)
				continue
			}
			if !can.CanChallenge {
				fmt.Printf("The challenge for validator %d cannot be submitted. Skipping.\n", validatorId)
				continue
			}

			// Assign max fees
			err = gas.AssignMaxFeeAndLimit(can.GasInfo, rp, yes)
			if err != nil {
				return err
			}

			challengeResp, err := rp.ChallengeMegapoolPerformance(megapoolAddress, validatorId, group.StartEpoch, group.Participation)
			if err != nil {
				return err
			}

			fmt.Printf("Submitting the performance challenge for validator %d...\n", validatorId)
			cliutils.PrintTransactionHash(rp, challengeResp.TxHash)
			if _, err = rp.WaitForTransaction(challengeResp.TxHash); err != nil {
				return err
			}
			fmt.Printf("Successfully challenged validator %d.\n", validatorId)
		}
	}

	return nil
}
