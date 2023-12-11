package input

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/tyler-smith/go-bip39"
	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/types"
	hexutils "github.com/rocket-pool/smartnode/shared/utils/hex"
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
func ValidateArgCount(c *cli.Context, count int) error {
	if len(c.Args()) != count {
		return fmt.Errorf("Incorrect argument count; usage: %s", c.Command.UsageText)
	}
	return nil
}

// Validate a big int
func ValidateBigInt(name, value string) (*big.Int, error) {
	val, success := big.NewInt(0).SetString(value, 0)
	if !success {
		return nil, fmt.Errorf("Invalid %s '%s'", name, value)
	}
	return val, nil
}

// Validate an array of big ints
func ValidateBigInts(name, value string) ([]*big.Int, error) {
	elements := strings.Split(value, ",")
	ints := make([]*big.Int, len(elements))
	for i, element := range elements {
		val, success := big.NewInt(0).SetString(element, 0)
		if !success {
			return nil, fmt.Errorf("Invalid int %d in %s '%s'", i, name, value)
		}
		ints = append(ints, val)
	}
	return ints, nil
}

// Validate a boolean value
func ValidateBool(name, value string) (bool, error) {
	val := strings.ToLower(value)
	if !(val == "true" || val == "yes" || val == "false" || val == "no") {
		return false, fmt.Errorf("Invalid %s '%s' - valid values are 'true', 'yes', 'false' and 'no'", name, value)
	}
	if val == "true" || val == "yes" {
		return true, nil
	}
	return false, nil
}

// Validate an unsigned integer value
func ValidateUint(name, value string) (uint64, error) {
	val, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid %s '%s'", name, value)
	}
	return val, nil
}

// Validate an unsigned integer value
func ValidateUint32(name, value string) (uint32, error) {
	val, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("Invalid %s '%s'", name, value)
	}
	return uint32(val), nil
}

// Validate an address
func ValidateAddress(name, value string) (common.Address, error) {
	if !common.IsHexAddress(value) {
		return common.Address{}, fmt.Errorf("Invalid %s '%s'", name, value)
	}
	return common.HexToAddress(value), nil
}

// Validate a collection of addresses
func ValidateAddresses(name, value string) ([]common.Address, error) {
	elements := strings.Split(value, ",")
	addresses := make([]common.Address, len(elements))
	for i, element := range elements {
		if !common.IsHexAddress(element) {
			return nil, fmt.Errorf("Invalid address %d in %s: '%s'", i, name, element)
		}
		addresses[i] = common.HexToAddress(element)
	}
	return addresses, nil
}

// Validate a wei amount
func ValidateWeiAmount(name, value string) (*big.Int, error) {
	val := new(big.Int)
	if _, ok := val.SetString(value, 10); !ok {
		return nil, fmt.Errorf("Invalid %s '%s'", name, value)
	}
	return val, nil
}

// Validate an ether amount
func ValidateEthAmount(name, value string) (float64, error) {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid %s '%s'", name, value)
	}
	return val, nil
}

// Validate a fraction
func ValidateFraction(name, value string) (float64, error) {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil || val < 0 || val > 1 {
		return 0, fmt.Errorf("Invalid %s '%s' - must be a number between 0 and 1", name, value)
	}
	return val, nil
}

