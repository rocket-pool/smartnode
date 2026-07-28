package eip712

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
)

type Components struct {
	V uint8    `json:"v"`
	R [32]byte `json:"r"`
	S [32]byte `json:"s"`
}

func (c *Components) UnmarshalText(text []byte) error {
	signature := string(text)

	if c == nil {
		return fmt.Errorf("Components is nil")
	}

	if len(signature) != 132 || signature[:2] != "0x" {
		return fmt.Errorf("Invalid 129 byte 0x-prefixed EIP-712 signature while parsing: '%s'", signature)
	}
	signature = signature[2:]
	if !regexp.MustCompile("^[A-Fa-f0-9]+$").MatchString(signature) {
		return fmt.Errorf("Invalid 129 byte 0x-prefixed EIP-712 signature while parsing: '%s'", signature)
	}

	// Slice signature string into v, r, s component of a signature giving node permission to use the given signer
	str_v := signature[len(signature)-2:]
	str_r := signature[:64]
	str_s := signature[64:128]

	// Convert v to uint8 and v,s to [32]byte
	bytes_r, err := hex.DecodeString(str_r)
	if err != nil {
		return fmt.Errorf("error decoding r: %v", err)
	}
	bytes_s, err := hex.DecodeString(str_s)
	if err != nil {
		return fmt.Errorf("error decoding s: %v", err)
	}

	int_v, err := strconv.ParseUint(str_v, 16, 8)
	if err != nil {
		return fmt.Errorf("error parsing v: %v", err)
	}

	c.V = uint8(int_v)
	c.R = ([32]byte)(bytes_r)
	c.S = ([32]byte)(bytes_s)

	return nil
}

func (c *Components) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("0x%x%x%x", c.R, c.S, c.V)), nil
}
