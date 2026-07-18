package services

import (
	"bytes"
	"crypto/sha256"
	"math/big"
	"math/bits"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/types/eth2/fork/fulu"
	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
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

// ---------------------------------------------------------------------------
// Participation witness chain assembly
// ---------------------------------------------------------------------------

// newProofTestFuluState builds a minimal but SSZ-valid fulu beacon state with
// numValidators validators and per-validator previous-epoch participation
// flags of (i % 8).
func newProofTestFuluState(t *testing.T, numValidators int, slot uint64) *fulu.BeaconState {
	t.Helper()

	validators := make([]*generic.Validator, numValidators)
	balances := make([]uint64, numValidators)
	inactivityScores := make([]uint64, numValidators)
	previousParticipation := make([]byte, numValidators)
	currentParticipation := make([]byte, numValidators)
	for i := range validators {
		validators[i] = &generic.Validator{
			Pubkey:                make([]byte, 48),
			WithdrawalCredentials: make([]byte, 32),
			EffectiveBalance:      32e9,
		}
		validators[i].Pubkey[0] = byte(i + 1)
		balances[i] = 32e9
		previousParticipation[i] = byte(i % 8)
		currentParticipation[i] = byte((i + 1) % 8)
	}

	randaoMixes := make([][]byte, 65536)
	for i := range randaoMixes {
		randaoMixes[i] = make([]byte, 32)
	}

	syncCommittee := func() *generic.SyncCommittee {
		pubkeys := make([][]byte, 512)
		for i := range pubkeys {
			pubkeys[i] = make([]byte, 48)
		}
		return &generic.SyncCommittee{PubKeys: pubkeys}
	}

	parentRoot := make([]byte, 32)
	parentRoot[0] = 0xaa
	bodyRoot := make([]byte, 32)
	bodyRoot[0] = 0xbb

	return &fulu.BeaconState{
		GenesisValidatorsRoot: make([]byte, 32),
		Slot:                  slot,
		Fork: &generic.Fork{
			PreviousVersion: make([]byte, 4),
			CurrentVersion:  make([]byte, 4),
		},
		LatestBlockHeader: &generic.BeaconBlockHeader{
			Slot:          slot,
			ProposerIndex: 1,
			ParentRoot:    parentRoot,
			StateRoot:     make([]byte, 32),
			BodyRoot:      bodyRoot,
		},
		HistoricalRoots: [][]byte{},
		Eth1Data: &generic.Eth1Data{
			DepositRoot: make([]byte, 32),
			BlockHash:   make([]byte, 32),
		},
		Eth1DataVotes:                []*generic.Eth1Data{},
		Validators:                   validators,
		Balances:                     balances,
		RandaoMixes:                  randaoMixes,
		Slashings:                    make([]uint64, 8192),
		PreviousEpochParticipation:   previousParticipation,
		CurrentEpochParticipation:    currentParticipation,
		PreviousJustifiedCheckpoint:  &generic.Checkpoint{Root: make([]byte, 32)},
		CurrentJustifiedCheckpoint:   &generic.Checkpoint{Root: make([]byte, 32)},
		FinalizedCheckpoint:          &generic.Checkpoint{Root: make([]byte, 32)},
		InactivityScores:             inactivityScores,
		CurrentSyncCommittee:         syncCommittee(),
		NextSyncCommittee:            syncCommittee(),
		LatestExecutionPayloadHeader: &generic.ExecutionPayloadHeader{},
		HistoricalSummaries:          []*generic.HistoricalSummary{},
		ProposerLookahead:            make([]uint64, 64),
	}
}

// concatGid appends a child generalized index (rooted at 1 in its own
// subtree) onto a parent generalized index, mirroring the on-chain SSZ.concat
// path construction. The combined index exceeds 64 bits for full
// participation witness chains, hence big.Int.
func concatGid(parent *big.Int, child uint64) *big.Int {
	depth := bits.Len64(child) - 1
	out := new(big.Int).Lsh(parent, uint(depth))
	return out.Or(out, new(big.Int).SetUint64(child-1<<uint(depth)))
}

// walkWitnessChain mirrors the on-chain SSZ.restoreMerkleRoot: walk the
// generalized index from the leaf to the root, consuming the flat witness
// array leaf-first. Enforces the SSZ.length witness count check.
func walkWitnessChain(t *testing.T, leaf []byte, gid *big.Int, witnesses [][]byte) []byte {
	t.Helper()
	if len(witnesses) != gid.BitLen()-1 {
		t.Fatalf("witness count %d does not match gindex depth %d", len(witnesses), gid.BitLen()-1)
	}
	value := leaf
	g := new(big.Int).Set(gid)
	for _, witness := range witnesses {
		var pair [64]byte
		if g.Bit(0) == 1 {
			copy(pair[:32], witness)
			copy(pair[32:], value)
		} else {
			copy(pair[:32], value)
			copy(pair[32:], witness)
		}
		hashed := sha256.Sum256(pair[:])
		value = hashed[:]
		g.Rsh(g, 1)
	}
	return value
}

// anchorBlockRoot computes the block root the contract retrieves via
// EIP-4788: the anchor state's latest block header with its state root
// filled in.
func anchorBlockRoot(t *testing.T, state *fulu.BeaconState) []byte {
	t.Helper()
	stateRoot, err := state.HashTreeRoot()
	if err != nil {
		t.Fatalf("failed to hash anchor state: %v", err)
	}
	header := *state.LatestBlockHeader
	header.StateRoot = stateRoot[:]
	blockRoot, err := header.HashTreeRoot()
	if err != nil {
		t.Fatalf("failed to hash anchor block header: %v", err)
	}
	return blockRoot[:]
}

func TestBuildRecentParticipationWitnesses(t *testing.T) {
	const numValidators = 100
	const validatorIndex = uint64(70)
	const participationSlot = uint64(105003*32 - 1)
	const anchorSlot = participationSlot + 100

	participationState := newProofTestFuluState(t, numValidators, participationSlot)
	anchorState := newProofTestFuluState(t, numValidators, anchorSlot)

	// Wire the anchor's state_roots vector to commit to the participation state
	participationRoot, err := participationState.HashTreeRoot()
	if err != nil {
		t.Fatalf("failed to hash participation state: %v", err)
	}
	anchorState.StateRoots[participationSlot%generic.SlotsPerHistoricalRoot] = participationRoot

	if err := verifyParticipationStateLink(anchorState, participationState, participationSlot); err != nil {
		t.Fatalf("state link check failed: %v", err)
	}

	chunk, chunkProof, err := participationState.PreviousEpochParticipationChunkProof(validatorIndex)
	if err != nil {
		t.Fatalf("failed to build the chunk proof: %v", err)
	}
	witnesses, err := buildRecentParticipationWitnesses(anchorState, participationSlot, chunkProof)
	if err != nil {
		t.Fatalf("failed to build the recent witnesses: %v", err)
	}
	if len(witnesses) != 64 {
		t.Fatalf("recent witness count = %d, want 64", len(witnesses))
	}

	// Combined gindex, mirroring BeaconStateVerifier.verifyParticipation:
	// header -> state_root ++ state -> state_roots[n] ++ state -> chunk
	stateRootsGid := (uint64(1)*64+generic.BeaconStateStateRootsFieldIndex)*generic.BeaconStateBlockRootsMaxLength + participationSlot%generic.SlotsPerHistoricalRoot
	chunkGid := generic.GetGeneralizedIndexForParticipationChunk(validatorIndex/32, fulu.GetGeneralizedIndexForPreviousEpochParticipation())
	gid := big.NewInt(int64(generic.BeaconBlockHeaderStateRootGeneralizedIndex))
	gid = concatGid(gid, stateRootsGid)
	gid = concatGid(gid, chunkGid)

	root := walkWitnessChain(t, chunk[:], gid, witnesses)
	if !bytes.Equal(root, anchorBlockRoot(t, anchorState)) {
		t.Fatalf("restored root %x does not match the anchor block root", root)
	}
}

func TestBuildHistoricalParticipationWitnesses(t *testing.T) {
	const numValidators = 100
	const validatorIndex = uint64(33)
	const capellaOffset = uint64(0)
	// Participation slot in era 1, anchored more than 8192 slots later
	const participationSlot = uint64(generic.SlotsPerHistoricalRoot + 5000)
	const eraBoundarySlot = (participationSlot/generic.SlotsPerHistoricalRoot + 1) * generic.SlotsPerHistoricalRoot
	const anchorSlot = uint64(5*generic.SlotsPerHistoricalRoot + 77)
	const entry = participationSlot/generic.SlotsPerHistoricalRoot - capellaOffset

	participationState := newProofTestFuluState(t, numValidators, participationSlot)
	eraState := newProofTestFuluState(t, numValidators, eraBoundarySlot)
	anchorState := newProofTestFuluState(t, numValidators, anchorSlot)

	// Wire the era boundary state's state_roots vector to commit to the
	// participation state
	participationRoot, err := participationState.HashTreeRoot()
	if err != nil {
		t.Fatalf("failed to hash participation state: %v", err)
	}
	eraState.StateRoots[participationSlot%generic.SlotsPerHistoricalRoot] = participationRoot

	if err := verifyParticipationStateLink(eraState, participationState, participationSlot); err != nil {
		t.Fatalf("state link check failed: %v", err)
	}

	// Wire the anchor's historical_summaries to commit to the era boundary
	// state's roots vectors
	hsls := generic.HistoricalSummaryLists{
		BlockRoots: eraState.BlockRoots,
		StateRoots: eraState.StateRoots,
	}
	hslsTree, err := hsls.GetTree()
	if err != nil {
		t.Fatalf("failed to get historical summary lists tree: %v", err)
	}
	blockSummaryNode, err := hslsTree.Get(2)
	if err != nil {
		t.Fatalf("failed to get block summary node: %v", err)
	}
	stateSummaryNode, err := hslsTree.Get(3)
	if err != nil {
		t.Fatalf("failed to get state summary node: %v", err)
	}
	summaries := make([]*generic.HistoricalSummary, entry+1)
	for i := range summaries {
		summaries[i] = &generic.HistoricalSummary{}
		summaries[i].BlockSummaryRoot[0] = byte(i + 1)
	}
	copy(summaries[entry].BlockSummaryRoot[:], blockSummaryNode.Hash())
	copy(summaries[entry].StateSummaryRoot[:], stateSummaryNode.Hash())
	anchorState.HistoricalSummaries = summaries

	chunk, chunkProof, err := participationState.PreviousEpochParticipationChunkProof(validatorIndex)
	if err != nil {
		t.Fatalf("failed to build the chunk proof: %v", err)
	}
	witnesses, err := buildHistoricalParticipationWitnesses(anchorState, eraState, participationSlot, capellaOffset, chunkProof)
	if err != nil {
		t.Fatalf("failed to build the historical witnesses: %v", err)
	}
	if len(witnesses) != 90 {
		t.Fatalf("historical witness count = %d, want 90", len(witnesses))
	}

	// Combined gindex, mirroring BeaconStateVerifier.verifyParticipation:
	// header -> state_root ++ state -> historical_summaries[n] ++
	// HistoricalSummary -> state_summary_root ++ state_roots -> [n] ++
	// state -> chunk
	summaryElementGid := (uint64(1)*64+generic.BeaconStateHistoricalSummariesFieldIndex)*2*generic.BeaconStateHistoricalSummariesMaxLength + entry
	stateRootsVectorGid := generic.SlotsPerHistoricalRoot + participationSlot%generic.SlotsPerHistoricalRoot
	chunkGid := generic.GetGeneralizedIndexForParticipationChunk(validatorIndex/32, fulu.GetGeneralizedIndexForPreviousEpochParticipation())
	gid := big.NewInt(int64(generic.BeaconBlockHeaderStateRootGeneralizedIndex))
	gid = concatGid(gid, summaryElementGid)
	gid = concatGid(gid, 3) // HistoricalSummary -> state_summary_root
	gid = concatGid(gid, stateRootsVectorGid)
	gid = concatGid(gid, chunkGid)

	root := walkWitnessChain(t, chunk[:], gid, witnesses)
	if !bytes.Equal(root, anchorBlockRoot(t, anchorState)) {
		t.Fatalf("restored root %x does not match the anchor block root", root)
	}
}
