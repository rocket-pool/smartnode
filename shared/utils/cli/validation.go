package cli

import (
    "fmt"
    "math/big"
    "regexp"
    "strconv"
    "strings"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/passwords"
)


// Check command argument count
func CheckArgCount(c *cli.Context, count int) error {
    if len(c.Args()) != count {
        return cli.NewExitError(fmt.Sprintf("USAGE:\n   %s", c.Command.UsageText), 1)
    }
    return nil
}


// Check API command argument count
func CheckAPIArgCount(c *cli.Context, count int) error {
    if len(c.Args()) != count {
        return fmt.Errorf("Incorrect argument count; usage: %s", c.Command.UsageText)
    }
    return nil
}


// Validate an address
func ValidateAddress(name, value string) (common.Address, error) {
    if !common.IsHexAddress(value) {
        return common.Address{}, fmt.Errorf("Invalid %s '%s'", name, value)
    }
    return common.HexToAddress(value), nil
}


// Validate a token amount
func ValidateTokenAmount(name, value string) (*big.Int, error) {
    val, err := strconv.ParseFloat(value, 64)
    if err != nil {
        return nil, fmt.Errorf("Invalid %s '%s'", name, value)
    }
    return eth.EthToWei(val), nil
}


// Validate a deposit amount
func ValidateDepositAmount(name, value string) (*big.Int, error) {
    val, err := strconv.ParseFloat(value, 64)
    if err != nil || !(val == 0 || val == 16 || val == 32) {
        return nil, fmt.Errorf("Invalid %s '%s' - valid values are 0, 16 and 32", name, value)
    }
    return eth.EthToWei(val), nil
}


// Validate a token type
func ValidateTokenType(name, value string) (string, error) {
    val := strings.ToLower(value)
    if !(val == "eth" || val == "neth") {
        return "", fmt.Errorf("Invalid %s '%s' - valid types are 'ETH' and 'nETH'", name, value)
    }
    return val, nil
}


// Validate a burnable token type
func ValidateBurnableType(name, value string) (string, error) {
    val := strings.ToLower(value)
    if !(val == "neth") {
        return "", fmt.Errorf("Invalid %s '%s' - valid types are 'nETH'", name, value)
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


// Validate a node password
func ValidatePassword(name, value string) (string, error) {
    if len(value) < passwords.MinPasswordLength {
        return "", fmt.Errorf("Invalid %s '%s' - must be at least %d characters long", name, value, passwords.MinPasswordLength)
    }
    return value, nil
}


// Validate a timezone location
func ValidateTimezoneLocation(name, value string) (string, error) {
    if !regexp.MustCompile("^\\w{2,}\\/\\w{2,}$").MatchString(value) {
        return "", fmt.Errorf("Invalid %s '%s' - must be in the format 'Country/City'", name, value)
    }
    return value, nil
}

