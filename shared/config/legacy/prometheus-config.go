package config

// Constants
const prometheusTag string = "prom/prometheus:v2.47.1"

// Defaults
const defaultPrometheusPort uint16 = 9091
const defaultPrometheusOpenPort string = string(RPC_Closed)

// Configuration for Prometheus
type PrometheusConfig struct {
	Title string `yaml:"-"`

	// The port to serve metrics on
	Port Parameter `yaml:"port,omitempty"`

	// Toggle for forwarding the API port outside of Docker
	OpenPort Parameter `yaml:"openPort,omitempty"`

	// The Docker Hub tag for Prometheus
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags Parameter `yaml:"additionalFlags,omitempty"`
}

// Generates a new Prometheus config
func NewPrometheusConfig(cfg *RocketPoolConfig) *PrometheusConfig {
	rpcPortModes := PortModes("")
	return &PrometheusConfig{
		Title: "Prometheus Settings",

		Port: Parameter{
			ID:                 "port",
			Name:               "Prometheus Port",
			Description:        "The port Prometheus should make its statistics available on.",
			Type:               ParameterType_Uint16,
			Default:            map[Network]interface{}{Network_All: defaultPrometheusPort},
			AffectsContainers:  []ContainerID{ContainerID_Prometheus},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		OpenPort: Parameter{
			ID:                 "openPort",
			Name:               "Expose Prometheus Port",
			Description:        "Expose the Prometheus's port to other processes on your machine, or to your local network so other machines can access it too.",
			Type:               ParameterType_Choice,
			Default:            map[Network]interface{}{Network_All: defaultPrometheusOpenPort},
			AffectsContainers:  []ContainerID{ContainerID_Prometheus},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options:            rpcPortModes,
		},

		ContainerTag: Parameter{
			ID:                 "containerTag",
			Name:               "Prometheus Container Tag",
			Description:        "The tag name of the Prometheus container you want to use on Docker Hub.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: prometheusTag},
			AffectsContainers:  []ContainerID{ContainerID_Prometheus},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		AdditionalFlags: Parameter{
			ID:                 "additionalFlags",
			Name:               "Additional Prometheus Flags",
			Description:        "Additional custom command line flags you want to pass to Prometheus, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Grafana},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Get the parameters for this config
func (cfg *PrometheusConfig) GetParameters() []*Parameter {
	return []*Parameter{
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
