package apcupsd

import (
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const (
	containerTag string = "rocketpool/graffiti-wall-addon:v1.0.1"
)

// Configuration for the Graffiti Wall Writer
type ApcupsdConfig struct {
	Title string `yaml:"-"`

	Enabled    config.Parameter `yaml:"enabled,omitempty"`
	MountPoint config.Parameter `yaml:"mountPoint,omitempty"`
	Debug      config.Parameter `yaml:"debug,omitempty"`
	PollCron   config.Parameter `yaml:"pollCron,omitempty"`
	Timeout    config.Parameter `yaml:"timeout,omitempty"`
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
			AffectsContainers:    []config.ContainerID{ContainerID_Apcupsd},
			EnvironmentVariables: []string{"ADDON_APCUPSD_ENABLED"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
		MountPoint: config.Parameter{
			ID:                   "mountPoint",
			Name:                 "APC USB Mount Location",
			Description:          "The USB mount point for your APC device. This must be set correctly for the container to read data from your UPC. To determine the mount point on your system:\n1. Unplug the USB cable of your UPS and plug it back in.\n2. When your server detects the device an entry will show up when you run `sudo dmesg | grep usb`.\n3. Identify the mount point for your UPS. Often it is named `hiddev*` e.g. `hiddev0`,`hiddev1`... but may vary depending on how many peripherals you have connected.\n4. Verify the mount point for your distribution. Often this maps to `/dev/usb/hiddev*`\n This is the value to enter in the field below. NOTE: If you reconnect your UPC this value may need to be updated.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			Regex:                "[^\\0]+",
			AffectsContainers:    []config.ContainerID{ContainerID_Apcupsd},
			EnvironmentVariables: []string{"ADDON_APCUPSD_MOUNT_POINT"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
		Debug: config.Parameter{
			ID:                   "debug",
			Name:                 "Debug",
			Description:          "Output debug logs for APCUPSD monitoring",
			Type:                 config.ParameterType_Bool,
			Default:              map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:    []config.ContainerID{ContainerID_Apcupsd},
			EnvironmentVariables: []string{"ADDON_APCUPSD_DEBUG"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
		Timeout: config.Parameter{
			ID:                   "timeout",
			Name:                 "Timeout",
			Description:          "How long to wait for a connection to the UPS (ms) before timing out. Defaults to \"30000ms\".",
			Type:                 config.ParameterType_Int,
			Default:              map[config.Network]interface{}{config.Network_All: 0},
			AffectsContainers:    []config.ContainerID{ContainerID_Apcupsd},
			EnvironmentVariables: []string{"ADDON_APCUPSD_TIMEOUT"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
		PollCron: config.Parameter{
			ID:                   "pollCron",
			Name:                 "Update Interval",
			Description:          "CRON interval to poll stats from the UPC. Uses node-cron format, see https://www.npmjs.com/package/node-cron for details. Defaults to \"* * * * *\"",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			Regex:                "^?:\\d+|\\*|\\*\\/\\d+$",
			AffectsContainers:    []config.ContainerID{ContainerID_Apcupsd},
			EnvironmentVariables: []string{"ADDON_APCUPSD_POLL_INTERVAL"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (cfg *ApcupsdConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.Enabled,
		&cfg.MountPoint,
		&cfg.PollCron,
		&cfg.Debug,
		// &cfg.Timeout,
	}
}

// The the title for the config
func (cfg *ApcupsdConfig) GetConfigTitle() string {
	return cfg.Title
}
