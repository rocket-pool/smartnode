package utils

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/cli/input"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// Config
const (
	MinDaoMemberIDLength int = 3
	MinPasswordLength    int = 12
)

type EIP712Components struct {
	V uint8    `json:"v"`
	R [32]byte `json:"r"`
	S [32]byte `json:"s"`
}

// Validate command argument count - only used by the CLI
// TODO: refactor CLI arg validation and move it out of shared
func ValidateArgCount(c *cli.Context, expectedCount int) {
	err := input.ValidateArgCount(c.Args().Len(), expectedCount)
	if err != nil {
		// Handle invalid arg count
		var argCountErr *input.InvalidArgCountError
		if errors.As(err, &argCountErr) {
			fmt.Fprintf(os.Stderr, "%s%s%s\n\n", terminal.ColorRed, err.Error(), terminal.ColorReset)
			cli.ShowSubcommandHelpAndExit(c, 1)
		}

		// Handle other errors
		fmt.Fprintf(os.Stderr, "%s%s%s\n\n", terminal.ColorRed, err.Error(), terminal.ColorReset)
		os.Exit(1)
	}
}

// Validate a token type
func ValidateTokenType(name, value string) (string, error) {
	// Check if this is a token address
	// This was taken from the Ethereum library: https://github.com/ethereum/go-ethereum/blob/master/common/types.go
	if strings.HasPrefix(value, "0x") {
		// Remove the 0x prefix
		val := value[2:]

		// Zero pad if it's an odd number of chars
		if len(val)%2 == 1 {
			val = "0" + val
		}

		// Attempt parsing
		_, err := hex.DecodeString(val)
		if err != nil {
			return "", fmt.Errorf("Invalid %s '%s' - could not parse address: %w", name, value, err)
		}

		// If it passes, return the original value
		return value, nil
	}

	// Not a token address, check against the well-known names
	val := strings.ToLower(value)
	if !(val == "eth" || val == "rpl" || val == "fsrpl" || val == "reth") {
		return "", fmt.Errorf("Invalid %s '%s' - valid types are 'ETH', 'RPL', 'fsRPL', and 'rETH'", name, value)
	}
	return val, nil
}

// Validate a proposal type
func ValidateProposalType(name, value string) (string, error) {
	val := strings.ToLower(value)
	if !(val == "pending" || val == "active" || val == "succeeded" || val == "executed" || val == "cancelled" || val == "defeated" || val == "expired" || val == "all") {
		return "", fmt.Errorf("Invalid %s '%s' - valid types are 'pending', 'active', 'succeeded', 'executed', 'cancelled', 'defeated', 'expired', and 'all'", name, value)
	}
	return val, nil
}

// Validate a burnable token type
func ValidateBurnableTokenType(name, value string) (string, error) {
	val := strings.ToLower(value)
	if !(val == "reth") {
		return "", fmt.Errorf("Invalid %s '%s' - valid types are 'rETH'", name, value)
	}
	return val, nil
}

// Validate a DAO member ID
func ValidateDaoMemberID(name, value string) (string, error) {
	val := strings.TrimSpace(value)
	if len(val) < MinDaoMemberIDLength {
		return "", fmt.Errorf("Invalid %s '%s' - must be at least %d characters long", name, val, MinDaoMemberIDLength)
	}
	return val, nil
}

// Validate a vote direction
func ValidateVoteDirection(name, value string) (types.VoteDirection, error) {
	val, exists := api.VoteDirectionMap[value]
	if !exists {
		return 0, fmt.Errorf("Invalid %s '%s': not a valid vote direction name", name, value)
	}
	return val, nil
}

// Expects a 129 byte 0x-prefixed EIP-712 signature and returns v/r/s as v uint8 and r, s [32]byte
func ParseEIP712(signature string) (*EIP712Components, error) {
	if len(signature) != 132 || signature[:2] != "0x" {
		return nil, fmt.Errorf("Invalid 129 byte 0x-prefixed EIP-712 signature while parsing: '%s'", signature)
	}
	signature = signature[2:]
	if !regexp.MustCompile("^[A-Fa-f0-9]+$").MatchString(signature) {
		return &EIP712Components{}, fmt.Errorf("Invalid 129 byte 0x-prefixed EIP-712 signature while parsing: '%s'", signature)
	}

	// Slice signature string into v, r, s component of a signature giving node permission to use the given signer
	str_v := signature[len(signature)-2:]
	str_r := signature[:64]
	str_s := signature[64:128]

	// Convert v to uint8 and v,s to [32]byte
	bytes_r, err := hex.DecodeString(str_r)
	if err != nil {
		return &EIP712Components{}, fmt.Errorf("error decoding r: %v", err)
	}
	bytes_s, err := hex.DecodeString(str_s)
	if err != nil {
		return &EIP712Components{}, fmt.Errorf("error decoding s: %v", err)
	}

	int_v, err := strconv.ParseUint(str_v, 16, 8)
	if err != nil {
		return &EIP712Components{}, fmt.Errorf("error parsing v: %v", err)
	}

	return &EIP712Components{uint8(int_v), ([32]byte)(bytes_r), ([32]byte)(bytes_s)}, nil
}
