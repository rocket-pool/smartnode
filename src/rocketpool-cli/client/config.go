package client

import (
	"fmt"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client/template"
	"github.com/rocket-pool/smartnode/v2/shared/config"
)

const (
	prometheusConfigTemplate    string = "prometheus-cfg.tmpl"
	prometheusConfigTarget      string = "prometheus.yml"
	grafanaConfigTemplate       string = "grafana-prometheus-datasource.tmpl"
	grafanaConfigTarget         string = "grafana-prometheus-datasource.yml"
	alertmanagerConfigTemplate  string = "alerting/alertmanager.tmpl"
	alertmanagerConfigFile      string = "alerting/alertmanager.yml"
	alertingRulesConfigTemplate string = "alerting/rules/default.metatmpl"
	alertingRulesConfigFile     string = "alerting/rules/default.yml"
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
func (c *Client) UpdatePrometheusConfiguration(cfg *config.SmartNodeConfig) error {
	t, err := c.createTemplateBinding(prometheusConfigTemplate, prometheusConfigTarget, "Prometheus config")
	if err != nil {
		return err
	}
	return t.Write(cfg)
}

// Load the Grafana config template, do a template variable substitution, and save it
func (c *Client) UpdateGrafanaDatabaseConfiguration(cfg *config.SmartNodeConfig) error {
	t, err := c.createTemplateBinding(grafanaConfigTemplate, grafanaConfigTarget, "Grafana config")
	if err != nil {
		return err
	}
	return t.Write(cfg)
}

// Load the alerting configuration templates, do the template variable substitutions, and save them.
func (c *Client) UpdateAlertmanagerConfiguration(cfg *config.SmartNodeConfig) error {
	// Config
	t, err := c.createTemplateBinding(alertmanagerConfigTemplate, alertmanagerConfigFile, "alertmanager config")
	if err != nil {
		return err
	}
	err = t.WriteWithDelims(cfg, "{{", "}}")
	if err != nil {
		return err
	}

	// Rules
	t, err = c.createTemplateBinding(alertingRulesConfigTemplate, alertingRulesConfigFile, "alerting rules")
	if err != nil {
		return err
	}
	err = t.WriteWithDelims(cfg, "{{{", "}}}")
	if err != nil {
		return err
	}

	return nil
}

// Create the binding for a template file to be converted in the templating engine
func (c *Client) createTemplateBinding(templateFile string, targetFile string, description string) (template.Template, error) {
	templatePath, err := homedir.Expand(filepath.Join(templatesDir, templateFile))
	if err != nil {
		return template.Template{}, fmt.Errorf("error expanding %s template path: %w", description, err)
	}

	targetPath, err := homedir.Expand(filepath.Join(c.Context.ConfigPath, targetFile))
	if err != nil {
		return template.Template{}, fmt.Errorf("error expanding %s target path: %w", description, err)
	}

	t := template.Template{
		Src: templatePath,
		Dst: targetPath,
	}
	return t, nil
}
