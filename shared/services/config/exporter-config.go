package config

// Constants
const exporterTag string = "prom/node-exporter:v1.3.1"

// Defaults
const defaultExporterRootFs bool = false

// Configuration for Exporter
type ExporterConfig struct {
	Title string `yaml:"-"`

	// Toggle for enabling access to the root filesystem (for multiple disk usage metrics)
	RootFs Parameter `yaml:"rootFs,omitempty"`

	// The Docker Hub tag for Prometheus
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags Parameter `yaml:"additionalFlags,omitempty"`
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

		AdditionalFlags: Parameter{
			ID:                   "additionalFlags",
			Name:                 "Additional Exporter Flags",
			Description:          "Additional custom command line flags you want to pass to the Node Exporter, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Grafana},
			EnvironmentVariables: []string{},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (config *ExporterConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.RootFs,
		&config.ContainerTag,
		&config.AdditionalFlags,
	}
}

// The the title for the config
func (config *ExporterConfig) GetConfigTitle() string {
	return config.Title
}
