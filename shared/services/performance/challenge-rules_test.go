package performance

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func missedWords(period, misses uint64) []*big.Int {
	words := make([]*big.Int, (period+255)/256)
	for i := range words {
		words[i] = new(big.Int)
	}
	for i := uint64(0); i < misses; i++ {
		words[i/256].SetBit(words[i/256], int(i%256), 1)
	}
	return words
}

func TestValidateChallengeBitmapContractBoundaries(t *testing.T) {
	params := ChallengeParams{ExitsEnabled: true, PeriodEpochs: 1000, ProofBufferEpochs: 225, ThresholdWei: big.NewInt(940000000000000000)}
	for _, test := range []struct {
		name           string
		current, start uint64
		words          []*big.Int
		valid          bool
	}{
		{
			name:    "exact minimum in partial period",
			current: 160,
			start:   100,
			words:   missedWords(1000, 60),
			valid:   true,
		},
		{
			name:    "one below minimum",
			current: 160,
			start:   100,
			words:   missedWords(1000, 59),
			valid:   false,
		},
		{
			name:    "current epoch marked missed",
			current: 159,
			start:   100,
			words:   missedWords(1000, 60),
			valid:   false,
		},
		{
			name:    "start equals current",
			current: 100,
			start:   100,
			words:   missedWords(1000, 60),
			valid:   false,
		},
		{
			name:    "future start",
			current: 99,
			start:   100,
			words:   missedWords(1000, 60),
			valid:   false,
		},
		{
			name:    "last valid buffer epoch",
			current: 1324,
			start:   100,
			words:   missedWords(1000, 60),
			valid:   true,
		},
		{
			name:    "buffer boundary",
			current: 1325,
			start:   100,
			words:   missedWords(1000, 60),
			valid:   false,
		},
		{
			name:    "truncated partial period bitmap",
			current: 160,
			start:   100,
			words:   missedWords(60, 60),
			valid:   false,
		},
		{
			name:    "padding bit",
			current: 1200,
			start:   100,
			words:   missedWords(1000, 1001),
			valid:   false,
		},
		{
			name:    "missing word",
			current: 1200,
			start:   100,
			words:   []*big.Int{nil, big.NewInt(0), big.NewInt(0), big.NewInt(0)},
			valid:   false,
		},
		{
			name:    "negative word",
			current: 1200,
			start:   100,
			words:   []*big.Int{big.NewInt(-1), big.NewInt(0), big.NewInt(0), big.NewInt(0)},
			valid:   false,
		},
		{
			name:    "uint256 overflow",
			current: 1200,
			start:   100,
			words:   []*big.Int{new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(0), big.NewInt(0), big.NewInt(0)},
			valid:   false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidateChallengeBitmap(params, test.current, test.start, test.words); (err == nil) != test.valid {
				t.Fatalf("valid=%t, err=%v", test.valid, err)
			}
		})
	}
	params.PeriodEpochs = 1001 // ceil(60.06) = 61, not 60
	if ValidateChallengeBitmap(params, 200, 100, missedWords(1001, 60)) == nil {
		t.Fatal("fractional minimum rounded down")
	}
	if err := ValidateChallengeBitmap(params, 200, 100, missedWords(1001, 61)); err != nil {
		t.Fatal(err)
	}
	params.ExitsEnabled = false
	if ValidateChallengeBitmap(params, 200, 100, missedWords(1001, 61)) == nil {
		t.Fatal("exits disabled")
	}
}

func TestPartialWindowUsesFullPeriodThreshold(t *testing.T) {
	params := ChallengeParams{ExitsEnabled: true, PeriodEpochs: 1000, ProofBufferEpochs: 225, ThresholdWei: big.NewInt(940000000000000000)}
	for _, misses := range []uint64{59, 60} {
		resp := &api.VerifyPerformanceResponse{StartEpoch: 100, EndEpoch: 199, Participation: missedWords(100, misses)}
		SetChallengeability(resp, params, 202)
		if resp.Challengeable != (misses == 60) || len(resp.Participation) != 4 {
			t.Fatalf("incorrect partial window (%d misses): %+v", misses, resp)
		}
		resp.InactiveEpochs = 1
		SetChallengeability(resp, params, 202)
		if resp.Challengeable {
			t.Fatal("inactive member would allow activation defense")
		}
	}
}

func TestMeasurementWaitsForFollowingEpoch(t *testing.T) {
	for _, test := range []struct {
		head, start, end uint64
		valid            bool
	}{{12, 10, 10, true}, {11, 10, 10, false}, {10, 10, 10, false}, {0, 0, 0, false}, {12, 10, 9, false}, {^uint64(0), 0, ^uint64(0), false}} {
		if err := ValidateMeasurementRange(test.head, test.start, test.end); (err == nil) != test.valid {
			t.Fatalf("%+v: %v", test, err)
		}
	}
}

func TestDenebTargetInclusionWindow(t *testing.T) {
	for _, inclusion := range []uint64{321, 352, 360, 383, 384} {
		c := newEpochCache(nil, beacon.Eth2Config{SlotsPerEpoch: 32}, nil, nil)
		root := common.HexToHash("0x01")
		duty := attestationDuty{slot: 320, committeeIndex: 0, position: 0, committeeSizesAtDay: map[uint64]int{0: 1}}
		c.epochDuties[10] = map[string]attestationDuty{"1": duty}
		c.targetRoots[10] = rootResult{root: root}
		for slot := uint64(321); slot <= 384; slot++ {
			c.attestations[slot] = cachedAttestations{}
		}
		bits := bitfield.NewBitlist(1)
		bits.SetBitAt(0, true)
		att := beacon.AttestationInfo{SlotIndex: 320, TargetEpoch: 10, TargetRoot: root, AggregationBits: bits}
		att.Committees = bitfield.NewBitvector64()
		att.Committees.SetBitAt(0, true)
		c.attestations[inclusion] = cachedAttestations{exists: true, attestations: []beacon.AttestationInfo{att}}
		got, err := c.evaluateEpoch("1", 1, 10)
		if err != nil {
			t.Fatal(err)
		}
		if (got == epochResultTimely) != (inclusion < 384) {
			t.Fatalf("inclusion %d: got %v", inclusion, got)
		}
	}
}
