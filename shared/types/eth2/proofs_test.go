package eth2

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"io"
	"math/bits"

	// Don't get tripped up by the sha256 import
	// See https://github.com/ethereum/consensus-specs/pull/779
	"crypto/sha256"
	_ "embed"
	"testing"

	ssz "github.com/ferranbt/fastssz"
)

// Test state - deneb fork. Hoodi genesis state.
//
//go:embed testdata/hoodi_genesis.ssz.gz
var testState []byte

//go:embed testdata/block_11544444.ssz
var testBlock []byte

func init() {
	decompressor, err := gzip.NewReader(bytes.NewReader(testState))
	if err != nil {
		panic(err)
	}
	out, err := io.ReadAll(decompressor)
	if err != nil {
		panic(err)
	}
	testState = out
}

func hash(a, b []byte, isReversed bool) []byte {
	tmp := [64]byte{}
	if isReversed {
		copy(tmp[:32], b)
		copy(tmp[32:], a)
	} else {
		copy(tmp[:32], a)
		copy(tmp[32:], b)
	}
	out := sha256.Sum256(tmp[:])
	return out[:]
}

func offsetGidRoot(gid uint64, newRoot uint64) uint64 {
	mulp := gid
	if bits.OnesCount64(gid) == 1 {
		// gid is a power of 2, so to offset it we just multiply by the new root
		return mulp * newRoot
	}
	add := uint64(0)
	// Get the highest power of 2 that is less than, or equal to, gid
	clz := bits.LeadingZeros64(gid)
	mulp = uint64(0x8000000000000000) >> clz
	add = gid - mulp
	// mulp is the highest power of 2 that is less than, or equal to, gid
	// add is the difference between gid and mulp
	// Multiply mulp by the new root and add the remainder
	return mulp*newRoot + add
}

func TestOffsetGidRoot(t *testing.T) {
	type c struct {
		oldGid      uint64
		newRoot     uint64
		expectedGid uint64
	}
	testCases := []c{
		{oldGid: 1, newRoot: 1, expectedGid: 1},
		{oldGid: 2, newRoot: 1, expectedGid: 2},
		{oldGid: 3, newRoot: 1, expectedGid: 3},
		{oldGid: 4, newRoot: 1, expectedGid: 4},
		{oldGid: 5, newRoot: 1, expectedGid: 5},

		{oldGid: 1, newRoot: 2, expectedGid: 2},
		{oldGid: 2, newRoot: 2, expectedGid: 4},
		{oldGid: 3, newRoot: 2, expectedGid: 5},
		{oldGid: 4, newRoot: 2, expectedGid: 8},
		{oldGid: 5, newRoot: 2, expectedGid: 9},

		{oldGid: 1, newRoot: 303, expectedGid: 303},
		{oldGid: 2, newRoot: 303, expectedGid: 606},
		{oldGid: 3, newRoot: 303, expectedGid: 607},
		{oldGid: 4, newRoot: 303, expectedGid: 1212},
		{oldGid: 5, newRoot: 303, expectedGid: 1213},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("oldGid: %d, newRoot: %d, expectedGid: %d", tc.oldGid, tc.newRoot, tc.expectedGid), func(t *testing.T) {
			result := offsetGidRoot(tc.oldGid, tc.newRoot)
			if result != tc.expectedGid {
				t.Fatalf("expected gid: %d, got: %d", tc.expectedGid, result)
			}
		})
	}
}

