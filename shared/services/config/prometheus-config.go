package config

import (
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const prometheusTag string = "prom/prometheus:v2.42.0"

// Defaults
const defaultPrometheusPort uint16 = 9091
const defaultPrometheusOpenPort string = "closed"

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
	rpcPortModes := []config.ParameterOption{{
		Name:        "Closed",
		Description: "Do not allow connections to the RPC port",
		Value:       config.RPC_Closed,
	}, {
		Name:        "Open to Localhost",
		Description: "Allow connections from this host only",
		Value:       config.RPC_OpenLocalhost,
	}, {
		Name:        "Open to External hosts",
		Description: "Allow connections from external hosts. This is safe if you're running your node on your local network. If you're a VPS user, this would expose your node to the internet",
		Value:       config.RPC_OpenExternal,
	}}
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
			Description:          "Expose the Prometheus's port to your local network, so other machines can access it too.",
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
