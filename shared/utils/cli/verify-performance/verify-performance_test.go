package verifyperformance

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// challengeableResult builds a challengeable verify result for the tests.
func challengeableResult(validatorId uint32, startEpoch uint64, missedEpochs []uint64, participation []*big.Int) api.VerifyPerformanceResult {
	return api.VerifyPerformanceResult{
		ValidatorId: validatorId,
		Performance: &api.VerifyPerformanceResponse{
			StartEpoch:      startEpoch,
			MissedEpochList: missedEpochs,
			Participation:   participation,
			Challengeable:   true,
		},
	}
}

func TestGroupChallengeable(t *testing.T) {
	missedA := []uint64{105000, 105003}
	missedB := []uint64{105001}
	participationA := []*big.Int{big.NewInt(9)} // bits 0 and 3
	participationB := []*big.Int{big.NewInt(2)} // bit 1

	passing := api.VerifyPerformanceResult{
		ValidatorId: 90,
		Performance: &api.VerifyPerformanceResponse{
			StartEpoch:      105000,
			MissedEpochList: []uint64{},
			Challengeable:   false,
		},
	}
	notChallengeable := api.VerifyPerformanceResult{
		ValidatorId: 91,
		Performance: &api.VerifyPerformanceResponse{
			StartEpoch:      105000,
			MissedEpochList: missedA,
			Challengeable:   false,
		},
	}
	errored := api.VerifyPerformanceResult{
		ValidatorId: 92,
		Error:       "validator not found",
	}

	tests := []struct {
		name    string
		results []api.VerifyPerformanceResult
		want    []ChallengeGroup
	}{
		{
			name:    "no results",
			results: []api.VerifyPerformanceResult{},
			want:    []ChallengeGroup{},
		},
		{
			name:    "errored, passing and non-challengeable results are excluded",
			results: []api.VerifyPerformanceResult{passing, notChallengeable, errored},
			want:    []ChallengeGroup{},
		},
		{
			name: "identical missed epochs merge into one group",
			results: []api.VerifyPerformanceResult{
				challengeableResult(3, 105000, missedA, participationA),
				challengeableResult(7, 105000, missedA, participationA),
			},
			want: []ChallengeGroup{
				{
					ValidatorIds:  []uint32{3, 7},
					StartEpoch:    105000,
					Participation: participationA,
					MissedEpochs:  missedA,
				},
			},
		},
		{
			name: "different missed epochs form separate groups in first-seen order",
			results: []api.VerifyPerformanceResult{
				challengeableResult(3, 105000, missedA, participationA),
				challengeableResult(5, 105000, missedB, participationB),
				errored,
				challengeableResult(7, 105000, missedA, participationA),
			},
			want: []ChallengeGroup{
				{
					ValidatorIds:  []uint32{3, 7},
					StartEpoch:    105000,
					Participation: participationA,
					MissedEpochs:  missedA,
				},
				{
					ValidatorIds:  []uint32{5},
					StartEpoch:    105000,
					Participation: participationB,
					MissedEpochs:  missedB,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GroupChallengeable(tc.results)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("GroupChallengeable() = %+v, want %+v", got, tc.want)
			}
		})
	}
}