func validateStateProof(t *testing.T, leaf []byte, proof [][]byte, gid uint64, state *BeaconStateDeneb) ([]byte, []byte) {
	// First, offset the gid to account for the fact that state proofs are actually beacon block header proofs
	gid = offsetGidRoot(gid, beaconBlockHeaderStateRootGeneralizedIndex)
	currentHash := leaf

	// The gid is now the index of the leaf
	for i, proofRow := range proof {
		// The last neighbor must have a gid of either 2 or 3
		if i == len(proof)-1 {
			if gid != 2 && gid != 3 {
				t.Fatalf("last node/neighbor gid must be 2 or 3, got: %d", gid)
			}
		}
		// If the current gid is odd, the neighbor is on the left, otherwise it's on the right
		t.Logf("iter: %d, gid: %d, currentHash: %x, proofRow: %x", i, gid, currentHash, proofRow)
		neighborIsLeft := gid%2 == 1
		gid /= 2

		// hash the neighbor and the current hash together
		currentHash = hash(currentHash, proofRow, neighborIsLeft)
	}

	// Shallow copy the LatestBlockHeader
	lbh := *state.LatestBlockHeader
	// Restore when done
	defer func() {
		state.LatestBlockHeader = &lbh
	}()
	// Set the state root in LatestBlockHeader before calculating the hash
	stateRoot, err := state.HashTreeRoot()
	if err != nil {
		t.Fatalf("Failed to get state root: %v", err)
	}
	state.LatestBlockHeader.StateRoot = stateRoot[:]

	// The final hash must be the block root
	finalHash, err := state.LatestBlockHeader.HashTreeRoot()
	if err != nil {
		t.Fatalf("Failed to get block root: %v", err)
	}
	if !bytes.Equal(currentHash, finalHash[:]) {
		t.Logf("currentHash: %x", currentHash)
		t.Logf("block root: %x", finalHash)
		t.Fatalf("final hash does not match block root")
	}

	return stateRoot[:], finalHash[:]
}

func validateValidatorProof(t *testing.T, leaf []byte, proof [][]byte, gid uint64, state *BeaconStateDeneb) ([]byte, []byte) {
	gid *= 4
	return validateStateProof(t, leaf, proof, gid, state)
}

func validateWithdrawableEpochProof(t *testing.T, leaf []byte, proof [][]byte, gid uint64, state *BeaconStateDeneb) ([]byte, []byte) {
	gid = offsetGidRoot(beaconStateValidatorWithdrawableEpochGeneralizedIndex, gid)
	return validateStateProof(t, leaf, proof, gid, state)
}

func getValidatorLeaf(t *testing.T, validator *Validator) []byte {
	// The leaf for a validator is the parent hash of its first two fields.
	// this is at gid 4
	tree, err := validator.GetTree()
	if err != nil {
		t.Fatalf("Failed to get validator tree: %v", err)
	}

	leafNode, err := tree.Get(4)
	if err != nil {
		t.Fatalf("Failed to get validator leaf node: %v", err)
	}

	return leafNode.Hash()
}

func TestWithdrawalCredentialsStateProof(t *testing.T) {
	state := &BeaconStateDeneb{}
	err := state.UnmarshalSSZ(testState)
	if err != nil {
		t.Fatalf("Failed to unmarshal test state: %v", err)
	}

	type tc struct {
		validatorIndex uint64
		gid            uint64
	}

	testCases := []tc{
		{validatorIndex: 0, gid: 94557999988736},
		{validatorIndex: 111111, gid: 94558000099847},
		{validatorIndex: 555555, gid: 94558000544291},
	}

	// check a few of the proofs using the ssz library
	for _, tc := range testCases {
		proof, err := state.ValidatorCredentialsProof(tc.validatorIndex)
		if err != nil {
			t.Fatalf("Failed to get validator credentials proof: %v", err)
		}
		gid := getGeneralizedIndexForValidator(tc.validatorIndex)
		t.Logf("gid: %v", gid)
		if gid != tc.gid {
			t.Fatalf("expected gid: %v, got: %v", tc.gid, gid)
		}

		stateRoot, blockRoot := validateValidatorProof(t, getValidatorLeaf(t, state.Validators[tc.validatorIndex]), proof, tc.gid, state)
		t.Logf("stateRoot: %x", stateRoot)
		t.Logf("blockRoot: %x", blockRoot)
		expectedStateRoot, err := hex.DecodeString("2683ebc120f91f740c7bed4c866672d01e1ba51b4cc360297138465ee5df40f0")
		if err != nil {
			panic(err)
		}
		expectedBlockRoot, err := hex.DecodeString("376450cd7fb9f05ade82a7f88565ac57af449ac696b1a6ac5cc7dac7d467b7d6")
		if err != nil {
			panic(err)
		}
		if !bytes.Equal(stateRoot, expectedStateRoot) {
			t.Fatalf("expected state root: %x, got: %x", expectedStateRoot, stateRoot)
		}
		if !bytes.Equal(blockRoot, expectedBlockRoot) {
			t.Fatalf("expected block root: %x, got: %x", expectedBlockRoot, blockRoot)
		}
	}
}

