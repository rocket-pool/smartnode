package node

import (
	"math/big"
	"testing"
)

func TestChooseMegapoolStake(t *testing.T) {
	deadline := big.NewInt(1_000)
	eligible := megapoolStakeChoice{
		inPrestake:        true,
		prooflessEnabled:  true,
		signatureRetained: true,
		deadline:          deadline,
		headTimestamp:     1_000,
	}

	if got := chooseMegapoolStake(eligible); got != stakeWithoutProof {
		t.Fatalf("deadline equal to head timestamp chose %d", got)
	}
	eligible.headTimestamp = 1_001
	if got := chooseMegapoolStake(eligible); got != stakeWithoutProof {
		t.Fatalf("deadline before head timestamp chose %d", got)
	}

	waiting := eligible
	waiting.headTimestamp = 999
	if got := chooseMegapoolStake(waiting); got != stakeSkip {
		t.Fatalf("deadline one second in the future chose %d", got)
	}

	inState := waiting
	inState.inBeaconState = true
	if got := chooseMegapoolStake(inState); got != stakeWithProof {
		t.Fatalf("validator in the beacon state before the deadline chose %d", got)
	}
	inState.prooflessEnabled = false
	inState.signatureRetained = false
	inState.deadline = nil
	if got := chooseMegapoolStake(inState); got != stakeWithProof {
		t.Fatalf("validator in the beacon state with proofless staking disabled chose %d", got)
	}

	notPrestake := eligible
	notPrestake.inPrestake = false
	notPrestake.inBeaconState = true
	if got := chooseMegapoolStake(notPrestake); got != stakeSkip {
		t.Fatalf("validator that is not prestaked chose %d", got)
	}

	disabled := eligible
	disabled.prooflessEnabled = false
	if got := chooseMegapoolStake(disabled); got != stakeSkip {
		t.Fatalf("proofless staking disabled chose %d", got)
	}
	missingSignature := eligible
	missingSignature.signatureRetained = false
	if got := chooseMegapoolStake(missingSignature); got != stakeSkip {
		t.Fatalf("missing retained signature chose %d", got)
	}
	unsetDeadline := eligible
	unsetDeadline.deadline = big.NewInt(0)
	if got := chooseMegapoolStake(unsetDeadline); got != stakeSkip {
		t.Fatalf("zero deadline chose %d", got)
	}
	if got := chooseMegapoolStake(megapoolStakeChoice{
		inPrestake:        true,
		prooflessEnabled:  true,
		signatureRetained: true,
		deadline:          nil,
		headTimestamp:     1_000,
	}); got != stakeSkip {
		t.Fatalf("nil deadline chose %d", got)
	}

	huge := new(big.Int).Lsh(big.NewInt(1), 64)
	if challengeDeadlineElapsed(huge, ^uint64(0)) {
		t.Fatal("deadline above uint64 was treated as elapsed")
	}
}

func TestValidatorInSelectedBeaconState(t *testing.T) {
	included, err := validatorInSelectedBeaconState("", 10)
	if err != nil || included {
		t.Fatalf("empty index = %v, %v", included, err)
	}
	included, err = validatorInSelectedBeaconState("2", 3)
	if err != nil || !included {
		t.Fatalf("index inside the state = %v, %v", included, err)
	}
	included, err = validatorInSelectedBeaconState("2", 2)
	if err != nil || included {
		t.Fatalf("index at the end of the state = %v, %v", included, err)
	}
	if _, err := validatorInSelectedBeaconState("nope", 2); err == nil {
		t.Fatal("non-numeric index did not fail")
	}
}
