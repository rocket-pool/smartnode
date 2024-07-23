package eip712

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type EIP712Components struct {
	R [32]byte `json:"r"`
	S [32]byte `json:"s"`
	V uint8    `json:"v"`
}

const EIP712Length = 65

// Pretty print for EIP712Components
func (e *EIP712Components) Print() {
	fmt.Println("EIP712 Components:")
	fmt.Printf("R: %x\n", e.R)
	fmt.Printf("S: %x\n", e.S)
	fmt.Printf("V: %d\n", e.V)
}

// String returns a hexadecimal string representation of EIP712Components
func (e *EIP712Components) String() string {
	out, err := e.MarshalText()
	if err != nil {
		// MarshalText should never return an error
		panic(err)
	}
	return fmt.Sprintf("%x", out)

}

// UnmarshalText verifies the length of a decoded EIP-712 signature and assigns the appropriate bytes to R/S/V
func (e *EIP712Components) UnmarshalText(inp []byte) error {
	if len(inp) != EIP712Length {
		return fmt.Errorf("error decoding EIP-712 signature string: invalid length %d bytes (expected %d bytes)", len(inp), EIP712Length)
	}

	copy(e.R[:], inp[0:32])
	copy(e.S[:], inp[32:64])
	e.V = inp[64]

	return nil
}

// MarshalText initializes an empty byte slice, copies fields R/S/V into signatureBytes then returns it
func (e *EIP712Components) MarshalText() ([]byte, error) {
	signatureBytes := make([]byte, EIP712Length)

	copy(signatureBytes[0:32], e.R[:])
	copy(signatureBytes[32:64], e.S[:])
	signatureBytes[64] = e.V

	return signatureBytes, nil
}

// Validate recovers the address of a signer from a message and signature then
// compares the recovered signer address to the expected signer address
func (e *EIP712Components) Validate(msg []byte, expectedSigner common.Address) error {

	hash := accounts.TextHash(msg)

	// Convert the EIP712Components to a signature
	sig := make([]byte, 65)
	copy(sig[0:32], e.R[:])
	copy(sig[32:64], e.S[:])
	sig[64] = e.V

	// V (Recovery ID) must by 27 or 28, so we subtract 27 from 0 or 1 to get the recovery ID.
	sig[crypto.RecoveryIDOffset] -= 27

	// Recover the public key from the signature
	pubKey, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return fmt.Errorf("error recovering public key: %v", err)
	}

	// Restore V to its original value
	sig[crypto.RecoveryIDOffset] += 27

	// Derive the signer address from the public key
	recoveredSigner := crypto.PubkeyToAddress(*pubKey)

	// Compare the recovered signer address with the expected address
	if recoveredSigner != expectedSigner {
		return fmt.Errorf("signature does not match the expected signer: got %s, expected %s", recoveredSigner.Hex(), expectedSigner.Hex())
	}

	return nil
}
