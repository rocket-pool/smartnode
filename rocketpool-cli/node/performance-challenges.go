package node

import (
	"fmt"
	"math/big"
	"time"

	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func validatePerformanceChallengeID(value string) (*big.Int, error) {
	id, ok := new(big.Int).SetString(value, 10)
	if !ok || id.Sign() <= 0 || id.BitLen() > 256 {
		return nil, fmt.Errorf("challenge ID must be a positive uint256")
	}
	return id, nil
}

func getPerformanceChallenges(id *big.Int) error {
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	response, err := rp.GetPerformanceChallenges(id)
	if err != nil {
		return err
	}
	if !response.Saturn2Deployed {
		fmt.Println(cliutils.Saturn2NotDeployedMessage)
		return nil
	}
	if len(response.Challenges) == 0 {
		fmt.Println("No matching performance challenges.")
		return nil
	}

	for _, details := range response.Challenges {
		challenge := details.Challenge
		status := details.State
		bond := new(big.Rat).SetFrac(status.BondAmount, big.NewInt(1e18)).FloatString(18)
		deadline := time.Unix(status.ResponseDeadline.Int64(), 0).UTC().Format(time.RFC3339)

		fmt.Printf("Challenge %s; start epoch %d\n", challenge.ChallengeId, challenge.StartEpoch)
		if len(challenge.MinipoolAddresses) > 0 {
			fmt.Printf("  Node: %s; minipools: %v\n", challenge.NodeAddress, challenge.MinipoolAddresses)
		} else {
			fmt.Printf("  Megapool: %s; validators: %v\n", challenge.MegapoolAddress, challenge.ValidatorIds)
		}
		fmt.Printf("  Proposer: %s; responder: %s\n", status.Proposer, status.Responder)
		fmt.Printf("  Response deadline: %s\n", deadline)
		fmt.Printf("  Defeated: %t; finalised: %t; bond settled: %t; recorded bond: %s RPL\n", status.Responded, status.Finalised, status.BondSettled, bond)
	}
	return nil
}

func actOnPerformanceChallenge(id *big.Int, defend bool, yes bool) error {
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	canAct := rp.CanFinalisePerformanceChallenge
	submit := rp.FinalisePerformanceChallenge
	action := "Finalise"
	if defend {
		canAct = rp.CanDefendPerformanceChallenge
		submit = rp.DefendPerformanceChallenge
		action = "Defend"
	}

	response, err := canAct(id)
	if err != nil {
		return err
	}
	if !response.CanAct {
		fmt.Println(response.Reason)
		return nil
	}
	if !defend {
		fmt.Println("Finalisation requests exits for eligible challenged validators. Bond settlement is separate.")
	}

	if err := gas.AssignMaxFeeAndLimit(response.GasLimits, rp, yes); err != nil {
		return err
	}
	if prompt.Declined(yes, "%s performance challenge %s?", action, id) {
		fmt.Println("Cancelled.")
		return nil
	}

	result, err := submit(id)
	if err != nil {
		return err
	}
	cliutils.PrintTransactionHash(rp, result.TxHash)
	if _, err := rp.WaitForTransaction(result.TxHash); err != nil {
		return err
	}
	fmt.Printf("Challenge %s: %s completed.\n", id, action)
	return nil
}
