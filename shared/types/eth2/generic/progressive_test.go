package generic

import (
	"bytes"
	"testing"

	"github.com/pk910/dynamic-ssz/treeproof"
)

func TestConcatGindices(t *testing.T) {
	if got := ConcatGindices(); got != 1 {
		t.Fatalf("empty: got %d want 1", got)
	}
	if got := ConcatGindices(1); got != 1 {
		t.Fatalf("root only: got %d want 1", got)
	}
	// Left of root, then left again: 1 -> 2 -> 4
	if got := ConcatGindices(2, 2); got != 4 {
		t.Fatalf("concat(2,2): got %d want 4", got)
	}
	// Progressive field 1: progressive tree at 2, chunk gindex 24
	if got := ConcatGindices(2, 24); got != 40 {
		t.Fatalf("concat(2,24): got %d want 40", got)
	}
}

func TestProgressiveChunkGindex(t *testing.T) {
	cases := []struct {
		index uint64
		want  uint64
	}{
		{0, 2},
		{1, 24},
		{2, 25},
		{3, 26},
		{4, 27},
		{5, 224},
		{11, 230},
		{27, 1926},
		{44, 1943},
	}
	for _, tc := range cases {
		if got := ProgressiveChunkGindex(tc.index); got != tc.want {
			t.Errorf("ProgressiveChunkGindex(%d): got %d want %d", tc.index, got, tc.want)
		}
	}
}

func TestProgressiveContainerFieldGindex(t *testing.T) {
	// Known Gloas BeaconState field gindices (ssz-index → absolute from root).
	cases := []struct {
		field uint64
		want  uint64
	}{
		{0, 4},     // genesis_time
		{1, 40},    // genesis_validators_root
		{2, 41},    // slot
		{5, 352},   // block_roots
		{11, 358},  // validators
		{27, 2950}, // historical_summaries
		{44, 2967}, // payload_expected_withdrawals
	}
	for _, tc := range cases {
		if got := ProgressiveContainerFieldGindex(tc.field); got != tc.want {
			t.Errorf("ProgressiveContainerFieldGindex(%d): got %d want %d", tc.field, got, tc.want)
		}
	}
}

func TestProgressiveListElementGindex(t *testing.T) {
	// Element 0: left of progressive tree at list root → concat(2, 2) = 4
	if got := ProgressiveListElementGindex(0); got != 4 {
		t.Fatalf("element 0: got %d want 4", got)
	}
	// Element 1: concat(2, 24) = 40
	if got := ProgressiveListElementGindex(1); got != 40 {
		t.Fatalf("element 1: got %d want 40", got)
	}
}

// progressiveItem is a fixed composite so each ProgressiveList element is one
// Merkle leaf (basic types pack multiple values per chunk).
type progressiveItem struct {
	X uint64
	Y uint64
}

// progressiveProbe is a ProgressiveContainer with a ProgressiveList of composites
// so we can verify gindices against real treeproof nodes.
type progressiveProbe struct {
	A uint64             `ssz-index:"0"`
	B uint64             `ssz-index:"1"`
	C uint64             `ssz-index:"2"`
	D []*progressiveItem `ssz-type:"progressive-list" ssz-max:"1024" ssz-index:"5"`
}

func TestProgressiveGindicesMatchTree(t *testing.T) {
	obj := &progressiveProbe{
		A: 0x11,
		B: 0x22,
		C: 0x33,
		D: []*progressiveItem{
			{X: 100, Y: 1},
			{X: 200, Y: 2},
			{X: 300, Y: 3},
		},
	}

	tree, err := SSZ.GetTree(obj)
	if err != nil {
		t.Fatalf("GetTree: %v", err)
	}

	// Field A (ssz-index 0)
	gidA := ProgressiveContainerFieldGindex(0)
	proofA, err := tree.Prove(int(gidA))
	if err != nil {
		t.Fatalf("Prove field A (gid=%d): %v", gidA, err)
	}
	if ok, err := treeproof.VerifyProof(tree.Hash(), proofA); err != nil || !ok {
		t.Fatalf("verify field A: ok=%v err=%v leaf=%x", ok, err, proofA.Leaf)
	}

	// Field B (ssz-index 1)
	gidB := ProgressiveContainerFieldGindex(1)
	proofB, err := tree.Prove(int(gidB))
	if err != nil {
		t.Fatalf("Prove field B (gid=%d): %v", gidB, err)
	}
	if ok, err := treeproof.VerifyProof(tree.Hash(), proofB); err != nil || !ok {
		t.Fatalf("verify field B: ok=%v err=%v", ok, err)
	}

	// Field D element 1 (ssz-index 5 progressive list of composites)
	gidD := ProgressiveContainerFieldGindex(5)
	gidD1 := GetGeneralizedIndexForProgressiveListElement(gidD, 1)
	proofD1, err := tree.Prove(int(gidD1))
	if err != nil {
		t.Fatalf("Prove D[1] (gid=%d): %v", gidD1, err)
	}
	itemRoot, err := SSZ.HashTreeRoot(obj.D[1])
	if err != nil {
		t.Fatalf("HTR D[1]: %v", err)
	}
	if !bytes.Equal(proofD1.Leaf, itemRoot[:]) {
		t.Fatalf("D[1] leaf mismatch: got %x want %x", proofD1.Leaf, itemRoot)
	}
	if ok, err := treeproof.VerifyProof(tree.Hash(), proofD1); err != nil || !ok {
		t.Fatalf("verify D[1]: ok=%v err=%v", ok, err)
	}

	// Sanity: progressive field gindex must differ from old power-of-two container
	// layout (ceil(6 fields)=8 + 5 = 13).
	oldStyle := uint64(8 + 5)
	if gidD == oldStyle {
		t.Fatalf("progressive field gindex unexpectedly equals old-style %d", oldStyle)
	}
}
