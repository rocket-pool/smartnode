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
func (c *Client) LoadGlobalConfig_Legacy(globalConfigPath string) (config.LegacyRocketPoolConfig, error) {
	return c.loadConfig_Legacy(globalConfigPath)
}

// Load/save the user config
func (c *Client) LoadUserConfig_Legacy(userConfigPath string) (config.LegacyRocketPoolConfig, error) {
	return c.loadConfig_Legacy(fmt.Sprintf("%s/%s", c.configPath, LegacyUserConfigFile))
}

// Load the merged global & user config
func (c *Client) LoadMergedConfig_Legacy(globalConfigPath string, userConfigPath string) (config.LegacyRocketPoolConfig, error) {
	globalConfig, err := c.LoadGlobalConfig_Legacy(globalConfigPath)
	if err != nil {
		return config.LegacyRocketPoolConfig{}, err
	}
	userConfig, err := c.LoadUserConfig_Legacy(userConfigPath)
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
