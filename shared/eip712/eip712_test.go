package eip712

import (
	"testing"
)

func TestDecodeAndEncode(t *testing.T) {

	// signer := "0x18eea3fbe5008d6f7a95d963a4be403e82d35758"
	signature := "0xba283b21f7168e53b082ad552d974591abe0f4db5b7032374abbcdcf09e0eadc2c0530ff4ac1d63e19c1ceca2d14b374c86b6c84f46bbd57747b48c21388c4e71c"
	eip712Components := new(EIP712Components)

	// Decode string
	err := eip712Components.Decode(signature)
	if err != nil {
		t.Fatalf("Failed to decode signature: %v", err)
	}

	eip712Components.String()

	// Encode eip712Components back to a string
	encodedSig := eip712Components.Encode()

	if encodedSig != signature {
		t.Fatalf("Expected %s but got %s", signature, encodedSig)
	}

}

func TestInvalidSignature(t *testing.T) {
	invalidSignature := "0xinvalidsignature"
	eip712Components := new(EIP712Components)

	err := eip712Components.Decode(invalidSignature)
	if err == nil {
		t.Fatal("Expected error for invalid signature but got none")
	}
}

func TestEmptySignature(t *testing.T) {
	emptySignature := ""
	eip712Components := new(EIP712Components)

	err := eip712Components.Decode(emptySignature)
	if err == nil {
		t.Fatal("Expected error for empty signature but got none")
	}
}
