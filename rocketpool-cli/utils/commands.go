package utils

import (
	"fmt"
	"math/big"
	"time"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/input"
	"github.com/urfave/cli/v2"
)

const (
	boolUsage             string = "specify 'true', 'false', 'yes', or 'no'"
	floatEthUsage         string = "specify an amount of ETH (e.g., '16.0')"
	floatRplUsage         string = "specify an amount of RPL (e.g., '16.0')"
	blockCountUsage       string = "specify a number, in blocks (e.g., '40000')"
	percentUsage          string = "specify a percentage between 0 and 1 (e.g., '0.51' for 51%)"
	unboundedPercentUsage string = "specify a percentage that can go over 100% (e.g., '1.5' for 150%)"
	uintUsage             string = "specify an integer (e.g., '50')"
	durationUsage         string = "specify a duration using hours, minutes, and seconds (e.g., '20m' or '72h0m0s')"
)

// Valid types for setting values
type SettingType interface {
	bool | *big.Int | uint64 | time.Duration
}

// Creates a command that acts as a category container for individual setter subcommands
func CreateSetterCategory(name string, formalName string, alias string, contract rocketpool.ContractName) *cli.Command {
	return &cli.Command{
		Name:    name,
		Aliases: []string{alias},
		Usage:   fmt.Sprintf("%s settings (under %s)", formalName, contract),
	}
}

// Creates a command to propose a new setting that takes a percentage value
func CreateBoolSetter[SettingNameType any](
	name string,
	alias string,
	contractName rocketpool.ContractName,
	settingName SettingNameType,
	runner func(c *cli.Context, contractName rocketpool.ContractName, settingName SettingNameType, value bool) error,
) *cli.Command {
	return createSettingCommandStub(name, alias, contractName, settingName, boolUsage, false,
		func(c *cli.Context) (bool, error) {
			return input.ValidateBool("value", c.Args().Get(0))
		},
		runner,
	)
}

// Creates a command to propose a new setting that takes an ETH value
func CreateEthSetter[SettingNameType any](
	name string,
	alias string,
	contractName rocketpool.ContractName,
	settingName SettingNameType,
	runner func(c *cli.Context, contractName rocketpool.ContractName, settingName SettingNameType, value *big.Int) error,
) *cli.Command {
	return createSettingCommandStub(name, alias, contractName, settingName, floatEthUsage, true,
		func(c *cli.Context) (*big.Int, error) {
			return ParseFloat(c, "value", c.Args().Get(0), false)
		},
		runner,
	)
}

// Creates a command to propose a new setting that takes an RPL value
func CreateRplSetter[SettingNameType any](
	name string,
	alias string,
	contractName rocketpool.ContractName,
	settingName SettingNameType,
	runner func(c *cli.Context, contractName rocketpool.ContractName, settingName SettingNameType, value *big.Int) error,
) *cli.Command {
	return createSettingCommandStub(name, alias, contractName, settingName, floatRplUsage, true,
		func(c *cli.Context) (*big.Int, error) {
			return ParseFloat(c, "value", c.Args().Get(0), false)
		},
		runner,
	)
}

// Creates a command to propose a new setting that takes an integer value (number of blocks)
func CreateBlockCountSetter[SettingNameType any](
	name string,
	alias string,
	contractName rocketpool.ContractName,
	settingName SettingNameType,
	runner func(c *cli.Context, contractName rocketpool.ContractName, settingName SettingNameType, value uint64) error,
) *cli.Command {
	return createSettingCommandStub(name, alias, contractName, settingName, boolUsage, false,
		func(c *cli.Context) (uint64, error) {
			return input.ValidatePositiveUint("value", c.Args().Get(0))
		},
		runner,
	)
}

// Creates a command to propose a new setting that takes a percentage value
func CreatePercentSetter[SettingNameType any](
	name string,
	alias string,
	contractName rocketpool.ContractName,
	settingName SettingNameType,
	runner func(c *cli.Context, contractName rocketpool.ContractName, settingName SettingNameType, value *big.Int) error,
) *cli.Command {
	return createSettingCommandStub(name, alias, contractName, settingName, percentUsage, true,
		func(c *cli.Context) (*big.Int, error) {
			return ParseFloat(c, "value", c.Args().Get(0), true)
		},
		runner,
	)
}

// Creates a command to propose a new setting that takes a percentage value but can go over 100%
func CreateUnboundedPercentSetter[SettingNameType any](
	name string,
	alias string,
	contractName rocketpool.ContractName,
	settingName SettingNameType,
	runner func(c *cli.Context, contractName rocketpool.ContractName, settingName SettingNameType, value *big.Int) error,
) *cli.Command {
	return createSettingCommandStub(name, alias, contractName, settingName, unboundedPercentUsage, true,
		func(c *cli.Context) (*big.Int, error) {
			return ParseFloat(c, "value", c.Args().Get(0), false)
		},
		runner,
	)
}

// Creates a command to propose a new setting that takes an unsigned integer value
func CreateUintSetter[SettingNameType any](
	name string,
	alias string,
	contractName rocketpool.ContractName,
	settingName SettingNameType,
	runner func(c *cli.Context, contractName rocketpool.ContractName, settingName SettingNameType, value uint64) error,
) *cli.Command {
	return createSettingCommandStub(name, alias, contractName, settingName, uintUsage, false,
		func(c *cli.Context) (uint64, error) {
			return input.ValidatePositiveUint("value", c.Args().Get(0))
		},
		runner,
	)
}

// Creates a command to propose a new setting that takes a time.Duration value
func CreateDurationSetter[SettingNameType any](
	name string,
	alias string,
	contractName rocketpool.ContractName,
	settingName SettingNameType,
	runner func(c *cli.Context, contractName rocketpool.ContractName, settingName SettingNameType, value time.Duration) error,
) *cli.Command {

	return createSettingCommandStub(name, alias, contractName, settingName, durationUsage, false,
		func(c *cli.Context) (time.Duration, error) {
			return input.ValidateDuration("value", c.Args().Get(0))
		},
		runner,
	)
}

// Internal body for the bulk of setting command creation
func createSettingCommandStub[ValueType SettingType, SettingNameType any](
	name string,
	alias string,
	contractName rocketpool.ContractName,
	settingName SettingNameType,
	usage string,
	includeRawFlag bool,
	parser func(c *cli.Context) (ValueType, error),
	runner func(c *cli.Context, contractName rocketpool.ContractName, settingName SettingNameType, value ValueType) error,
) *cli.Command {
	// Set up the flags
	flags := []cli.Flag{
		YesFlag,
	}
	if includeRawFlag {
		flags = append(flags, RawFlag)
	}

	// Create the command
	return &cli.Command{
		Name:      name,
		Aliases:   []string{alias},
		Usage:     fmt.Sprintf("Propose updating the %s setting; %s", settingName, usage),
		ArgsUsage: "value",
		Flags:     flags,
		Action: func(c *cli.Context) error {
			// Validate args
			if err := input.ValidateArgCount(c, 1); err != nil {
				return err
			}

			// Parse the value
			value, err := parser(c)
			if err != nil {
				return err
			}

			// Run the command body
			return runner(c, contractName, settingName, value)
		},
	}
}
