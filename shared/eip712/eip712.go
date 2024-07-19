package eip712

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

type EIP712Components struct {
	V uint8    `json:"v"`
	R [32]byte `json:"r"`
	S [32]byte `json:"s"`
}

const EIP712Length = 65

// Pretty print for EIP712Components
func (e EIP712Components) String() {
	fmt.Printf("EIP712 Components:\n")
	fmt.Printf("V: %d\n", e.V)
	fmt.Printf("R: %x\n", e.R)
	fmt.Printf("S: %x\n", e.S)
}

// SanitizeHexInput checks for and strips a 0x prefix, returning an error if one is not found.
// It returns the byte slice representing the decoded hex, or a decoding error
func SanitizeHex(inp string) ([]byte, error) {
	if !strings.HasPrefix(inp, "0x") {
		return nil, errors.New("expected hex prefix")
	}
	return hex.DecodeString(strings.TrimPrefix(inp, "0x"))
}

// SanitizeEIP712String checks that a 0x-prefixed hex string is the proper length to be
// parsed as a EIP-712 signature, and parses it, returning an error if the length is invalid
// or the hex cannot be decoded.
func SanitizeEIP712String(inp string) ([]byte, error) {
	decoded, err := SanitizeHex(inp)
	if err != nil {
		return nil, fmt.Errorf("error sanitizing EIP-712 signature string as hex: %w", err)
	}
	if len(decoded) != EIP712Length {
		return nil, fmt.Errorf("error sanitizing EIP-712 signature string: invalid length %d (expected %d)", len(decoded), EIP712Length)
	}
	return decoded, nil
}

// ParseEIP712Components expects a sanitized 65 byte EIP-712 signature
// and returns v/r/s: v as uint8 and r, s as [32]byte
func ParseEIP712Components(data []byte) (EIP712Components, error) {
	if len(data) != EIP712Length {
		return EIP712Components{}, fmt.Errorf("error parsing EIP-712 signature string: invalid length %d (expected %d)", len(data), EIP712Length)
	}

	var v uint8
	var r [32]byte
	var s [32]byte

	v = data[64]
	copy(r[:], data[0:32])
	copy(s[:], data[32:64])

	return EIP712Components{
		V: v,
		R: r,
		S: s,
	}, nil
}
