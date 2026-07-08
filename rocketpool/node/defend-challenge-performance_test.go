package node

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/rocket-pool/smartnode/shared/services/performance"
)

func TestGetChallengedEpochs(t *testing.T) {
	tests := []struct {
		name       string
		startEpoch uint64
		words      []*big.Int
		want       []uint64
	}{
		{
			name:       "no words",
			startEpoch: 100,
			words:      []*big.Int{},
			want:       []uint64{},
		},
		{
			name:       "single word, no bits set",
			startEpoch: 100,
			words:      []*big.Int{big.NewInt(0)},
			want:       []uint64{},
		},
		{
			name:       "single word, single bit set",
			startEpoch: 100,
			words:      []*big.Int{big.NewInt(1)},
			want:       []uint64{100},
		},
		{
			name:       "single word, scattered bits",
			startEpoch: 100,
			// bits 0, 3, and 5 set -> 1 + 8 + 32 = 41
			words: []*big.Int{big.NewInt(41)},
			want:  []uint64{100, 103, 105},
		},
		{
			name:       "single word, non-zero start offset within word",
			startEpoch: 50,
			words:      []*big.Int{new(big.Int).Lsh(big.NewInt(1), 10)}, // bit 10 set
			want:       []uint64{60},
		},
		{
			name:       "multiple words, bits in each",
			startEpoch: 105000,
			words: []*big.Int{
				big.NewInt(1),                        // bit 0 of word 0 -> epoch 105000
				new(big.Int).Lsh(big.NewInt(1), 5),   // bit 5 of word 1 -> epoch 105000 + 256 + 5
				new(big.Int).Lsh(big.NewInt(1), 255), // bit 255 of word 2 -> epoch 105000 + 512 + 255
			},
			want: []uint64{105000, 105261, 105767},
		},
		{
			name:       "all bits set in a word",
			startEpoch: 0,
			words: []*big.Int{
				new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 200), big.NewInt(1)), // bits 0..199 set
			},
			want: allEpochsInRange(0, 200),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			challenge := megapoolPerformanceChallenge{
				startEpoch:            tc.startEpoch,
				participationCallData: tc.words,
			}
			got := challenge.getChallengedEpochs()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("getChallengedEpochs() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestParticipationBitsetRoundTrip pins the encoder in the performance
// package and the decoder here to the same bit-layout convention: encoding a
// missed-epoch list and decoding it back must return the original list.
func TestParticipationBitsetRoundTrip(t *testing.T) {
	tests := []struct {
		name         string
		startEpoch   uint64
		endEpoch     uint64
		missedEpochs []uint64
	}{
		{
			name:         "no missed epochs",
			startEpoch:   105000,
			endEpoch:     105500,
			missedEpochs: []uint64{},
		},
		{
			name:         "scattered epochs across word boundaries",
			startEpoch:   105000,
			endEpoch:     106000,
			missedEpochs: []uint64{105000, 105255, 105256, 105511, 105767, 106000},
		},
		{
			name:         "every epoch in a small range",
			startEpoch:   200,
			endEpoch:     209,
			missedEpochs: allEpochsInRange(200, 10),
		},
		{
			name:         "every epoch in a large range",
			startEpoch:   1,
			endEpoch:     44032,
			missedEpochs: allEpochsInRange(1, 44032),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			challenge := megapoolPerformanceChallenge{
				startEpoch:            tc.startEpoch,
				participationCallData: performance.EncodeParticipationBitset(tc.startEpoch, tc.endEpoch, tc.missedEpochs),
			}
			got := challenge.getChallengedEpochs()
			if !reflect.DeepEqual(got, tc.missedEpochs) {
				t.Errorf("round trip = %v, want %v", got, tc.missedEpochs)
			}
		})
	}
}

// allEpochsInRange returns [start, start+count) as a slice, matching the
// order getChallengedEpochs produces for a fully-set bitmap word.
func allEpochsInRange(start uint64, count int) []uint64 {
	epochs := make([]uint64, count)
	for i := 0; i < count; i++ {
		epochs[i] = start + uint64(i)
	}
	return epochs
}