func TestValidatorWithdrawableEpochProof(t *testing.T) {
	state := &BeaconStateDeneb{}
	err := state.UnmarshalSSZ(testState)
	if err != nil {
		t.Fatalf("Failed to unmarshal test state: %v", err)
	}

	type tc struct {
		validatorIndex uint64
		gid            uint64
	}

	testCases := []tc{
		{validatorIndex: 0, gid: 94557999988736},
		{validatorIndex: 111111, gid: 94558000099847},
		{validatorIndex: 555555, gid: 94558000544291},
	}

	for _, tc := range testCases {
		proof, err := state.ValidatorWithdrawableEpochProof(tc.validatorIndex)
		if err != nil {
			t.Fatalf("Failed to get validator withdrawable epoch proof: %v", err)
		}
		gid := getGeneralizedIndexForValidator(tc.validatorIndex)
		t.Logf("gid: %v", gid)
		if gid != tc.gid {
			t.Fatalf("expected gid: %v, got: %v", tc.gid, gid)
		}

		expectedWithdrawableEpoch := state.Validators[tc.validatorIndex].WithdrawableEpoch
		leaf := ssz.LeafFromUint64(expectedWithdrawableEpoch)

		stateRoot, blockRoot := validateWithdrawableEpochProof(t, leaf.Hash(), proof, tc.gid, state)
		t.Logf("stateRoot: %x", stateRoot)
		t.Logf("blockRoot: %x", blockRoot)
		expectedStateRoot, err := hex.DecodeString("2683ebc120f91f740c7bed4c866672d01e1ba51b4cc360297138465ee5df40f0")
		if err != nil {
			panic(err)
		}
		expectedBlockRoot, err := hex.DecodeString("376450cd7fb9f05ade82a7f88565ac57af449ac696b1a6ac5cc7dac7d467b7d6")
		if err != nil {
			panic(err)
		}
		if !bytes.Equal(stateRoot, expectedStateRoot) {
			t.Fatalf("expected state root: %x, got: %x", expectedStateRoot, stateRoot)
		}
		if !bytes.Equal(blockRoot, expectedBlockRoot) {
			t.Fatalf("expected block root: %x, got: %x", expectedBlockRoot, blockRoot)
		}
	}
}

func validateBlockProof(t *testing.T, leaf [32]byte, proof [][]byte, gid uint64, block *BeaconBlockDeneb) []byte {
	savedExepectedBlockRoot, err := hex.DecodeString("8442138d973483bfeaba9082f28217234e2879dedb5202e67ef68e2349db9a31")
	if err != nil {
		panic(err)
	}
	expectedBlockRoot, err := block.HashTreeRoot()
	if err != nil {
		t.Fatalf("Failed to get block root: %v", err)
	}

	if !bytes.Equal(expectedBlockRoot[:], savedExepectedBlockRoot) {
		t.Fatalf("expected block root: %x, got: %x", savedExepectedBlockRoot, expectedBlockRoot)
	}

	currentHash := leaf[:]
	for i, proofRow := range proof {
		// The last neighbor must have a gid of either 2 or 3
		if i == len(proof)-1 {
			if gid != 2 && gid != 3 {
				t.Fatalf("last node/neighbor gid must be 2 or 3, got: %d", gid)
			}
		}
		// If the current gid is odd, the neighbor is on the left, otherwise it's on the right
		t.Logf("iter: %d, gid: %d, currentHash: %x, proofRow: %x", i, gid, currentHash, proofRow)
		neighborIsLeft := gid%2 == 1
		gid /= 2

		// hash the neighbor and the current hash together
		currentHash = hash(currentHash, proofRow, neighborIsLeft)
	}

	if !bytes.Equal(currentHash, expectedBlockRoot[:]) {
		t.Fatalf("final hash does not match block root")
	} else {
		t.Logf("final hash %x matches expected block root %x", currentHash, expectedBlockRoot)
	}

	return currentHash
}

