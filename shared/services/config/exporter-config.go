package config

import (
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const exporterTag string = "prom/node-exporter:v1.8.0"

// Defaults
const defaultExporterRootFs bool = false

// Configuration for Exporter
type ExporterConfig struct {
	Title string `yaml:"-"`

	// Toggle for enabling access to the root filesystem (for multiple disk usage metrics)
	RootFs config.Parameter `yaml:"rootFs,omitempty"`

	// The Docker Hub tag for Prometheus
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags config.Parameter `yaml:"additionalFlags,omitempty"`
}

// Generates a new Exporter config
func NewExporterConfig(cfg *RocketPoolConfig) *ExporterConfig {
	return &ExporterConfig{
		Title: "Node Exporter Settings",

		RootFs: config.Parameter{
			ID:                 "enableRootFs",
			Name:               "Allow Root Filesystem Access",
			Description:        "Give Prometheus's Node Exporter permission to view your root filesystem instead of being limited to its own Docker container.\nThis is needed if you want the Grafana dashboard to report the used disk space of a second SSD.",
			Type:               config.ParameterType_Bool,
			Default:            map[config.Network]interface{}{config.Network_All: defaultExporterRootFs},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Exporter},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ContainerTag: config.Parameter{
			ID:                 "containerTag",
			Name:               "Exporter Container Tag",
			Description:        "The tag name of the Prometheus Node Exporter container you want to use on Docker Hub.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: exporterTag},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Exporter},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		AdditionalFlags: config.Parameter{
			ID:                 "additionalFlags",
			Name:               "Additional Exporter Flags",
			Description:        "Additional custom command line flags you want to pass to the Node Exporter, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Grafana},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Get the parameters for this config
func (cfg *ExporterConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.RootFs,
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// The the title for the config
func (cfg *ExporterConfig) GetConfigTitle() string {
	return cfg.Title
}
