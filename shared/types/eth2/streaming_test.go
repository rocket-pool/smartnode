package eth2

import (
	"bytes"
	"io"
	"testing"

	"github.com/rocket-pool/smartnode/shared/types/eth2/fork/deneb"
	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
)

// Verify that the streaming decode path produces a state identical to the
// buffered decode path
func TestBeaconStateStreamingDecode(t *testing.T) {
	buffered := &deneb.BeaconState{}
	if err := generic.SSZ.UnmarshalSSZ(buffered, testState); err != nil {
		t.Fatalf("Failed to unmarshal test state: %v", err)
	}
	bufferedRoot, err := generic.SSZ.HashTreeRoot(buffered)
	if err != nil {
		t.Fatalf("Failed to get buffered state root: %v", err)
	}

	streamed := &deneb.BeaconState{}
	if err := generic.SSZ.UnmarshalSSZReader(streamed, bytes.NewReader(testState), len(testState)); err != nil {
		t.Fatalf("Failed to stream-unmarshal test state: %v", err)
	}
	streamedRoot, err := generic.SSZ.HashTreeRoot(streamed)
	if err != nil {
		t.Fatalf("Failed to get streamed state root: %v", err)
	}

	if bufferedRoot != streamedRoot {
		t.Fatalf("streamed state root %x does not match buffered state root %x", streamedRoot, bufferedRoot)
	}
}

// Verify that decodeSSZ streams when the size is known and falls back to
// buffering when it isn't, producing identical results
func TestDecodeSSZSizeFallback(t *testing.T) {
	streamed := &deneb.SignedBeaconBlock{}
	if err := decodeSSZ(streamed, io.NopCloser(bytes.NewReader(testBlock)), int64(len(testBlock))); err != nil {
		t.Fatalf("Failed to decode block with known size: %v", err)
	}

	// A size of -1 mimics a response without Content-Length
	buffered := &deneb.SignedBeaconBlock{}
	if err := decodeSSZ(buffered, io.NopCloser(bytes.NewReader(testBlock)), -1); err != nil {
		t.Fatalf("Failed to decode block with unknown size: %v", err)
	}

	streamedRoot, err := generic.SSZ.HashTreeRoot(streamed.Block)
	if err != nil {
		t.Fatalf("Failed to get streamed block root: %v", err)
	}
	bufferedRoot, err := generic.SSZ.HashTreeRoot(buffered.Block)
	if err != nil {
		t.Fatalf("Failed to get buffered block root: %v", err)
	}
	if streamedRoot != bufferedRoot {
		t.Fatalf("streamed block root %x does not match buffered block root %x", streamedRoot, bufferedRoot)
	}
}

func BenchmarkBeaconStateDecodeStreaming(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		state := &deneb.BeaconState{}
		if err := generic.SSZ.UnmarshalSSZReader(state, bytes.NewReader(testState), len(testState)); err != nil {
			b.Fatalf("Failed to stream-unmarshal test state: %v", err)
		}
	}
}

func BenchmarkBeaconStateDecodeBuffered(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		state := &deneb.BeaconState{}
		if err := generic.SSZ.UnmarshalSSZ(state, testState); err != nil {
			b.Fatalf("Failed to unmarshal test state: %v", err)
		}
	}
}

func BenchmarkValidatorProof(b *testing.B) {
	state := &deneb.BeaconState{}
	if err := generic.SSZ.UnmarshalSSZ(state, testState); err != nil {
		b.Fatalf("Failed to unmarshal test state: %v", err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := state.ValidatorProof(555555); err != nil {
			b.Fatalf("Failed to get validator proof: %v", err)
		}
	}
}
