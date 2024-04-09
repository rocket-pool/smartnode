package config

import (
	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/smartnode/v2/shared/config/ids"
)

type MetricsConfig struct {
	*config.MetricsConfig
	EnableOdaoMetrics config.Parameter[bool]
}

// Generates a new metrics config
func NewMetricsConfig() *MetricsConfig {
	cfg := config.NewMetricsConfig()
	cfg.BitflyNodeMetrics.MachineName.Default = map[config.Network]string{
		config.Network_All: "Smart Node",
	}
	cfg.EnableMetrics.AffectsContainers = append(cfg.EnableMetrics.AffectsContainers, ContainerID_Alertmanager)

	return &MetricsConfig{
		MetricsConfig: cfg,
		EnableOdaoMetrics: config.Parameter[bool]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.MetricsEnableOdaoID,
				Name:               "Enable Oracle DAO Metrics",
				Description:        "Enable the tracking of Oracle DAO performance metrics, such as prices and balances submission participation.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_Daemon},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]bool{
				config.Network_All: false,
			},
		},
	}
}

// The title for the config
func (cfg *MetricsConfig) GetTitle() string {
	return cfg.MetricsConfig.GetTitle()
}

// Get the parameters for this config
func (cfg *MetricsConfig) GetParameters() []config.IParameter {
	params := cfg.MetricsConfig.GetParameters()
	params = append(params,
		&cfg.EnableOdaoMetrics,
	)
	return params
}

// Get the sections underneath this one
func (cfg *MetricsConfig) GetSubconfigs() map[string]config.IConfigSection {
	return cfg.MetricsConfig.GetSubconfigs()
}
