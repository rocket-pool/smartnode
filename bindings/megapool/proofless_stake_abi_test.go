package megapool

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

func TestProoflessStakeABISelectors(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(prooflessStakeABI))
	if err != nil {
		t.Fatal(err)
	}
	for _, signature := range []string{
		"stakeWithoutProof(uint32)",
		"getValidatorChallengeDeadline(uint32)",
		"getValidatorPrestakeSignature(uint32)",
	} {
		name := signature[:strings.Index(signature, "(")]
		method, ok := parsed.Methods[name]
		if !ok {
			t.Fatalf("missing %s", name)
		}
		want := crypto.Keccak256([]byte(signature))[:4]
		if !bytes.Equal(method.ID, want) {
			t.Fatalf("%s selector = %x, want %x", signature, method.ID, want)
		}
	}

	packed, err := parsed.Pack("stakeWithoutProof", uint32(4))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(packed[:4], crypto.Keccak256([]byte("stakeWithoutProof(uint32)"))[:4]) {
		t.Fatalf("packed selector = %x", packed[:4])
	}
	var arg [32]byte
	arg[31] = 4
	if !bytes.Equal(packed[4:], arg[:]) {
		t.Fatalf("packed validator id = %x", packed[4:])
	}
}

func TestMegapoolABIMergesProoflessStakeMethods(t *testing.T) {
	for _, encoded := range []string{megapoolV1EncodedAbi, megapoolV2EncodedAbi} {
		decoded, err := rocketpool.DecodeAbi(encoded)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := decoded.Methods["stakeWithoutProof"]; ok {
			t.Fatal("encoded megapool ABI already contains stakeWithoutProof")
		}
		merged, err := withProoflessStakeMethods(decoded)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := decoded.Methods["stakeWithoutProof"]; ok {
			t.Fatal("merge mutated the decoded ABI")
		}
		packed, err := merged.Pack("stakeWithoutProof", uint32(1))
		if err != nil {
			t.Fatal(err)
		}
		want := crypto.Keccak256([]byte("stakeWithoutProof(uint32)"))[:4]
		if !bytes.Equal(packed[:4], want) {
			t.Fatalf("merged selector = %x, want %x", packed[:4], want)
		}
		if _, err := merged.Pack("getValidatorChallengeDeadline", uint32(1)); err != nil {
			t.Fatal(err)
		}
		if _, err := merged.Pack("getValidatorPrestakeSignature", uint32(1)); err != nil {
			t.Fatal(err)
		}
	}
}
