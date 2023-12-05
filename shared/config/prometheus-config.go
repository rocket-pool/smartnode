package config

import (
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const prometheusTag string = "prom/prometheus:v2.47.1"

// Defaults
const defaultPrometheusPort uint16 = 9091
const defaultPrometheusOpenPort string = string(config.RPC_Closed)

// Configuration for Prometheus
type PrometheusConfig struct {
	Title string `yaml:"-"`

	// The port to serve metrics on
	Port config.Parameter `yaml:"port,omitempty"`

	// Toggle for forwarding the API port outside of Docker
	OpenPort config.Parameter `yaml:"openPort,omitempty"`

	// The Docker Hub tag for Prometheus
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags config.Parameter `yaml:"additionalFlags,omitempty"`
}

// Generates a new Prometheus config
func NewPrometheusConfig(cfg *RocketPoolConfig) *PrometheusConfig {
	rpcPortModes := config.PortModes("")
	return &PrometheusConfig{
		Title: "Prometheus Settings",

		Port: config.Parameter{
			ID:                   "port",
			Name:                 "Prometheus Port",
			Description:          "The port Prometheus should make its statistics available on.",
			Type:                 config.ParameterType_Uint16,
			Default:              map[config.Network]interface{}{config.Network_All: defaultPrometheusPort},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Prometheus},
			EnvironmentVariables: []string{"PROMETHEUS_PORT"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		OpenPort: config.Parameter{
			ID:                   "openPort",
			Name:                 "Expose Prometheus Port",
			Description:          "Expose the Prometheus's port to other processes on your machine, or to your local network so other machines can access it too.",
			Type:                 config.ParameterType_Choice,
			Default:              map[config.Network]interface{}{config.Network_All: defaultPrometheusOpenPort},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Prometheus},
			EnvironmentVariables: []string{"PROMETHEUS_OPEN_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options:              rpcPortModes,
		},

		ContainerTag: config.Parameter{
			ID:                   "containerTag",
			Name:                 "Prometheus Container Tag",
			Description:          "The tag name of the Prometheus container you want to use on Docker Hub.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: prometheusTag},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Prometheus},
			EnvironmentVariables: []string{"PROMETHEUS_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalFlags: config.Parameter{
			ID:                   "additionalFlags",
			Name:                 "Additional Prometheus Flags",
			Description:          "Additional custom command line flags you want to pass to Prometheus, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Grafana},
			EnvironmentVariables: []string{},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (cfg *PrometheusConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.Port,
		&cfg.OpenPort,
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// The the title for the config
func (cfg *PrometheusConfig) GetConfigTitle() string {
	return cfg.Title
}
