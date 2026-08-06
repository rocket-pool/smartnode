package gloas

import (
	"bytes"
	"testing"

	"github.com/pk910/dynamic-ssz/treeproof"

	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
)

func TestGloasFieldGindices(t *testing.T) {
	// Stable progressive field gindices for Gloas BeaconState (ssz-index → absolute).
	cases := []struct {
		name string
		got  uint64
		want uint64
	}{
		{"slot", GetGeneralizedIndexForSlot(), 41},
		{"block_roots", GetGeneralizedIndexForBlockRoots(), 352},
		{"validators", GetGeneralizedIndexForValidators(), 358},
		{"historical_summaries", GetGeneralizedIndexForHistoricalSummaries(), 2950},
		{"validators[0]", GetGeneralizedIndexForValidator(0), 1432},
		{"validators[1]", GetGeneralizedIndexForValidator(1), 11464},
		{"previous_epoch_participation", GetGeneralizedIndexForPreviousEpochParticipation(), 362},
		{"participation chunk 0", GetGeneralizedIndexForParticipationChunk(0), 1448},
		{"participation chunk 1", GetGeneralizedIndexForParticipationChunk(1), 11592},
		{"participation chunk 4", GetGeneralizedIndexForParticipationChunk(4), 11595},
		{"participation chunk 5", GetGeneralizedIndexForParticipationChunk(5), 92768},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s: got %d want %d", tc.name, tc.got, tc.want)
		}
	}
}

func TestValidatorAndSlotProofProgressive(t *testing.T) {
	state := minimalBeaconState()
	state.Slot = 100
	state.Validators = []*generic.Validator{
		{
			Pubkey:                bytes.Repeat([]byte{0x01}, 48),
			WithdrawalCredentials: bytes.Repeat([]byte{0x02}, 32),
			EffectiveBalance:      32_000_000_000,
		},
		{
			Pubkey:                bytes.Repeat([]byte{0x03}, 48),
			WithdrawalCredentials: bytes.Repeat([]byte{0x04}, 32),
			EffectiveBalance:      31_000_000_000,
		},
	}
	state.Balances = []uint64{32_000_000_000, 31_000_000_000}

	// Prove validators[1] and slot against the state tree (without block-header extension).
	stateTree, err := generic.SSZ.GetTree(state)
	if err != nil {
		t.Fatalf("GetTree: %v", err)
	}
	stateRoot := stateTree.Hash()

	validatorGid := GetGeneralizedIndexForValidator(1)
	validatorProof, err := stateTree.Prove(int(validatorGid))
	if err != nil {
		t.Fatalf("Prove validator: %v", err)
	}
	validatorLeaf, err := generic.SSZ.HashTreeRoot(state.Validators[1])
	if err != nil {
		t.Fatalf("validator HTR: %v", err)
	}
	if !bytes.Equal(validatorProof.Leaf, validatorLeaf[:]) {
		t.Fatalf("validator leaf mismatch")
	}
	if ok, err := treeproof.VerifyProof(stateRoot, validatorProof); err != nil || !ok {
		t.Fatalf("verify validator proof: ok=%v err=%v", ok, err)
	}

	slotProof, err := stateTree.Prove(int(GetGeneralizedIndexForSlot()))
	if err != nil {
		t.Fatalf("Prove slot: %v", err)
	}
	if ok, err := treeproof.VerifyProof(stateRoot, slotProof); err != nil || !ok {
		t.Fatalf("verify slot proof: ok=%v err=%v", ok, err)
	}

	// Full ValidatorAndSlotProof (state + block header extension) should succeed.
	vProof, sProof, err := state.ValidatorAndSlotProof(1)
	if err != nil {
		t.Fatalf("ValidatorAndSlotProof: %v", err)
	}
	if len(vProof) == 0 || len(sProof) == 0 {
		t.Fatalf("expected non-empty proofs, got validator=%d slot=%d", len(vProof), len(sProof))
	}
	// Block-header extension adds the same suffix to both proofs.
	if len(vProof) <= len(validatorProof.Hashes) {
		t.Fatalf("expected block-header extension on validator proof")
	}
}

