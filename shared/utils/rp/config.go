package rp

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alessio/shellescape"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"gopkg.in/yaml.v2"
)

const (
	upgradeFlagFile string = ".firstrun"
)

// Loads a config without updating it if it exists
func LoadConfigFromFile(path string) (*config.RocketPoolConfig, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, nil
	}

	cfg, err := config.LoadFromFile(path)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Saves a config and removes the upgrade flag file
func SaveConfig(cfg *config.RocketPoolConfig, path string) error {

	settings := cfg.Serialize()
	configBytes, err := yaml.Marshal(settings)
	if err != nil {
		return fmt.Errorf("could not serialize settings file: %w", err)
	}

	if err := os.WriteFile(path, configBytes, 0664); err != nil {
		return fmt.Errorf("could not write Rocket Pool config to %s: %w", shellescape.Quote(path), err)
	}

	return nil

}

// Checks if this is the first run of the configurator after an install
func IsFirstRun(configDir string) bool {
	upgradeFilePath := filepath.Join(configDir, upgradeFlagFile)

	// Load the config normally if the upgrade flag file isn't there
	_, err := os.Stat(upgradeFilePath)
	if os.IsNotExist(err) {
		return false
	}

	return true
}

// Remove the upgrade flag file
func RemoveUpgradeFlagFile(configDir string) error {

	// Check for the upgrade flag file
	upgradeFilePath := filepath.Join(configDir, upgradeFlagFile)
	_, err := os.Stat(upgradeFilePath)
	if os.IsNotExist(err) {
		return nil
	}

	// Delete the upgrade flag file
	err = os.Remove(upgradeFilePath)
	if err != nil {
		return fmt.Errorf("error removing upgrade flag file: %w", err)
	}

	return nil

}
