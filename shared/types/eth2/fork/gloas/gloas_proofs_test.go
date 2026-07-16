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
