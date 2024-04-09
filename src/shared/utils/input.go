package utils

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// Config
const (
	MinDaoMemberIDLength int = 3
	MinPasswordLength    int = 12
)

//
// General types
//

// Validate command argument count
func ValidateArgCount(c *cli.Context, expectedCount int) error {
	return input.ValidateArgCount(c.Args().Len(), expectedCount)
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
