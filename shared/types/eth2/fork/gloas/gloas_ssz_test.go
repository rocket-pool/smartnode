package gloas

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
)

// minimalBeaconState returns a Gloas BeaconState with required fixed-size
// vectors populated so SSZ marshal/HTR succeed.
func minimalBeaconState() *BeaconState {
	state := &BeaconState{
		GenesisTime:           1,
		GenesisValidatorsRoot: make([]byte, 32),
		Slot:                  42,
		Fork: &generic.Fork{
			PreviousVersion: make([]byte, 4),
			CurrentVersion:  make([]byte, 4),
			Epoch:           0,
		},
		LatestBlockHeader: &generic.BeaconBlockHeader{
			ParentRoot: make([]byte, 32),
			StateRoot:  make([]byte, 32),
			BodyRoot:   make([]byte, 32),
		},
		Eth1Data: &generic.Eth1Data{
			DepositRoot: make([]byte, 32),
			BlockHash:   make([]byte, 32),
		},
		RandaoMixes:       make([][]byte, 65536),
		Slashings:         make([]uint64, 8192),
		JustificationBits: [1]byte{0},
		PreviousJustifiedCheckpoint: &generic.Checkpoint{
			Root: make([]byte, 32),
		},
		CurrentJustifiedCheckpoint: &generic.Checkpoint{
			Root: make([]byte, 32),
		},
		FinalizedCheckpoint: &generic.Checkpoint{
			Root: make([]byte, 32),
		},
		CurrentSyncCommittee: &generic.SyncCommittee{
			PubKeys: make([][]byte, 512),
		},
		NextSyncCommittee: &generic.SyncCommittee{
			PubKeys: make([][]byte, 512),
		},
		ProposerLookahead:            make([]uint64, 64),
		ExecutionPayloadAvailability: make([]byte, 1024),
		LatestExecutionPayloadBid: &ExecutionPayloadBid{
			BlobKzgCommitments: nil,
		},
	}
	for i := range state.RandaoMixes {
		state.RandaoMixes[i] = make([]byte, 32)
	}
	for i := range state.CurrentSyncCommittee.PubKeys {
		state.CurrentSyncCommittee.PubKeys[i] = make([]byte, 48)
	}
	for i := range state.NextSyncCommittee.PubKeys {
		state.NextSyncCommittee.PubKeys[i] = make([]byte, 48)
	}
	for i := range state.BuilderPendingPayments {
		state.BuilderPendingPayments[i] = &BuilderPendingPayment{
			Withdrawal: &BuilderPendingWithdrawal{},
		}
	}
	return state
}

func TestBeaconStateRoundTrip(t *testing.T) {
	state := minimalBeaconState()
	state.Validators = []*generic.Validator{
		{
			Pubkey:                make([]byte, 48),
			WithdrawalCredentials: make([]byte, 32),
		},
	}
	state.Balances = []uint64{32_000_000_000}
	state.PayloadExpectedWithdrawals = []*generic.Withdrawal{
		{
			Index:          1,
			ValidatorIndex: 0,
			Address:        [20]byte{},
			Amount:         1_000_000_000,
		},
	}

	encoded, err := generic.SSZ.MarshalSSZ(state)
	if err != nil {
		t.Fatalf("MarshalSSZ: %v", err)
	}

	decoded := &BeaconState{}
	if err := generic.SSZ.UnmarshalSSZ(decoded, encoded); err != nil {
		t.Fatalf("UnmarshalSSZ: %v", err)
	}

	if decoded.Slot != state.Slot {
		t.Fatalf("slot: got %d want %d", decoded.Slot, state.Slot)
	}
	if len(decoded.Validators) != 1 {
		t.Fatalf("validators len: got %d want 1", len(decoded.Validators))
	}
	if len(decoded.PayloadExpectedWithdrawals) != 1 {
		t.Fatalf("payload_expected_withdrawals len: got %d want 1", len(decoded.PayloadExpectedWithdrawals))
	}
	if decoded.PayloadExpectedWithdrawals[0].Amount != 1_000_000_000 {
		t.Fatalf("withdrawal amount: got %d", decoded.PayloadExpectedWithdrawals[0].Amount)
	}

	origRoot, err := generic.SSZ.HashTreeRoot(state)
	if err != nil {
		t.Fatalf("HashTreeRoot original: %v", err)
	}
	decodedRoot, err := generic.SSZ.HashTreeRoot(decoded)
	if err != nil {
		t.Fatalf("HashTreeRoot decoded: %v", err)
	}
	if !bytes.Equal(origRoot[:], decodedRoot[:]) {
		t.Fatalf("HTR mismatch after round-trip:\n  orig=%x\n  dec =%x", origRoot, decodedRoot)
	}
}

func TestSignedBeaconBlockRoundTrip(t *testing.T) {
	block := &SignedBeaconBlock{
		Block: &BeaconBlock{
			Slot:       7,
			ParentRoot: [32]byte{1},
			StateRoot:  [32]byte{2},
			Body: &BeaconBlockBody{
				RandaoReveal: make([]byte, 96),
				Eth1Data: &generic.Eth1Data{
					DepositRoot: make([]byte, 32),
					BlockHash:   make([]byte, 32),
				},
				SyncAggregate: &generic.SyncAggregate{
					SyncCommiteeBits: make([]byte, 64),
				},
				SignedExecutionPayloadBid: &SignedExecutionPayloadBid{
					Message: &ExecutionPayloadBid{},
				},
				ParentExecutionRequests: &ExecutionRequests{},
			},
		},
		Signature: make([]byte, 96),
	}

	encoded, err := generic.SSZ.MarshalSSZ(block)
	if err != nil {
		t.Fatalf("MarshalSSZ: %v", err)
	}

	decoded := &SignedBeaconBlock{}
	if err := generic.SSZ.UnmarshalSSZ(decoded, encoded); err != nil {
		t.Fatalf("UnmarshalSSZ: %v", err)
	}

	if decoded.Block.Slot != block.Block.Slot {
		t.Fatalf("slot: got %d want %d", decoded.Block.Slot, block.Block.Slot)
	}

	origRoot, err := generic.SSZ.HashTreeRoot(block)
	if err != nil {
		t.Fatalf("HashTreeRoot original: %v", err)
	}
	decodedRoot, err := generic.SSZ.HashTreeRoot(decoded)
	if err != nil {
		t.Fatalf("HashTreeRoot decoded: %v", err)
	}
	if !bytes.Equal(origRoot[:], decodedRoot[:]) {
		t.Fatalf("HTR mismatch after round-trip:\n  orig=%x\n  dec =%x", origRoot, decodedRoot)
	}
}

func TestProgressiveTypesValidate(t *testing.T) {
	// Ensure dynamic-ssz accepts progressive container/list annotations on Gloas types.
	for _, typ := range []any{
		(*BeaconState)(nil),
		(*BeaconBlockBody)(nil),
		(*Attestation)(nil),
		(*ExecutionPayloadBid)(nil),
		(*ExecutionRequests)(nil),
		(*IndexedAttestation)(nil),
		(*PayloadAttestation)(nil),
	} {
		rt := reflect.TypeOf(typ).Elem()
		if err := generic.SSZ.ValidateType(rt); err != nil {
			t.Fatalf("ValidateType(%s): %v", rt, err)
		}
	}
}
