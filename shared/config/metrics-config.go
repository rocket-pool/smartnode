package config

import (
	"github.com/rocket-pool/node-manager-core/config"
)

// Generates a new metrics config
func NewMetricsConfig() *config.MetricsConfig {
	cfg := config.NewMetricsConfig()
	cfg.BitflyNodeMetrics.MachineName.Default = map[config.Network]string{
		config.Network_All: "Smart Node",
	}
	return cfg
}
