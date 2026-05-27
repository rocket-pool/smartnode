package watchtower

import "testing"

func TestGetIndexToSubmit_CountEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		blockNumber uint64
		count       uint64
		expected    uint64
	}{
		{
			name:        "zero members returns zero",
			blockNumber: 0,
			count:       0,
			expected:    0,
		},
		{
			name:        "one member always returns zero",
			blockNumber: 123456789,
			count:       1,
			expected:    0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := getIndexToSubmit(test.blockNumber, test.count)
			if actual != test.expected {
				t.Fatalf("wrong index: got %d, want %d", actual, test.expected)
			}
		})
	}
}

func TestGetIndexToSubmit_Count10_PerEpochPermutationAndFairness(t *testing.T) {
	const (
		count  = uint64(10)
		epochs = uint64(200)
	)

	totalTurnsByMember := make([]uint64, count)
	firstTurnByMember := make([]uint64, count)

	for epoch := uint64(0); epoch < epochs; epoch++ {
		seenThisEpoch := make([]bool, count)

		for position := uint64(0); position < count; position++ {
			turn := epoch*count + position
			turnStartBlock := turn * BlocksPerTurn
			midTurnBlock := turnStartBlock + (BlocksPerTurn / 2)

			indexFromStart := getIndexToSubmit(turnStartBlock, count)
			indexFromMidTurn := getIndexToSubmit(midTurnBlock, count)

			if indexFromStart != indexFromMidTurn {
				t.Fatalf("index changed inside turn (epoch %d, position %d): start=%d mid=%d",
					epoch, position, indexFromStart, indexFromMidTurn)
			}

			if indexFromStart >= count {
				t.Fatalf("index out of range (epoch %d, position %d): got %d, count=%d",
					epoch, position, indexFromStart, count)
			}

			if seenThisEpoch[indexFromStart] {
				t.Fatalf("duplicate member in epoch %d: member %d appears more than once",
					epoch, indexFromStart)
			}
			seenThisEpoch[indexFromStart] = true
			totalTurnsByMember[indexFromStart]++

			if position == 0 {
				firstTurnByMember[indexFromStart]++
			}
		}

		for member := uint64(0); member < count; member++ {
			if !seenThisEpoch[member] {
				t.Fatalf("epoch %d is missing member %d", epoch, member)
			}
		}
	}

	// Every epoch is a full permutation, so every member appears exactly once per
	// epoch and therefore exactly `epochs` times overall.
	for member := uint64(0); member < count; member++ {
		if totalTurnsByMember[member] != epochs {
			t.Fatalf("member %d fairness mismatch: got %d turns, want %d",
				member, totalTurnsByMember[member], epochs)
		}
	}

}