func TestWithdrawalProof(t *testing.T) {
	block := &SignedBeaconBlockDeneb{}
	err := block.UnmarshalSSZ(testBlock)
	if err != nil {
		t.Fatalf("Failed to unmarshal test block: %v", err)
	}

	for idx, withdrawal := range block.Block.Body.ExecutionPayload.Withdrawals {
		proof, err := block.Block.ProveWithdrawal(uint64(idx))
		if err != nil {
			t.Fatalf("Failed to get withdrawal proof: %v", err)
		}

		if testing.Verbose() {
			for i, p := range proof {
				t.Logf("proof[%d]: %x", i, p)
			}
		}
		gid := uint64(1)
		gid = gid*beaconBlockDenebChunksCeil + beaconBlockDenebBodyIndex
		gid = gid*beaconBlockDenebBodyChunksCeil + beaconBlockDenebBodyExecutionPayloadIndex
		gid = gid*beaconBlockDenebBodyExecutionPayloadChunksCeil + beaconBlockDenebBodyExecutionPayloadWithdrawalsIndex
		gid = gid * 2
		gid = gid*beaconBlockDenebWithdrawalsArrayMax + uint64(idx)
		leaf, err := withdrawal.HashTreeRoot()
		if err != nil {
			t.Fatalf("Failed to get withdrawal leaf: %v", err)
		}
		validateBlockProof(t, leaf, proof, gid, block.Block)
	}
}

func TestBlockRootProof(t *testing.T) {
	state := &BeaconStateDeneb{}
	err := state.UnmarshalSSZ(testState)
	if err != nil {
		t.Fatalf("Failed to unmarshal test state: %v", err)
	}

	for _, slot := range []uint64{0, 10, 100, 1000} {

		proof, err := state.BlockRootProof(slot)
		if err != nil {
			t.Fatalf("Failed to get block root proof: %v", err)
		}

		gid := uint64(1)
		gid = gid*beaconStateChunkCeil + beaconStateBlockRootsFieldIndex
		gid = gid*beaconStateBlockRootsMaxLength + (slot % slotsPerHistoricalRoot)

		stateRoot, blockRoot := validateStateProof(t, state.BlockRoots[slot%slotsPerHistoricalRoot][:], proof, gid, state)
		t.Logf("stateRoot: %x", stateRoot)
		t.Logf("blockRoot: %x", blockRoot)
		expectedStateRoot, err := hex.DecodeString("2683ebc120f91f740c7bed4c866672d01e1ba51b4cc360297138465ee5df40f0")
		if err != nil {
			panic(err)
		}
		expectedBlockRoot, err := hex.DecodeString("376450cd7fb9f05ade82a7f88565ac57af449ac696b1a6ac5cc7dac7d467b7d6")
		if err != nil {
			panic(err)
		}
		if !bytes.Equal(stateRoot, expectedStateRoot) {
			t.Fatalf("expected state root: %x, got: %x", expectedStateRoot, stateRoot)
		}
		if !bytes.Equal(blockRoot, expectedBlockRoot) {
			t.Fatalf("expected block root: %x, got: %x", expectedBlockRoot, blockRoot)
		}
	}
}
