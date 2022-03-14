package rp

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/alessio/shellescape"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"gopkg.in/yaml.v2"
)

const upgradeFlagFile string = ".firstrun"

// Loads a config, handling setting upgrades if necessary
func LoadAndUpgradeConfigFromFile(path string) (*config.RocketPoolConfig, error) {
	configDir := filepath.Dir(path)
	upgradeFilePath := filepath.Join(configDir, upgradeFlagFile)

	// Load the config normally if the upgrade flag file isn't there
	_, err := os.Stat(upgradeFilePath)
	if os.IsNotExist(err) {
		return config.LoadFromFile(path, false)
	} else {
		// Upgrade the config
		cfg, err := config.LoadFromFile(path, true)
		if err != nil {
			return nil, err
		}

		return cfg, nil
	}
}

// Saves a config, removing the upgrade flag file if present.
func SaveConfig(cfg *config.RocketPoolConfig, path string) error {

	settings := cfg.Serialize()
	configBytes, err := yaml.Marshal(settings)
	if err != nil {
		return fmt.Errorf("could not serialize settings file: %w", err)
	}

	if err := ioutil.WriteFile(path, configBytes, 0664); err != nil {
		return fmt.Errorf("could not write Rocket Pool config to %s: %w", shellescape.Quote(path), err)
	}

	// Check for the upgrade flag file
	configDir := filepath.Dir(path)
	upgradeFilePath := filepath.Join(configDir, upgradeFlagFile)
	_, err = os.Stat(upgradeFilePath)
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