func TestBlockRootProofProgressive(t *testing.T) {
	state := minimalBeaconState()
	// Slot far enough that block_roots[slot % 8192] is still "recent".
	state.Slot = 100
	// Plant a distinctive root at index 10.
	var planted [32]byte
	for i := range planted {
		planted[i] = byte(i + 1)
	}
	state.BlockRoots[10] = planted

	tree, err := generic.SSZ.GetTree(state)
	if err != nil {
		t.Fatalf("GetTree: %v", err)
	}

	// Prove block_roots[10] via the helper (uses progressive field gindex + vector).
	proofHashes, err := state.BlockRootProof(10)
	if err != nil {
		t.Fatalf("BlockRootProof: %v", err)
	}

	gid := generic.GetGeneralizedIndexForVectorElement(
		GetGeneralizedIndexForBlockRoots(),
		generic.BeaconStateBlockRootsMaxLength,
		10,
	)
	direct, err := tree.Prove(int(gid))
	if err != nil {
		t.Fatalf("direct Prove: %v", err)
	}
	if !bytes.Equal(direct.Leaf, planted[:]) {
		t.Fatalf("leaf mismatch: got %x want %x", direct.Leaf, planted)
	}
	if ok, err := treeproof.VerifyProof(tree.Hash(), direct); err != nil || !ok {
		t.Fatalf("verify block root: ok=%v err=%v", ok, err)
	}
	if len(proofHashes) != len(direct.Hashes) {
		t.Fatalf("helper proof length %d != direct %d", len(proofHashes), len(direct.Hashes))
	}
	for i := range proofHashes {
		if !bytes.Equal(proofHashes[i], direct.Hashes[i]) {
			t.Fatalf("proof hash[%d] mismatch", i)
		}
	}
}

func TestHistoricalSummaryProofProgressive(t *testing.T) {
	state := minimalBeaconState()
	// Historical: slot + 8192 <= state.Slot
	state.Slot = generic.SlotsPerHistoricalRoot + 100
	summary := &generic.HistoricalSummary{
		BlockSummaryRoot: [32]byte{0xaa},
		StateSummaryRoot: [32]byte{0xbb},
	}
	state.HistoricalSummaries = []*generic.HistoricalSummary{summary}

	tree, err := generic.SSZ.GetTree(state)
	if err != nil {
		t.Fatalf("GetTree: %v", err)
	}

	// Prove historical_summaries[0] for slot 0 with capellaOffset 0.
	arrayIndex := uint64(0)
	gid := generic.GetGeneralizedIndexForListElement(
		GetGeneralizedIndexForHistoricalSummaries(),
		generic.BeaconStateHistoricalSummariesMaxLength,
		arrayIndex,
	)
	proof, err := tree.Prove(int(gid))
	if err != nil {
		t.Fatalf("Prove historical summary: %v", err)
	}
	summaryRoot, err := generic.SSZ.HashTreeRoot(summary)
	if err != nil {
		t.Fatalf("summary HTR: %v", err)
	}
	if !bytes.Equal(proof.Leaf, summaryRoot[:]) {
		t.Fatalf("summary leaf mismatch")
	}
	if ok, err := treeproof.VerifyProof(tree.Hash(), proof); err != nil || !ok {
		t.Fatalf("verify historical summary: ok=%v err=%v", ok, err)
	}

	// Helper should also succeed (includes block-header extension).
	hashes, err := state.HistoricalSummaryProof(0, 0)
	if err != nil {
		t.Fatalf("HistoricalSummaryProof: %v", err)
	}
	if len(hashes) <= len(proof.Hashes) {
		t.Fatalf("expected block-header extension on historical summary proof")
	}
}

func TestGloasWithdrawalProofGindices(t *testing.T) {
	// Stable progressive field gindices for the Gloas withdrawal-proof path
	// (ssz-index → absolute).
	cases := []struct {
		name string
		got  uint64
		want uint64
	}{
		{"state_roots", GetGeneralizedIndexForStateRoots(), 353},
		{"payload_expected_withdrawals", GetGeneralizedIndexForPayloadExpectedWithdrawals(), 2967},
		{"payload_expected_withdrawals[0]", GetGeneralizedIndexForExpectedWithdrawal(0), 11868},
		{"payload_expected_withdrawals[1]", GetGeneralizedIndexForExpectedWithdrawal(1), 94952},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s: got %d want %d", tc.name, tc.got, tc.want)
		}
	}
}

