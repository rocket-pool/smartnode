package apcupsd

import (
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const (
	containerTag         string = "gersilex/apcupsd:v1.0.0"
	exporterContainerTag string = "jangrewe/apcupsd-exporter:latest"
)

type Mode string

const (
	Mode_Network   Mode = "network"
	Mode_Container Mode = "docker"
)

// Configuration for the Graffiti Wall Writer
type ApcupsdConfig struct {
	Title string `yaml:"-"`

	Enabled                     config.Parameter `yaml:"enabled,omitempty"`
	ApcupsdContainerTag         config.Parameter `yaml:"apcupsdContainerTag,omitempty"`
	ApcupsdExporterContainerTag config.Parameter `yaml:"apcupsdExporterContainerTag,omitempty"`
	MetricsPort                 config.Parameter `yaml:"metricsPort,omitempty"`
	MountPoint                  config.Parameter `yaml:"mountPoint,omitempty"`
	Mode                        config.Parameter `yaml:"mode,omitempty"`
	NetworkAddress              config.Parameter `yaml:"NetworkAddress,omitempty"`
}

// Creates a new configuration instance
func NewConfig() *ApcupsdConfig {
	return &ApcupsdConfig{
		Title: "APCUPSD Settings",

		Enabled: config.Parameter{
			ID:                   "enabled",
			Name:                 "Enabled",
			Description:          "Enable APCUPSD monitoring",
			Type:                 config.ParameterType_Bool,
			Default:              map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:    []config.ContainerID{ContainerID_Apcupsd, ContainerID_ApcupsdExporter},
			EnvironmentVariables: []string{"ADDON_APCUPSD_ENABLED"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
		Mode: config.Parameter{
			ID:                   "mode",
			Name:                 "Mode",
			Description:          "How would you like to run APCUPSD?\n Select `Container` if you'd like smart node to run apcupsd inside a container for you.\nSelect `network` mode if you want to connect to an instance of apcupsd running on your host machine or on your network.",
			Type:                 config.ParameterType_Choice,
			Default:              map[config.Network]interface{}{config.Network_All: Mode_Container},
			AffectsContainers:    []config.ContainerID{ContainerID_Apcupsd, ContainerID_ApcupsdExporter},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []config.ParameterOption{{
				Name:        "Container",
				Description: "Let the smart node run APCUPSD inside a container for you",
				Value:       Mode_Container,
			}, {
				Name:        "Network",
				Description: "Connect the APCUPSD exporter to an instance of APCUSD running on your host machine or on your network",
				Value:       Mode_Network,
			}},
		},
		ApcupsdContainerTag: config.Parameter{
			ID:                   "containerTag",
			Name:                 "APCUPSD Container Tag",
			Description:          "The container tag name of the APCUPSD container.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: containerTag},
			AffectsContainers:    []config.ContainerID{ContainerID_Apcupsd},
			EnvironmentVariables: []string{"ADDON_APCUPSD_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},
		ApcupsdExporterContainerTag: config.Parameter{
			ID:                   "exporterContainerTag",
			Name:                 "APCUPSD Exporter Container Tag",
			Description:          "The container tag name of the APCUPSD Prometheus Exporter.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: exporterContainerTag},
			AffectsContainers:    []config.ContainerID{ContainerID_ApcupsdExporter},
			EnvironmentVariables: []string{"ADDON_APCUPSD_EXPORTER_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},
		MetricsPort: config.Parameter{
			ID:                   "metricsPort",
			Name:                 "APCUPSD Exporter Metrics Port",
			Description:          "The port the exporter should use to provide metrics to prometheus.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: "9162"},
			AffectsContainers:    []config.ContainerID{ContainerID_ApcupsdExporter, config.ContainerID_Prometheus},
			EnvironmentVariables: []string{"ADDON_APCUPSD_METRICS_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
		MountPoint: config.Parameter{
			ID:                   "mountPoint",
			Name:                 "APC USB Mount Location",
			Description:          "The USB mount point for your APC device. This must be set correctly for the container to read data from your UPC. To determine the mount point on your system:\n1. Unplug the USB cable of your UPS and plug it back in.\n2. When your server detects the device an entry will show up when you run `sudo dmesg | grep usb`.\n3. Identify the mount point for your UPS. Often it is named `hiddev*` e.g. `hiddev0`,`hiddev1`... but may vary depending on how many peripherals you have connected.\n4. Verify the mount point for your distribution. Often this maps to `/dev/usb/hiddev*`\nThis is the value to enter in the field below. NOTE: If you reconnect your UPC this value may need to be updated.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: "/dev/usb/hiddev0"},
			AffectsContainers:    []config.ContainerID{ContainerID_Apcupsd},
			EnvironmentVariables: []string{"ADDON_APCUPSD_MOUNT_POINT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
		NetworkAddress: config.Parameter{
			ID:                   "networkAddress",
			Name:                 "APCUPSD Network Address",
			Description:          "The network address and port that should be used to connect to APCUPSD.\nIf you have apcupsd installed on your host you should use the default host.docker.internal:3551.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: "host.docker.internal:3551"},
			AffectsContainers:    []config.ContainerID{ContainerID_ApcupsdExporter},
			EnvironmentVariables: []string{"ADDON_APCUPSD_NETWORK_ADDRESS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (cfg *ApcupsdConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{&cfg.Enabled, &cfg.Mode, &cfg.ApcupsdExporterContainerTag, &cfg.ApcupsdContainerTag, &cfg.MetricsPort, &cfg.MountPoint, &cfg.NetworkAddress}

}

// The the title for the config
func (cfg *ApcupsdConfig) GetConfigTitle() string {
	return cfg.Title
}
