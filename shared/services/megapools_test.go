package services

import (
	"crypto/sha256"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// stubFinder returns a findInQueueFunc that always yields the given position.
// Pass nil to simulate a validator that is not found in the queue.
func stubFinder(pos *big.Int) findInQueueFunc {
	return func(
		_ *rocketpool.RocketPool,
		_ common.Address,
		_ uint32,
		_ string,
		_ *big.Int,
		_ *big.Int,
	) (*big.Int, error) {
		return pos, nil
	}
}

// makeQueueDetails builds an api.QueueDetails from plain uint64 values.
func makeQueueDetails(queueIndex, expressLen, standardLen, expressRate uint64) api.QueueDetails {
	return api.QueueDetails{
		QueueIndex:          new(big.Int).SetUint64(queueIndex),
		ExpressQueueLength:  new(big.Int).SetUint64(expressLen),
		StandardQueueLength: new(big.Int).SetUint64(standardLen),
		ExpressQueueRate:    expressRate,
	}
}

// estimatePosition calls calculatePositionInQueue and returns the uint64 result.
// zeroBasedPos is the 0-based index returned by findInQueue (0 = head of queue).
func estimatePosition(t *testing.T, qd api.QueueDetails, zeroBasedPos uint64, queueKey string) uint64 {
	t.Helper()
	result, err := calculatePositionInQueue(
		nil,
		qd,
		common.Address{},
		0,
		queueKey,
		stubFinder(new(big.Int).SetUint64(zeroBasedPos)),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result, got nil")
	}
	return result.Uint64()
}

// ---------------------------------------------------------------------------
// Contract cycle recap (expressQueueRate = 2, queueInterval = 3):
//   queueIndex % 3 == 0 → express
//   queueIndex % 3 == 1 → express
//   queueIndex % 3 == 2 → standard
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Express queue
// ---------------------------------------------------------------------------

// TestExpressQueue_FirstInEmptyQueue
// pos=1, queueIndex=0, expressRemainingInCycle=2.
// pos(1) <= remaining(2) → 0 standard entries before us.
// overallPosition = 1 + 0 = 1.
func TestExpressQueue_FirstInEmptyQueue(t *testing.T) {
	qd := makeQueueDetails(0, 1, 0, 2)
	if got := estimatePosition(t, qd, 0, "deposit.queue.express"); got != 1 {
		t.Errorf("want 1, got %d", got)
	}
}

// TestExpressQueue_SpansIntoNextCycle
// pos=3, queueIndex=0, expressRemainingInCycle=2.
// pos(3) > remaining(2) → standardEntriesBefore = ceil((3-2)/2) = 1.
// overallPosition = 3 + 1 = 4.
func TestExpressQueue_SpansIntoNextCycle(t *testing.T) {
	qd := makeQueueDetails(0, 10, 5, 2)
	if got := estimatePosition(t, qd, 2, "deposit.queue.express"); got != 4 {
		t.Errorf("want 4, got %d", got)
	}
}

// TestExpressQueue_MidCycle_NoSpill
// pos=1, queueIndex=1, expressUsedInCycle=1, expressRemainingInCycle=1.
// pos(1) <= remaining(1) → 0 standard entries before us.
// overallPosition = 1 + 0 = 1.
func TestExpressQueue_MidCycle_NoSpill(t *testing.T) {
	qd := makeQueueDetails(1, 5, 5, 2)
	if got := estimatePosition(t, qd, 0, "deposit.queue.express"); got != 1 {
		t.Errorf("want 1, got %d", got)
	}
}

// TestExpressQueue_MidCycle_Spills
// pos=2, queueIndex=1, expressRemainingInCycle=1.
// pos(2) > remaining(1) → standardEntriesBefore = ceil((2-1)/2) = 1.
// overallPosition = 2 + 1 = 3.
func TestExpressQueue_MidCycle_Spills(t *testing.T) {
	qd := makeQueueDetails(1, 10, 5, 2)
	if got := estimatePosition(t, qd, 1, "deposit.queue.express"); got != 3 {
		t.Errorf("want 3, got %d", got)
	}
}

// TestExpressQueue_StandardQueueCapped
// pos=10, queueIndex=0, uncapped standardEntriesBefore = ceil((10-2)/2) = 4,
// but standardQueueLength=1 → cap to 1.
// overallPosition = 10 + 1 = 11.
func TestExpressQueue_StandardQueueCapped(t *testing.T) {
	qd := makeQueueDetails(0, 20, 1, 2)
	if got := estimatePosition(t, qd, 9, "deposit.queue.express"); got != 11 {
		t.Errorf("want 11, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// Standard queue
// ---------------------------------------------------------------------------

// TestStandardQueue_FirstEntry_StartOfCycle
// pos=1, queueIndex=0, expressRemainingInCycle=2.
// expressEntriesBefore = (1-1)*2 + 2 = 2.
// overallPosition = 1 + 2 = 3.
func TestStandardQueue_FirstEntry_StartOfCycle(t *testing.T) {
	qd := makeQueueDetails(0, 10, 5, 2)
	if got := estimatePosition(t, qd, 0, "deposit.queue.standard"); got != 3 {
		t.Errorf("want 3, got %d", got)
	}
}

// TestStandardQueue_SecondEntry_StartOfCycle
// pos=2, queueIndex=0.
// expressEntriesBefore = (2-1)*2 + 2 = 4.
// overallPosition = 2 + 4 = 6.
func TestStandardQueue_SecondEntry_StartOfCycle(t *testing.T) {
	qd := makeQueueDetails(0, 10, 5, 2)
	if got := estimatePosition(t, qd, 1, "deposit.queue.standard"); got != 6 {
		t.Errorf("want 6, got %d", got)
	}
}

// TestStandardQueue_MidCycle
// pos=1, queueIndex=1, expressUsedInCycle=1, expressRemainingInCycle=1.
// expressEntriesBefore = 0*2 + 1 = 1.
// overallPosition = 1 + 1 = 2.
func TestStandardQueue_MidCycle(t *testing.T) {
	qd := makeQueueDetails(1, 10, 5, 2)
	if got := estimatePosition(t, qd, 0, "deposit.queue.standard"); got != 2 {
		t.Errorf("want 2, got %d", got)
	}
}

// TestStandardQueue_JustAfterStandardSlot
// queueIndex=3 wraps back to slot 0 of the next cycle (3%3=0).
// Same as start-of-cycle: expressRemainingInCycle=2.
// pos=1: expressEntriesBefore = 0 + 2 = 2.
// overallPosition = 1 + 2 = 3.
func TestStandardQueue_JustAfterStandardSlot(t *testing.T) {
	qd := makeQueueDetails(3, 10, 5, 2)
	if got := estimatePosition(t, qd, 0, "deposit.queue.standard"); got != 3 {
		t.Errorf("want 3, got %d", got)
	}
}

// TestStandardQueue_ExpressQueueCapped
// pos=5, queueIndex=0, uncapped expressEntriesBefore = (5-1)*2 + 2 = 10,
// but expressQueueLength=3 → cap to 3.
// overallPosition = 5 + 3 = 8.
func TestStandardQueue_ExpressQueueCapped(t *testing.T) {
	qd := makeQueueDetails(0, 3, 10, 2)
	if got := estimatePosition(t, qd, 4, "deposit.queue.standard"); got != 8 {
		t.Errorf("want 8, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

// TestNotInQueue — findInQueue returns nil → result must be nil, no error.
func TestNotInQueue(t *testing.T) {
	qd := makeQueueDetails(0, 5, 5, 2)
	result, err := calculatePositionInQueue(
		nil, qd, common.Address{}, 99, "deposit.queue.express", stubFinder(nil),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for validator not in queue, got %v", result)
	}
}

// TestHighExpressQueueRate — sanity check with expressQueueRate=4 (cycle of 5).
// Standard queue, pos=1, queueIndex=0.
// expressRemainingInCycle = 4.
// expressEntriesBefore = 0*4 + 4 = 4.
// overallPosition = 1 + 4 = 5.
func TestHighExpressQueueRate(t *testing.T) {
	qd := makeQueueDetails(0, 10, 5, 4)
	if got := estimatePosition(t, qd, 0, "deposit.queue.standard"); got != 5 {
		t.Errorf("want 5, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// Participation proof slot selection
// ---------------------------------------------------------------------------

// TestParticipationProofSlotRange — the proof state for challenged epoch E
// must come from epoch E+1, where E's flags live in
// previous_epoch_participation.
func TestParticipationProofSlotRange(t *testing.T) {
	tests := []struct {
		name            string
		challengedEpoch uint64
		slotsPerEpoch   uint64
		wantFirst       uint64
		wantLast        uint64
	}{
		{
			name:            "epoch zero",
			challengedEpoch: 0,
			slotsPerEpoch:   32,
			wantFirst:       32,
			wantLast:        63,
		},
		{
			name:            "mainnet-like epoch",
			challengedEpoch: 105000,
			slotsPerEpoch:   32,
			wantFirst:       105001 * 32,
			wantLast:        105002*32 - 1,
		},
		{
			name:            "single-slot epochs",
			challengedEpoch: 10,
			slotsPerEpoch:   1,
			wantFirst:       11,
			wantLast:        11,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			first, last := participationProofSlotRange(tc.challengedEpoch, tc.slotsPerEpoch)
			if first != tc.wantFirst || last != tc.wantLast {
				t.Errorf("participationProofSlotRange(%d, %d) = (%d, %d), want (%d, %d)",
					tc.challengedEpoch, tc.slotsPerEpoch, first, last, tc.wantFirst, tc.wantLast)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// participationBitmapWitness
// ---------------------------------------------------------------------------

// testHashTree mirrors RocketNetworkParticipation.hashTree: raw uint256 words
// as leaves, zero-padded to the next power of two, hashed pairwise with
// sha256.
func testHashTree(t *testing.T, leaves []*big.Int) [32]byte {
	t.Helper()
	width := 1
	for width < len(leaves) {
		width *= 2
	}
	tree := make([][32]byte, width)
	for i, leaf := range leaves {
		leaf.FillBytes(tree[i][:])
	}
	for width > 1 {
		for i := 0; i < width; i += 2 {
			var pair [64]byte
			copy(pair[:32], tree[i][:])
			copy(pair[32:], tree[i+1][:])
			tree[i/2] = sha256.Sum256(pair[:])
		}
		width /= 2
	}
	return tree[0]
}

// testRestoreMerkleRoot mirrors RocketNetworkParticipation.restoreMerkleRoot:
// walk the generalized index from the leaf to the root, hashing with the
// witnesses.
func testRestoreMerkleRoot(t *testing.T, leaf [32]byte, gindex uint64, witnesses []common.Hash) [32]byte {
	t.Helper()
	if 1<<(len(witnesses)+1) <= gindex {
		t.Fatalf("invalid witness length %d for gindex %d", len(witnesses), gindex)
	}
	value := leaf
	i := 0
	for gindex != 1 {
		var pair [64]byte
		if gindex%2 == 1 {
			copy(pair[:32], witnesses[i][:])
			copy(pair[32:], value[:])
		} else {
			copy(pair[:32], value[:])
			copy(pair[32:], witnesses[i][:])
		}
		value = sha256.Sum256(pair[:])
		gindex /= 2
		i++
	}
	return value
}

func TestParticipationBitmapWitness(t *testing.T) {
	for _, wordCount := range []int{1, 2, 3, 5, 8} {
		// Deterministic, distinct bitmap words
		participation := make([]*big.Int, wordCount)
		for i := range participation {
			word := new(big.Int).Lsh(big.NewInt(int64(i)+1), 13)
			word.Add(word, big.NewInt(int64(i)*7+1))
			participation[i] = word
		}

		root := testHashTree(t, participation)
		width := uint64(1)
		for width < uint64(wordCount) {
			width *= 2
		}

		for leafIndex := uint64(0); leafIndex < uint64(wordCount); leafIndex++ {
			witness, err := participationBitmapWitness(participation, leafIndex)
			if err != nil {
				t.Fatalf("wordCount %d leafIndex %d: unexpected error: %v", wordCount, leafIndex, err)
			}

			var leaf [32]byte
			participation[leafIndex].FillBytes(leaf[:])

			// The contract computes the generalized index as
			// nextPowerOfTwo(leafCount) + leafIndex
			gindex := width + leafIndex
			restored := testRestoreMerkleRoot(t, leaf, gindex, witness)
			if restored != root {
				t.Fatalf("wordCount %d leafIndex %d: restored root %x does not match tree root %x", wordCount, leafIndex, restored, root)
			}
		}
	}
}

func TestParticipationBitmapWitnessSingleWord(t *testing.T) {
	participation := []*big.Int{big.NewInt(0b1011)}
	witness, err := participationBitmapWitness(participation, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A single-word bitmap has an empty branch: the root is the word itself
	if len(witness) != 0 {
		t.Fatalf("expected empty witness for a single-word bitmap, got %d hashes", len(witness))
	}
}

func TestParticipationBitmapWitnessErrors(t *testing.T) {
	if _, err := participationBitmapWitness([]*big.Int{}, 0); err == nil {
		t.Fatal("expected an error for an empty bitmap")
	}
	if _, err := participationBitmapWitness([]*big.Int{big.NewInt(1)}, 1); err == nil {
		t.Fatal("expected an error for an out-of-bounds leaf index")
	}
	tooBig := new(big.Int).Lsh(big.NewInt(1), 256)
	if _, err := participationBitmapWitness([]*big.Int{tooBig}, 0); err == nil {
		t.Fatal("expected an error for a word larger than uint256")
	}
}
