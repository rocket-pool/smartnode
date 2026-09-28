package node

import (
	"fmt"
	"math/big"

	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func releaseChallengeBond(id *big.Int, yes bool) error {
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()
	can, err := rp.CanReleaseChallengeBond(id)
	if err != nil {
		return err
	}
	if !can.Saturn2Deployed {
		fmt.Println(cliutils.Saturn2NotDeployedMessage)
		return nil
	}
	if !can.CanSettle {
		fmt.Println(can.Reason)
		return nil
	}
	formatRPL := func(value *big.Int) string {
		return new(big.Rat).SetFrac(value, big.NewInt(1e18)).FloatString(18)
	}
	fmt.Printf("Challenge %s; proposer: %s\nRecorded bond: %s RPL\n", id, can.Proposer.Hex(), formatRPL(can.BondAmount))
	fmt.Println("This unlocks the proposer's existing staked RPL and does not finalise validator exits.")
	if err := gas.AssignMaxFeeAndLimit(can.GasLimits, rp, yes); err != nil {
		return err
	}
	if prompt.Declined(yes, "Release bond for challenge %s?", id) {
		fmt.Println("Cancelled.")
		return nil
	}
	result, err := rp.ReleaseChallengeBond(id)
	if err != nil {
		return err
	}
	cliutils.PrintTransactionHash(rp, result.TxHash)
	if _, err := rp.WaitForTransaction(result.TxHash); err != nil {
		return err
	}
	fmt.Printf("Challenge %s: Release bond completed.\n", id)
	return nil
}

func claimChallengeReward(id *big.Int, yes bool) error {
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()
	can, err := rp.CanClaimChallengeReward(id)
	if err != nil {
		return err
	}
	if !can.Saturn2Deployed {
		fmt.Println(cliutils.Saturn2NotDeployedMessage)
		return nil
	}
	if !can.CanSettle {
		fmt.Println(can.Reason)
		return nil
	}
	formatRPL := func(value *big.Int) string {
		return new(big.Rat).SetFrac(value, big.NewInt(1e18)).FloatString(18)
	}
	fmt.Printf("Challenge %s; proposer: %s\nRecorded bond: %s RPL\n", id, can.Proposer.Hex(), formatRPL(can.BondAmount))
	fmt.Printf("Estimated reward: %s staked RPL; burn: %s RPL. Amounts depend on the proposer's recoverable stake at execution.\n", formatRPL(can.EstimatedReward), formatRPL(can.EstimatedBurn))
	if err := gas.AssignMaxFeeAndLimit(can.GasLimits, rp, yes); err != nil {
		return err
	}
	if prompt.Declined(yes, "Claim defender reward for challenge %s?", id) {
		fmt.Println("Cancelled.")
		return nil
	}
	result, err := rp.ClaimChallengeReward(id)
	if err != nil {
		return err
	}
	cliutils.PrintTransactionHash(rp, result.TxHash)
	if _, err := rp.WaitForTransaction(result.TxHash); err != nil {
		return err
	}
	fmt.Printf("Challenge %s: Claim defender reward completed.\n", id)
	return nil
}
