package rocketpool

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

// Decode, decompress and parse a zlib-compressed, base64-encoded ABI
func DecodeAbi(abiEncoded string) (*abi.ABI, error) {

	// base64 decode
	abiCompressed, err := base64.StdEncoding.DecodeString(abiEncoded)
	if err != nil {
		return nil, fmt.Errorf("Could not decode base64 data: %w", err)
	}

	// zlib decompress
	byteReader := bytes.NewReader(abiCompressed)
	zlibReader, err := zlib.NewReader(byteReader)
	if err != nil {
		return nil, fmt.Errorf("Could not decompress zlib data: %w", err)
	}
	defer func() {
		_ = zlibReader.Close()
	}()

	// Parse ABI
	abiParsed, err := abi.JSON(zlibReader)
	if err != nil {
		return nil, fmt.Errorf("Could not parse JSON: %w", err)
	}

	// Return
	return &abiParsed, nil

}

// zlib-compress and base64-encode an ABI JSON string
func EncodeAbiStr(abiStr string) (string, error) {

	// zlib compress
	var abiCompressed bytes.Buffer
	zlibWriter := zlib.NewWriter(&abiCompressed)
	if _, err := zlibWriter.Write([]byte(abiStr)); err != nil {
		return "", fmt.Errorf("Could not zlib-compress ABI string: %w", err)
	}
	if err := zlibWriter.Flush(); err != nil {
		return "", fmt.Errorf("Could not zlib-compress ABI string: %w", err)
	}
	if err := zlibWriter.Close(); err != nil {
		return "", fmt.Errorf("Could not zlib-compress ABI string: %w", err)
	}

	// base64 encode & return
	return base64.StdEncoding.EncodeToString(abiCompressed.Bytes()), nil

}
