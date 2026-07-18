package eth2

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/rocket-pool/smartnode/shared/types/eth2/fork/fulu"
	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
	hexutils "github.com/rocket-pool/smartnode/shared/utils/hex"
)

// newTestFuluState builds a minimal but SSZ-valid fulu beacon state with
// numValidators validators and per-validator previous-epoch participation
// flags of (i % 8).
func newTestFuluState(t *testing.T, numValidators int, slot uint64) *fulu.BeaconState {
	t.Helper()

	validators := make([]*generic.Validator, numValidators)
	balances := make([]uint64, numValidators)
	inactivityScores := make([]uint64, numValidators)
	previousParticipation := make([]byte, numValidators)
	currentParticipation := make([]byte, numValidators)
	for i := range validators {
		pubkey := make([]byte, 48)
		pubkey[0] = byte(i + 1)
		validators[i] = &generic.Validator{
			Pubkey:                make([]byte, 48),
			WithdrawalCredentials: make([]byte, 32),
			EffectiveBalance:      32e9,
		}
		copy(validators[i].Pubkey, pubkey)
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

// walkProof walks a merkle branch from leaf to root, consuming the witnesses
// leaf-first, and returns the reconstructed root. Fails if the witness count
// doesn't match the gid depth.
func walkProof(t *testing.T, leaf []byte, proof [][]byte, gid uint64) []byte {
	t.Helper()

	currentHash := leaf
	for _, proofRow := range proof {
		if gid == 1 {
			t.Fatalf("too many witnesses for gid depth")
		}
		neighborIsLeft := gid%2 == 1
		gid /= 2
		currentHash = hash(currentHash, proofRow, neighborIsLeft)
	}
	if gid != 1 {
		t.Fatalf("too few witnesses for gid depth, remaining gid: %d", gid)
	}
	return currentHash
}

// validateFuluStateProofToStateRoot walks a state-internal proof from leaf to
// root and checks it lands on the state's hash tree root. Returns the state
// root.
func validateFuluStateProofToStateRoot(t *testing.T, leaf []byte, proof [][]byte, gid uint64, state *fulu.BeaconState) []byte {
	t.Helper()

	currentHash := walkProof(t, leaf, proof, gid)

	stateRoot, err := state.HashTreeRoot()
	if err != nil {
		t.Fatalf("Failed to get state root: %v", err)
	}
	if !bytes.Equal(currentHash, stateRoot[:]) {
		t.Fatalf("final hash %x does not match state root %x", currentHash, stateRoot)
	}
	return stateRoot[:]
}

func TestPreviousEpochParticipationChunkProof(t *testing.T) {
	const numValidators = 100
	const slot = uint64(105002*32 + 31) // last slot of some epoch
	state := newTestFuluState(t, numValidators, slot)

	// Validators in different chunks of the participation byte list (32 flag
	// bytes per chunk), including the last validator (partially filled chunk).
	for _, validatorIndex := range []uint64{0, 31, 32, 70, numValidators - 1} {
		chunk, participationBranch, err := state.PreviousEpochParticipationChunkProof(validatorIndex)
		if err != nil {
			t.Fatalf("PreviousEpochParticipationChunkProof(%d) failed: %v", validatorIndex, err)
		}

		// The chunk must hold the validator's flags byte at its in-chunk offset.
		chunkOffset := validatorIndex % 32
		if chunk[chunkOffset] != state.PreviousEpochParticipation[validatorIndex] {
			t.Fatalf("chunk byte %d = %x, want %x", chunkOffset, chunk[chunkOffset], state.PreviousEpochParticipation[validatorIndex])
		}

		// The participation branch must connect the chunk to the state root.
		chunkGid := generic.GetGeneralizedIndexForParticipationChunk(validatorIndex/32, fulu.GetGeneralizedIndexForPreviousEpochParticipation())
		validateFuluStateProofToStateRoot(t, chunk[:], participationBranch, chunkGid, state)

		// The witness count must match the on-chain path length:
		// 6 (state fields) + 1 (list length mixin) + 35 (chunk index) = 42
		if len(participationBranch) != 42 {
			t.Fatalf("participation branch length = %d, want 42", len(participationBranch))
		}
	}

	// Out-of-bounds validator index must error.
	if _, _, err := state.PreviousEpochParticipationChunkProof(numValidators); err == nil {
		t.Fatalf("expected an error for an out-of-bounds validator index")
	}
}

func TestStateRootProof(t *testing.T) {
	const numValidators = 10
	const slot = uint64(105002*32 + 31)
	state := newTestFuluState(t, numValidators, slot)

	// Fill state_roots with distinct values
	for i := range state.StateRoots {
		binary.LittleEndian.PutUint64(state.StateRoots[i][:], uint64(i)+1)
	}

	for _, targetSlot := range []uint64{slot - 1, slot - 5000, slot - generic.SlotsPerHistoricalRoot} {
		proof, err := state.StateRootProof(targetSlot)
		if err != nil {
			t.Fatalf("StateRootProof(%d) failed: %v", targetSlot, err)
		}

		// The witness count must match the on-chain path length:
		// 6 (state fields) + 13 (vector index) = 19
		if len(proof) != 19 {
			t.Fatalf("state root proof length = %d, want 19", len(proof))
		}

		idx := targetSlot % generic.SlotsPerHistoricalRoot
		gid := (uint64(1)*64+generic.BeaconStateStateRootsFieldIndex)*generic.BeaconStateBlockRootsMaxLength + idx
		validateFuluStateProofToStateRoot(t, state.StateRoots[idx][:], proof, gid, state)
	}

	// A slot not in the past must error
	if _, err := state.StateRootProof(slot); err == nil {
		t.Fatalf("expected an error for a slot not in the past")
	}
	// A slot more than 8192 slots in the past must error
	if _, err := state.StateRootProof(slot - generic.SlotsPerHistoricalRoot - 1); err == nil {
		t.Fatalf("expected an error for a historical slot")
	}
}

func TestHistoricalSummaryStateRootProof(t *testing.T) {
	// Reuse the mainnet fixture for the 8192 slot era that ended at slot
	// 11567103; the era boundary state is at slot 11567104
	var roots testRoots
	err := json.Unmarshal(testRootsJSON, &roots)
	if err != nil {
		t.Fatalf("Failed to unmarshal test roots: %v", err)
	}

	const eraBoundarySlot = uint64(11567104)
	if eraBoundarySlot%generic.SlotsPerHistoricalRoot != 0 {
		t.Fatalf("era boundary slot %d is not aligned", eraBoundarySlot)
	}
	state := newTestFuluState(t, 10, eraBoundarySlot)
	for i, blockRoot := range roots.BlockRoots {
		blockRootBytes, err := hex.DecodeString(hexutils.RemovePrefix(blockRoot))
		if err != nil {
			t.Fatalf("Failed to decode block root: %v", err)
		}
		copy(state.BlockRoots[i][:], blockRootBytes)
	}
	for i, stateRoot := range roots.StateRoots {
		stateRootBytes, err := hex.DecodeString(hexutils.RemovePrefix(stateRoot))
		if err != nil {
			t.Fatalf("Failed to decode state root: %v", err)
		}
		copy(state.StateRoots[i][:], stateRootBytes)
	}

	// A misaligned state must error
	misaligned := newTestFuluState(t, 10, eraBoundarySlot+1)
	if _, err := misaligned.HistoricalSummaryStateRootProof(int(eraBoundarySlot - 100)); err == nil {
		t.Fatalf("expected an error for a misaligned state")
	}

	// Prove a state root within the era [11558912, 11567103]
	const targetSlot = uint64(11560000)
	proof, err := state.HistoricalSummaryStateRootProof(int(targetSlot))
	if err != nil {
		t.Fatalf("HistoricalSummaryStateRootProof(%d) failed: %v", targetSlot, err)
	}

	// The witness count must match the on-chain path length:
	// 13 (vector index) + 1 (state_summary_root vs block_summary_root) = 14
	if len(proof) != 14 {
		t.Fatalf("historical summary state root proof length = %d, want 14", len(proof))
	}

	// Walk from state_roots[idx] up to the HistoricalSummary container root
	idx := targetSlot % generic.SlotsPerHistoricalRoot
	gid := (uint64(1)*2+1)*generic.SlotsPerHistoricalRoot + idx
	summaryRoot := walkProof(t, state.StateRoots[idx][:], proof, gid)

	// The HistoricalSummary root equals sha256(block_summary_root ++
	// state_summary_root); the fixture era's summary roots are known mainnet
	// values (see TestBlockRootProof)
	expectedBlockSummaryRoot, err := hex.DecodeString("9d73b29c6e80e8300cedfa9e53aff89523affb98f3bd3f6752ecc159a2058858")
	if err != nil {
		t.Fatalf("Failed to decode expected block summary root: %v", err)
	}
	expectedStateSummaryRoot, err := hex.DecodeString("8a38d8b000dc65641eff8f7ff2e0b6f3b129957410cb9ecf183537484630b289")
	if err != nil {
		t.Fatalf("Failed to decode expected state summary root: %v", err)
	}
	expectedSummaryRoot := hash(expectedBlockSummaryRoot, expectedStateSummaryRoot, false)
	if !bytes.Equal(summaryRoot, expectedSummaryRoot) {
		t.Fatalf("summary root %x does not match expected %x", summaryRoot, expectedSummaryRoot)
	}
}
