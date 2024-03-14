package client

import (
	"fmt"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/rocket-pool/smartnode/rocketpool-cli/client/template"
	"github.com/rocket-pool/smartnode/shared/config"
)

const (
	prometheusConfigTemplate string = "prometheus-cfg.tmpl"
	prometheusConfigTarget   string = "prometheus.yml"
	grafanaConfigTemplate    string = "grafana-prometheus-datasource.tmpl"
	grafanaConfigTarget      string = "grafana-prometheus-datasource.yml"
)

// Load the config
// Returns the RocketPoolConfig and whether or not it was newly generated
func (c *Client) LoadConfig() (*config.SmartNodeConfig, bool, error) {
	if c.cfg != nil {
		return c.cfg, c.isNewCfg, nil
	}

	settingsFilePath := filepath.Join(c.Context.ConfigPath, SettingsFile)
	expandedPath, err := homedir.Expand(settingsFilePath)
	if err != nil {
		return nil, false, fmt.Errorf("error expanding settings file path: %w", err)
	}

	cfg, err := LoadConfigFromFile(expandedPath)
	if err != nil {
		return nil, false, err
	}

	if cfg != nil {
		// A config was loaded, return it now
		c.cfg = cfg
		return cfg, false, nil
	}

	// Config wasn't loaded, but there was no error- we should create one.
	c.cfg = config.NewSmartNodeConfig(c.Context.ConfigPath, c.Context.NativeMode)
	c.isNewCfg = true
	return c.cfg, true, nil
}

// Load the backup config
func (c *Client) LoadBackupConfig() (*config.SmartNodeConfig, error) {
	settingsFilePath := filepath.Join(c.Context.ConfigPath, BackupSettingsFile)
	expandedPath, err := homedir.Expand(settingsFilePath)
	if err != nil {
		return nil, fmt.Errorf("error expanding backup settings file path: %w", err)
	}

	return LoadConfigFromFile(expandedPath)
}

// Save the config
func (c *Client) SaveConfig(cfg *config.SmartNodeConfig) error {
	settingsFileDirectoryPath, err := homedir.Expand(c.Context.ConfigPath)
	if err != nil {
		return err
	}
	return SaveConfig(cfg, settingsFileDirectoryPath, SettingsFile)
}

// Load the Prometheus template, do a template variable substitution, and save it
func (c *Client) UpdatePrometheusConfiguration(config *config.SmartNodeConfig) error {
	prometheusConfigTemplatePath, err := homedir.Expand(filepath.Join(templatesDir, prometheusConfigTemplate))
	if err != nil {
		return fmt.Errorf("error expanding Prometheus config template path: %w", err)
	}

	prometheusConfigTargetPath, err := homedir.Expand(filepath.Join(c.Context.ConfigPath, prometheusConfigTarget))
	if err != nil {
		return fmt.Errorf("error expanding Prometheus config target path: %w", err)
	}

	t := template.Template{
		Src: prometheusConfigTemplatePath,
		Dst: prometheusConfigTargetPath,
	}

	return t.Write(config)
}
