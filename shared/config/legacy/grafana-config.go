package config

// Constants
const grafanaTag string = "grafana/grafana:9.4.15"

// Defaults
const defaultGrafanaPort uint16 = 3100

// Configuration for Grafana
type GrafanaConfig struct {
	Title string `yaml:"-"`

	// The HTTP port to serve on
	Port Parameter `yaml:"port,omitempty"`

	// The Docker Hub tag for Grafana
	ContainerTag Parameter `yaml:"containerTag,omitempty"`
}

// Generates a new Grafana config
func NewGrafanaConfig(cfg *RocketPoolConfig) *GrafanaConfig {
	return &GrafanaConfig{
		Title: "Grafana Settings",

		Port: Parameter{
			ID:                 "port",
			Name:               "Grafana Port",
			Description:        "The port Grafana should run its HTTP server on - this is the port you will connect to in your browser.",
			Type:               ParameterType_Uint16,
			Default:            map[Network]interface{}{Network_All: defaultGrafanaPort},
			AffectsContainers:  []ContainerID{ContainerID_Grafana},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ContainerTag: Parameter{
			ID:                 "containerTag",
			Name:               "Grafana Container Tag",
			Description:        "The tag name of the Grafana container you want to use on Docker Hub.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: grafanaTag},
			AffectsContainers:  []ContainerID{ContainerID_Grafana},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},
	}
}

// Get the parameters for this config
func (cfg *GrafanaConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&cfg.Port,
		&cfg.ContainerTag,
	}
}

// The the title for the config
func (cfg *GrafanaConfig) GetConfigTitle() string {
	return cfg.Title
}
