package config

// Constants
const grafanaTag string = "grafana/grafana:8.3.2"

// Defaults
const defaultGrafanaPort uint16 = 3100

// Configuration for Grafana
type GrafanaConfig struct {
	// The HTTP port to serve on
	Port Parameter `yaml:"port,omitempty"`

	// The Docker Hub tag for Grafana
	ContainerTag Parameter `yaml:"containerTag,omitempty"`
}

// Generates a new Grafana config
func NewGrafanaConfig(config *RocketPoolConfig) *GrafanaConfig {
	return &GrafanaConfig{
		Port: Parameter{
			ID:                   "port",
			Name:                 "Grafana Port",
			Description:          "The port Grafana should run its HTTP server on - this is the port you will connect to in your browser.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultGrafanaPort},
			AffectsContainers:    []ContainerID{ContainerID_Grafana},
			EnvironmentVariables: []string{"GRAFANA_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: Parameter{
			ID:                   "containerTag",
			Name:                 "Grafana Container Tag",
			Description:          "The tag name of the Grafana container you want to use on Docker Hub.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: grafanaTag},
			AffectsContainers:    []ContainerID{ContainerID_Grafana},
			EnvironmentVariables: []string{"GRAFANA_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},
	}
}

// Get the parameters for this config
func (config *GrafanaConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.Port,
		&config.ContainerTag,
	}
}
