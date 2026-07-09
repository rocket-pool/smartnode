package eth2

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/rocket-pool/smartnode/shared/types/eth2/fork/fulu"
	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
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

// validateFuluStateProof walks a merged state+block-header proof from leaf to
// root and checks it lands on the state's block root. Returns the state root
// and block root.
func validateFuluStateProof(t *testing.T, leaf []byte, proof [][]byte, gid uint64, state *fulu.BeaconState) ([]byte, []byte) {
	t.Helper()

	// State proofs are merged with the block-header proof, so the effective
	// tree is rooted at the beacon block header.
	gid = offsetGidRoot(gid, generic.BeaconBlockHeaderStateRootGeneralizedIndex)
	currentHash := leaf

	for i, proofRow := range proof {
		// The last neighbor must have a gid of either 2 or 3
		if i == len(proof)-1 {
			if gid != 2 && gid != 3 {
				t.Fatalf("last node/neighbor gid must be 2 or 3, got: %d", gid)
			}
		}
		neighborIsLeft := gid%2 == 1
		gid /= 2
		currentHash = hash(currentHash, proofRow, neighborIsLeft)
	}

	// Compute the expected block root: the latest block header with the state
	// root filled in.
	stateRoot, err := state.HashTreeRoot()
	if err != nil {
		t.Fatalf("Failed to get state root: %v", err)
	}
	header := *state.LatestBlockHeader
	header.StateRoot = stateRoot[:]
	blockRoot, err := header.HashTreeRoot()
	if err != nil {
		t.Fatalf("Failed to get block root: %v", err)
	}

	if !bytes.Equal(currentHash, blockRoot[:]) {
		t.Fatalf("final hash %x does not match block root %x", currentHash, blockRoot)
	}
	return stateRoot[:], blockRoot[:]
}

func TestPreviousEpochParticipationAndSlotProof(t *testing.T) {
	const numValidators = 100
	const slot = uint64(105002*32 + 31) // last slot of some epoch
	state := newTestFuluState(t, numValidators, slot)

	// Validators in different chunks of the participation byte list (32 flag
	// bytes per chunk), including the last validator (partially filled chunk).
	for _, validatorIndex := range []uint64{0, 31, 32, 70, numValidators - 1} {
		chunk, chunkOffset, participationBranch, slotProof, err := state.PreviousEpochParticipationAndSlotProof(validatorIndex)
		if err != nil {
			t.Fatalf("PreviousEpochParticipationAndSlotProof(%d) failed: %v", validatorIndex, err)
		}

		if chunkOffset != validatorIndex%32 {
			t.Fatalf("chunkOffset = %d, want %d", chunkOffset, validatorIndex%32)
		}

		// The chunk must hold the validator's flags byte at its in-chunk offset.
		if chunk[chunkOffset] != state.PreviousEpochParticipation[validatorIndex] {
			t.Fatalf("chunk byte %d = %x, want %x", chunkOffset, chunk[chunkOffset], state.PreviousEpochParticipation[validatorIndex])
		}

		// The participation branch must connect the chunk to the block root.
		chunkGid := generic.GetGeneralizedIndexForParticipationChunk(validatorIndex/32, fulu.GetGeneralizedIndexForPreviousEpochParticipation())
		validateFuluStateProof(t, chunk[:], participationBranch, chunkGid, state)

		// The slot proof must connect the slot leaf to the same block root.
		slotLeaf := make([]byte, 32)
		binary.LittleEndian.PutUint64(slotLeaf, state.Slot)
		validateFuluStateProof(t, slotLeaf, slotProof, fulu.GetGeneralizedIndexForSlot(), state)
	}

	// Out-of-bounds validator index must error.
	if _, _, _, _, err := state.PreviousEpochParticipationAndSlotProof(numValidators); err == nil {
		t.Fatalf("expected an error for an out-of-bounds validator index")
	}
}
