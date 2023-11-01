package client

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/a8m/envsubst"
	"github.com/alessio/shellescape"
	"github.com/mitchellh/go-homedir"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
)

// Load the config
func (c *Client) LoadConfig() (*config.RocketPoolConfig, bool, error) {
	settingsFilePath := filepath.Join(c.configPath, SettingsFile)
	expandedPath, err := homedir.Expand(settingsFilePath)
	if err != nil {
		return nil, false, fmt.Errorf("error expanding settings file path: %w", err)
	}

	cfg, err := rp.LoadConfigFromFile(expandedPath)
	if err != nil {
		return nil, false, err
	}

	isNew := false
	if cfg == nil {
		cfg = config.NewRocketPoolConfig(c.configPath, c.daemonPath != "")
		isNew = true
	}
	return cfg, isNew, nil
}

// Load the backup config
func (c *Client) LoadBackupConfig() (*config.RocketPoolConfig, error) {
	settingsFilePath := filepath.Join(c.configPath, BackupSettingsFile)
	expandedPath, err := homedir.Expand(settingsFilePath)
	if err != nil {
		return nil, fmt.Errorf("error expanding backup settings file path: %w", err)
	}

	return rp.LoadConfigFromFile(expandedPath)
}

// Save the config
func (c *Client) SaveConfig(cfg *config.RocketPoolConfig) error {
	settingsFilePath := filepath.Join(c.configPath, SettingsFile)
	expandedPath, err := homedir.Expand(settingsFilePath)
	if err != nil {
		return err
	}
	return rp.SaveConfig(cfg, expandedPath)
}

// Remove the upgrade flag file
func (c *Client) RemoveUpgradeFlagFile() error {
	expandedPath, err := homedir.Expand(c.configPath)
	if err != nil {
		return err
	}
	return rp.RemoveUpgradeFlagFile(expandedPath)
}

// Returns whether or not this is the first run of the configurator since a previous installation
func (c *Client) IsFirstRun() (bool, error) {
	expandedPath, err := homedir.Expand(c.configPath)
	if err != nil {
		return false, fmt.Errorf("error expanding settings file path: %w", err)
	}
	return rp.IsFirstRun(expandedPath), nil
}

// Load the Prometheus template, do an environment variable substitution, and save it
func (c *Client) UpdatePrometheusConfiguration(settings map[string]string) error {
	prometheusTemplatePath, err := homedir.Expand(fmt.Sprintf("%s/%s", c.configPath, PrometheusConfigTemplate))
	if err != nil {
		return fmt.Errorf("Error expanding Prometheus template path: %w", err)
	}

	prometheusConfigPath, err := homedir.Expand(fmt.Sprintf("%s/%s", c.configPath, PrometheusFile))
	if err != nil {
		return fmt.Errorf("Error expanding Prometheus config file path: %w", err)
	}

	// Set the environment variables defined in the user settings for metrics
	oldValues := map[string]string{}
	for varName, varValue := range settings {
		oldValues[varName] = os.Getenv(varName)
		os.Setenv(varName, varValue)
	}

	// Read and substitute the template
	contents, err := envsubst.ReadFile(prometheusTemplatePath)
	if err != nil {
		return fmt.Errorf("Error reading and substituting Prometheus configuration template: %w", err)
	}

	// Unset the env vars
	for name, value := range oldValues {
		os.Setenv(name, value)
	}

	// Write the actual Prometheus config file
	err = os.WriteFile(prometheusConfigPath, contents, 0664)
	if err != nil {
		return fmt.Errorf("Could not write Prometheus config file to %s: %w", shellescape.Quote(prometheusConfigPath), err)
	}
	err = os.Chmod(prometheusConfigPath, 0664)
	if err != nil {
		return fmt.Errorf("Could not set Prometheus config file permissions: %w", shellescape.Quote(prometheusConfigPath), err)
	}

	return nil
}
