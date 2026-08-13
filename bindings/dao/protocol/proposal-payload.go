package protocol

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/types"
	strutils "github.com/rocket-pool/smartnode/bindings/utils/strings"
)

const proposalSettingMultiMethod = "proposalSettingMulti"

type DecodedProposalSetting struct {
	Contract string                    `json:"contract"`
	Path     string                    `json:"path"`
	Type     types.ProposalSettingType `json:"type"`
	Value    string                    `json:"value"`
}

func decodeProposalSettingMultiArgs(args []any) ([]DecodedProposalSetting, error) {
	if len(args) != 4 {
		return nil, fmt.Errorf("proposalSettingMulti expected 4 arguments, got %d", len(args))
	}

	contracts, err := asStringSlice(args[0])
	if err != nil {
		return nil, fmt.Errorf("contract names: %w", err)
	}
	paths, err := asStringSlice(args[1])
	if err != nil {
		return nil, fmt.Errorf("setting paths: %w", err)
	}
	settingTypes, err := asSettingTypes(args[2])
	if err != nil {
		return nil, fmt.Errorf("setting types: %w", err)
	}
	values, err := asBytesSlice(args[3])
	if err != nil {
		return nil, fmt.Errorf("setting values: %w", err)
	}

	if len(contracts) != len(paths) || len(paths) != len(settingTypes) || len(settingTypes) != len(values) {
		return nil, fmt.Errorf("proposalSettingMulti argument lengths do not match")
	}

	settings := make([]DecodedProposalSetting, len(contracts))
	for i := range contracts {
		value, err := decodeMultiSettingValue(settingTypes[i], values[i])
		if err != nil {
			return nil, fmt.Errorf("setting %s: %w", paths[i], err)
		}
		settings[i] = DecodedProposalSetting{
			Contract: contracts[i],
			Path:     paths[i],
			Type:     settingTypes[i],
			Value:    value,
		}
	}
	return settings, nil
}

func FormatProposalSettingMulti(settings []DecodedProposalSetting) string {
	parts := make([]string, len(settings))
	for i, setting := range settings {
		parts[i] = fmt.Sprintf("%s=%s", setting.Path, setting.Value)
	}
	return strutils.Sanitize(fmt.Sprintf("%s(%s)", proposalSettingMultiMethod, strings.Join(parts, ", ")))
}

func formatGenericProposalPayload(method *abi.Method, args []any) string {
	argStrs := make([]string, 0, len(args))
	for ai, arg := range args {
		switch method.Inputs[ai].Type.T {
		case abi.AddressTy:
			argStrs = append(argStrs, arg.(common.Address).Hex())
		case abi.HashTy:
			argStrs = append(argStrs, arg.(common.Hash).Hex())
		case abi.FixedBytesTy:
			fallthrough
		case abi.BytesTy:
			argStrs = append(argStrs, fmt.Sprintf("%x", arg.([]byte)))
		default:
			argStrs = append(argStrs, fmt.Sprintf("%v", arg))
		}
	}
	return strutils.Sanitize(fmt.Sprintf("%s(%s)", method.RawName, strings.Join(argStrs, ",")))
}

func decodeMultiSettingValue(settingType types.ProposalSettingType, data []byte) (string, error) {
	switch settingType {
	case types.ProposalSettingType_Uint256:
		return new(big.Int).SetBytes(data).String(), nil
	case types.ProposalSettingType_Bool:
		return fmt.Sprint(new(big.Int).SetBytes(data).Sign() != 0), nil
	case types.ProposalSettingType_Address:
		if len(data) >= common.AddressLength {
			return common.BytesToAddress(data[len(data)-common.AddressLength:]).Hex(), nil
		}
		return common.BytesToAddress(data).Hex(), nil
	default:
		return "", fmt.Errorf("unknown setting type %d", settingType)
	}
}

func asStringSlice(value any) ([]string, error) {
	switch typed := value.(type) {
	case []string:
		return typed, nil
	default:
		return nil, fmt.Errorf("expected []string, got %T", value)
	}
}

func asSettingTypes(value any) ([]types.ProposalSettingType, error) {
	switch typed := value.(type) {
	case []uint8:
		out := make([]types.ProposalSettingType, len(typed))
		for i, item := range typed {
			out[i] = types.ProposalSettingType(item)
		}
		return out, nil
	case []*big.Int:
		out := make([]types.ProposalSettingType, len(typed))
		for i, item := range typed {
			if item == nil {
				return nil, fmt.Errorf("nil setting type at index %d", i)
			}
			out[i] = types.ProposalSettingType(item.Uint64())
		}
		return out, nil
	case []uint16:
		out := make([]types.ProposalSettingType, len(typed))
		for i, item := range typed {
			out[i] = types.ProposalSettingType(item)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("expected uint8 array, got %T", value)
	}
}

func asBytesSlice(value any) ([][]byte, error) {
	switch typed := value.(type) {
	case [][]byte:
		return typed, nil
	default:
		return nil, fmt.Errorf("expected [][]byte, got %T", value)
	}
}
