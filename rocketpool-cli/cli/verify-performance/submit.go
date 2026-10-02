package verifyperformance

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// SubmitChallenges submits one bounded list per shared bitmap and node owner.
func SubmitChallenges(rp *rocketpool.Client, address common.Address, results api.VerifyPerformanceBatchResponse, minipools, yes bool) error {
	for _, group := range GroupChallengeable(results.Results) {
		var can api.CanChallengeMegapoolPerformanceResponse
		var err error
		var targets interface{} = group.ValidatorIds
		if minipools {
			targets = group.MinipoolAddresses
			can, err = rp.CanChallengeMinipoolPerformance(group.MinipoolAddresses, group.StartEpoch, group.Participation)
		} else {
			can, err = rp.CanChallengeMegapoolPerformance(address, group.ValidatorIds, group.StartEpoch, group.Participation)
		}
		if err != nil {
			return err
		}
		if !can.CanChallenge {
			fmt.Printf("Skipping %v: %s\n", targets, can.Reason)
			continue
		}
		bond := new(big.Rat).SetFrac(can.ChallengeBond, big.NewInt(1e18)).FloatString(6)
		fmt.Printf("\nChallenge targets: %v\nOne bond of %s staked RPL covers this entire list.\n", targets, bond)
		if err := gas.AssignMaxFeeAndLimit(can.GasLimits, rp, yes); err != nil {
			return err
		}
		if prompt.Declined(yes, "Submit this performance challenge and lock %s staked RPL?", bond) {
			fmt.Println("Skipped.")
			continue
		}
		var result api.ChallengeMegapoolPerformanceResponse
		if minipools {
			result, err = rp.ChallengeMinipoolPerformance(group.MinipoolAddresses, group.StartEpoch, group.Participation, can.ChallengeBond)
		} else {
			result, err = rp.ChallengeMegapoolPerformance(address, group.ValidatorIds, group.StartEpoch, group.Participation, can.ChallengeBond)
		}
		if err != nil {
			return err
		}
		cliutils.PrintTransactionHash(rp, result.TxHash)
		if _, err := rp.WaitForTransaction(result.TxHash); err != nil {
			return err
		}
		fmt.Printf("Successfully submitted the performance challenge for %v.\n", targets)
	}
	return nil
}
