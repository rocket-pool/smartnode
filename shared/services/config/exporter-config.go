package config

// Constants
const exporterTag string = "prom/node-exporter:v1.3.1"

// Defaults
const defaultExporterRootFs bool = false
const defaultExporterPort uint16 = 9103

// Configuration for Exporter
type ExporterConfig struct {
	Title string `yaml:"title,omitempty"`

	// Toggle for enabling access to the root filesystem (for multiple disk usage metrics)
	RootFs Parameter `yaml:"rootFs,omitempty"`

	// The port to serve metrics on
	Port Parameter `yaml:"port,omitempty"`

	// The Docker Hub tag for Prometheus
	ContainerTag Parameter `yaml:"containerTag,omitempty"`
}

// Generates a new Exporter config
func NewExporterConfig(config *RocketPoolConfig) *ExporterConfig {
	return &ExporterConfig{
		Title: "Node Exporter Settings",

		RootFs: Parameter{
			ID:                   "enableRootFs",
			Name:                 "Allow Root Filesystem Access",
			Description:          "Give Prometheus's Node Exporter permission to view your root filesystem instead of being limited to its own Docker container.\nThis is needed if you want the Grafana dashboard to report the used disk space of a second SSD.",
			Type:                 ParameterType_Bool,
			Default:              map[Network]interface{}{Network_All: defaultExporterRootFs},
			AffectsContainers:    []ContainerID{ContainerID_Exporter},
			EnvironmentVariables: []string{"EXPORTER_ROOT_FS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		Port: Parameter{
			ID:                   "port",
			Name:                 "Exporter Port",
			Description:          "The port Prometheus's Node Exporter should make its statistics available on.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultExporterPort},
			AffectsContainers:    []ContainerID{ContainerID_Exporter},
			EnvironmentVariables: []string{"EXPORTER_PORT"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: Parameter{
			ID:                   "containerTag",
			Name:                 "Exporter Container Tag",
			Description:          "The tag name of the Prometheus Node Exporter container you want to use on Docker Hub.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: exporterTag},
			AffectsContainers:    []ContainerID{ContainerID_Exporter},
			EnvironmentVariables: []string{"EXPORTER_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},
	}
}

// Get the parameters for this config
func (config *ExporterConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.RootFs,
		&config.Port,
		&config.ContainerTag,
	}
}

// The the title for the config
func (config *ExporterConfig) GetConfigTitle() string {
	return config.Title
}
