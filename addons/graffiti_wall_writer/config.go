package graffiti_wall_writer

import (
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const (
	containerTag string = "rocketpool/graffiti-wall-addon:v0.1.0"
)

// Configuration for the Graffiti Wall Writer
type GraffitiWallWriterConfig struct {
	Title string `yaml:"-"`

	Enabled config.Parameter `yaml:"enabled,omitempty"`

	InputURL config.Parameter `yaml:"inputUrl,omitempty"`

	UpdateWallTime config.Parameter `yaml:"updateWallTime,omitempty"`

	UpdateInputTime config.Parameter `yaml:"updateInputTime,omitempty"`

	UpdatePixelTime config.Parameter `yaml:"updatePixelTime,omitempty"`

	// The Docker Hub tag
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags config.Parameter `yaml:"additionalFlags,omitempty"`
}

// Creates a new configuration instance
func NewConfig() *GraffitiWallWriterConfig {
	return &GraffitiWallWriterConfig{
		Title: "Graffiti Wall Writer Settings",

		Enabled: config.Parameter{
			ID:                   "enabled",
			Name:                 "Enabled",
			Description:          "Enable the Graffiti Wall Writer",
			Type:                 config.ParameterType_Bool,
			Default:              map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:    []config.ContainerID{ContainerID_GraffitiWallWriter, config.ContainerID_Validator},
			EnvironmentVariables: []string{"ADDON_GWW_ENABLED"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		InputURL: config.Parameter{
			ID:                   "inputUrl",
			Name:                 "Input URL",
			Description:          "URL or path to file to source pixeldata from",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{ContainerID_GraffitiWallWriter},
			EnvironmentVariables: []string{"ADDON_GWW_INPUT_URL"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		UpdateWallTime: config.Parameter{
			ID:                   "updateWallTime",
			Name:                 "Wall Update Interval",
			Description:          "The time, in seconds, between updating the beaconcha.in graffiti wall canvas",
			Type:                 config.ParameterType_Uint,
			Default:              map[config.Network]interface{}{config.Network_All: uint64(600)},
			AffectsContainers:    []config.ContainerID{ContainerID_GraffitiWallWriter},
			EnvironmentVariables: []string{"ADDON_GWW_UPDATE_WALL_TIME"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		UpdateInputTime: config.Parameter{
			ID:                   "updateInputTime",
			Name:                 "Input Update Interval",
			Description:          "The time, in seconds, between input updates - only if remote URL is used. File will be instantly reloaded when changed.",
			Type:                 config.ParameterType_Uint,
			Default:              map[config.Network]interface{}{config.Network_All: uint64(600)},
			AffectsContainers:    []config.ContainerID{ContainerID_GraffitiWallWriter},
			EnvironmentVariables: []string{"ADDON_GWW_UPDATE_INPUT_TIME"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		UpdatePixelTime: config.Parameter{
			ID:                   "updatePixelTime",
			Name:                 "Pixel Update Interval",
			Description:          "The time, in seconds, between output updates.",
			Type:                 config.ParameterType_Uint,
			Default:              map[config.Network]interface{}{config.Network_All: uint64(60)},
			AffectsContainers:    []config.ContainerID{ContainerID_GraffitiWallWriter},
			EnvironmentVariables: []string{"ADDON_GWW_UPDATE_PIXEL_TIME"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: config.Parameter{
			ID:                   "containerTag",
			Name:                 "Container Tag",
			Description:          "The tag name of the container you want to use on Docker Hub.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: containerTag},
			AffectsContainers:    []config.ContainerID{ContainerID_GraffitiWallWriter},
			EnvironmentVariables: []string{"ADDON_GWW_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalFlags: config.Parameter{
			ID:                   "additionalFlags",
			Name:                 "Additional Flags",
			Description:          "Additional custom command line flags you want to pass to the addon, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{ContainerID_GraffitiWallWriter},
			EnvironmentVariables: []string{"ADDON_GWW_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (cfg *GraffitiWallWriterConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.Enabled,
		&cfg.InputURL,
		&cfg.UpdateWallTime,
		&cfg.UpdateInputTime,
		&cfg.UpdatePixelTime,
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// The the title for the config
func (cfg *GraffitiWallWriterConfig) GetConfigTitle() string {
	return cfg.Title
}