// Validate a percentage
func ValidatePercentage(name, value string) (float64, error) {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil || val < 0 || val > 100 {
		return 0, fmt.Errorf("Invalid %s '%s' - must be a number between 0 and 100", name, value)
	}
	return val, nil
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

//
// Command specific types
//

// Validate a positive unsigned integer value
func ValidatePositiveUint(name, value string) (uint64, error) {
	val, err := ValidateUint(name, value)
	if err != nil {
		return 0, err
	}
	if val == 0 {
		return 0, fmt.Errorf("Invalid %s '%s' - must be greater than 0", name, value)
	}
	return val, nil
}

// Validate a list of comma-separated positive unsigned integer values
func ValidatePositiveUints(name, value string) ([]uint64, error) {
	elements := strings.Split(value, ",")
	vals := []uint64{}
	for i, element := range elements {
		element = strings.TrimSpace(element)
		val, err := ValidateUint(name, element)
		if err != nil {
			return nil, fmt.Errorf("Invalid %s '%s' - element %d (%s) could not be parsed: %w", name, value, i, element, err)
		}
		if val == 0 {
			return nil, fmt.Errorf("Invalid %s '%s' - element %d (%s) must be greater than 0", name, value, i, element)
		}
		vals = append(vals, val)
	}
	return vals, nil
}

// Validate a positive 32-bit unsigned integer value
func ValidatePositiveUint32(name, value string) (uint32, error) {
	val, err := ValidateUint32(name, value)
	if err != nil {
		return 0, err
	}
	if val == 0 {
		return 0, fmt.Errorf("Invalid %s '%s' - must be greater than 0", name, value)
	}
	return val, nil
}

// Validate a positive wei amount
func ValidatePositiveWeiAmount(name, value string) (*big.Int, error) {
	val, err := ValidateWeiAmount(name, value)
	if err != nil {
		return nil, err
	}
	if val.Cmp(big.NewInt(0)) < 1 {
		return nil, fmt.Errorf("Invalid %s '%s' - must be greater than 0", name, value)
	}
	return val, nil
}

// Validate a positive or zero wei amount
func ValidatePositiveOrZeroWeiAmount(name, value string) (*big.Int, error) {
	val, err := ValidateWeiAmount(name, value)
	if err != nil {
		return nil, err
	}
	if val.Cmp(big.NewInt(0)) < 0 {
		return nil, fmt.Errorf("Invalid %s '%s' - must be greater or equal to 0", name, value)
	}
	return val, nil
}

// Validate a positive ether amount
func ValidatePositiveEthAmount(name, value string) (float64, error) {
	val, err := ValidateEthAmount(name, value)
	if err != nil {
		return 0, err
	}
	if val <= 0 {
		return 0, fmt.Errorf("Invalid %s '%s' - must be greater than 0", name, value)
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

// Validate a node password
func ValidateNodePassword(name string, value string) ([]byte, error) {
	if len(value) < MinPasswordLength {
		return nil, fmt.Errorf("Invalid %s '%s' - must be at least %d characters long", name, value, MinPasswordLength)
	}
	return []byte(value), nil
}

// Validate a wallet mnemonic phrase
func ValidateWalletMnemonic(name, value string) (string, error) {
	if !bip39.IsMnemonicValid(value) {
		return "", fmt.Errorf("Invalid %s '%s'", name, value)
	}
	return value, nil
}

// Validate a timezone location
func ValidateTimezoneLocation(name, value string) (string, error) {
	if !regexp.MustCompile("^([a-zA-Z_]{2,}\\/)+[a-zA-Z_]{2,}$").MatchString(value) {
		return "", fmt.Errorf("Invalid %s '%s' - must be in the format 'Country/City'", name, value)
	}
	return value, nil
}

// Validate a DAO member ID
func ValidateDAOMemberID(name, value string) (string, error) {
	val := strings.TrimSpace(value)
	if len(val) < MinDaoMemberIDLength {
		return "", fmt.Errorf("Invalid %s '%s' - must be at least %d characters long", name, val, MinDaoMemberIDLength)
	}
	return val, nil
}

// Validate a transaction hash
func ValidateTxHash(name, value string) (common.Hash, error) {
	// Remove a 0x prefix if present
	value = strings.TrimPrefix(value, "0x")

	// Hash should be 64 characters long
	if len(value) != hex.EncodedLen(common.HashLength) {
		return common.Hash{}, fmt.Errorf("Invalid %s '%s': it must have 64 characters.", name, value)
	}

	// Try to parse the string (removing the prefix)
	bytes, err := hex.DecodeString(value)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Invalid %s '%s': %w", name, value, err)
	}
	hash := common.BytesToHash(bytes)

	return hash, nil
}

// Validate TX info
func ValidateTxInfo(name string, value string) (*core.TransactionInfo, error) {
	// Remove a 0x prefix if present
	value = strings.TrimPrefix(value, "0x")

	// Try to parse the string
	bytes, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("Invalid %s '%s': %w", name, value, err)
	}

	// Deserialize it
	var info core.TransactionInfo
	err = json.Unmarshal(bytes, &info)
	if err != nil {
		return nil, fmt.Errorf("Deserializing %s failed: %w", name, err)
	}

	return &info, nil
}

// Validate a validator pubkey
func ValidatePubkey(name, value string) (types.ValidatorPubkey, error) {
	pubkey, err := types.HexToValidatorPubkey(hexutils.RemovePrefix(value))
	if err != nil {
		return types.ValidatorPubkey{}, fmt.Errorf("Invalid %s '%s': %w", name, value, err)
	}
	return pubkey, nil
}

// Validate a hex-encoded byte array
func ValidateByteArray(name, value string) ([]byte, error) {
	// Remove a 0x prefix if present
	if strings.HasPrefix(value, "0x") {
		value = value[2:]
	}

	// Try to parse the string (removing the prefix)
	bytes, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("Invalid %s '%s': %w", name, value, err)
	}

	return bytes, nil
}

// Validate a duration
func ValidateDuration(name, value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("Invalid %s '%s': %w", name, value, err)
	}
	return duration, nil
}

// Validate a vote direction
func ValidateVoteDirection(name, value string) (types.VoteDirection, error) {
	switch value {
	case "abstain":
		return types.VoteDirection_Abstain, nil
	case "for":
		return types.VoteDirection_For, nil
	case "against":
		return types.VoteDirection_Against, nil
	case "veto":
		return types.VoteDirection_AgainstWithVeto, nil
	}
	return 0, fmt.Errorf("Invalid %s '%s': not a valid vote direction name", name, value)
}

// Validate a timestamp using RFC3339
func ValidateTime(name, value string) (time.Time, error) {
	val, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("Invalid %s '%s': %w", name, value, err)
	}
	return val, nil
}
