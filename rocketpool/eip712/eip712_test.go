package eip712

import (
	"bytes"
	"encoding/hex"
	"testing"
)

const (
	sigStr = "0x659e8fd295febd9bdec44e1e357f57860498756080ff588ed4b47cfe66306ee94652244f44f4b0e6b059fadd07a9ca81361fc40af4391c4d54bbc40e471f79921b"

	expectedV = 27
	expectedR = "659e8fd295febd9bdec44e1e357f57860498756080ff588ed4b47cfe66306ee9"
	expectedS = "4652244f44f4b0e6b059fadd07a9ca81361fc40af4391c4d54bbc40e471f7992"
)

func TestUnmarshalText(t *testing.T) {
	sig := Components{}
	err := sig.UnmarshalText([]byte(sigStr))
	if err != nil {
		t.Fatalf("Failed to unmarshal text: %v", err)
	}

	if sig.V != expectedV {
		t.Fatalf("Expected V: %d, got: %d", expectedV, sig.V)
	}
	expectedRBytes, err := hex.DecodeString(expectedR)
	if err != nil {
		t.Fatalf("Failed to decode expected R: %v", err)
	}
	if !bytes.Equal(sig.R[:], expectedRBytes) {
		t.Fatalf("Expected R: %x, got: %x", expectedRBytes, sig.R[:])
	}
	expectedSBytes, err := hex.DecodeString(expectedS)
	if err != nil {
		t.Fatalf("Failed to decode expected S: %v", err)
	}
	if !bytes.Equal(sig.S[:], expectedSBytes) {
		t.Fatalf("Expected S: %x, got: %x", expectedSBytes, sig.S[:])
	}
}

func TestRoundTrip(t *testing.T) {
	sig := Components{}
	err := sig.UnmarshalText([]byte(sigStr))
	if err != nil {
		t.Fatalf("Failed to unmarshal text: %v", err)
	}

	marshalled, err := sig.MarshalText()
	if err != nil {
		t.Fatalf("Failed to marshal text: %v", err)
	}

	if string(marshalled) != sigStr {
		t.Fatalf("Expected marshalled: %s, got: %s", sigStr, string(marshalled))
	}
}