func TestProveExpectedWithdrawal(t *testing.T) {
	state := minimalBeaconState()
	state.Slot = 100
	state.PayloadExpectedWithdrawals = []*generic.Withdrawal{
		{Index: 10, ValidatorIndex: 5, Address: [20]byte{0xaa}, Amount: 32_000_000_000},
		{Index: 11, ValidatorIndex: 7, Address: [20]byte{0xbb}, Amount: 1_000_000_000},
	}

	stateTree, err := generic.SSZ.GetTree(state)
	if err != nil {
		t.Fatalf("GetTree: %v", err)
	}
	stateRoot := stateTree.Hash()

	// Prove payload_expected_withdrawals[1] via the helper.
	proofHashes, err := state.ProveExpectedWithdrawal(1)
	if err != nil {
		t.Fatalf("ProveExpectedWithdrawal: %v", err)
	}

	// Compare against a direct tree proof at the expected gindex.
	gid := GetGeneralizedIndexForExpectedWithdrawal(1)
	direct, err := stateTree.Prove(int(gid))
	if err != nil {
		t.Fatalf("direct Prove: %v", err)
	}
	leaf, err := generic.SSZ.HashTreeRoot(state.PayloadExpectedWithdrawals[1])
	if err != nil {
		t.Fatalf("withdrawal HTR: %v", err)
	}
	if !bytes.Equal(direct.Leaf, leaf[:]) {
		t.Fatalf("leaf mismatch")
	}
	if ok, err := treeproof.VerifyProof(stateRoot, direct); err != nil || !ok {
		t.Fatalf("verify expected withdrawal proof: ok=%v err=%v", ok, err)
	}
	if len(proofHashes) != len(direct.Hashes) {
		t.Fatalf("helper proof length %d != direct %d", len(proofHashes), len(direct.Hashes))
	}
	for i := range proofHashes {
		if !bytes.Equal(proofHashes[i], direct.Hashes[i]) {
			t.Fatalf("proof hash[%d] mismatch", i)
		}
	}

	// Out-of-bounds indices must error.
	if _, err := state.ProveExpectedWithdrawal(2); err == nil {
		t.Fatalf("expected out-of-bounds error")
	}
}

func TestStateRootProofProgressive(t *testing.T) {
	state := minimalBeaconState()
	// Slot far enough that state_roots[slot % 8192] is still "recent".
	state.Slot = 100
	// Plant a distinctive root at index 10.
	var planted [32]byte
	for i := range planted {
		planted[i] = byte(i + 42)
	}
	state.StateRoots[10] = planted

	tree, err := generic.SSZ.GetTree(state)
	if err != nil {
		t.Fatalf("GetTree: %v", err)
	}

	// Prove state_roots[10] via the helper (progressive field gindex + vector).
	proofHashes, err := state.StateRootProof(10)
	if err != nil {
		t.Fatalf("StateRootProof: %v", err)
	}

	gid := generic.GetGeneralizedIndexForVectorElement(
		GetGeneralizedIndexForStateRoots(),
		generic.BeaconStateStateRootsMaxLength,
		10,
	)
	direct, err := tree.Prove(int(gid))
	if err != nil {
		t.Fatalf("direct Prove: %v", err)
	}
	if !bytes.Equal(direct.Leaf, planted[:]) {
		t.Fatalf("leaf mismatch: got %x want %x", direct.Leaf, planted)
	}
	if ok, err := treeproof.VerifyProof(tree.Hash(), direct); err != nil || !ok {
		t.Fatalf("verify state root: ok=%v err=%v", ok, err)
	}
	if len(proofHashes) != len(direct.Hashes) {
		t.Fatalf("helper proof length %d != direct %d", len(proofHashes), len(direct.Hashes))
	}
	for i := range proofHashes {
		if !bytes.Equal(proofHashes[i], direct.Hashes[i]) {
			t.Fatalf("proof hash[%d] mismatch", i)
		}
	}

	// A slot that rolled out of the ring must error (use historical_summaries).
	state.Slot = generic.SlotsPerHistoricalRoot + 100
	if _, err := state.StateRootProof(0); err == nil {
		t.Fatalf("expected historical-slot error")
	}
}

