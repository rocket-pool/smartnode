package eip712

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func TestDecodeAndEncode(t *testing.T) {
	signature := "0xba283b21f7168e53b082ad552d974591abe0f4db5b7032374abbcdcf09e0eadc2c0530ff4ac1d63e19c1ceca2d14b374c86b6c84f46bbd57747b48c21388c4e71c"
	eip712Components := new(EIP712Components)

	// Decode string
	decoded, err := hexutil.Decode(signature)
	if err != nil {
		t.Fatalf("Failed to decode hex string: %v", err)
	}

	err = eip712Components.UnmarshalText(decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshall signature: %v", err)
	}

	decoded, err = eip712Components.MarshalText()
	if err != nil {
		t.Fatalf("Failed to marshall signature: %v", err)
	}

	// Convert the encoded byte slice back to a hex string
	encodedHexSig := hexutil.Encode(decoded)

	if encodedHexSig != signature {
		t.Fatalf("Expected %s but got %s", signature, decoded)
	}

}

func TestDecodeInvalid712Hex(t *testing.T) {
	invalidSignature := "0xinvalidsignature"
	eip712Components := new(EIP712Components)

	err := eip712Components.UnmarshalText([]byte(invalidSignature))
	if err == nil {
		t.Fatal("Expected error for invalid signature but got none")
	}
}

func TestDecodeEmptySignature(t *testing.T) {
	emptySignature := ""
	eip712Components := new(EIP712Components)

	err := eip712Components.UnmarshalText([]byte(emptySignature))
	if err == nil {
		t.Fatal("Expected error for empty signature but got none")
	}
}

func TestDecodeInvalidLength(t *testing.T) {
	// Create a hex-encoded signature with an invalid length (not 65 bytes)
	invalidLengthSignature := "0xba283b21f7168e53b082ad552d974591abe0f4db5b7032374abbcdcf09e0eadc" // 64 characters (32 bytes)
	eip712Components := new(EIP712Components)

	// Decode string
	decoded, err := hexutil.Decode(invalidLengthSignature)
	if err != nil {
		t.Fatalf("Failed to decode hex string: %v", err)
	}

	err = eip712Components.UnmarshalText(decoded)
	if err == nil {
		t.Fatal("Expected error for signature with invalid length but got none")
	}

	expectedErrMsg := fmt.Sprintf("error decoding EIP-712 signature string: invalid length %d bytes (expected %d bytes)", 32, EIP712Length)
	if err.Error() != expectedErrMsg {
		t.Fatalf("Expected error message: '%s' but got: '%s'", expectedErrMsg, err.Error())
	}
}

func TestValidateSuccess(t *testing.T) {
	// Create a valid EIP-712 signature
	signature := "0xba283b21f7168e53b082ad552d974591abe0f4db5b7032374abbcdcf09e0eadc2c0530ff4ac1d63e19c1ceca2d14b374c86b6c84f46bbd57747b48c21388c4e71c"
	decoded, err := hexutil.Decode(signature)
	if err != nil {
		t.Fatalf("Failed to decode hex string: %v", err)
	}

	eip712Components := new(EIP712Components)
	err = eip712Components.UnmarshalText(decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal signature: %v", err)
	}

	// Message to be signed
	message := "0xe8325f5f4486c2ff2ac7b522fbc9eb249d46c936 may delegate to me for Rocket Pool governance"
	msg := []byte(message)

	// Expected signer
	expectedSigner := common.HexToAddress("0x18eea3fBe5008d6f7a95d963a4BE403E82d35758")

	// Validate
	err = eip712Components.Validate(msg, expectedSigner)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}
}

func TestValidateInvalidSignature(t *testing.T) {
	// Create an invalid EIP-712 signature
	invalidSignature := "0xba283b21f7168e53b082ad552d974591abe0f4db5b7032374abbcdcf09e0eadc2c0530ff4ac1d63e19c1ceca2d14b374c86b6c84f46bbd57747b48c21388c4e71f" // Last byte is invalid
	decoded, _ := hexutil.Decode(invalidSignature)
	eip712Components := new(EIP712Components)
	err := eip712Components.UnmarshalText(decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal signature: %v", err)
	}

	// Message to be signed
	msg := []byte("0xe8325f5f4486c2ff2ac7b522fbc9eb249d46c936 may delegate to me for Rocket Pool governance")

	// Some arbitrary expected signer
	expectedSigner := common.HexToAddress("0x7f0bfc4a2d057dc75a7a2d3c9dc7eae2b3881e3e")

	// Validate
	err = eip712Components.Validate(msg, expectedSigner)
	if err == nil {
		t.Fatal("Expected error for invalid signature but got none")
	}
}

func TestValidateSignerMismatch(t *testing.T) {
	// Create a valid EIP-712 signature
	signature := "0xba283b21f7168e53b082ad552d974591abe0f4db5b7032374abbcdcf09e0eadc2c0530ff4ac1d63e19c1ceca2d14b374c86b6c84f46bbd57747b48c21388c4e71c"
	decoded, err := hexutil.Decode(signature)
	if err != nil {
		t.Fatalf("Failed to decode hex string: %v", err)
	}

	eip712Components := new(EIP712Components)
	err = eip712Components.UnmarshalText(decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal signature: %v", err)
	}

	// Message to be signed
	message := "0xe8325f5f4486c2ff2ac7b522fbc9eb249d46c936 may delegate to me for Rocket Pool governance"
	msg := []byte(message)

	// Provide a mismatched expected signer
	expectedSigner := common.HexToAddress("0x0000000000000000000000000000000000000000")

	// Validate
	err = eip712Components.Validate(msg, expectedSigner)
	if err == nil {
		t.Fatalf("Expected validation to fail, but it succeeded")
	}
}
