package megapool

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

func TestMegapoolABIIncludesProoflessStake(t *testing.T) {
	signatures := []string{
		"stakeWithoutProof(uint32)",
		"getValidatorChallengeDeadline(uint32)",
		"getValidatorPrestakeSignature(uint32)",
	}
	for _, encoded := range []string{megapoolV1EncodedAbi, megapoolV2EncodedAbi} {
		parsed, err := rocketpool.DecodeAbi(encoded)
		if err != nil {
			t.Fatal(err)
		}
		for _, signature := range signatures {
			name := signature[:bytes.IndexByte([]byte(signature), '(')]
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
		if _, err := parsed.Pack("getValidatorChallengeDeadline", uint32(1)); err != nil {
			t.Fatal(err)
		}
		if _, err := parsed.Pack("getValidatorPrestakeSignature", uint32(1)); err != nil {
			t.Fatal(err)
		}
	}
}
