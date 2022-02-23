package rocketpool

import (
	"fmt"
	"io/ioutil"

	"github.com/alessio/shellescape"
	"github.com/mitchellh/go-homedir"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// Config
const (
	LegacyGlobalConfigFile    = "config.yml"
	LegacyUserConfigFile      = "settings.yml"
	LegacyComposeFile         = "docker-compose.yml"
	LegacyMetricsComposeFile  = "docker-compose-metrics.yml"
	LegacyFallbackComposeFile = "docker-compose-fallback.yml"
)

// Load the global config
func (c *Client) LoadGlobalConfig_Legacy() (config.LegacyRocketPoolConfig, error) {
	return c.loadConfig_Legacy(fmt.Sprintf("%s/%s", c.configPath, LegacyGlobalConfigFile))
}

// Load/save the user config
func (c *Client) LoadUserConfig_Legacy() (config.LegacyRocketPoolConfig, error) {
	return c.loadConfig_Legacy(fmt.Sprintf("%s/%s", c.configPath, LegacyUserConfigFile))
}
func (c *Client) SaveUserConfig_Legacy(cfg config.LegacyRocketPoolConfig) error {
	return c.saveConfig_Legacy(cfg, fmt.Sprintf("%s/%s", c.configPath, LegacyUserConfigFile))
}

// Load the merged global & user config
func (c *Client) LoadMergedConfig_Legacy() (config.LegacyRocketPoolConfig, error) {
	globalConfig, err := c.LoadGlobalConfig_Legacy()
	if err != nil {
		return config.LegacyRocketPoolConfig{}, err
	}
	userConfig, err := c.LoadUserConfig_Legacy()
	if err != nil {
		return config.LegacyRocketPoolConfig{}, err
	}
	return config.Merge(&globalConfig, &userConfig)
}

// Load a config file
func (c *Client) loadConfig_Legacy(path string) (config.LegacyRocketPoolConfig, error) {
	expandedPath, err := homedir.Expand(path)
	if err != nil {
		return config.LegacyRocketPoolConfig{}, err
	}
	configBytes, err := ioutil.ReadFile(expandedPath)
	if err != nil {
		return config.LegacyRocketPoolConfig{}, fmt.Errorf("Could not read Rocket Pool config at %s: %w", shellescape.Quote(path), err)
	}
	return config.Parse(configBytes)
}

// Save a config file
func (c *Client) saveConfig_Legacy(cfg config.LegacyRocketPoolConfig, path string) error {
	configBytes, err := cfg.Serialize()
	if err != nil {
		return err
	}
	expandedPath, err := homedir.Expand(path)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(expandedPath, configBytes, 0); err != nil {
		return fmt.Errorf("Could not write Rocket Pool config to %s: %w", shellescape.Quote(expandedPath), err)
	}
	return nil
}
