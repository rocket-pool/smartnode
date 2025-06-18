package config

import (
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const grafanaTag string = "grafana/grafana:9.5.18"

// Defaults
const defaultGrafanaPort uint16 = 3100
const defaultGrafanaOpenPort string = string(config.RPC_OpenExternal)

// Configuration for Grafana
type GrafanaConfig struct {
	Title string `yaml:"-"`

	// The HTTP port to serve on
	Port config.Parameter `yaml:"port,omitempty"`

	// Toggle for forwarding the API port outside of Docker
	OpenPort config.Parameter `yaml:"openPort,omitempty"`

	// The Docker Hub tag for Grafana
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`
}

// Generates a new Grafana config
func NewGrafanaConfig(cfg *RocketPoolConfig) *GrafanaConfig {
	rpcPortModes := config.PortModes("")
	return &GrafanaConfig{
		Title: "Grafana Settings",

		Port: config.Parameter{
			ID:                 "port",
			Name:               "Grafana Port",
			Description:        "The port Grafana should run its HTTP server on - this is the port you will connect to in your browser.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultGrafanaPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Grafana},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		OpenPort: config.Parameter{
			ID:                 "openPort",
			Name:               "Expose Grafana Port",
			Description:        "Expose the Grafana's port to other processes on your machine, or to your local network so other machines can access it too.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: defaultGrafanaOpenPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Grafana},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options:            rpcPortModes,
		},

		ContainerTag: config.Parameter{
			ID:                 "containerTag",
			Name:               "Grafana Container Tag",
			Description:        "The tag name of the Grafana container you want to use on Docker Hub.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: grafanaTag},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Grafana},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},
	}
}

// Get the parameters for this config
func (cfg *GrafanaConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.Port,
		&cfg.OpenPort,
		&cfg.ContainerTag,
	}
}

// The title for the config
func (cfg *GrafanaConfig) GetConfigTitle() string {
	return cfg.Title
}
