package performance

import (
	"math/big"
	"reflect"
	"testing"
)

func TestEncodeParticipationBitset(t *testing.T) {
	tests := []struct {
		name         string
		startEpoch   uint64
		endEpoch     uint64
		missedEpochs []uint64
		want         []*big.Int
	}{
		{
			name:         "end before start",
			startEpoch:   100,
			endEpoch:     99,
			missedEpochs: []uint64{},
			want:         []*big.Int{},
		},
		{
			name:         "no missed epochs, range spanning two words",
			startEpoch:   100,
			endEpoch:     399, // 300 epochs -> 2 words
			missedEpochs: []uint64{},
			want:         []*big.Int{big.NewInt(0), big.NewInt(0)},
		},
		{
			name:         "single missed epoch at start",
			startEpoch:   100,
			endEpoch:     355, // exactly 256 epochs -> 1 word
			missedEpochs: []uint64{100},
			want:         []*big.Int{big.NewInt(1)},
		},
		{
			name:         "scattered bits within one word",
			startEpoch:   100,
			endEpoch:     200,
			missedEpochs: []uint64{100, 103, 105},
			// bits 0, 3, and 5 set -> 1 + 8 + 32 = 41
			want: []*big.Int{big.NewInt(41)},
		},
		{
			name:         "word boundary: offsets 255 and 256",
			startEpoch:   1000,
			endEpoch:     1511, // 512 epochs -> 2 words
			missedEpochs: []uint64{1255, 1256},
			want: []*big.Int{
				new(big.Int).Lsh(big.NewInt(1), 255), // bit 255 of word 0
				big.NewInt(1),                        // bit 0 of word 1
			},
		},
		{
			name:         "range not a multiple of the word size",
			startEpoch:   0,
			endEpoch:     256, // 257 epochs -> 2 words
			missedEpochs: []uint64{256},
			want:         []*big.Int{big.NewInt(0), big.NewInt(1)},
		},
		{
			name:         "epochs outside the range are ignored",
			startEpoch:   100,
			endEpoch:     200,
			missedEpochs: []uint64{99, 150, 201},
			want:         []*big.Int{new(big.Int).Lsh(big.NewInt(1), 50)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := EncodeParticipationBitset(tc.startEpoch, tc.endEpoch, tc.missedEpochs)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("EncodeParticipationBitset(%d, %d, %v) = %v, want %v",
					tc.startEpoch, tc.endEpoch, tc.missedEpochs, got, tc.want)
			}
		})
	}
}