func TestHistoricalSummaryStateRootProof(t *testing.T) {
	state := minimalBeaconState()
	// Era-aligned state: its state_roots ring covers the full previous era.
	state.Slot = generic.SlotsPerHistoricalRoot
	var planted [32]byte
	for i := range planted {
		planted[i] = byte(i + 77)
	}
	state.StateRoots[7] = planted

	proofHashes, err := state.HistoricalSummaryStateRootProof(7)
	if err != nil {
		t.Fatalf("HistoricalSummaryStateRootProof: %v", err)
	}

	// Direct proof against the HistoricalSummaryLists tree (gid 3 → state_roots).
	hsls := generic.HistoricalSummaryLists{
		BlockRoots: state.BlockRoots,
		StateRoots: state.StateRoots,
	}
	tree, err := generic.SSZ.GetTree(&hsls)
	if err != nil {
		t.Fatalf("GetTree: %v", err)
	}
	gid := uint64(1)
	gid = gid*2 + 1
	gid = gid * generic.SlotsPerHistoricalRoot
	gid = gid + 7
	direct, err := tree.Prove(int(gid))
	if err != nil {
		t.Fatalf("direct Prove: %v", err)
	}
	if !bytes.Equal(direct.Leaf, planted[:]) {
		t.Fatalf("leaf mismatch")
	}
	if ok, err := treeproof.VerifyProof(tree.Hash(), direct); err != nil || !ok {
		t.Fatalf("verify historical summary state root: ok=%v err=%v", ok, err)
	}
	if len(proofHashes) != len(direct.Hashes) {
		t.Fatalf("helper proof length %d != direct %d", len(proofHashes), len(direct.Hashes))
	}
	for i := range proofHashes {
		if !bytes.Equal(proofHashes[i], direct.Hashes[i]) {
			t.Fatalf("proof hash[%d] mismatch", i)
		}
	}

	// Non-aligned states must error.
	state.Slot = generic.SlotsPerHistoricalRoot + 1
	if _, err := state.HistoricalSummaryStateRootProof(7); err == nil {
		t.Fatalf("expected non-aligned state error")
	}
}

