package client

import (
	"fmt"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client/template"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
)

// Load the config
// Returns the RocketPoolConfig and whether or not it was newly generated
func (c *Client) LoadConfig() (*config.RocketPoolConfig, bool, error) {
	settingsFilePath := filepath.Join(c.Context.ConfigPath, SettingsFile)
	expandedPath, err := homedir.Expand(settingsFilePath)
	if err != nil {
		return nil, false, fmt.Errorf("error expanding settings file path: %w", err)
	}

	cfg, err := rp.LoadConfigFromFile(expandedPath)
	if err != nil {
		return nil, false, err
	}

	if cfg != nil {
		// A config was loaded, return it now
		return cfg, false, nil
	}

	// Config wasn't loaded, but there was no error- we should create one.
	return config.NewRocketPoolConfig(c.Context.ConfigPath, c.Context.ApiSocketPath != ""), true, nil
}

// Load the backup config
func (c *Client) LoadBackupConfig() (*config.RocketPoolConfig, error) {
	settingsFilePath := filepath.Join(c.Context.ConfigPath, BackupSettingsFile)
	expandedPath, err := homedir.Expand(settingsFilePath)
	if err != nil {
		return nil, fmt.Errorf("error expanding backup settings file path: %w", err)
	}

	return rp.LoadConfigFromFile(expandedPath)
}

// Save the config
func (c *Client) SaveConfig(cfg *config.RocketPoolConfig) error {
	settingsFileDirectoryPath, err := homedir.Expand(c.Context.ConfigPath)
	if err != nil {
		return err
	}
	return rp.SaveConfig(cfg, settingsFileDirectoryPath, SettingsFile)
}

// Remove the upgrade flag file
func (c *Client) RemoveUpgradeFlagFile() error {
	expandedPath, err := homedir.Expand(c.Context.ConfigPath)
	if err != nil {
		return err
	}
	return rp.RemoveUpgradeFlagFile(expandedPath)
}

// Returns whether or not this is the first run of the configurator since a previous installation
func (c *Client) IsFirstRun() (bool, error) {
	expandedPath, err := homedir.Expand(c.Context.ConfigPath)
	if err != nil {
		return false, fmt.Errorf("error expanding settings file path: %w", err)
	}
	return rp.IsFirstRun(expandedPath), nil
}

// Load the Prometheus template, do a template variable substitution, and save it
func (c *Client) UpdatePrometheusConfiguration(config *config.RocketPoolConfig) error {
	prometheusTemplatePath, err := homedir.Expand(fmt.Sprintf("%s/%s", c.Context.ConfigPath, PrometheusConfigTemplate))
	if err != nil {
		return fmt.Errorf("Error expanding Prometheus template path: %w", err)
	}

	prometheusConfigPath, err := homedir.Expand(fmt.Sprintf("%s/%s", c.Context.ConfigPath, PrometheusFile))
	if err != nil {
		return fmt.Errorf("Error expanding Prometheus config file path: %w", err)
	}

	t := template.Template{
		Src: prometheusTemplatePath,
		Dst: prometheusConfigPath,
	}

	return t.Write(config)
}
