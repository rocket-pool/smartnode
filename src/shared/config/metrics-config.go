package config

import (
	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/smartnode/shared/config/ids"
)

type MetricsConfig struct {
	base                  *config.MetricsConfig
	EnableOdaoMetrics     config.Parameter[bool]
	WatchtowerMetricsPort config.Parameter[uint16]
}

// Generates a new metrics config
func NewMetricsConfig() *MetricsConfig {
	cfg := config.NewMetricsConfig()
	cfg.BitflyNodeMetrics.MachineName.Default = map[config.Network]string{
		config.Network_All: "Smart Node",
	}

	return &MetricsConfig{
		base: cfg,
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

		WatchtowerMetricsPort: config.Parameter[uint16]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.MetricsWatchtowerPortID,
				Name:               "Daemon Metrics Port",
				Description:        "The port the Watchtower container should expose its metrics on.",
				AffectsContainers:  []config.ContainerID{ContainerID_Watchtower, config.ContainerID_Prometheus},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]uint16{
				config.Network_All: 9104,
			},
		},
	}
}

// The title for the config
func (cfg *MetricsConfig) GetTitle() string {
	return cfg.base.GetTitle()
}

// Get the parameters for this config
func (cfg *MetricsConfig) GetParameters() []config.IParameter {
	params := cfg.base.GetParameters()
	params = append(params,
		&cfg.EnableOdaoMetrics,
		&cfg.WatchtowerMetricsPort,
	)
	return params
}

// Get the sections underneath this one
func (cfg *MetricsConfig) GetSubconfigs() map[string]config.IConfigSection {
	return cfg.base.GetSubconfigs()
}
