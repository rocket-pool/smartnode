package services

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestPerformanceChallengeRequestValidation(t *testing.T) {
	valid32 := make([]uint32, 32)
	for i := range valid32 {
		valid32[i] = uint32(i)
	}
	for _, test := range []struct {
		name    string
		request PerformanceChallengeRequest
		valid   bool
	}{
		{
			name:    "empty",
			request: PerformanceChallengeRequest{},
			valid:   false,
		},
		{
			name:    "validator zero",
			request: PerformanceChallengeRequest{ValidatorIds: []uint32{0}},
			valid:   true,
		},
		{
			name:    "32 validators",
			request: PerformanceChallengeRequest{ValidatorIds: valid32},
			valid:   true,
		},
		{
			name:    "33 validators",
			request: PerformanceChallengeRequest{ValidatorIds: append(append([]uint32{}, valid32...), 32)},
			valid:   false,
		},
		{
			name:    "duplicate validator",
			request: PerformanceChallengeRequest{ValidatorIds: []uint32{0, 0}},
			valid:   false,
		},
		{
			name:    "minipool",
			request: PerformanceChallengeRequest{MinipoolAddresses: []common.Address{{1}}},
			valid:   true,
		},
		{
			name:    "zero minipool",
			request: PerformanceChallengeRequest{MinipoolAddresses: []common.Address{{}}},
			valid:   false,
		},
		{
			name:    "duplicate minipool",
			request: PerformanceChallengeRequest{MinipoolAddresses: []common.Address{{1}, {1}}},
			valid:   false,
		},
		{
			name:    "mixed types",
			request: PerformanceChallengeRequest{ValidatorIds: []uint32{0}, MinipoolAddresses: []common.Address{{1}}},
			valid:   false,
		},
		{
			name:    "mixed addresses",
			request: PerformanceChallengeRequest{MegapoolAddress: common.Address{1}, MinipoolAddresses: []common.Address{{1}}},
			valid:   false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := test.request.Validate(); (err == nil) != test.valid {
				t.Fatalf("valid=%t, err=%v", test.valid, err)
			}
		})
	}
}

func TestAvailablePerformanceBondUsesUnlockedStake(t *testing.T) {
	for _, test := range []struct{ stake, locked, want int64 }{{100, 40, 60}, {100, 100, 0}, {90, 100, 0}, {0, 0, 0}} {
		stake, locked := big.NewInt(test.stake), big.NewInt(test.locked)
		got := AvailablePerformanceBond(stake, locked)
		if got.Int64() != test.want || stake.Int64() != test.stake || locked.Int64() != test.locked {
			t.Fatalf("incorrect available stake for %+v: %s", test, got)
		}
	}
}

func TestDefenseSearchContinuesPastUnavailableHistory(t *testing.T) {
	var attempted []uint64
	proof, err := findParticipationDefense([]uint64{10, 11, 12, 13}, func(epoch uint64) (bool, error) {
		if epoch == 10 {
			return false, fmt.Errorf("pruned block")
		}
		return epoch != 11, nil
	}, func(epoch uint64) (PerformanceDefenseProofs, error) {
		attempted = append(attempted, epoch)
		if epoch == 12 {
			return PerformanceDefenseProofs{}, fmt.Errorf("pruned SSZ state")
		}
		return PerformanceDefenseProofs{Offset: 3}, nil
	})
	if err != nil || proof == nil || proof.Offset != 3 || !reflect.DeepEqual(attempted, []uint64{12, 13}) {
		t.Fatalf("later valid proof lost: %+v, %v, %v", proof, err, attempted)
	}
	proof, err = findParticipationDefense([]uint64{10}, func(uint64) (bool, error) { return false, fmt.Errorf("pruned") }, nil)
	if proof != nil || err == nil {
		t.Fatal("missing history must not be reported as proven absence of a defense")
	}
	proof, err = findParticipationDefense([]uint64{10}, func(uint64) (bool, error) { return false, nil }, nil)
	if proof != nil || err != nil {
		t.Fatal("valid search with no defense must be distinguishable from unavailable data")
	}
}
