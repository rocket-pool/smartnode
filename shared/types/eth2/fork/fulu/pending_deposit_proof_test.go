package fulu

import (
	"errors"
	"testing"

	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
)

func TestPendingDepositProofsUnsupported(t *testing.T) {
	_, _, err := (&BeaconState{}).PendingDepositProofs([]uint64{0})
	if !errors.Is(err, generic.ErrPendingDepositProofUnsupported) {
		t.Fatalf("expected unsupported proof error, got %v", err)
	}
}
