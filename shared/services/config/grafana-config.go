package config

// Constants
const grafanaTag string = "grafana/grafana:8.3.2"

// Defaults
const defaultGrafanaPort uint16 = 3100

// Configuration for Grafana
type GrafanaConfig struct {
	// The HTTP port to serve on
	Port Parameter

	// The Docker Hub tag for Grafana
	ContainerTag Parameter
}

// Generates a new Grafana config
func NewGrafanaConfig(config *MasterConfig) *GrafanaConfig {
	return &GrafanaConfig{
		Port: Parameter{
			ID:                   "port",
			Name:                 "HTTP Port",
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
			Name:                 "Container Tag",
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

// Handle a network change on all of the parameters
func (config *GrafanaConfig) changeNetwork(oldNetwork Network, newNetwork Network) {
	changeNetworkForParameter(&config.Port, oldNetwork, newNetwork)
	changeNetworkForParameter(&config.ContainerTag, oldNetwork, newNetwork)
}