func TestPreviousEpochParticipationChunkProofProgressive(t *testing.T) {
	state := minimalBeaconState()

	// 180 flag bytes span packed chunks 0-5: chunks 0-4 fill progressive
	// levels 0 and 1, chunk 5 is the first (partial) chunk of level 2.
	const numValidators = 180
	state.PreviousEpochParticipation = make([]byte, numValidators)
	for i := range state.PreviousEpochParticipation {
		state.PreviousEpochParticipation[i] = byte(i % 8)
	}

	tree, err := generic.SSZ.GetTree(state)
	if err != nil {
		t.Fatalf("GetTree: %v", err)
	}

	cases := []struct {
		validatorIndex uint64
		wantDepth      int
	}{
		{0, 10},   // chunk 0 (progressive level 0)
		{31, 10},  // chunk 0, last byte
		{32, 13},  // chunk 1, first chunk of level 1
		{70, 13},  // chunk 2, mid level 1
		{127, 13}, // chunk 3, last byte
		{128, 13}, // chunk 4, last chunk of level 1
		{160, 16}, // chunk 5, first chunk of level 2
		{179, 16}, // chunk 5, tail of the partial chunk
	}
	for _, tc := range cases {
		chunk, hashes, err := state.PreviousEpochParticipationChunkProof(tc.validatorIndex)
		if err != nil {
			t.Fatalf("chunk proof for validator %d: %v", tc.validatorIndex, err)
		}
		if got, want := chunk[tc.validatorIndex%32], state.PreviousEpochParticipation[tc.validatorIndex]; got != want {
			t.Fatalf("validator %d: chunk byte %d != flags %d", tc.validatorIndex, got, want)
		}
		if len(hashes) != tc.wantDepth {
			t.Fatalf("validator %d: proof depth %d, want %d", tc.validatorIndex, len(hashes), tc.wantDepth)
		}

		gid := GetGeneralizedIndexForParticipationChunk(tc.validatorIndex / 32)
		direct, err := tree.Prove(int(gid))
		if err != nil {
			t.Fatalf("direct Prove gid %d: %v", gid, err)
		}
		if !bytes.Equal(direct.Leaf, chunk[:]) {
			t.Fatalf("validator %d: direct leaf %x != chunk %x", tc.validatorIndex, direct.Leaf, chunk)
		}
		if ok, err := treeproof.VerifyProof(tree.Hash(), direct); err != nil || !ok {
			t.Fatalf("validator %d: verify participation chunk: ok=%v err=%v", tc.validatorIndex, ok, err)
		}
		if len(hashes) != len(direct.Hashes) {
			t.Fatalf("validator %d: helper proof length %d != direct %d", tc.validatorIndex, len(hashes), len(direct.Hashes))
		}
		for i := range hashes {
			if !bytes.Equal(hashes[i], direct.Hashes[i]) {
				t.Fatalf("validator %d: proof hash[%d] mismatch", tc.validatorIndex, i)
			}
		}
	}

	// Out-of-bounds index must error.
	if _, _, err := state.PreviousEpochParticipationChunkProof(numValidators); err == nil {
		t.Fatalf("expected out-of-bounds error")
	}

	// An empty participation list must reject every index.
	empty := minimalBeaconState()
	if _, _, err := empty.PreviousEpochParticipationChunkProof(0); err == nil {
		t.Fatalf("expected error for empty participation list")
	}
}

func TestPreviousEpochParticipationChunkProofLargeIndex(t *testing.T) {
	state := minimalBeaconState()

	// A registry of over a million validators: index 1,048,600 sits in packed
	// chunk 32768, which lives in progressive level 8 of the list, so the
	// proof must be 10 + 3*8 = 34 hashes deep.
	const validatorIndex = uint64(1_048_600)
	const numValidators = validatorIndex + 50
	state.PreviousEpochParticipation = make([]byte, numValidators)
	for i := range state.PreviousEpochParticipation {
		state.PreviousEpochParticipation[i] = byte(i % 8)
	}

	tree, err := generic.SSZ.GetTree(state)
	if err != nil {
		t.Fatalf("GetTree: %v", err)
	}

	chunk, hashes, err := state.PreviousEpochParticipationChunkProof(validatorIndex)
	if err != nil {
		t.Fatalf("chunk proof for validator %d: %v", validatorIndex, err)
	}
	if got, want := chunk[validatorIndex%32], state.PreviousEpochParticipation[validatorIndex]; got != want {
		t.Fatalf("chunk byte %d != flags %d", got, want)
	}
	if len(hashes) != 34 {
		t.Fatalf("proof depth %d, want 34", len(hashes))
	}

	gid := GetGeneralizedIndexForParticipationChunk(validatorIndex / 32)
	direct, err := tree.Prove(int(gid))
	if err != nil {
		t.Fatalf("direct Prove gid %d: %v", gid, err)
	}
	if !bytes.Equal(direct.Leaf, chunk[:]) {
		t.Fatalf("direct leaf %x != chunk %x", direct.Leaf, chunk)
	}
	if ok, err := treeproof.VerifyProof(tree.Hash(), direct); err != nil || !ok {
		t.Fatalf("verify participation chunk: ok=%v err=%v", ok, err)
	}
	if len(hashes) != len(direct.Hashes) {
		t.Fatalf("helper proof length %d != direct %d", len(hashes), len(direct.Hashes))
	}
	for i := range hashes {
		if !bytes.Equal(hashes[i], direct.Hashes[i]) {
			t.Fatalf("proof hash[%d] mismatch", i)
		}
	}
}
