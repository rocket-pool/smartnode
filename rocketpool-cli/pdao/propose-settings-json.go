package pdao

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// writeSettingToBatchJSON creates or appends a setting change to a batch proposal JSON file.
// If the same contract+setting is already present, its value is replaced.
func writeSettingToBatchJSON(path string, contract string, setting string, value string) error {
	settingType, err := protocol.GetProposalSettingType(contract, setting)
	if err != nil {
		return err
	}

	entry := api.PDAOBatchSetting{
		Contract: contract,
		Setting:  setting,
		Type:     protocol.ProposalSettingTypeName(settingType),
		Value:    value,
	}

	settings, err := readBatchSettingsFileOptional(path, true)
	if err != nil {
		return err
	}

	replaced := false
	for i, existing := range settings {
		if existing.Contract == contract && existing.Setting == setting {
			settings[i] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		settings = append(settings, entry)
	}

	if err := writeBatchSettingsFile(path, settings); err != nil {
		return err
	}

	if replaced {
		fmt.Printf("Updated %s in %s (%d setting(s) total).\n", setting, path, len(settings))
	} else {
		fmt.Printf("Added %s to %s (%d setting(s) total).\n", setting, path, len(settings))
	}
	return nil
}

func readBatchSettingsFile(path string) ([]api.PDAOBatchSetting, error) {
	return readBatchSettingsFileOptional(path, false)
}

func readBatchSettingsFileOptional(path string, allowMissing bool) ([]api.PDAOBatchSetting, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if allowMissing {
				return nil, nil
			}
			return nil, fmt.Errorf("JSON file %s does not exist", path)
		}
		return nil, fmt.Errorf("could not read JSON file %s: %w", path, err)
	}
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil, nil
	}

	var settings []api.PDAOBatchSetting
	if err := json.Unmarshal([]byte(trimmed), &settings); err != nil {
		return nil, fmt.Errorf("could not parse JSON file %s: %w", path, err)
	}
	return settings, nil
}

func writeBatchSettingsFile(path string, settings []api.PDAOBatchSetting) error {
	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("could not encode JSON file %s: %w", path, err)
	}
	out = append(out, '\n')
	if err := os.WriteFile(path, out, 0644); err != nil {
		return fmt.Errorf("could not write JSON file %s: %w", path, err)
	}
	return nil
}
