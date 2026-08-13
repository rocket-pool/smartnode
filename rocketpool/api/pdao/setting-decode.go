package pdao

import (
	"fmt"
	"strings"

	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/types"
	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type decodedBatchSettings struct {
	contractNames []string
	settingPaths  []string
	settingTypes  []types.ProposalSettingType
	values        []any
	message       string
}

func decodeBatchSettings(entries []api.PDAOBatchSetting, customMessage string) (*decodedBatchSettings, error) {
	if len(entries) == 0 {
		return nil, protocol.ErrEmptyBatchSettings
	}

	decoded := &decodedBatchSettings{
		contractNames: make([]string, 0, len(entries)),
		settingPaths:  make([]string, 0, len(entries)),
		settingTypes:  make([]types.ProposalSettingType, 0, len(entries)),
		values:        make([]any, 0, len(entries)),
	}

	seen := make(map[string]struct{}, len(entries))
	for i, entry := range entries {
		if entry.Contract == "" || entry.Setting == "" {
			return nil, fmt.Errorf("setting %d is missing contract or setting path", i)
		}

		key := entry.Contract + "\x00" + entry.Setting
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("%w: %s / %s", protocol.ErrDuplicateBatchSetting, entry.Contract, entry.Setting)
		}
		seen[key] = struct{}{}

		settingType, err := protocol.GetProposalSettingType(entry.Contract, entry.Setting)
		if err != nil {
			return nil, err
		}
		if entry.Type != "" {
			namedType, err := protocol.ParseProposalSettingTypeName(entry.Type)
			if err != nil {
				return nil, fmt.Errorf("setting %s: %w", entry.Setting, err)
			}
			if namedType != settingType {
				return nil, fmt.Errorf("setting %s has type %s but JSON claims %s", entry.Setting, protocol.ProposalSettingTypeName(settingType), entry.Type)
			}
		}

		value, err := decodeSettingValue(settingType, entry.Value)
		if err != nil {
			return nil, fmt.Errorf("setting %s: %w", entry.Setting, err)
		}

		decoded.contractNames = append(decoded.contractNames, entry.Contract)
		decoded.settingPaths = append(decoded.settingPaths, entry.Setting)
		decoded.settingTypes = append(decoded.settingTypes, settingType)
		decoded.values = append(decoded.values, value)
	}

	if customMessage != "" {
		decoded.message = customMessage
	} else {
		decoded.message = "set " + strings.Join(decoded.settingPaths, ", ")
	}
	return decoded, nil
}

func decodeSettingValue(settingType types.ProposalSettingType, value string) (any, error) {
	switch settingType {
	case types.ProposalSettingType_Bool:
		return cliutils.ValidateBool("value", value)
	case types.ProposalSettingType_Uint256:
		return cliutils.ValidateBigInt("value", value)
	case types.ProposalSettingType_Address:
		address, err := cliutils.ValidateAddress("value", value)
		if err != nil {
			return nil, err
		}
		return address, nil
	default:
		return nil, fmt.Errorf("unsupported setting type %v", settingType)
	}
}
