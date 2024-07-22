package eip712

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type EIP712Components struct {
	R [32]byte `json:"r"`
	S [32]byte `json:"s"`
	V uint8    `json:"v"`
}

const EIP712Length = 65

// Pretty print for EIP712Components
func (e EIP712Components) String() {
	fmt.Println("EIP712 Components:")
	fmt.Printf("R: %x\n", e.R)
	fmt.Printf("S: %x\n", e.S)
	fmt.Printf("V: %d\n", e.V)
}

// Decode decodes a hex-encoded EIP-712 signature string, verifies the length
// and assigns the appropriate bytes to R/S/V
func (e *EIP712Components) Decode(inp string) error {
	decodedBytes, err := hexutil.Decode(inp)
	if err != nil {
		return err
	}

	if len(decodedBytes) != EIP712Length {
		return fmt.Errorf("error decoding EIP-712 signature string: invalid length %d bytes (expected %d bytes)", len(decodedBytes), EIP712Length)
	}

	copy(e.R[:], decodedBytes[0:32])
	copy(e.S[:], decodedBytes[32:64])
	e.V = decodedBytes[64]

	return nil
}

// Encode initializes an empty byte slice, copies fields R/S/V into it,
// encodes the byte slice into a hex string then returns it
func (e *EIP712Components) Encode() string {
	encodedBytes := make([]byte, EIP712Length)

	copy(encodedBytes[0:32], e.R[:])
	copy(encodedBytes[32:64], e.S[:])
	encodedBytes[64] = e.V

	return hexutil.Encode(encodedBytes)
}
