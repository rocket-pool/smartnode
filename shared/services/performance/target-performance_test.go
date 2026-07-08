package performance

import (
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func TestIsChallengeable(t *testing.T) {
	// Mainnet timing: 32 slots * 12s = 384s per epoch, so a 24h proof buffer
	// spans 86400 / 384 = 225 epochs.
	cfg := beacon.Eth2Config{SecondsPerEpoch: 384}
	params := ChallengeParams{
		ExitsEnabled: true,
		PeriodEpochs: 1000,
		ProofBuffer:  24 * time.Hour,
	}
	// The challenge window is period + proofBufferEpochs = 1225 epochs, so with
	// currentEpoch = 10000 the oldest challengeable start epoch is 8776.

	tests := []struct {
		name         string
		params       ChallengeParams
		cfg          beacon.Eth2Config
		currentEpoch uint64
		startEpoch   uint64
		endEpoch     uint64
		want         bool
	}{
		{
			name:         "recent full period",
			params:       params,
			cfg:          cfg,
			currentEpoch: 10000,
			startEpoch:   9000,
			endEpoch:     9999,
			want:         true,
		},
		{
			name:         "exits disabled",
			params:       ChallengeParams{ExitsEnabled: false, PeriodEpochs: 1000, ProofBuffer: 24 * time.Hour},
			cfg:          cfg,
			currentEpoch: 10000,
			startEpoch:   9000,
			endEpoch:     9999,
			want:         false,
		},
		{
			name:         "range one epoch too long",
			params:       params,
			cfg:          cfg,
			currentEpoch: 10000,
			startEpoch:   9000,
			endEpoch:     10000,
			want:         false,
		},
		{
			name:         "range one epoch too short",
			params:       params,
			cfg:          cfg,
			currentEpoch: 10000,
			startEpoch:   9000,
			endEpoch:     9998,
			want:         false,
		},
		{
			name:         "start epoch just inside the window",
			params:       params,
			cfg:          cfg,
			currentEpoch: 10000,
			startEpoch:   8776,
			endEpoch:     9775,
			want:         true,
		},
		{
			name:         "start epoch at the window boundary",
			params:       params,
			cfg:          cfg,
			currentEpoch: 10000,
			startEpoch:   8775,
			endEpoch:     9774,
			want:         false,
		},
		{
			name:         "current epoch smaller than the window",
			params:       params,
			cfg:          cfg,
			currentEpoch: 1000,
			startEpoch:   0,
			endEpoch:     999,
			want:         true,
		},
		{
			name:         "invalid beacon config",
			params:       params,
			cfg:          beacon.Eth2Config{},
			currentEpoch: 10000,
			startEpoch:   9000,
			endEpoch:     9999,
			want:         false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsChallengeable(tc.params, tc.cfg, tc.currentEpoch, tc.startEpoch, tc.endEpoch)
			if got != tc.want {
				t.Errorf("IsChallengeable(%+v, current %d, [%d, %d]) = %v, want %v",
					tc.params, tc.currentEpoch, tc.startEpoch, tc.endEpoch, got, tc.want)
			}
		})
	}
}

func TestExceedsChallengeThreshold(t *testing.T) {
	tests := []struct {
		name         string
		totalEpochs  uint64
		missedEpochs uint64
		thresholdPct float64
		want         bool
	}{
		{
			name:         "no epochs checked",
			totalEpochs:  0,
			missedEpochs: 0,
			thresholdPct: 94.0,
			want:         false,
		},
		{
			name:         "no missed epochs",
			totalEpochs:  1024,
			missedEpochs: 0,
			thresholdPct: 94.0,
			want:         false,
		},
		// 16/256 = 6.25% missed; the allowed slack is 100% - threshold.
		{
			name:         "missed share exactly at the allowed slack",
			totalEpochs:  256,
			missedEpochs: 16,
			thresholdPct: 93.75,
			want:         false,
		},
		{
			name:         "missed share above the allowed slack",
			totalEpochs:  256,
			missedEpochs: 16,
			thresholdPct: 94.0,
			want:         true,
		},
		{
			name:         "missed share below the allowed slack",
			totalEpochs:  256,
			missedEpochs: 16,
			thresholdPct: 93.0,
			want:         false,
		},
		{
			name:         "every epoch missed",
			totalEpochs:  256,
			missedEpochs: 256,
			thresholdPct: 94.0,
			want:         true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := &api.VerifyPerformanceResponse{
				TotalEpochs:             tc.totalEpochs,
				MissedEpochs:            tc.missedEpochs,
				PerformanceThresholdPct: tc.thresholdPct,
			}
			got := ExceedsChallengeThreshold(resp)
			if got != tc.want {
				t.Errorf("ExceedsChallengeThreshold(%d missed of %d, threshold %.2f%%) = %v, want %v",
					tc.missedEpochs, tc.totalEpochs, tc.thresholdPct, got, tc.want)
			}
		})
	}
}

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
