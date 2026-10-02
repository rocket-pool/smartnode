package services

import (
	"bytes"
	"context"
	"io"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
)

type headProofExecutionClient struct {
	rocketpool.ExecutionClient
	header *ethtypes.Header
}

func (c *headProofExecutionClient) HeaderByNumber(_ context.Context, number *big.Int) (*ethtypes.Header, error) {
	if number != nil {
		panic("head proof must use execution head")
	}
	return c.header, nil
}

type headProofBeaconClient struct {
	beacon.Client
	header        beacon.BeaconBlockHeader
	state         []byte
	found         bool
	requestedRoot string
	requestedSlot uint64
	index         string
}

func (c *headProofBeaconClient) GetBeaconBlockHeader(root string) (beacon.BeaconBlockHeader, bool, error) {
	c.requestedRoot = root
	return c.header, c.found, nil
}

func (c *headProofBeaconClient) GetBeaconStateSSZ(slot uint64) (*beacon.BeaconStateSSZ, error) {
	c.requestedSlot = slot
	return &beacon.BeaconStateSSZ{Data: io.NopCloser(bytes.NewReader(c.state)), Size: int64(len(c.state)), Fork: "fulu"}, nil
}

func (c *headProofBeaconClient) GetValidatorIndex(rptypes.ValidatorPubkey) (string, error) {
	return c.index, nil
}

func TestHeadStakeProofUsesExecutionAnchor(t *testing.T) {
	// Slot 100 can be newer than finality and several slots behind beacon head.
	// A timestamp from a later execution block also handles skipped payloads.
	state := newProofTestFuluState(t, 1, 100)
	root := common.BytesToHash(anchorBlockRoot(t, state))
	stateRoot, err := generic.SSZ.HashTreeRoot(state)
	if err != nil {
		t.Fatal(err)
	}
	data, err := generic.SSZ.MarshalSSZ(state)
	if err != nil {
		t.Fatal(err)
	}
	ec := &headProofExecutionClient{header: &ethtypes.Header{ParentBeaconRoot: &root, Time: 2000}}
	bc := &headProofBeaconClient{header: beacon.BeaconBlockHeader{Root: root, StateRoot: common.Hash(stateRoot), Slot: 100}, state: data, found: true, index: "0"}
	anchor, timestamp, err := GetHeadBeaconState(bc, ec)
	if err != nil {
		t.Fatal(err)
	}
	if bc.requestedRoot != root.Hex() || bc.requestedSlot != 100 || timestamp != 2000 {
		t.Fatalf("wrong anchor: root=%s slot=%d timestamp=%d", bc.requestedRoot, bc.requestedSlot, timestamp)
	}
	var pubkey rptypes.ValidatorPubkey
	copy(pubkey[:], state.Validators[0].Pubkey)
	validator, slot, err := GetValidatorProofFromState(bc, pubkey, anchor)
	if err != nil {
		t.Fatal(err)
	}
	if validator.ValidatorIndex.Uint64() != 0 || slot.Slot != 100 {
		t.Fatal("staking proof did not preserve the selected state and timestamp")
	}
	// Restore the slot proof to the exact EIP-4788 block root.
	var leaf [32]byte
	leaf[0] = 100
	gindex := new(big.Int).SetUint64(fuluSlotProofGindex())
	witnesses := make([][]byte, len(slot.Witnesses))
	for i := range slot.Witnesses {
		witnesses[i] = slot.Witnesses[i][:]
	}
	if !bytes.Equal(walkWitnessChain(t, leaf[:], gindex, witnesses), root[:]) {
		t.Fatal("slot proof does not match execution head's parent beacon root")
	}
	bc.index = "1"
	if _, _, err := GetValidatorProofFromState(bc, pubkey, anchor); err == nil || !strings.Contains(err.Error(), "index not found") {
		t.Fatalf("expected missing validator to wait for the proof state, got %v", err)
	}
	bc.index = "0"
	pubkey[0]++
	if _, _, err := GetValidatorProofFromState(bc, pubkey, anchor); err == nil || !strings.Contains(err.Error(), "pubkey does not match") {
		t.Fatalf("expected changed validator identity to be rejected, got %v", err)
	}
	for _, tt := range []struct {
		name   string
		header beacon.BeaconBlockHeader
		found  bool
		want   string
	}{
		{name: "unavailable block", header: bc.header, found: false, want: "not available"},
		{name: "different block", header: beacon.BeaconBlockHeader{Root: common.HexToHash("0xff")}, found: true, want: "block root does not match"},
		{name: "state reorg", header: beacon.BeaconBlockHeader{Root: root, StateRoot: common.HexToHash("0xff"), Slot: 100}, found: true, want: "state root does not match"},
		{name: "different slot", header: beacon.BeaconBlockHeader{Root: root, StateRoot: common.Hash(stateRoot), Slot: 101}, found: true, want: "state slot does not match"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			bc.header, bc.found = tt.header, tt.found
			if _, _, err := GetHeadBeaconState(bc, ec); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q, got %v", tt.want, err)
			}
		})
	}
	ec.header.ParentBeaconRoot = nil
	if _, _, err := GetHeadBeaconState(bc, ec); err == nil || !strings.Contains(err.Error(), "no parent beacon block root") {
		t.Fatalf("expected missing root error, got %v", err)
	}
}

func fuluSlotProofGindex() uint64 {
	// BeaconBlockHeader.state_root (11) followed by BeaconState.slot (66).
	return (11 << 6) | 2
}
