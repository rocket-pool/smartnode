package config

// Constants
const exporterTag string = "prom/node-exporter:v1.3.1"

// Defaults
const defaultExporterRootFs bool = false
const defaultExporterPort uint16 = 9103

// Configuration for Exporter
type ExporterConfig struct {
	// The master configuration this belongs to
	MasterConfig *Configuration

	// Toggle for enabling access to the root filesystem (for multiple disk usage metrics)
	RootFs *Parameter

	// The port to serve metrics on
	Port *Parameter

	// The Docker Hub tag for Prometheus
	ContainerTag *Parameter
}

// Generates a new Exporter config
func NewExporterConfig(config *Configuration) *ExporterConfig {
	return &ExporterConfig{
		MasterConfig: config,

		RootFs: &Parameter{
			ID:                   "enableRootFs",
			Name:                 "Allow Root Filesystem Access",
			Description:          "Give the exporter permission to view your root filesystem instead of being limited to its own Docker container.\nThis is needed if you want the Grafana dashboard to report the used disk space of a second SSD.",
			Type:                 ParameterType_Bool,
			Default:              defaultExporterRootFs,
			AffectsContainers:    []ContainerID{ContainerID_Exporter},
			EnvironmentVariables: []string{"EXPORTER_ROOT_FS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		Port: &Parameter{
			ID:                   "port",
			Name:                 "API Port",
			Description:          "The port the Exporter should make its statistics available on.",
			Type:                 ParameterType_Uint16,
			Default:              defaultExporterPort,
			AffectsContainers:    []ContainerID{ContainerID_Exporter},
			EnvironmentVariables: []string{"EXPORTER_PORT"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: &Parameter{
			ID:                   "containerTag",
			Name:                 "Container Tag",
			Description:          "The tag name of the Exporter container you want to use on Docker Hub.",
			Type:                 ParameterType_String,
			Default:              exporterTag,
			AffectsContainers:    []ContainerID{ContainerID_Exporter},
			EnvironmentVariables: []string{"EXPORTER_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},
	}
}
