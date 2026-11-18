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

	pastDeadline := eligible
	pastDeadline.inBeaconState = true
	if got := chooseMegapoolStake(pastDeadline); got != stakeWithoutProof {
		t.Fatalf("validator in the beacon state after the deadline chose %d", got)
	}
	pastDeadline.prooflessEnabled = false
	if got := chooseMegapoolStake(pastDeadline); got != stakeWithProof {
		t.Fatalf("validator in the beacon state after the deadline with proofless staking disabled chose %d", got)
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

func TestProoflessStakeSupported(t *testing.T) {
	if prooflessStakeSupported(1) {
		t.Fatal("delegate version 1 was treated as supporting proofless staking")
	}
	if !prooflessStakeSupported(2) {
		t.Fatal("delegate version 2 was treated as missing proofless staking")
	}
}

func TestValidatorNeedsStakeProof(t *testing.T) {
	deadline := big.NewInt(1_000)
	choice := megapoolStakeChoice{
		inPrestake:        true,
		inBeaconState:     true,
		prooflessEnabled:  true,
		signatureRetained: true,
		deadline:          deadline,
		headTimestamp:     1_000,
	}
	if validatorNeedsStakeProof(choice, true) {
		t.Fatal("a validator past the deadline needs a beacon state")
	}
	if validatorNeedsStakeProof(choice, false) {
		t.Fatal("a validator the beacon has not listed needs a beacon state")
	}

	choice.headTimestamp = 999
	if !validatorNeedsStakeProof(choice, true) {
		t.Fatal("a validator before the deadline with a beacon index does not need a beacon state")
	}
	if validatorNeedsStakeProof(choice, false) {
		t.Fatal("a validator before the deadline and not listed by the beacon needs a beacon state")
	}

	choice.headTimestamp = 1_000
	choice.signatureRetained = false
	if !validatorNeedsStakeProof(choice, true) {
		t.Fatal("a validator with no retained signature does not need a beacon state")
	}
	choice.signatureRetained = true
	choice.prooflessEnabled = false
	if !validatorNeedsStakeProof(choice, true) {
		t.Fatal("a listed validator with proofless staking disabled does not need a beacon state")
	}
	if validatorNeedsStakeProof(choice, false) {
		t.Fatal("an unlisted validator with proofless staking disabled needs a beacon state")
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
